# HTTPTestStub 示例配置

本目录包含 HTTPTestStub 的各种使用示例，展示了不同的功能和配置方式。

## 示例文件列表

### 1. 静态响应 (01-static-response.yaml)
最简单的示例，返回固定的静态响应。

```bash
kubectl apply -f 01-static-response.yaml
```

### 2. 通配符匹配 (02-pattern-response.yaml)
演示如何使用通配符匹配 URL 模式。

```bash
kubectl apply -f 02-pattern-response.yaml
```

### 3. 计数器响应 (03-counter-response.yaml)
演示如何根据请求次数返回不同的响应。

```bash
kubectl apply -f 03-counter-response.yaml
```

### 4. 脚本响应 (04-script-response.yaml)
演示如何使用 Shell 脚本动态生成响应。

```bash
kubectl apply -f 04-script-response.yaml
```

**注意**：使用脚本响应需要先部署脚本文件，参考 `scripts/` 目录。

### 5. 延迟响应 (05-script-delay-response.yaml)
演示如何实现延迟响应。

```bash
kubectl apply -f 05-script-delay-response.yaml
```

### 6. 脚本响应规则 (06-script-rule-response.yaml)
演示如何结合计数器规则和脚本响应，实现更灵活的响应逻辑。

```bash
kubectl apply -f 06-script-rule-response.yaml
```

## 使用脚本功能

如果示例中使用了脚本响应，需要：

1. 确保脚本文件存在于 `/scripts` 目录中
2. 在部署 Operator 时启用脚本功能：

```bash
helm install k8s-http-fake-operator ./charts/k8s-http-fake-operator \
  --set operator.scripts.enabled=true \
  --set operator.scripts.hostPath=/path/to/scripts/on/host
```

3. 将脚本文件复制到指定的宿主机目录：

```bash
cp scripts/*.sh /path/to/scripts/on/host/
```

## 测试示例

部署示例后，可以使用 curl 或其他 HTTP 客户端测试：

```bash
# 测试静态响应
curl http://<service-ip>:8080/api/health

# 测试通配符匹配
curl http://<service-ip>:8080/api/users/123

# 测试计数器响应
curl -X POST http://<service-ip>:8080/api/counter

# 测试脚本响应
curl http://<service-ip>:8080/api/script

# 测试延迟响应
curl -X POST http://<service-ip>:8080/api/delay

# 测试脚本响应规则
curl http://<service-ip>:8080/api/users/456/details
```

## 清理示例

删除所有示例：

```bash
kubectl delete -f examples/
```

或删除单个示例：

```bash
kubectl delete -f 01-static-response.yaml
```