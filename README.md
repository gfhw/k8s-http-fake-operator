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

### Helm Chart 配置

#### Service IP 配置

在 `values.yaml` 文件中，你可以通过 `service.clusterIP` 字段指定自定义的集群 IP 地址：

```yaml
service:
  type: ClusterIP
  clusterIP: "10.96.0.100"  # 自定义的集群 IP（支持 IPv4 和 IPv6）
  httpPort: 8080
  httpsPort: 8443
  healthPort: 8081
```

**配置说明**：
- **`clusterIP`**：可选字段，指定自定义的集群 IP 地址
  - 留空时，Kubernetes 会自动分配 IP
  - 填写时，必须使用集群 CIDR 范围内的有效 IPv4 或 IPv6 地址
  - 不能使用已被其他 Service 使用的 IP

**验证规则**：
- 必须是有效的 IPv4 或 IPv6 地址格式（如 `10.96.0.100` 或 `2001:db8::1`）
- 如果格式不正确，部署会失败并显示错误信息

#### API Group 配置（多部署支持）

通过动态 API Group 配置，你可以在同一 Kubernetes 集群中部署多套 k8s-http-fake-operator 实例，每套实例监听不同的 CRD，实现完全隔离的多服务打桩。

**配置方法**：

```yaml
apiGroup:
  # 使用 suffix（推荐）
  suffix: "service1"  # 将生成 httpteststub.service1.com
  
  # 或使用完整名称
  fullName: "my-custom-api-group.io"  # 完全自定义
  
  # 如果都留空，默认使用 httpteststub.example.com
```

**部署示例**：

```bash
# 部署第一套实例（API Group: httpteststub.service1.com）
helm install http-stub-service1 ./charts/k8s-http-fake-operator --set apiGroup.suffix=service1

# 部署第二套实例（API Group: httpteststub.service2.com）
helm install http-stub-service2 ./charts/k8s-http-fake-operator --set apiGroup.suffix=service2

# 部署第三套实例（使用默认 API Group: httpteststub.example.com）
helm install http-stub-default ./charts/k8s-http-fake-operator
```

**CR 创建示例**：

```yaml
# 第一套实例的 CR
apiVersion: httpteststub.service1.com/v1
kind: HTTPTestStub
metadata:
  name: stub1
  namespace: default
spec:
  protocol: http
  request:
    method: GET
    url:
      type: exact
      pattern: /api/service1
  response:
    type: static
    static:
      status: 200
      body: {"service": "service1"}

---
# 第二套实例的 CR
apiVersion: httpteststub.service2.com/v1
kind: HTTPTestStub
metadata:
  name: stub1
  namespace: default
spec:
  protocol: http
  request:
    method: GET
    url:
      type: exact
      pattern: /api/service2
  response:
    type: static
    static:
      status: 200
      body: {"service": "service2"}
```

#### 必要参数验证

**必须配置的参数**：

| 参数 | 说明 | 示例值 |
|------|------|--------|
| `image.repository` | Docker 镜像仓库 | `k8s-http-fake-operator` |
| `service.httpPort` | HTTP 服务端口 | `8080` |
| `service.httpsPort` | HTTPS 服务端口 | `8443` |
| `service.healthPort` | 健康检查端口 | `8081` |
| `operator.server.httpPort` | Operator HTTP 端口 | `8080` |
| `operator.server.httpsPort` | Operator HTTPS 端口 | `8443` |

**条件性必须参数**：

当 `tls.enabled` 为 `true` 时，以下参数也必须配置：

| 参数 | 说明 | 示例值 |
|------|------|--------|
| `tls.certSecretName` | TLS 证书密钥名称 | `my-tls-secret` |

**错误提示**：

如果缺少必要参数，部署时会显示详细的错误信息，例如：

```
ERROR: image.repository is required. Please set the Docker image repository in values.yaml.
```

#### Helm 自定义配置示例

**示例 1：默认配置（自动分配 IP）**

