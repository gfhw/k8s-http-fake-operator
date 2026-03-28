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

**直接应用配置**：
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

支持多个通配符的复杂模式匹配：

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

**复杂模式示例**：

```yaml
# 匹配 /api/users/123/aa, /api/users/456/aa 等
url:
  type: pattern
  pattern: /api/users/*/aa

# 匹配 /api/v1/users/123/details, /api/v2/users/456/details 等
url:
  type: pattern
  pattern: /api/*/users/*/details
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

### 6. 脚本响应

支持通过 Shell 脚本动态生成响应，实现更复杂的业务逻辑：

```yaml
apiVersion: httpteststub.example.com/v1
kind: HTTPTestStub
metadata:
  name: script-response
  namespace: default
spec:
  protocol: http
  request:
    method: GET
    url:
      type: exact
      pattern: /api/script
  response:
    type: script
    script:
      name: example-script
      type: shell
      path: example.sh
      timeout: 10
      args:
        - "/scripts"
      env:
        CUSTOM_VAR: "custom_value"
```

**脚本输出格式**：

脚本输出单行 JSON 格式，包含响应体、响应头和状态码：

```json
{
  "body": {"message": "Hello from script!", "timestamp": 1234567890},
  "headers": {"Content-Type": "application/json"},
  "status": 200
}
```

**字段说明**：
- `body`: 响应体内容（JSON 格式）
- `headers`: 响应头（可选）
- `status`: HTTP 状态码（可选，默认 200）

**示例脚本**（`example.sh`）：

```bash
#!/bin/bash

echo '{"body": {"message": "Hello from script!", "timestamp": '$(date +%s)', "env_var": "'$CUSTOM_VAR'"}, "headers": {"Content-Type": "application/json"}, "status": 200}'

**延迟响应示例**：

```yaml
apiVersion: httpteststub.example.com/v1
kind: HTTPTestStub
metadata:
  name: script-delay-response
  namespace: default
spec:
  protocol: http
  request:
    method: POST
    url:
      type: exact
      pattern: /api/delay
  response:
    type: script
    script:
      name: delay-script
      type: shell
      path: delay.sh
      timeout: 15
      args:
        - "5"
```

**延迟脚本示例**（`delay.sh`）：

```bash
#!/bin/bash

DELAY_SECONDS=${1:-5}

echo "Sleeping for $DELAY_SECONDS seconds..."
sleep $DELAY_SECONDS

echo '{"body": {"message": "Delayed response", "delay": "'$DELAY_SECONDS'", "timestamp": '$(date +%s)'}, "headers": {"Content-Type": "application/json"}, "status": 200}'
```

**脚本配置说明**：

| 字段 | 类型 | 说明 | 示例值 |
|------|------|------|--------|
| `name` | string | 脚本名称 | `example-script` |
| `type` | string | 脚本类型（shell, python, py） | `shell` |
| `path` | string | 脚本文件路径（相对或绝对路径） | `example.sh` |
| `timeout` | int | 超时时间（秒） | `10` |
| `args` | []string | 脚本参数 | `["5"]` |
| `env` | map[string]string | 环境变量 | `{"CUSTOM_VAR": "value"}` |
| `content` | string | 内联脚本内容（可选） | `#!/bin/bash\necho "hello"` |

**环境变量**：

脚本执行时会自动注入以下环境变量：
- `REQUEST_METHOD`：HTTP 请求方法
- `REQUEST_PATH`：请求路径
- `REQUEST_HEADERS`：请求头（JSON 格式）
- `REQUEST_BODY`：请求体

**启用脚本功能**：

在 `values.yaml` 中配置：

```yaml
operator:
  scripts:
    enabled: true
    directory: "/scripts"
    hostPath: "/path/to/scripts/on/host"
    readOnly: true
```

部署时指定脚本目录：

