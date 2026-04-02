# k8s-http-fake-operator

一个基于 Kubernetes Operator 的 HTTP 测试桩服务，使用 Beego 框架提供高性能的 HTTP/HTTPS 服务。

## 功能特性

- **高性能**：使用 Beego 框架 + 内存缓存，支持高并发请求
- **灵活匹配**：支持 URL 精确匹配、通配符匹配和正则表达式匹配
- **多种响应**：支持静态响应、计数器响应（按请求次数返回不同响应）
- **脚本支持**：支持 Shell 脚本动态生成响应
- **资源管理**：内置速率限制和资源监控
- **配置热更新**：CR 变更自动同步到内存缓存
- **云原生**：完整的 Helm Chart 支持，支持多部署隔离

## 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                    k8s-http-fake-operator                    │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │   Beego      │  │  Controller  │  │  Memory Cache    │  │
│  │   Server     │  │  Reconciler  │  │  (sync.Map)      │  │
│  │ HTTP: 8080   │  │ Watches CR   │  │ HTTPTestStub     │  │
│  │ HTTPS: 8443  │  │ Updates      │  │ Resources        │  │
│  │ Health: 8081 │  │ Cache        │  │                  │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## 快速开始

### 前置条件

- Kubernetes 集群 (>= 1.20)
- kubectl 已配置
- Helm 3

### 部署

```bash
# 安装 Operator（CRD 会自动创建）
helm install k8s-http-fake-operator ./charts/k8s-http-fake-operator

# 查看部署状态
kubectl get pods -l app.kubernetes.io/name=k8s-http-fake-operator
```

### 创建第一个 Stub

```yaml
apiVersion: httpteststub.example.com/v1
kind: HTTPTestStub
metadata:
  name: health-check
spec:
  protocol: http
  request:
    method: GET
    url:
      type: exact
      pattern: /api/health
  response:
    type: static
    static:
      status: 200
      headers:
        Content-Type: application/json
      body:
        status: "healthy"
```

```bash
kubectl apply -f - <<EOF
apiVersion: httpteststub.example.com/v1
kind: HTTPTestStub
metadata:
  name: health-check
spec:
  protocol: http
  request:
    method: GET
    url:
      type: exact
      pattern: /api/health
  response:
    type: static
    static:
      status: 200
      headers:
        Content-Type: application/json
      body:
        status: "healthy"
EOF
```

### 测试

```bash
# 端口转发
kubectl port-forward svc/k8s-http-fake-operator 8080:8080

# 测试请求
curl http://localhost:8080/api/health
```

## URL 匹配

| 类型 | 字段 | 示例 | 说明 |
|------|------|------|------|
| `exact` | `pattern` | `/api/health` | 精确匹配 |
| `pattern` | `pattern` | `/api/users/*/aa` | 通配符匹配，`*` 匹配任意字符串 |
| `regex` | `regex` | `/api/users/\d+` | 正则表达式匹配 |

**示例**：

```yaml
# 精确匹配
url:
  type: exact
  pattern: /api/health

# 通配符匹配 - 匹配 /api/users/123, /api/users/456 等
url:
  type: pattern
  pattern: /api/users/*

# 多通配符 - 匹配 /api/users/123/aa 等
url:
  type: pattern
  pattern: /api/users/*/aa

# 正则匹配 - 匹配 /api/users/123, /api/users/456 等
url:
  type: regex
  regex: /api/users/[0-9]+
```

## 响应类型

### 1. 静态响应

```yaml
spec:
  response:
    type: static
    static:
      status: 200
      headers:
        Content-Type: application/json
      body:
        message: "Hello World"
```

### 2. 计数器响应（按请求次数返回不同响应）

```yaml
spec:
  stubs:
    - request:
        method: POST
        url:
          type: exact
          pattern: /api/counter
      responseRules:
        - rule:
            type: range
            start: 1
            end: 3
          response:
            status: 200
            body:
              message: "First three requests"
        - rule:
            type: default
          response:
            status: 200
            body:
              message: "Default response"
      counter:
        reset: true
        resetAfter: 10
```

### 3. 脚本响应

```yaml
spec:
  response:
    type: script
    script:
      name: example-script
      type: shell
      path: example.sh
      timeout: 10
      env:
        CUSTOM_VAR: "value"
```

**脚本输出格式**：

```json
{
  "body": {"message": "Hello from script!", "timestamp": 1234567890},
  "headers": {"Content-Type": "application/json"},
  "status": 200
}
```

**示例脚本**（`example.sh`）：

