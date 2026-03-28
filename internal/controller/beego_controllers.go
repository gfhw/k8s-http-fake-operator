package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	beego "github.com/beego/beego/v2/server/web"
	httpteststubv1 "httpteststub.example.com/api/v1"
)

type StubController struct {
	beego.Controller
}

func (c *StubController) Get() {
	c.handleRequest("GET")
}

func (c *StubController) Post() {
	c.handleRequest("POST")
}

func (c *StubController) Put() {
	c.handleRequest("PUT")
}

func (c *StubController) Delete() {
	c.handleRequest("DELETE")
}

func (c *StubController) Patch() {
	c.handleRequest("PATCH")
}

func (c *StubController) handleRequest(method string) {
	path := c.Ctx.Input.URL()

	stub := GetMatchingStub(method, path)
	if stub == nil {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Ctx.Output.JSON(map[string]string{
			"error":  "No matching stub found",
			"method": method,
			"path":   path,
		}, false, false)
		return
	}

	stubKey := fmt.Sprintf("%s/%s", stub.Namespace, stub.Name)

	if !CheckResourceLimit(stubKey, 1000) {
		c.Ctx.Output.SetStatus(http.StatusTooManyRequests)
		c.Ctx.Output.JSON(map[string]string{
			"error": "Rate limit exceeded",
		}, false, false)
		return
	}

	c.handleStubResponse(stub, method, path)
}

func (c *StubController) handleStubResponse(stub *httpteststubv1.HTTPTestStub, method, path string) {
	stubKey := fmt.Sprintf("%s/%s", stub.Namespace, stub.Name)

	for _, s := range stub.Spec.Stubs {
		if matchRequest(method, path, &s.Request) {
			c.handleStubRules(stubKey, &s)
			return
		}
	}

	if matchRequest(method, path, &stub.Spec.Request) {
		c.handleResponse(&stub.Spec.Response)
		return
	}

	c.Ctx.Output.SetStatus(http.StatusNotFound)
	c.Ctx.Output.JSON(map[string]string{
		"error": "No matching rule found",
	}, false, false)
}

func (c *StubController) handleStubRules(stubKey string, stub *httpteststubv1.Stub) {
	count := IncrementCounter(stubKey)

	if stub.Counter.ResetAfter > 0 {
		ResetCounter(stubKey, stub.Counter.ResetAfter)
	}

	for _, rule := range stub.ResponseRules {
		if rule.Rule.Type == "range" && count >= rule.Rule.Start && count <= rule.Rule.End {
			c.sendStaticResponse(&rule.Response)
			return
		} else if rule.Rule.Type == "default" {
			c.sendStaticResponse(&rule.Response)
			return
		}
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Ctx.Output.JSON(map[string]interface{}{
		"result": true,
		"count":  count,
	}, false, false)
}

func (c *StubController) handleResponse(response *httpteststubv1.Response) {
	switch response.Type {
	case "static":
		if response.Static != nil {
			c.sendStaticResponse(response.Static)
		}
	case "script":
		if response.Script != nil {
			c.sendScriptResponse(response.Script)
		}
	default:
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Ctx.Output.JSON(map[string]string{
			"error": "Unknown response type",
		}, false, false)
	}
}

func (c *StubController) sendStaticResponse(static *httpteststubv1.Static) {
	for key, value := range static.Headers {
		c.Ctx.Output.Header(key, value)
	}

	c.Ctx.Output.SetStatus(static.Status)

	bodyBytes, err := json.Marshal(static.Body)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Ctx.Output.JSON(map[string]string{
			"error": "Failed to serialize response body",
		}, false, false)
		return
	}

	c.Ctx.Output.Body(bodyBytes)
}

func (c *StubController) sendScriptResponse(script *httpteststubv1.Script) {
	c.Ctx.Output.SetStatus(http.StatusNotImplemented)
	c.Ctx.Output.JSON(map[string]string{
		"error":  "Script response not yet implemented",
		"script": script.Name,
	}, false, false)
}

type HealthController struct {
	beego.Controller
}

func (c *HealthController) Get() {
	stubs := GetAllStubs()

	stubStats := make([]map[string]interface{}, 0)
	for _, stub := range stubs {
		stubKey := fmt.Sprintf("%s/%s", stub.Namespace, stub.Name)
		requestsPerMinute, totalRequests, uptime := GetResourceStats(stubKey)

		stats := map[string]interface{}{
			"name":                stub.Name,
			"namespace":           stub.Namespace,
			"status":              stub.Status.Status,
			"requests_per_minute": requestsPerMinute,
			"total_requests":      totalRequests,
			"uptime_seconds":      int64(uptime.Seconds()),
		}
		stubStats = append(stubStats, stats)
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	c.Ctx.Output.JSON(map[string]interface{}{
		"status":      "healthy",
		"timestamp":   time.Now().Unix(),
		"stubs_count": len(stubs),
		"stubs":       stubStats,
		"system": map[string]interface{}{
			"goroutines":      runtime.NumGoroutine(),
			"memory_alloc_mb": memStats.Alloc / 1024 / 1024,
			"memory_sys_mb":   memStats.Sys / 1024 / 1024,
			"gc_cycles":       memStats.NumGC,
		},
	}, false, false)
}

type ReadyController struct {
	beego.Controller
}

func (c *ReadyController) Get() {
	c.Ctx.Output.JSON(map[string]string{
		"status": "ready",
	}, false, false)
}