```yaml
image:
  repository: k8s-http-fake-operator
  tag: latest

service:
  type: ClusterIP
  clusterIP: ""  # 留空，自动分配
  httpPort: 8080
  httpsPort: 8443
  healthPort: 8081

operator:
  server:
    httpPort: 8080
    httpsPort: 8443
```

**示例 2：自定义 Cluster IP**

```yaml
image:
  repository: k8s-http-fake-operator
  tag: v1.0.0

service:
  type: ClusterIP
  clusterIP: "10.96.0.100"  # 自定义 IPv4
  httpPort: 80
  httpsPort: 443
  healthPort: 8081

operator:
  server:
    httpPort: 80
    httpsPort: 443
```

**示例 3：使用 IPv6**

```yaml
service:
  type: ClusterIP
  clusterIP: "2001:db8::1"  # 自定义 IPv6
  httpPort: 8080
  httpsPort: 8443
  healthPort: 8081
```

**示例 4：启用 TLS**

```yaml
image:
  repository: k8s-http-fake-operator
  tag: latest

service:
  type: ClusterIP
  httpPort: 8080
  httpsPort: 8443
  healthPort: 8081

tls:
  enabled: true
  certSecretName: my-tls-secret  # 必须配置

operator:
  server:
    httpPort: 8080
    httpsPort: 8443
```

**示例 5：多部署配置**

```yaml
# 第一套部署
apiGroup:
  suffix: "service1"

service:
  type: ClusterIP
  clusterIP: "10.96.0.100"
  httpPort: 8080
  httpsPort: 8443

# 第二套部署
apiGroup:
  suffix: "service2"

service:
  type: ClusterIP
  clusterIP: "10.96.0.101"
  httpPort: 8080
  httpsPort: 8443
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

### Service IP 配置故障排除

#### 1. 部署失败

**症状**：`helm install` 命令失败，显示错误信息

**原因**：
- 缺少必要参数
- Cluster IP 格式不正确
- Cluster IP 已被占用

**解决方案**：
- 检查 values.yaml 中的必要参数
- 确保 Cluster IP 格式正确且未被占用
- 留空 clusterIP 字段，让 Kubernetes 自动分配

#### 2. Service IP 未生效

**症状**：部署成功，但 Service IP 不是配置的值

**原因**：
- 配置的 Cluster IP 不在集群 CIDR 范围内
- 配置的 Cluster IP 已被其他 Service 使用

**解决方案**：
- 使用 `kubectl cluster-info dump | grep -E 'cluster-cidr|service-cluster-ip-range'` 查看集群 CIDR
- 选择 CIDR 范围内的未使用 IP
- 或留空 clusterIP 字段

#### 3. 健康检查失败

**症状**：Pod 状态为 `CrashLoopBackOff` 或 `Unready`

**原因**：
- 端口配置错误
- 服务未正常启动

**解决方案**：
- 检查 `service.httpPort`、`service.httpsPort` 和 `service.healthPort` 配置
- 确保这些端口与 `operator.server` 中的配置一致
- 查看 Pod 日志：`kubectl logs <pod-name>`

### 多部署配置故障排除

#### 1. CRD 冲突

**症状**：部署时提示 CRD 已存在

**解决方案**：
- 确保每个部署使用不同的 API Group
- 检查现有的 CRD：`kubectl get crd | grep httpteststub`

#### 2. Service IP 冲突

**症状**：Service 创建失败

**解决方案**：
- 检查 IP 是否已被使用：`kubectl get svc --all-namespaces`
- 使用自动分配 IP（留空 `clusterIP`）

#### 3. CR 无法创建

**症状**：创建 CR 时提示 API Group 不存在

**解决方案**：
- 确保部署已成功创建 CRD：`kubectl get crd`
- 检查 CR 的 `apiVersion` 是否与部署的 API Group 匹配

#### 4. 请求路由错误

**症状**：请求发送到错误的部署实例

**解决方案**：
- 检查 Service 的 IP 和端口配置
- 确认 CR 的 `apiVersion` 与目标部署的 API Group 匹配
- 使用 `kubectl get svc` 查看所有 Service 的 IP 地址

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
