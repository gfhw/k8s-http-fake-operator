#!/bin/bash

set -e

COLOR_GREEN='\033[0;32m'
COLOR_RED='\033[0;31m'
COLOR_YELLOW='\033[1;33m'
COLOR_BLUE='\033[0;34m'
COLOR_RESET='\033[0m'

log_info() {
    echo -e "${COLOR_BLUE}[INFO]${COLOR_RESET} $1"
}

log_success() {
    echo -e "${COLOR_GREEN}[SUCCESS]${COLOR_RESET} $1"
}

log_error() {
    echo -e "${COLOR_RED}[ERROR]${COLOR_RESET} $1"
}

log_warn() {
    echo -e "${COLOR_YELLOW}[WARN]${COLOR_RESET} $1"
}

check_operator_status() {
    log_info "检查 k8s-http-fake-operator 组件状态..."
    
    local pod_status
    pod_status=$(kubectl get pods -l app.kubernetes.io/name=k8s-http-fake-operator -n default -o jsonpath='{.items[0].status.phase}' 2>/dev/null || echo "NotFound")
    
    if [ "$pod_status" != "Running" ]; then
        log_error "Operator Pod 状态异常: $pod_status"
        log_info "Pod 详细信息:"
        kubectl get pods -l app.kubernetes.io/name=k8s-http-fake-operator -n default -o wide
        log_info "Pod 日志:"
        kubectl logs -l app.kubernetes.io/name=k8s-http-fake-operator -n default --tail=50
        return 1
    fi
    
    log_success "Operator Pod 运行正常"
    return 0
}

wait_for_stub_ready() {
    local stub_name=$1
    local max_wait=30
    local count=0
    
    log_info "等待 Stub '$stub_name' 就绪..."
    
    while [ $count -lt $max_wait ]; do
        local phase
        phase=$(kubectl get httpteststub "$stub_name" -n default -o jsonpath='{.status.phase}' 2>/dev/null || echo "NotFound")
        
        if [ "$phase" == "Running" ]; then
            log_success "Stub '$stub_name' 已就绪"
            return 0
        fi
        
        sleep 1
        count=$((count + 1))
    done
    
    log_error "Stub '$stub_name' 就绪超时"
    log_info "Stub 详细信息:"
    kubectl get httpteststub "$stub_name" -n default -o yaml
    return 1
}

test_http_request() {
    local method=$1
    local path=$2
    local expected_status=$3
    local data=$4
    local headers=$5
    
    log_info "测试 HTTP 请求: $method $path"
    
    local curl_cmd="curl -s -o /tmp/response_body.txt -w '%{http_code}' -X $method"
    
    if [ -n "$data" ]; then
        curl_cmd="$curl_cmd -d '$data'"
    fi
    
    if [ -n "$headers" ]; then
        curl_cmd="$curl_cmd $headers"
    fi
    
    curl_cmd="$curl_cmd http://localhost:8080$path"
    
    log_info "执行命令: $curl_cmd"
    
    local actual_status
    actual_status=$(eval "$curl_cmd" 2>&1) || actual_status="000"
    
    log_info "响应状态码: $actual_status"
    
    if [ -f /tmp/response_body.txt ] && [ -s /tmp/response_body.txt ]; then
        log_info "响应内容:"
        cat /tmp/response_body.txt | head -c 500
        echo ""
    fi
    
    if [ "$actual_status" == "$expected_status" ]; then
        log_success "HTTP 请求测试通过 (期望: $expected_status, 实际: $actual_status)"
        return 0
    else
        log_error "HTTP 请求测试失败 (期望: $expected_status, 实际: $actual_status)"
        return 1
    fi
}