```bash
#!/bin/bash
echo '{"body": {"message": "Hello from script!", "timestamp": '$(date +%s)'}, "headers": {"Content-Type": "application/json"}, "status": 200}'
```

## 配置参考

### Helm Chart 配置

| 配置类别 | 字段 | 说明 | 示例值 |
|---------|------|------|--------|
| **Service** | `service.type` | Service 类型 | `ClusterIP` |
|  | `service.clusterIP` | 自定义集群 IP | `10.96.0.100` |
|  | `service.httpPort` | HTTP 端口 | `8080` |
| **API Group** | `apiGroup.suffix` | API Group 后缀 | `service1` |
| **镜像** | `image.repository` | 镜像仓库 | `k8s-http-fake-operator` |
|  | `image.tag` | 镜像标签 | `latest` |
| **TLS** | `tls.enabled` | 启用 TLS | `true` |
|  | `tls.autoGenerate` | 自动生成自签名证书 | `true` |
|  | `tls.certSecretName` | TLS 证书 Secret | `my-tls-secret` |
|  | `tls.certPath` | 容器内证书路径 | `/etc/tls/tls.crt` |
| **脚本** | `operator.scripts.enabled` | 启用脚本功能 | `true` |
|  | `operator.scripts.hostPath` | 宿主机脚本目录 | `/path/to/scripts` |

### 配置示例

**1. 自定义 Cluster IP**：

```yaml
service:
  type: ClusterIP
  clusterIP: "10.96.0.100"
  httpPort: 8080
```

**2. 启用 TLS（使用自签名证书）**：

```yaml
tls:
  enabled: true
  autoGenerate: true  # 默认使用自签名证书
```

**3. 使用自定义证书**：

```bash
# 创建 TLS Secret
kubectl create secret tls my-cert --cert=cert.crt --key=cert.key

# 部署时指定
helm install k8s-http-fake-operator ./charts/k8s-http-fake-operator \
  --set tls.autoGenerate=false \
  --set tls.certSecretName=my-cert
```

## 协议分离

同时监听 HTTP (8080) 和 HTTPS (8443) 端口。通过 CR 的 `spec.protocol` 控制：

| protocol | 说明 | HTTP:8080 | HTTPS:8443 |
|----------|------|-----------|------------|
| `http` | 仅 HTTP | ✅ | ❌ 404 |
| `https` | 仅 HTTPS | ❌ 404 | ✅ |
| `both` 或空 | 两者 | ✅ | ✅ |

**示例（HTTPS only）**：

```yaml
spec:
  protocol: https
  request:
    method: GET
    url:
      type: exact
      pattern: /api/secure
  response:
    type: static
    static:
      status: 200
      headers:
        Content-Type: application/json
      body:
        message: "Secure endpoint"
```

**测试**：

```bash
curl -k https://<service>:8443/api/secure   # HTTPS 正常响应
curl http://<service>:8080/api/secure       # HTTP 返回 404
```

**3. 启用脚本功能**：

```yaml
operator:
  scripts:
    enabled: true
    hostPath: /path/to/scripts/on/host
```

**4. 多部署配置**（不同服务隔离）：

```yaml
# 部署命令：helm install http-stub-servicea ./charts/k8s-http-fake-operator -f service-a-values.yaml
apiGroup:
  suffix: "servicea"

service:
  clusterIP: "10.96.0.100"
```

对应的 CR 使用 `apiVersion: httpteststub.servicea.com/v1`

### 多部署说明

通过动态 API Group 配置，可以在同一集群中部署多套实例，每套实例完全隔离：

- 部署 1：`httpteststub.servicea.com`
- 部署 2：`httpteststub.serviceb.com`

每套部署只监听属于自己的 CR，互不干扰。

## 高级功能

### 脚本与计数器结合

```yaml
spec:
  stubs:
    - responseRules:
        - rule:
            type: range
            start: 1
            end: 3
          script:
            name: first-script
            type: shell
            path: first_three.sh
        - rule:
            type: default
          script:
            name: default-script
            type: shell
            path: default.sh
      counter:
        reset: true
        resetAfter: 10
```

### 环境变量

脚本执行时会自动注入以下环境变量：

- `REQUEST_METHOD`：HTTP 请求方法
- `REQUEST_PATH`：请求路径
- `REQUEST_HEADERS`：请求头（JSON 格式）
- `REQUEST_BODY`：请求体

## 测试服务

### 端口转发

