# Service IP 配置和参数验证

## 概述

本文档介绍如何配置 Kubernetes Service 的 IP 地址，以及 Helm Chart 中的参数验证机制。

## Service IP 配置

### 1. 配置自定义 Cluster IP

在 `values.yaml` 文件中，你可以通过 `service.clusterIP` 字段指定自定义的集群 IP 地址：

```yaml
service:
  type: ClusterIP
  clusterIP: "10.96.0.100"  # 自定义的集群 IP
  httpPort: 8080
  httpsPort: 8443
  healthPort: 8081
```

### 2. 配置说明

- **`clusterIP`**：可选字段，指定自定义的集群 IP 地址
  - 留空时，Kubernetes 会自动分配 IP
  - 填写时，必须使用集群 CIDR 范围内的有效 IPv4 地址
  - 不能使用已被其他 Service 使用的 IP

### 3. 验证规则

Helm Chart 会自动验证 `clusterIP` 格式是否正确：
- 必须是有效的 IPv4 地址格式（如 `10.96.0.100`）
- 如果格式不正确，部署会失败并显示错误信息

## 必要参数验证

### 1. 必须配置的参数

以下参数是必须配置的，如果未配置，部署会失败：

| 参数 | 说明 | 示例值 |
|------|------|--------|
| `image.repository` | Docker 镜像仓库 | `k8s-http-fake-operator` |
| `service.httpPort` | HTTP 服务端口 | `8080` |
| `service.httpsPort` | HTTPS 服务端口 | `8443` |
| `service.healthPort` | 健康检查端口 | `8081` |
| `operator.server.httpPort` | Operator HTTP 端口 | `8080` |
| `operator.server.httpsPort` | Operator HTTPS 端口 | `8443` |

### 2. 条件性必须参数

当 `tls.enabled` 为 `true` 时，以下参数也必须配置：

| 参数 | 说明 | 示例值 |
|------|------|--------|
| `tls.certSecretName` | TLS 证书密钥名称 | `my-tls-secret` |

### 3. 错误提示

如果缺少必要参数，部署时会显示详细的错误信息，例如：

```
ERROR: image.repository is required. Please set the Docker image repository in values.yaml.
```

## 部署流程

### 1. 准备配置

1. **编辑 values.yaml**：
   - 设置必要的参数
   - 可选：配置自定义 Cluster IP

2. **验证配置**：
   - 确保所有必要参数都已设置
   - 确保 Cluster IP 格式正确（如果指定）

### 2. 部署

```bash
# 部署 Helm Chart
helm install k8s-http-fake-operator ./charts/k8s-http-fake-operator

# 或使用自定义 values 文件
helm install k8s-http-fake-operator ./charts/k8s-http-fake-operator -f my-values.yaml
```

### 3. 验证部署

```bash
# 检查 Service IP
kubectl get svc k8s-http-fake-operator

# 检查 Pod 状态
kubectl get pods

# 查看日志
kubectl logs deployment/k8s-http-fake-operator
```

## 配置示例

### 示例 1：默认配置（自动分配 IP）

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

### 示例 2：自定义 Cluster IP

```yaml
image:
  repository: k8s-http-fake-operator
  tag: v1.0.0

service:
  type: ClusterIP
  clusterIP: "10.96.0.100"  # 自定义 IP
  httpPort: 80
  httpsPort: 443
  healthPort: 8081

operator:
  server:
    httpPort: 80
    httpsPort: 443
```

### 示例 3：启用 TLS

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

## 故障排除

### 1. 部署失败

**症状**：`helm install` 命令失败，显示错误信息

**原因**：
- 缺少必要参数
- Cluster IP 格式不正确
- Cluster IP 已被占用

**解决方案**：
- 检查 values.yaml 中的必要参数
- 确保 Cluster IP 格式正确且未被占用
- 留空 clusterIP 字段，让 Kubernetes 自动分配

### 2. Service IP 未生效

**症状**：部署成功，但 Service IP 不是配置的值

**原因**：
- 配置的 Cluster IP 不在集群 CIDR 范围内
- 配置的 Cluster IP 已被其他 Service 使用

**解决方案**：
- 使用 `kubectl cluster-info dump | grep -E 'cluster-cidr|service-cluster-ip-range'` 查看集群 CIDR
- 选择 CIDR 范围内的未使用 IP
- 或留空 clusterIP 字段

### 3. 健康检查失败

**症状**：Pod 状态为 `CrashLoopBackOff` 或 `Unready`

**原因**：
- 端口配置错误
- 服务未正常启动

**解决方案**：
- 检查 `service.httpPort`、`service.httpsPort` 和 `service.healthPort` 配置
- 确保这些端口与 `operator.server` 中的配置一致
- 查看 Pod 日志：`kubectl logs <pod-name>`

## 总结

- **Service IP 配置**：通过 `service.clusterIP` 字段可以自定义 Service 的 IP 地址
- **参数验证**：Helm Chart 会自动验证必要参数，确保部署成功
- **错误提示**：缺少必要参数时会显示详细的错误信息
- **灵活配置**：可以根据需要选择自动分配或自定义 IP

通过这些配置和验证机制，你可以确保 k8s-http-fake-operator 服务的稳定部署和运行。