        "09-error-response")
            # 先删除可能存在的 Stub，确保计数器重置
            kubectl delete httpteststub error-response -n default --ignore-not-found=true
            sleep 1
            # 重新应用配置
            kubectl apply -f "$example_file" -n default
            sleep 2
            # 测试第一次请求（应该返回 500）
            test_http_request "GET" "/api/error" "500" "" ""
            test_result=$?
            ;;