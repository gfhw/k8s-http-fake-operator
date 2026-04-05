        "09-error-response")
            # 彻底清理资源，确保计数器完全重置
            log_info "彻底清理 error-response 资源..."
            kubectl delete httpteststub error-response -n default --ignore-not-found=true
            sleep 3  # 等待资源完全删除
            
            # 重新应用配置
            log_info "重新应用 error-response 配置..."
            kubectl apply -f "$example_file" -n default
            sleep 5  # 等待 Stub 完全就绪和初始化
            
            # 再次等待 Stub 状态为 Running
            if ! wait_for_stub_ready "error-response"; then
                test_result=1
                break
            fi
            
            # 测试第一次请求（应该返回 500）
            test_http_request "GET" "/api/error" "500" "" ""
            test_result=$?
            ;;
