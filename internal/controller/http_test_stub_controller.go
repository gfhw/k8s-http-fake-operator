/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	httpteststubv1 "httpteststub.example.com/api/v1"
)

// HTTPTestStubReconciler reconciles a HTTPTestStub object
type HTTPTestStubReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	servers  sync.Map // key: name, value: *http.Server
	counters sync.Map // key: stub key, value: int (request count)
}

// +kubebuilder:rbac:groups=httpteststub.example.com,resources=httpteststubs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=httpteststub.example.com,resources=httpteststubs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=httpteststub.example.com,resources=httpteststubs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *HTTPTestStubReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// 获取 HTTPTestStub 资源
	var httpTestStub httpteststubv1.HTTPTestStub
	if err := r.Get(ctx, req.NamespacedName, &httpTestStub); err != nil {
		log.Error(err, "unable to fetch HTTPTestStub")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 处理服务创建/更新
	key := fmt.Sprintf("%s/%s", httpTestStub.Namespace, httpTestStub.Name)
	if server, ok := r.servers.Load(key); ok {
		// 停止旧服务器
		if s, ok := server.(*http.Server); ok {
			if err := s.Shutdown(ctx); err != nil {
				log.Error(err, "error shutting down server")
			}
			r.servers.Delete(key)
		}
	}

	// 启动新服务器
	go func() {
		if err := r.startServer(&httpTestStub); err != nil {
			log.Error(err, "error starting server")
		}
	}()

	// 更新状态
	httpTestStub.Status.Status = "Running"
	httpTestStub.Status.Address = fmt.Sprintf("%s:%d", httpTestStub.Name, httpTestStub.Spec.BaseInfo.Port)
	if err := r.Status().Update(ctx, &httpTestStub); err != nil {
		log.Error(err, "unable to update HTTPTestStub status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HTTPTestStubReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&httpteststubv1.HTTPTestStub{}).
		Named("httpteststub").
		Complete(r)
}

// startServer starts the HTTP/HTTPS server for the HTTPTestStub
func (r *HTTPTestStubReconciler) startServer(stub *httpteststubv1.HTTPTestStub) error {
	log := logf.Log.WithName("server")

	// 创建 HTTP 处理器
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.handleRequest(w, req, stub)
	})

	// 配置服务器
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", stub.Spec.BaseInfo.Port),
		Handler: handler,
	}

	// 存储服务器实例
	key := fmt.Sprintf("%s/%s", stub.Namespace, stub.Name)
	r.servers.Store(key, server)

	// 启动服务器
	log.Info("starting server", "address", server.Addr, "protocol", stub.Spec.BaseInfo.Protocol)

	if stub.Spec.BaseInfo.Protocol == "https" && stub.Spec.TLS.Enabled {
		log.Info("TLS enabled", "secretName", stub.Spec.TLS.SecretName)
		// TODO: 实现从 Kubernetes Secret 中获取 TLS 证书
		// 暂时使用自签名证书
		return server.ListenAndServeTLS("/tmp/tls.crt", "/tmp/tls.key")
	} else {
		return server.ListenAndServe()
	}
}

// handleRequest handles incoming HTTP requests
func (r *HTTPTestStubReconciler) handleRequest(w http.ResponseWriter, req *http.Request, stub *httpteststubv1.HTTPTestStub) {
	log := logf.Log.WithName("request")
	log.Info("received request", "method", req.Method, "url", req.URL.Path)

	// 首先尝试匹配单个 stub
	for _, s := range stub.Spec.Stubs {
		if r.matchRequest(req, &s.Request) {
			r.handleStubResponse(w, req, &s)
			return
		}
	}

	// 然后尝试匹配全局请求规则
	if r.matchRequest(req, &stub.Spec.Request) {
		r.handleResponse(w, req, &stub.Spec.Response)
		return
	}

	// 默认响应
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"error": "No matching stub found"}`))
}

// matchRequest matches the incoming request against the configured request rules
func (r *HTTPTestStubReconciler) matchRequest(req *http.Request, request *httpteststubv1.Request) bool {
	// 匹配 HTTP 方法
	if req.Method != request.Method {
		return false
	}

	// 匹配 URL
	urlPath := req.URL.Path
	switch request.URL.Type {
	case "exact":
		return urlPath == request.URL.Pattern
	case "pattern":
		return r.matchPattern(urlPath, request.URL.Pattern)
	case "regex":
		return r.matchRegex(urlPath, request.URL.Regex)
	default:
		return false
	}
}

// matchPattern matches the URL against a pattern with wildcards
func (r *HTTPTestStubReconciler) matchPattern(url, pattern string) bool {
	// 简单的通配符匹配实现
	// TODO: 实现更复杂的通配符匹配
	return url == pattern || pattern == "*"
}

// matchRegex matches the URL against a regular expression
func (r *HTTPTestStubReconciler) matchRegex(url, regex string) bool {
	// 简单的正则表达式匹配实现
	// TODO: 实现完整的正则表达式匹配
	return url == regex
}

// handleResponse handles the response based on the configured response rules
func (r *HTTPTestStubReconciler) handleResponse(w http.ResponseWriter, req *http.Request, response *httpteststubv1.Response) {
	log := logf.Log.WithName("response")

	switch response.Type {
	case "static":
		if response.Static != nil {
			r.sendStaticResponse(w, response.Static)
		}
	case "script":
		if response.Script != nil {
			log.Info("script response not yet implemented")
			// TODO: 实现脚本响应
			r.sendDefaultResponse(w)
		}
	default:
		log.Info("unknown response type", "type", response.Type)
		r.sendDefaultResponse(w)
	}
}

// handleStubResponse handles the response for a specific stub
func (r *HTTPTestStubReconciler) handleStubResponse(w http.ResponseWriter, req *http.Request, stub *httpteststubv1.Stub) {
	// 生成 stub 唯一标识
	stubKey := fmt.Sprintf("%s:%s", req.Method, req.URL.Path)

	// 获取并增加计数器
	count := 1
	if val, ok := r.counters.Load(stubKey); ok {
		count = val.(int) + 1
	}
	r.counters.Store(stubKey, count)

	// 检查是否需要重置计数器
	if stub.Counter.ResetAfter > 0 && count > stub.Counter.ResetAfter {
		r.counters.Store(stubKey, 1)
		count = 1
	}

	// 根据计数器值匹配响应规则
	for _, rule := range stub.ResponseRules {
		if rule.Rule.Type == "range" && count >= rule.Rule.Start && count <= rule.Rule.End {
			r.sendStaticResponse(w, &rule.Response)
			return
		} else if rule.Rule.Type == "default" {
			r.sendStaticResponse(w, &rule.Response)
			return
		}
	}

	// 默认响应
	r.sendDefaultResponse(w)
}

// sendStaticResponse sends a static response
func (r *HTTPTestStubReconciler) sendStaticResponse(w http.ResponseWriter, static *httpteststubv1.Static) {
	// 设置响应头
	for key, value := range static.Headers {
		w.Header().Set(key, value)
	}

	// 设置状态码
	w.WriteHeader(static.Status)

	// 发送响应体
	// TODO: 实现 JSON 序列化
	w.Write([]byte(`{"result": true}`))
}

// sendDefaultResponse sends a default response
func (r *HTTPTestStubReconciler) sendDefaultResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"result": true}`))
}
