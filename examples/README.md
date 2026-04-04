# HTTP Test Stub 示例

本目录包含各种 HTTP Test Stub 配置示例，涵盖了所有支持的打桩功能。

## 示例列表

### 基础响应类型

| 示例文件 | 功能描述 | 关键特性 |
|---------|---------|---------|
| [01-static-response.yaml](01-static-response.yaml) | 静态响应 | 返回固定的 JSON 响应 |
| [02-pattern-response.yaml](02-pattern-response.yaml) | 通配符匹配 | 使用 `*` 匹配 URL 中的任意字符串 |
| [03-counter-response.yaml](03-counter-response.yaml) | 计数器响应 | 按请求次数返回不同响应 |
| [04-script-response.yaml](04-script-response.yaml) | 脚本响应 | 通过 Shell 脚本动态生成响应 |

### 高级匹配功能

| 示例文件 | 功能描述 | 关键特性 |
|---------|---------|---------|
| [05-script-delay-response.yaml](05-script-delay-response.yaml) | 脚本延迟响应 | 脚本执行后延迟返回 |
| [06-script-rule-response.yaml](06-script-rule-response.yaml) | 脚本+计数器组合 | 根据请求次数执行不同脚本 |
| [07-regex-response.yaml](07-regex-response.yaml) | 正则匹配 | 使用正则表达式匹配 URL |
| [08-inline-script.yaml](08-inline-script.yaml) | 内联脚本 | 直接在配置中定义脚本逻辑 |
| [09-error-response.yaml](09-error-response.yaml) | 错误响应 | 返回 4xx/5xx 错误状态码 |
| [10-complex-json.yaml](10-complex-json.yaml) | 复杂 JSON | 返回嵌套的复杂 JSON 结构 |

### 延迟模拟

| 示例文件 | 功能描述 | 关键特性 |
|---------|---------|---------|
| [11-delay-fixed.yaml](11-delay-fixed.yaml) | 固定延迟 | 固定延迟 2 秒后返回响应 |
| [12-delay-random.yaml](12-delay-random.yaml) | 随机延迟 | 100-1000ms 随机延迟 |

### 请求体匹配

| 示例文件 | 功能描述 | 关键特性 |
|---------|---------|---------|
| [13-body-match-equal.yaml](13-body-match-equal.yaml) | 精确匹配 | 请求体完全匹配指定内容 |
| [14-body-match-contains.yaml](14-body-match-contains.yaml) | 包含匹配 | 请求体包含指定字符串 |
| [15-body-match-regex.yaml](15-body-match-regex.yaml) | 正则匹配 | 使用正则表达式匹配请求体 |

### 请求头匹配

| 示例文件 | 功能描述 | 关键特性 |
|---------|---------|---------|
| [16-header-match.yaml](16-header-match.yaml) | 请求头匹配 | 匹配 Authorization 等请求头 |
| [17-header-user-agent.yaml](17-header-user-agent.yaml) | User-Agent 匹配 | 使用正则匹配 User-Agent |

### 代理转发

| 示例文件 | 功能描述 | 关键特性 |
|---------|---------|---------|
| [18-proxy-forward.yaml](18-proxy-forward.yaml) | 简单代理 | 将请求转发到目标服务 |
| [19-proxy-transform.yaml](19-proxy-transform.yaml) | 代理+转换 | 转发请求并添加自定义头信息 |

### HTTP 方法支持

| 示例文件 | 功能描述 | 关键特性 |
|---------|---------|---------|
| [20-head-method.yaml](20-head-method.yaml) | HEAD 方法 | 处理 HEAD 请求 |
| [21-options-method.yaml](21-options-method.yaml) | OPTIONS 方法 | 处理 CORS 预检请求 |

## 快速使用

```bash
# 应用示例配置
kubectl apply -f examples/01-static-response.yaml

# 查看创建的 Stub
kubectl get httpteststub

# 测试端点
kubectl port-forward svc/k8s-http-fake-operator 8080:8080
curl http://localhost:8080/api/health
```

## 功能覆盖

这些示例覆盖了以下所有功能：

- ✅ 静态响应
- ✅ 脚本响应（Shell）
- ✅ 计数器响应（按请求次数）
- ✅ URL 精确匹配
- ✅ URL 通配符匹配
- ✅ URL 正则匹配
- ✅ 请求体精确匹配
- ✅ 请求体包含匹配
- ✅ 请求体正则匹配
- ✅ 请求头匹配
- ✅ 固定延迟模拟
- ✅ 随机延迟模拟
- ✅ 代理转发
- ✅ 代理请求/响应头转换
- ✅ GET/POST/PUT/DELETE/PATCH 方法
- ✅ HEAD 方法
- ✅ OPTIONS 方法
- ✅ HTTP/HTTPS 协议
- ✅ 错误响应（4xx/5xx）