test_example() {
    local example_file=$1
    local example_name=$(basename "$example_file" .yaml)
    
    echo ""
    echo "========================================"
    log_info "开始测试示例: $example_name"
    echo "========================================"
    
    log_info "应用配置文件: $example_file"
    
    if ! kubectl apply -f "$example_file" -n default; then
        log_error "应用配置失败"
        return 1
    fi
    
    local stub_name
    stub_name=$(grep "name:" "$example_file" | head -1 | awk '{print $2}')
    
    if [ -z "$stub_name" ]; then
        log_error "无法从配置文件中提取 Stub 名称"
        return 1
    fi
    
    log_info "Stub 名称: $stub_name"
    
    if ! wait_for_stub_ready "$stub_name"; then
        return 1
    fi
    
    log_info "启动端口转发..."
    kubectl port-forward svc/k8s-http-fake-operator 8080:8080 -n default > /dev/null 2>&1 &
    local port_forward_pid=$!
    sleep 3
    
    if ! kill -0 $port_forward_pid 2>/dev/null; then
        log_error "端口转发启动失败"
        return 1
    fi
    
    log_success "端口转发已启动 (PID: $port_forward_pid)"
    
    local test_result=0
    
    case "$example_name" in
        "01-static-response")
            test_http_request "GET" "/api/health" "200" "" ""
            test_result=$?
            ;;
        "02-pattern-response")
            test_http_request "GET" "/api/users/123" "200" "" ""
            test_result=$?
            ;;
        "03-counter-response")
            test_http_request "POST" "/api/counter" "200" "" ""
            test_http_request "POST" "/api/counter" "200" "" ""
            test_http_request "POST" "/api/counter" "200" "" ""
            test_http_request "POST" "/api/counter" "200" "" ""
            test_result=$?
            ;;
        "04-script-response")
            test_http_request "GET" "/api/script" "200" "" ""
            test_result=$?
            ;;
        "05-script-delay-response")
            test_http_request "GET" "/api/script-delay" "200" "" ""
            test_result=$?
            ;;
        "06-script-rule-response")
            test_http_request "GET" "/api/users/test/status" "200" "" ""
            test_result=$?
            ;;
        "07-regex-response")
            test_http_request "GET" "/api/users/123" "200" "" ""
            test_result=$?
            ;;
        "08-inline-script")
            test_http_request "POST" "/api/inline-script" "200" "{\"test\": \"data\"}" ""
            test_result=$?
            ;;
        "09-error-response")
            test_http_request "GET" "/api/error" "500" "" ""
            test_result=$?
            ;;
        "10-complex-json")
            test_http_request "GET" "/api/complex" "200" "" ""
            test_result=$?
            ;;
        "11-delay-fixed")
            log_info "测试固定延迟（预期 2 秒延迟）..."
            local start_time=$(date +%s%N)
            test_http_request "GET" "/api/delay/fixed" "200" "" ""
            local end_time=$(date +%s%N)
            local duration=$(( (end_time - start_time) / 1000000 ))
            log_info "实际响应时间: ${duration}ms"
            test_result=$?
            ;;
        "12-delay-random")
            log_info "测试随机延迟（预期 100-1000ms 延迟）..."
            local start_time=$(date +%s%N)
            test_http_request "GET" "/api/delay/random" "200" "" ""
            local end_time=$(date +%s%N)
            local duration=$(( (end_time - start_time) / 1000000 ))
            log_info "实际响应时间: ${duration}ms"
            test_result=$?
            ;;
        "13-body-match-equal")
            test_http_request "POST" "/api/login" "200" '{"username": "admin", "password": "123456"}' "-H 'Content-Type: application/json'"
            test_result=$?
            ;;
        "14-body-match-contains")
            test_http_request "POST" "/api/search" "200" '{"keyword": "test"}' "-H 'Content-Type: application/json'"
            test_result=$?
            ;;
        "15-body-match-regex")
            test_http_request "POST" "/api/validate" "200" "test@example.com" "-H 'Content-Type: text/plain'"
            test_result=$?
            ;;
        "16-header-match")
            test_http_request "GET" "/api/secure" "200" "" "-H 'Authorization: Bearer token123' -H 'X-API-Version: v1'"
            test_result=$?
            ;;
        "17-header-user-agent")
            test_http_request "GET" "/api/browser" "200" "" "-H 'User-Agent: Mozilla/5.0'"
            test_result=$?
            ;;
        "18-proxy-forward")
            log_warn "代理转发测试需要真实的目标服务，跳过实际请求测试"
            test_result=0
            ;;
        "19-proxy-transform")
            log_warn "代理转发测试需要真实的目标服务，跳过实际请求测试"
            test_result=0
            ;;
        "20-head-method")
            curl -s -I http://localhost:8080/api/resource | head -5
            test_result=$?
            ;;
        "21-options-method")
            curl -s -X OPTIONS -I http://localhost:8080/api/resource | head -5
            test_result=$?
            ;;
        *)
            log_warn "未定义测试用例，执行默认 GET 请求"
            test_http_request "GET" "/api/test" "200" "" ""
            test_result=$?
            ;;
    esac
    
    log_info "停止端口转发..."
    kill $port_forward_pid 2>/dev/null || true
    sleep 2
    
    log_info "清理测试资源: $stub_name"
    kubectl delete httpteststub "$stub_name" -n default --ignore-not-found=true
    
    if [ $test_result -eq 0 ]; then
        log_success "示例 '$example_name' 测试通过 ✓"
    else
        log_error "示例 '$example_name' 测试失败 ✗"
    fi
    
    return $test_result
}

main() {
    echo ""
    echo "========================================"
    echo "  k8s-http-fake-operator 示例测试脚本"
    echo "========================================"
    echo ""
    
    log_info "测试环境信息:"
    log_info "  - Kubernetes 版本: $(kubectl version --short 2>/dev/null | grep Server | awk '{print $3}')"
    log_info "  - 当前命名空间: default"
    log_info "  - 测试时间: $(date '+%Y-%m-%d %H:%M:%S')"
    echo ""
    
    if ! check_operator_status; then
        log_error "Operator 状态检查失败，退出测试"
        exit 1
    fi
    
    echo ""
    
    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local examples_dir="$script_dir"
    
    if [ ! -d "$examples_dir" ]; then
        log_error "示例目录不存在: $examples_dir"
        exit 1
    fi
    
    local total_tests=0
    local passed_tests=0
    local failed_tests=0
    
    local example_files=$(find "$examples_dir" -name "*.yaml" -type f | grep -E "[0-9]{2}-" | sort)
    
    if [ -z "$example_files" ]; then
        log_error "未找到示例文件"
        exit 1
    fi
    
    log_info "找到以下示例文件:"
    echo "$example_files" | while read file; do
        echo "  - $(basename "$file")"
    done
    echo ""
    
    for example_file in $example_files; do
        total_tests=$((total_tests + 1))
        
        if test_example "$example_file"; then
            passed_tests=$((passed_tests + 1))
        else
            failed_tests=$((failed_tests + 1))
        fi
        
        sleep 2
    done
    
    echo ""
    echo "========================================"
    echo "  测试结果汇总"
    echo "========================================"
    echo ""
    log_info "总测试数: $total_tests"
    log_success "通过: $passed_tests"
    log_error "失败: $failed_tests"
    echo ""
    
    if [ $failed_tests -eq 0 ]; then
        log_success "所有测试通过! 🎉"
        exit 0
    else
        log_error "部分测试失败，请检查日志"
        exit 1
    fi
}

main "$@"
