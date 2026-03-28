# 多部署配置指南

## 概述

通过动态 API Group 配置，你可以在同一 Kubernetes 集群中部署多套 k8s-http-fake-operator 实例，每套实例监听不同的 CRD，实现完全隔离的多服务打桩。

## 核心概念

### API Group 动态配置

每套部署使用不同的 API Group，例如：
- 部署 1：`httpteststub.service1.com`
- 部署 2：`httpteststub.service2.com`

这样每套部署只监听属于自己的 CR，互不干扰。

## 配置方法

### 方法一：使用 suffix（推荐）

在 `values.yaml` 中配置：

```yaml
apiGroup:
  suffix: "service1"  # 将生成 httpteststub.service1.com
```

### 方法二：使用完整名称

```yaml
apiGroup:
  fullName: "my-custom-api-group.io"  # 完全自定义
```

### 方法三：使用默认值

如果不配置，默认使用 `httpteststub.example.com`。

## 部署示例

### 场景：两个不同的服务需要打桩

#### 1. 部署第一个实例（Service A）

创建 `service-a-values.yaml`：

```yaml
apiGroup:
  suffix: "servicea"

service:
  type: ClusterIP
  clusterIP: "10.96.0.100"  # 自定义 IP
  httpPort: 8080
  httpsPort: 8443

fullnameOverride: "http-stub-servicea"
```

部署：

```bash
helm install http-stub-servicea ./charts/k8s-http-fake-operator -f service-a-values.yaml
```

创建 CR（使用 `httpteststub.servicea.com/v1`）：

```yaml
apiVersion: httpteststub.servicea.com/v1
kind: HTTPTestStub
metadata:
  name: service-a-stub
  namespace: default
spec:
  protocol: http
  request:
    method: GET
    url:
      type: exact
      pattern: /api/service-a/health
  response:
    type: static
    static:
      status: 200
      headers:
        Content-Type: application/json
      body: '{"service": "A", "status": "healthy"}'
```

#### 2. 部署第二个实例（Service B）

创建 `service-b-values.yaml`：

```yaml
apiGroup:
  suffix: "serviceb"

service:
  type: ClusterIP
  clusterIP: "10.96.0.101"  # 不同的 IP
  httpPort: 8080
  httpsPort: 8443

fullnameOverride: "http-stub-serviceb"
```

部署：

```bash
helm install http-stub-serviceb ./charts/k8s-http-fake-operator -f service-b-values.yaml
```

创建 CR（使用 `httpteststub.serviceb.com/v1`）：

```yaml
apiVersion: httpteststub.serviceb.com/v1
kind: HTTPTestStub
metadata:
  name: service-b-stub
  namespace: default
spec:
  protocol: http
  request:
    method: GET
    url:
      type: exact
      pattern: /api/service-b/health
  response:
    type: static
    static:
      status: 200
      headers:
        Content-Type: application/json
      body: '{"service": "B", "status": "healthy"}'
```

### 3. 测试验证

```bash
# 测试 Service A
kubectl port-forward svc/http-stub-servicea 8080:8080
curl http://localhost:8080/api/service-a/health
# 输出: {"service": "A", "status": "healthy"}

# 测试 Service B
kubectl port-forward svc/http-stub-serviceb 8080:8080
curl http://localhost:8080/api/service-b/health
# 输出: {"service": "B", "status": "healthy"}
```

## 架构图

```
┌─────────────────────────────────────────────────────────────┐
│                     Kubernetes Cluster                       │
│                                                              │
│  ┌─────────────────────┐      ┌─────────────────────┐       │
│  │  http-stub-servicea │      │  http-stub-serviceb │       │
│  │                     │      │                     │       │
│  │  API Group:         │      │  API Group:         │       │
│  │  httpteststub.      │      │  httpteststub.      │       │
│  │  servicea.com       │      │  serviceb.com       │       │
│  │                     │      │                     │       │
│  │  IP: 10.96.0.100    │      │  IP: 10.96.0.101    │       │
│  └──────────┬──────────┘      └──────────┬──────────┘       │
│             │                            │                  │
│             ▼                            ▼                  │
│  ┌─────────────────────┐      ┌─────────────────────┐       │
│  │  CRD:               │      │  CRD:               │       │
│  │  httpteststubs.     │      │  httpteststubs.     │       │
│  │  servicea.com       │      │  serviceb.com       │       │
│  └─────────────────────┘      └─────────────────────┘       │
│                                                              │
│  ┌─────────────────────┐      ┌─────────────────────┐       │
│  │  CR (Service A)     │      │  CR (Service B)     │       │
│  │  apiVersion:        │      │  apiVersion:        │       │
│  │  httpteststub.      │      │  httpteststub.      │       │
│  │  servicea.com/v1    │      │  serviceb.com/v1    │       │
│  └─────────────────────┘      └─────────────────────┘       │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## 完整配置示例

### 部署 1：支付服务打桩

```yaml
# payment-values.yaml
apiGroup:
  suffix: "payment"

