package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	httpteststubv1 "httpteststub.example.com/api/v1"
)

var stubCache sync.Map
var stubCounters sync.Map
var resourceLimits sync.Map

type HTTPTestStubReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type StubCounter struct {
	mu      sync.Mutex
	count   int
	resetAt int
}

type ResourceLimit struct {
	mu                   sync.Mutex
	requestCount         int
	lastResetTime        time.Time
	maxRequestsPerMinute int
	totalRequests        int
	startTime            time.Time
}

// StubStats 用于跟踪 stub 的请求统计信息
type StubStats struct {
	mu              sync.Mutex
	totalRequests   int64
	totalErrors     int64
	lastRequestTime time.Time
	totalDuration   time.Duration
}

// stubStatsMap 存储每个 stub 的统计信息
var stubStatsMap sync.Map

func (r *HTTPTestStubReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var httpTestStub httpteststubv1.HTTPTestStub
	if err := r.Get(ctx, req.NamespacedName, &httpTestStub); err != nil {
		log.Error(err, "unable to fetch HTTPTestStub")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	key := fmt.Sprintf("%s/%s", httpTestStub.Namespace, httpTestStub.Name)

	stubCache.Store(key, &httpTestStub)

	if _, exists := stubCounters.Load(key); !exists {
		stubCounters.Store(key, &StubCounter{})
	}

	if _, exists := resourceLimits.Load(key); !exists {
		resourceLimits.Store(key, &ResourceLimit{
			lastResetTime: time.Now(),
		})
	}

	// 更新状态为 Running
	httpTestStub.Status.Phase = "Running"
	if err := r.Status().Update(ctx, &httpTestStub); err != nil {
		log.Error(err, "unable to update HTTPTestStub status")
		return ctrl.Result{}, err
	}

	log.Info("HTTPTestStub loaded into cache", "name", httpTestStub.Name, "namespace", httpTestStub.Namespace)

	return ctrl.Result{}, nil
}

func (r *HTTPTestStubReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&httpteststubv1.HTTPTestStub{}).
		Named("httpteststub").
		Complete(r)
}

func GetMatchingStub(method, path, protocol string) *httpteststubv1.HTTPTestStub {
	var matchedStub *httpteststubv1.HTTPTestStub

	stubCache.Range(func(key, value interface{}) bool {
		stub := value.(*httpteststubv1.HTTPTestStub)

		// 检查 protocol 是否匹配
		// 如果 CR 的 protocol 为空或 "both"，则匹配所有协议
		// 否则必须完全匹配
		stubProtocol := stub.Spec.Protocol
		if stubProtocol != "" && stubProtocol != "both" && stubProtocol != protocol {
			return true // 跳过不匹配的 protocol
		}

		for _, s := range stub.Spec.Stubs {
			if matchRequest(method, path, &s.Request) {
				matchedStub = stub
				return false
			}
		}

		if matchRequest(method, path, &stub.Spec.Request) {
			matchedStub = stub
			return false
		}

		return true
	})

	return matchedStub
}

func matchRequest(method, path string, request *httpteststubv1.Request) bool {
	if method != request.Method {
		return false
	}

	switch request.URL.Type {
	case "exact":
		return path == request.URL.Pattern
	case "pattern":
		return matchPattern(path, request.URL.Pattern)
	case "regex":
		return matchRegex(path, request.URL.Regex)
	default:
		return false
	}
}

func matchPattern(url, pattern string) bool {
	if pattern == "*" {
		return true
	}
	if pattern == url {
		return true
	}

	// 处理多个通配符的情况
	patternParts := strings.Split(pattern, "*")
	if len(patternParts) == 1 {
		// 没有通配符，直接比较
		return url == pattern
	}

	// 检查前缀
	if patternParts[0] != "" && !strings.HasPrefix(url, patternParts[0]) {
		return false
	}

	// 检查后缀
	if patternParts[len(patternParts)-1] != "" && !strings.HasSuffix(url, patternParts[len(patternParts)-1]) {
		return false
	}

	// 检查中间部分
	currentIndex := len(patternParts[0])
	for i := 1; i < len(patternParts)-1; i++ {
		part := patternParts[i]
		if part == "" {
			continue
		}
		index := strings.Index(url[currentIndex:], part)
		if index == -1 {
			return false
		}
		currentIndex += index + len(part)
	}

	return true
}

func matchRegex(url, regex string) bool {
	matched, err := regexp.MatchString(regex, url)
	if err != nil {
		return false
	}
	return matched
}

func IncrementCounter(stubKey string) int {
	if counter, ok := stubCounters.Load(stubKey); ok {
		c := counter.(*StubCounter)
		c.mu.Lock()
		defer c.mu.Unlock()

		c.count++
		return c.count
	}
	return 0
}