```bash
# HTTP
kubectl port-forward svc/k8s-http-fake-operator 8080:8080

# HTTPS
kubectl port-forward svc/k8s-http-fake-operator 8443:8443

# 健康检查
kubectl port-forward svc/k8s-http-fake-operator 8081:8081
```

### 测试请求

```bash
curl http://localhost:8080/api/health
curl http://localhost:8080/api/users/123
curl -X POST http://localhost:8080/api/counter
```

### 资源统计

```bash
curl http://localhost:8081/healthz | jq .
```

## 开发指南

### 本地开发

```bash
# 安装依赖
go mod tidy

# 本地运行
go run ./cmd/main.go
```

### 构建镜像

```bash
# 构建
docker build -t k8s-http-fake-operator:latest .

# 推送
docker push your-registry/k8s-http-fake-operator:latest
```

### 容器持续运行

容器使用无限循环机制确保进程持续运行，崩溃后会自动重启（间隔 5 秒）。

### 生成代码

```bash
# 生成 DeepCopy 方法
go run sigs.k8s.io/controller-tools/cmd/controller-gen@latest object paths=./api/v1/
```

## 故障排查

### 常见问题

| 问题 | 可能原因 | 解决方案 |
|------|---------|---------|
| Stub 未生效 | CR 状态不是 `Running` | 检查 CR 状态和 Operator 日志 |
| 请求返回 404 | URL 匹配规则错误 | 确认请求方法和 URL 与配置匹配 |
| helm install 失败 | Cluster IP 冲突或格式错误 | 检查 IP 是否可用或留空自动分配 |
| CR 无法创建 | API Group 不存在 | 确认部署已创建 CRD 且 `apiVersion` 匹配 |
| 容器不断重启 | manager 崩溃 | 查看日志：`kubectl logs <pod-name>` |

### 排查命令

```bash
# 查看 Operator 日志
kubectl logs -l app.kubernetes.io/name=k8s-http-fake-operator -f

# 检查 CR 状态
kubectl get httpteststub -A
kubectl describe httpteststub <name>

# 检查 Pod 状态
kubectl get pods -A | grep http-fake
```

## 与其他方案对比

| 特性 | k8s-http-fake-operator | WireMock | MockServer | Hoverfly | Mountebank | Postman Mock |
|------|------------------------|----------|------------|----------|------------|--------------|
| **部署方式** | Kubernetes Operator | JVM 应用 | Node.js/Java | Go 二进制 | Node.js | SaaS/本地 |
| **配置方式** | CRD + YAML | JSON/API | JSON/API | JSON/API | JSON/API | GUI/API |
| **动态更新** | ✅ 热更新 | ❌ 需要重启 | ❌ 需要重启 | ✅ API 更新 | ❌ 需要重启 | ❌ 需要重启 |
| **Kubernetes 集成** | ✅ 原生 | ❌ 需额外部署 | ❌ 需额外部署 | ❌ 需额外部署 | ❌ 需额外部署 | ❌ 不支持 |
| **GitOps 支持** | ✅ 原生 | ⚠️ 有限 | ⚠️ 有限 | ⚠️ 有限 | ⚠️ 有限 | ❌ |
| **脚本支持** | ✅ Shell/Python | ⚠️ Java | ⚠️ JavaScript | ✅ JavaScript | ✅ JavaScript | ❌ |
| **计数器规则** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **正则匹配** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **HTTPS 支持** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **学习曲线** | 低（YAML） | 中 | 中 | 中 | 中 | 低 |
| **资源占用** | 低 | 高 | 中 | 低 | 中 | - |
| **多租户隔离** | ✅ 命名空间 | ❌ | ❌ | ❌ | ❌ | ❌ |
| **CRD 管理** | ✅ 原生 | ❌ | ❌ | ❌ | ❌ | ❌ |

### 核心优势

1. **云原生设计**
   - 专为 Kubernetes 设计，充分利用 CRD 生态
   - 与 Kubernetes RBAC、Policy、网络策略无缝集成
   - 支持多命名空间隔离，实现多租户

2. **GitOps 原生**
   - 配置即代码，版本化管理
   - CR 变更自动同步，无需重启
   - 可使用 ArgoCD、Flux 等 GitOps 工具管理

3. **计数器规则**
   - 业界独有功能！可按请求次数返回不同响应
   - 适用于模拟限流、降级、灰度发布等场景
   - 支持自动重置

4. **轻量高效**
   - 基于 Beego 框架，性能高
   - 资源占用低，适合边缘计算场景
   - 单二进制部署，无外部依赖

## 许可证

MIT License

Copyright (c) 2024 k8s-http-fake-operator

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
