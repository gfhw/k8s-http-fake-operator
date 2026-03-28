# k8s-http-fake-operator

一个基于 Kubernetes Operator 的 HTTP 测试桩服务，使用 Beego 框架提供高性能的 HTTP/HTTPS 服务。

## 功能特性

- **高性能**：使用 Beego 框架 + 内存缓存，支持高并发请求
- **灵活匹配**：支持 URL 精确匹配、通配符匹配和正则表达式匹配
- **多种响应**：支持静态响应、计数器响应（按请求次数返回不同响应）
- **资源管理**：内置速率限制和资源监控
- **配置热更新**：CR 变更自动同步到内存缓存
- **云原生**：完整的 Helm Chart 和 Docker 支持

## 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                    k8s-http-fake-operator                    │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │   Beego      │  │  Controller  │  │  Memory Cache    │  │
│  │   Server     │  │  Reconciler  │  │  (sync.Map)      │  │
│  │              │  │              │  │                  │  │
│  │ HTTP: 8080   │  │ Watches CR   │  │ ┌──────────────┐ │  │
│  │ HTTPS: 8443  │  │ Updates      │  │ │ HTTPTestStub │ │  │
│  │ Health: 8081 │  │ Cache        │  │ │ Resources    │ │  │
│  └──────────────┘  └──────────────┘  │ └──────────────┘ │  │
│          │                  │         └──────────────────┘  │
│          └──────────────────┘                               │
│                    │                                        │
└────────────────────┼────────────────────────────────────────┘
                     │
              ┌──────▼──────┐
              │ Kubernetes  │
              │ API Server  │
              └─────────────┘
```

## 快速开始

### 前置条件

- Kubernetes 集群 (>= 1.20)
- kubectl 已配置
- Helm 3 (可选，用于 Helm 部署)
- Docker (用于构建镜像)

### 安装 CRD

```bash
kubectl apply -f config/crd/bases/httpteststub.example.com_httpteststubs.yaml
```

### 部署方式一：使用 Helm Chart（推荐）

```bash
# 添加 Helm 仓库（如果需要）
# helm repo add k8s-http-fake-operator <repo-url>

# 安装 Operator
helm install k8s-http-fake-operator ./charts/k8s-http-fake-operator

# 查看部署状态
kubectl get pods -l app.kubernetes.io/name=k8s-http-fake-operator
```

### 部署方式二：使用 Kubectl

```bash
# 构建镜像（可选，或使用预构建镜像）
docker build -t k8s-http-fake-operator:latest .

# 部署
kubectl apply -f config/manager/manager.yaml
```

## 使用示例

### 1. 基本静态响应

```yaml
apiVersion: httpteststub.example.com/v1
kind: HTTPTestStub
metadata:
  name: health-check
  namespace: default
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
        timestamp: "2024-01-01T00:00:00Z"
```

应用配置：
```bash
kubectl apply -f - <<EOF
apiVersion: httpteststub.example.com/v1
kind: HTTPTestStub
metadata:
  name: health-check
  namespace: default
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

### 2. 通配符匹配

```yaml
apiVersion: httpteststub.example.com/v1
kind: HTTPTestStub
metadata:
  name: user-api
  namespace: default
spec:
  protocol: http
  request:
    method: GET
    url:
      type: pattern
      pattern: /api/users/*
  response:
    type: static
    static:
      status: 200
      headers:
        Content-Type: application/json
      body:
        users:
          - id: 1
            name: "John Doe"
          - id: 2
            name: "Jane Smith"
```

### 3. 正则表达式匹配

```yaml
apiVersion: httpteststub.example.com/v1
kind: HTTPTestStub
metadata:
  name: user-by-id
  namespace: default
spec:
  protocol: http
  request:
    method: GET
    url:
      type: regex
      regex: /api/users/[0-9]+
  response:
    type: static
    static:
      status: 200
      headers:
        Content-Type: application/json
      body:
        id: 123
        name: "Test User"
        email: "test@example.com"
```

### 4. 计数器响应（按请求次数返回不同响应）