func ResetCounter(stubKey string, resetAfter int) {
	if counter, ok := stubCounters.Load(stubKey); ok {
		c := counter.(*StubCounter)
		c.mu.Lock()
		defer c.mu.Unlock()

		if c.count >= resetAfter {
			c.count = 0
		}
	}
}

func CheckResourceLimit(stubKey string, maxRequestsPerMinute int) bool {
	if limit, ok := resourceLimits.Load(stubKey); ok {
		rl := limit.(*ResourceLimit)
		rl.mu.Lock()
		defer rl.mu.Unlock()

		now := time.Now()
		if now.Sub(rl.lastResetTime) > time.Minute {
			rl.requestCount = 0
			rl.lastResetTime = now
		}

		if rl.maxRequestsPerMinute == 0 {
			rl.maxRequestsPerMinute = maxRequestsPerMinute
		}

		if rl.startTime.IsZero() {
			rl.startTime = now
		}

		if rl.requestCount >= rl.maxRequestsPerMinute {
			return false
		}

		rl.requestCount++
		rl.totalRequests++
		return true
	}
	return true
}

func GetResourceStats(stubKey string) (requestsPerMinute, totalRequests int, uptime time.Duration) {
	if limit, ok := resourceLimits.Load(stubKey); ok {
		rl := limit.(*ResourceLimit)
		rl.mu.Lock()
		defer rl.mu.Unlock()

		requestsPerMinute = rl.requestCount
		totalRequests = rl.totalRequests
		if !rl.startTime.IsZero() {
			uptime = time.Since(rl.startTime)
		}
	}
	return
}

func SerializeResponse(response interface{}) ([]byte, error) {
	return json.Marshal(response)
}

func GetStubFromCache(key string) *httpteststubv1.HTTPTestStub {
	if value, ok := stubCache.Load(key); ok {
		return value.(*httpteststubv1.HTTPTestStub)
	}
	return nil
}

func GetAllStubs() []*httpteststubv1.HTTPTestStub {
	var stubs []*httpteststubv1.HTTPTestStub

	stubCache.Range(func(key, value interface{}) bool {
		stubs = append(stubs, value.(*httpteststubv1.HTTPTestStub))
		return true
	})

	return stubs
}

func RemoveStubFromCache(key string) {
	stubCache.Delete(key)
	stubCounters.Delete(key)
	resourceLimits.Delete(key)
	stubStatsMap.Delete(key)
}

// RecordRequestStats 记录请求统计信息
func RecordRequestStats(stubKey string, duration time.Duration, isError bool) {
	stats, exists := stubStatsMap.Load(stubKey)
	if !exists {
		stats = &StubStats{}
		stubStatsMap.Store(stubKey, stats)
	}

	s := stats.(*StubStats)
	s.mu.Lock()
	defer s.mu.Unlock()

	s.totalRequests++
	s.lastRequestTime = time.Now()
	s.totalDuration += duration
	if isError {
		s.totalErrors++
	}
}

// GetStubStats 获取 stub 的统计信息
func GetStubStats(stubKey string) (totalRequests, totalErrors int64, lastRequestTime time.Time, avgResponseTime int64, errorRate int) {
	stats, exists := stubStatsMap.Load(stubKey)
	if !exists {
		return 0, 0, time.Time{}, 0, 0
	}

	s := stats.(*StubStats)
	s.mu.Lock()
	defer s.mu.Unlock()

	totalRequests = s.totalRequests
	totalErrors = s.totalErrors
	lastRequestTime = s.lastRequestTime

	if s.totalRequests > 0 {
		avgResponseTime = int64(s.totalDuration.Milliseconds() / time.Duration(s.totalRequests).Milliseconds())
		errorRate = int(float64(s.totalErrors) * 100 / float64(s.totalRequests))
	}

	return
}

// UpdateStubStatus 更新 stub 的 status 到 CR
func (r *HTTPTestStubReconciler) UpdateStubStatus(ctx context.Context, stub *httpteststubv1.HTTPTestStub) error {
	key := fmt.Sprintf("%s/%s", stub.Namespace, stub.Name)

	totalRequests, totalErrors, lastRequestTime, avgResponseTime, errorRate := GetStubStats(key)

	stub.Status.Phase = "Running"
	stub.Status.TotalRequests = totalRequests
	stub.Status.TotalErrors = totalErrors
	stub.Status.AvgResponseTime = avgResponseTime
	stub.Status.ErrorRate = errorRate

	if !lastRequestTime.IsZero() {
		t := metav1.NewTime(lastRequestTime)
		stub.Status.LastRequestTime = &t
	}

	return r.Status().Update(ctx, stub)
}