image:
  repository: k8s-http-fake-operator
  tag: latest

service:
  type: ClusterIP
  clusterIP: "10.96.1.100"
  httpPort: 8080
  httpsPort: 8443
  healthPort: 8081

fullnameOverride: "http-stub-payment"

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

```yaml
# payment-stub.yaml
apiVersion: httpteststub.payment.com/v1
kind: HTTPTestStub
metadata:
  name: payment-gateway-stub
  namespace: default
spec:
  protocol: https
  request:
    method: POST
    url:
      type: exact
      pattern: /api/v1/payment/process
  response:
    type: static
    static:
      status: 200
      headers:
        Content-Type: application/json
      body: '{"transaction_id": "txn_123456", "status": "success", "amount": 100.00}'
```

### 部署 2：用户服务打桩

```yaml
# user-values.yaml
apiGroup:
  suffix: "user"

image:
  repository: k8s-http-fake-operator
  tag: latest

service:
  type: ClusterIP
  clusterIP: "10.96.2.100"
  httpPort: 8080
  httpsPort: 8443
  healthPort: 8081

fullnameOverride: "http-stub-user"

resources:
  limits:
    cpu: 300m
    memory: 256Mi
  requests:
    cpu: 50m
    memory: 64Mi
```

```yaml
# user-stub.yaml
apiVersion: httpteststub.user.com/v1
kind: HTTPTestStub
metadata:
  name: user-service-stub
  namespace: default
spec:
  protocol: http
  request:
    method: GET
    url:
      type: pattern
      pattern: /api/v1/users/*
  response:
    type: static
    static:
      status: 200
      headers:
        Content-Type: application/json
      body: '{"user_id": "123", "username": "testuser", "email": "test@example.com"}'
```

## 管理命令

### 查看所有部署

```bash
# 查看所有 Helm releases
helm list

# 查看所有 Services
kubectl get svc

# 查看所有 CRDs
kubectl get crd | grep httpteststub

# 查看特定 API Group 的 CR
kubectl get httpteststubs.httpteststub.servicea.com
kubectl get httpteststubs.httpteststub.serviceb.com
```

### 升级部署

```bash
# 升级 Service A
helm upgrade http-stub-servicea ./charts/k8s-http-fake-operator -f service-a-values.yaml

# 升级 Service B
helm upgrade http-stub-serviceb ./charts/k8s-http-fake-operator -f service-b-values.yaml
```

### 删除部署

```bash
# 删除 Service A（包括 CRD 和 CR）
helm uninstall http-stub-servicea

# 删除 Service B
helm uninstall http-stub-serviceb
```

## 注意事项

1. **API Group 唯一性**：确保每个部署的 API Group 是唯一的，避免冲突
2. **Service IP 唯一性**：如果使用自定义 Cluster IP，确保 IP 不冲突
3. **资源隔离**：可以为每个部署配置不同的资源限制
4. **命名空间隔离**：建议在不同的命名空间中部署，或使用 `fullnameOverride` 区分资源名称
5. **CRD 清理**：删除部署时，相关的 CRD 和 CR 也会被删除

## 故障排除

### 问题：CRD 冲突

**症状**：部署时提示 CRD 已存在

**解决方案**：
- 确保每个部署使用不同的 API Group
- 检查现有的 CRD：`kubectl get crd | grep httpteststub`

### 问题：Service IP 冲突

**症状**：Service 创建失败

**解决方案**：
- 检查 IP 是否已被使用：`kubectl get svc --all-namespaces`
- 使用自动分配 IP（留空 `clusterIP`）

### 问题：CR 无法创建

**症状**：创建 CR 时提示 API Group 不存在

**解决方案**：
- 确保部署已成功创建 CRD：`kubectl get crd`
- 检查 CR 的 `apiVersion` 是否与部署的 API Group 匹配

## 总结

通过动态 API Group 配置，你可以：

- ✅ 在同一集群部署多套实例
- ✅ 每套实例完全隔离，互不干扰
- ✅ 为不同服务配置不同的打桩规则
- ✅ 独立管理每套实例的生命周期
- ✅ 灵活扩展，支持任意数量的服务

这种设计使得 k8s-http-fake-operator 成为一个强大的多服务测试打桩平台！