```yaml
apiVersion: httpteststub.example.com/v1
kind: HTTPTestStub
metadata:
  name: counter-example
  namespace: default
spec:
  protocol: http
  request:
    method: POST
    url:
      type: exact
      pattern: /api/counter
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
            headers:
              Content-Type: application/json
            body:
              message: "First three requests"
              phase: "initial"
        - rule:
            type: default
          response:
            status: 200
            headers:
              Content-Type: application/json
            body:
              message: "Default response"
              phase: "normal"
      counter:
        reset: true
        resetAfter: 10
```

### 5. 多方法支持

同一个 stub 可以处理多种 HTTP 方法：

```yaml
apiVersion: httpteststub.example.com/v1
kind: HTTPTestStub
metadata:
  name: rest-api
  namespace: default
spec:
  protocol: http
  request:
    method: GET
    url:
      type: exact
      pattern: /api/resource
  response:
    type: static
    static:
      status: 200
      body:
        message: "GET response"
---
apiVersion: httpteststub.example.com/v1
kind: HTTPTestStub
metadata:
  name: rest-api-post
  namespace: default
spec:
  protocol: http
  request:
    method: POST
    url:
      type: exact
      pattern: /api/resource
  response:
    type: static
    static:
      status: 201
      body:
        message: "Created successfully"
```

## 测试服务

### 端口转发

```bash
# HTTP 服务
kubectl port-forward svc/k8s-http-fake-operator 8080:8080

# HTTPS 服务
kubectl port-forward svc/k8s-http-fake-operator 8443:8443

# 健康检查
kubectl port-forward svc/k8s-http-fake-operator 8081:8081
```

### 测试请求

```bash
# 测试健康检查端点
curl http://localhost:8081/healthz

# 测试就绪检查端点
curl http://localhost:8081/readyz

# 测试静态响应
curl http://localhost:8080/api/health

# 测试通配符匹配
curl http://localhost:8080/api/users/123

# 测试 POST 请求
curl -X POST http://localhost:8080/api/counter
```

### 查看资源统计

```bash
curl http://localhost:8081/healthz | jq .
```

返回示例：
```json
{
  "status": "healthy",
  "timestamp": 1704067200,
  "stubs_count": 3,
  "stubs": [
    {
      "name": "health-check",
      "namespace": "default",
      "status": "Running",
      "requests_per_minute": 5,
      "total_requests": 100,
      "uptime_seconds": 3600
    }
  ],
  "system": {
    "goroutines": 15,
    "memory_alloc_mb": 12,
    "memory_sys_mb": 64,
    "gc_cycles": 10
  }
}
```

## 配置参考

### HTTPTestStub CRD 规范

#### Spec 字段

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `protocol` | string | 是 | 协议类型：`http` 或 `https` |
| `request` | Request | 是 | 请求匹配规则 |
| `response` | Response | 否 | 响应配置（与 stubs 二选一） |
| `stubs` | []Stub | 否 | 多个 stub 配置（高级用法） |
| `scripts` | []Script | 否 | 脚本配置（预留） |
| `tls` | TLSConfig | 否 | TLS 配置 |

#### Request 字段

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `method` | string | 是 | HTTP 方法：`GET`, `POST`, `PUT`, `DELETE`, `PATCH` |
| `url` | URL | 是 | URL 匹配规则 |

#### URL 字段

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `type` | string | 是 | 匹配类型：`exact`, `pattern`, `regex` |
| `pattern` | string | 否 | 匹配模式（type=exact 或 pattern 时使用） |
| `regex` | string | 否 | 正则表达式（type=regex 时使用） |

#### Response 字段

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `type` | string | 是 | 响应类型：`static`, `script` |
| `static` | Static | 否 | 静态响应配置 |
| `script` | Script | 否 | 脚本响应配置 |

#### Static 字段

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `status` | int | 是 | HTTP 状态码 |
| `headers` | map[string]string | 否 | 响应头 |
| `body` | interface{} | 否 | 响应体（任意 JSON 类型） |

## 高级配置

### 启用 HTTPS

```yaml
apiVersion: httpteststub.example.com/v1
kind: HTTPTestStub
metadata:
  name: secure-api
  namespace: default
spec:
  protocol: https
  tls:
    enabled: true
    secretName: tls-secret
  request:
    method: GET
    url:
      type: exact
      pattern: /secure/endpoint
  response:
    type: static
    static:
      status: 200
      body:
        message: "Secure connection"
```

### Helm 自定义配置