```bash
helm install k8s-http-fake-operator ./charts/k8s-http-fake-operator \
  --set operator.scripts.enabled=true \
  --set operator.scripts.hostPath=/path/to/scripts/on/host
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

#### ResponseRule 字段

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `rule` | Rule | 是 | 规则条件 |
| `response` | Static | 否 | 静态响应配置（与 script 二选一） |
| `script` | Script | 否 | 脚本响应配置（与 response 二选一） |

#### Rule 字段

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `type` | string | 是 | 规则类型：`range`, `default` |
| `start` | int | 否 | 范围开始（type=range 时使用） |
| `end` | int | 否 | 范围结束（type=range 时使用） |

### Helm Chart 配置

#### 核心配置项

| 配置类别 | 字段 | 说明 | 示例值 |
|---------|------|------|--------|
| **Service 配置** | `service.type` | Service 类型 | `ClusterIP` |
|  | `service.clusterIP` | 自定义集群 IP（可选，支持 IPv4/IPv6） | `10.96.0.100` 或 `2001:db8::1` |
|  | `service.httpPort` | HTTP 服务端口 | `8080` |
|  | `service.httpsPort` | HTTPS 服务端口 | `8443` |
|  | `service.healthPort` | 健康检查端口 | `8081` |
| **API Group 配置** | `apiGroup.suffix` | API Group 后缀（生成 `httpteststub.<suffix>.com`） | `service1` |
|  | `apiGroup.fullName` | 完整 API Group 名称（覆盖 suffix） | `my-custom-api-group.io` |
| **镜像配置** | `image.repository` | Docker 镜像仓库 | `k8s-http-fake-operator` |
|  | `image.tag` | 镜像标签 | `latest` |
| **TLS 配置** | `tls.enabled` | 是否启用 TLS | `true` |
|  | `tls.certSecretName` | TLS 证书密钥名称（启用 TLS 时必填） | `my-tls-secret` |
| **Operator 配置** | `operator.server.httpPort` | Operator HTTP 端口 | `8080` |
|  | `operator.server.httpsPort` | Operator HTTPS 端口 | `8443` |
| **脚本配置** | `operator.scripts.enabled` | 是否启用脚本功能 | `true` |
|  | `operator.scripts.directory` | 脚本目录路径 | `/scripts` |
|  | `operator.scripts.hostPath` | 宿主机脚本目录路径 | `/path/to/scripts` |
|  | `operator.scripts.readOnly` | 脚本目录只读挂载 | `true` |

#### 配置示例

**1. 默认配置（自动分配 IP）**

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

**2. 自定义 Cluster IP**

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

**3. 多部署配置**

```yaml
# 部署命令：helm install http-stub-service1 ./charts/k8s-http-fake-operator --set apiGroup.suffix=service1
apiGroup:
  suffix: "service1"  # 生成 httpteststub.service1.com

service:
  type: ClusterIP
  clusterIP: "10.96.0.100"
  httpPort: 8080
  httpsPort: 8443
```

**4. 启用 TLS**

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

**5. 启用脚本功能**

```yaml
operator:
  scripts:
    enabled: true
    directory: "/scripts"
    hostPath: "/path/to/scripts/on/host"
    readOnly: true
```

**6. 脚本响应规则**

支持结合计数器规则使用脚本响应，实现更灵活的响应逻辑：

```yaml
apiVersion: httpteststub.example.com/v1
kind: HTTPTestStub
metadata:
  name: script-rule-response
  namespace: default
spec:
  protocol: http
  request:
    method: GET
    url:
      type: pattern
      pattern: /api/users/*/details
  stubs:
    - request:
        method: GET
        url:
          type: pattern
          pattern: /api/users/*/details
      responseRules:
        - rule:
            type: range
            start: 1
            end: 3
          script:
            name: first-three-script
            type: shell
            path: first_three.sh
            timeout: 10
            args:
              - "first"
        - rule:
            type: default
          script:
            name: default-script
            type: shell
            path: default.sh
            timeout: 10
            args:
              - "default"
      counter:
        reset: true
        resetAfter: 10
```

**7. 多部署配置**

**部署命令**：
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

### 基础排查命令

```bash
# 查看 Operator 日志
kubectl logs -l app.kubernetes.io/name=k8s-http-fake-operator -f

# 检查 CR 状态
kubectl get httpteststub -A
kubectl describe httpteststub <name> -n <namespace>

# 检查 Service 状态
kubectl get svc -A | grep http-fake

# 检查 Pod 状态
kubectl get pods -A | grep http-fake
```

### 常见问题排查

| 问题类型 | 症状 | 可能原因 | 解决方案 |
|---------|------|---------|---------|
| **Stub 配置** | Stub 未生效 | CR 状态不是 `Running` | 检查 CR 状态和 Operator 日志 |
|  | 请求返回 404 | URL 匹配规则错误 | 确认请求方法和 URL 与配置匹配 |
|  |  | 多个 stub 冲突 | 检查是否有重复的 URL 配置 |
| **部署问题** | `helm install` 失败 | 缺少必要参数 | 检查 values.yaml 中的必填参数 |
|  |  | Cluster IP 格式错误 | 使用有效的 IPv4/IPv6 地址 |
|  |  | Cluster IP 被占用 | 选择未使用的 IP 或留空自动分配 |
| **Service 问题** | Service IP 未生效 | IP 不在集群 CIDR 范围内 | 查看集群 CIDR 并选择范围内的 IP |
|  |  | IP 已被其他 Service 使用 | 选择未使用的 IP 或留空自动分配 |
| **Pod 问题** | `CrashLoopBackOff` 或 `Unready` | 端口配置错误 | 确保 `service` 和 `operator.server` 端口一致 |
|  |  | 服务未正常启动 | 查看 Pod 日志：`kubectl logs <pod-name>` |
| **多部署问题** | CRD 冲突 | 部署时提示 CRD 已存在 | 确保每个部署使用不同的 API Group |
|  | CR 无法创建 | 提示 API Group 不存在 | 确认部署已创建 CRD 且 `apiVersion` 匹配 |
|  | 请求路由错误 | 请求发送到错误实例 | 检查 Service IP 配置和 CR 的 `apiVersion` |
| **性能问题** | 请求响应慢 | 资源限制达到上限 | 查看 `/healthz` 端点的系统资源使用情况 |
|  |  | 并发请求过多 | 检查速率限制配置 |

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
