# HTTP Test Stub 示例

本目录包含了一系列 HTTP Test Stub 的示例配置，展示了 operator 的各种功能。

## 示例列表

### 基础功能

1. **01-static-response.yaml** - 静态响应
   - 最简单的示例，返回固定的 JSON 响应
   - 测试: `curl http://<service-ip>:8080/api/health`

2. **02-pattern-response.yaml** - 模式匹配响应
   - 使用通配符 `*` 匹配 URL 模式
   - 测试: `curl http://<service-ip>:8080/api/users/123`

3. **07-regex-response.yaml** - 正则表达式匹配
   - 使用正则表达式匹配复杂的 URL 模式
   - 测试: `curl http://<service-ip>:8080/api/v1/users/123`

### 计数器和规则

4. **03-counter-response.yaml** - 计数器响应规则
   - 根据请求次数返回不同的响应
   - 前3次请求返回初始响应，之后返回默认响应
   - 测试: `curl -X POST http://<service-ip>:8080/api/counter` (执行多次)

5. **06-script-rule-response.yaml** - 响应规则（静态响应版）
   - 根据请求次数返回不同的静态响应
   - 测试: `curl http://<service-ip>:8080/api/users/123/details` (执行多次)

### 脚本响应

6. **04-script-response.yaml** - 脚本响应
   - 使用外部脚本动态生成响应
   - 需要提前准备好脚本文件
   - 测试: `curl http://<service-ip>:8080/api/script`

7. **05-script-delay-response.yaml** - 延迟脚本响应
   - 模拟延迟响应，用于测试超时处理
   - 测试: `curl -X POST http://<service-ip>:8080/api/delay`

8. **08-inline-script.yaml** - 内联脚本响应
   - 在 CRD 中直接定义脚本内容
   - 不需要外部脚本文件
   - 测试: `curl http://<service-ip>:8080/api/inline`

### 高级功能

9. **09-error-response.yaml** - 错误响应
   - 模拟各种 HTTP 错误状态码
   - 500, 404, 429 错误轮询
   - 测试: `curl http://<service-ip>:8080/api/error` (执行多次)

10. **10-complex-json.yaml** - 复杂 JSON 响应
    - 返回复杂的嵌套 JSON 数据结构
    - 包含自定义响应头
    - 测试: `curl -X POST http://<service-ip>:8080/api/complex`

## 快速开始

### 部署示例

```bash
# 部署单个示例
kubectl apply -f examples/01-static-response.yaml

# 部署所有示例
kubectl apply -f examples/
```

### 获取服务地址

```bash
# 获取 ClusterIP
kubectl get service k8s-http-fake-operator

# 在集群内测试
curl http://<cluster-ip>:8080/api/health
```

### 从集群外测试

```bash
# 端口转发
kubectl port-forward svc/k8s-http-fake-operator 8080:8080

# 本地测试
curl http://localhost:8080/api/health
```

## 字段说明

### URL 匹配类型

- `exact` - 精确匹配
- `pattern` - 通配符匹配，支持 `*` 匹配任意字符
- `regex` - 正则表达式匹配

### 响应类型

- `static` - 静态响应，直接返回配置的 body
- `script` - 脚本响应，执行脚本生成响应内容

### 计数器规则

- `range` - 在指定请求次数范围内生效
- `default` - 默认规则，当其他规则不匹配时生效

## 注意事项

1. 脚本响应需要提前准备好脚本文件并挂载到容器中
2. 内联脚本会在运行时写入临时文件并执行
3. 计数器会在达到 `resetAfter` 后自动重置
4. 所有示例都假设 operator 部署在 `default` 命名空间