创建 `custom-values.yaml`：

```yaml
replicaCount: 2

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 200m
    memory: 256Mi

service:
  type: LoadBalancer
  httpPort: 80
  httpsPort: 443

tls:
  enabled: true
  certSecretName: my-tls-secret
```

部署：
```bash
helm install k8s-http-fake-operator ./charts/k8s-http-fake-operator -f custom-values.yaml
```

## 开发指南

### 本地开发

```bash
# 克隆仓库
git clone https://github.com/gfhw/k8s-http-fake-operator.git
cd k8s-http-fake-operator

# 安装依赖
go mod tidy

# 本地运行（需要连接 Kubernetes 集群）
go run ./cmd/main.go

# 或者使用 Makefile
make run
```

### 构建镜像

```bash
# 构建 Docker 镜像
docker build -t k8s-http-fake-operator:latest .

# 推送到镜像仓库
docker tag k8s-http-fake-operator:latest your-registry/k8s-http-fake-operator:latest
docker push your-registry/k8s-http-fake-operator:latest
```

### 运行测试

```bash
# 单元测试
make test

# 端到端测试
make test-e2e
```

### 生成代码

```bash
# 生成 DeepCopy 方法
go run sigs.k8s.io/controller-tools/cmd/controller-gen@latest object paths=./api/v1/

# 生成 CRD
make manifests
```

## 性能优化

### 资源限制

每个 stub 都有独立的资源限制：
- 每分钟最大请求数：1000（可配置）
- 自动速率限制和熔断

### 并发处理

- 使用 Go 协程处理并发请求
- sync.Map 实现线程安全的内存缓存
- 无锁计数器优化

### 监控指标

通过 `/healthz` 端点可以获取：
- 每个 stub 的请求统计
- 系统资源使用情况
- Goroutine 数量
- GC 统计信息

## 故障排查

### 查看 Operator 日志

```bash
kubectl logs -l app.kubernetes.io/name=k8s-http-fake-operator -f
```

### 检查 CR 状态

```bash
kubectl get httpteststub -A
kubectl describe httpteststub <name> -n <namespace>
```

### 常见问题

1. **Stub 未生效**
   - 检查 CR 状态是否为 `Running`
   - 确认 URL 匹配规则是否正确
   - 查看 Operator 日志是否有错误

2. **请求返回 404**
   - 确认请求方法和 URL 与 stub 配置匹配
   - 检查是否有多个 stub 冲突

3. **性能问题**
   - 检查资源限制是否达到上限
   - 查看 `/healthz` 端点的系统资源使用情况

## 架构对比

### 与传统 Mock 服务对比

| 特性 | k8s-http-fake-operator | 传统 Mock 服务 |
|------|------------------------|----------------|
| 部署方式 | Kubernetes Native | 独立服务 |
| 配置管理 | CRD + GitOps | 配置文件/数据库 |
| 动态更新 | 热更新，无需重启 | 需要重启服务 |
| 资源隔离 | 多租户支持 | 单租户 |
| 可观测性 | 内置监控 | 需额外配置 |

### 与 WireMock/Mountebank 对比

| 特性 | k8s-http-fake-operator | WireMock | Mountebank |
|------|------------------------|----------|------------|
| 学习曲线 | 低（YAML 配置） | 中 | 中 |
| Kubernetes 集成 | 原生支持 | 需额外部署 | 需额外部署 |
| 性能 | 高（内存缓存） | 中 | 中 |
| 脚本支持 | 预留 | 完整 | 完整 |
| 社区生态 | 新兴 | 成熟 | 成熟 |

## 贡献指南

欢迎提交 Issue 和 Pull Request！

### 提交 Issue

- 描述清楚问题和复现步骤
- 提供相关的日志和配置
- 标注版本信息

### 提交 PR

1. Fork 仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 路线图

- [ ] 脚本响应支持（JavaScript/Python）
- [ ] 请求录制和回放
- [ ] 分布式部署支持
- [ ] Prometheus 指标导出
- [ ] Web UI 管理界面
- [ ] OpenAPI 导入支持

## 许可证

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

## 联系方式

- GitHub: [https://github.com/gfhw/k8s-http-fake-operator](https://github.com/gfhw/k8s-http-fake-operator)
- 维护者: gfhw
