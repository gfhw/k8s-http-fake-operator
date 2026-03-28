# ConfigMap 配置说明

## 概述

现在 k8s-http-fake-operator 使用 ConfigMap 来管理所有启动参数，这样可以通过 Helm values.yaml 统一配置，无需手动修改命令行参数。

## 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                    Helm Chart                             │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ values.yaml  │  │ ConfigMap    │  │  Deployment      │  │
│  │              │  │ Template     │  │                  │  │
│  │ operator:    │  │              │  │  mounts ConfigMap │  │
│  │   metrics:   │  │ start.sh:    │  │  executes script  │  │
│  │     enabled  │  │   --metrics  │  │                  │  │
│  │   server:    │  │   --http    │  │  /config/start.sh │  │
│  │     httpPort │  │   --https   │  │                  │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
│          │                  │                    │              │
│          └──────────────────┴────────────────────┘              │
│                              │                                 │
└──────────────────────────────┼─────────────────────────────────┘
                               │
                        ┌──────▼──────┐
                        │   Manager   │
                        │  Process    │
                        └─────────────┘
```

## 配置文件结构

### 1. values.yaml

```yaml
operator:
  metrics:
    enabled: true           # 是否启用指标
    bindAddress: ":8080"    # 指标绑定地址
    secure: true             # 是否使用 HTTPS
    certPath: ""            # 指标证书路径
    certName: "tls.crt"      # 证书文件名
    keyName: "tls.key"        # 密钥文件名
  
  healthProbe:
    bindAddress: ":8081"    # 健康检查绑定地址
  
  leaderElection:
    enabled: false           # 是否启用 Leader 选举
  
  webhook:
    certPath: ""            # Webhook 证书路径
    certName: "tls.crt"      # Webhook 证书文件名
    keyName: "tls.key"        # Webhook 密钥文件名
  
  http2:
    enabled: false           # 是否启用 HTTP/2
  
  server:
    httpPort: 8080           # HTTP 服务端口
    httpsPort: 8443          # HTTPS 服务端口
    tlsCertFile: "/etc/tls/tls.crt"  # TLS 证书文件
    tlsKeyFile: "/etc/tls/tls.key"    # TLS 密钥文件
```

### 2. ConfigMap Template

ConfigMap 包含一个 `start.sh` 脚本，将 values.yaml 中的配置转换为命令行参数：

```yaml
data:
  start.sh: |
    #!/bin/sh
    
    ARGS=""
    
    # Metrics configuration
    {{- if .Values.operator.metrics.enabled }}
    ARGS="$ARGS --metrics-bind-address={{ .Values.operator.metrics.bindAddress }}"
    {{- end }}
    
    # ... 其他配置
    
    # Execute manager with all arguments
    exec /manager $ARGS
```

### 3. Deployment 配置

Deployment 挂载 ConfigMap 并执行启动脚本：

```yaml
containers:
  - name: manager
    image: "k8s-http-fake-operator:latest"
    ports:
      - name: http
        containerPort: 8080
    volumeMounts:
      - name: config
        mountPath: /config
        readOnly: true

volumes:
  - name: config
    configMap:
      name: k8s-http-fake-operator-config
```

## 使用方法

### 方法 1：使用默认配置

```bash
# 部署
helm install k8s-http-fake-operator ./charts/k8s-http-fake-operator

# 查看生成的 ConfigMap
kubectl get configmap k8s-http-fake-operator-config -o yaml
```

### 方法 2：自定义配置

创建 `custom-values.yaml`：

```yaml
operator:
  metrics:
    enabled: true
    bindAddress: ":9090"
  
  server:
    httpPort: 80
    httpsPort: 443
  
  leaderElection:
    enabled: true
```

部署：

```bash
helm install k8s-http-fake-operator ./charts/k8s-http-fake-operator -f custom-values.yaml
```

### 方法 3：生产环境配置

```yaml
operator:
  metrics:
    enabled: true
    bindAddress: ":8080"
    secure: true
  
  leaderElection:
    enabled: true
  
  server:
    httpPort: 80
    httpsPort: 443
    tlsCertFile: "/etc/tls/tls.crt"
    tlsKeyFile: "/etc/tls/tls.key"

tls:
  enabled: true
  certSecretName: production-tls

replicaCount: 3

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 200m
    memory: 256Mi
```

## 配置验证

### 1. 检查 ConfigMap

```bash
kubectl get configmap k8s-http-fake-operator-config -o yaml
```

查看 `start.sh` 内容，确认参数是否正确。

### 2. 检查 Pod 启动参数

```bash
kubectl exec -it <pod-name> -- cat /config/start.sh
```

### 3. 检查实际运行的命令

```bash
kubectl logs <pod-name> | grep "starting manager"
```

## 配置热更新

修改 `values.yaml` 后，重新部署即可：

```bash
helm upgrade k8s-http-fake-operator ./charts/k8s-http-fake-operator -f updated-values.yaml
```

ConfigMap 会自动更新，Pod 会自动重启以应用新配置。

## 常见配置场景

### 场景 1：开发环境

```yaml
operator:
  metrics:
    enabled: false
  
  leaderElection:
    enabled: false
  
  server:
    httpPort: 8080
    httpsPort: 8443
```

### 场景 2：测试环境

```yaml
operator:
  metrics:
    enabled: true
    bindAddress: ":8080"
  
  server:
    httpPort: 8080
    httpsPort: 8443

replicaCount: 1
```

### 场景 3：生产环境

```yaml
operator:
  metrics:
    enabled: true
    bindAddress: ":8080"
    secure: true
  
  leaderElection:
    enabled: true
  
  server:
    httpPort: 80
    httpsPort: 443

replicaCount: 3

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 200m
    memory: 256Mi
```

### 场景 4：启用 Webhook

```yaml
operator:
  webhook:
    certPath: "/etc/webhook-certs"
    certName: "tls.crt"
    keyName: "tls.key"
```

需要在 Deployment 中添加相应的 Secret 挂载。

## 故障排查

### 问题 1：ConfigMap 未生效

**症状**：Pod 使用默认参数启动

**解决**：
```bash
# 检查 ConfigMap 是否存在
kubectl get configmap k8s-http-fake-operator-config

# 检查 Pod 是否挂载 ConfigMap
kubectl describe pod <pod-name> | grep -A 5 "Mounts"

# 检查启动脚本
kubectl exec -it <pod-name> -- cat /config/start.sh
```

### 问题 2：参数不生效

**症状**：修改 values.yaml 后，参数未更新

**解决**：
```bash
# 重新部署
helm upgrade k8s-http-fake-operator ./charts/k8s-http-fake-operator -f values.yaml

# 强制重启 Pod
kubectl rollout restart deployment k8s-http-fake-operator
```

### 问题 3：脚本执行失败

**症状**：Pod 启动失败

**解决**：
```bash
# 查看 Pod 日志
kubectl logs <pod-name>

# 手动测试脚本
kubectl exec -it <pod-name> -- sh /config/start.sh
```

## 优势

1. **统一管理**：所有配置集中在 values.yaml
2. **版本控制**：配置变更可以通过 Git 追踪
3. **环境隔离**：不同环境使用不同的 values 文件
4. **热更新**：修改配置后自动重启 Pod
5. **易于维护**：无需手动管理复杂的命令行参数

## 迁移指南

如果你之前使用命令行参数，迁移步骤：

1. 将命令行参数转换为 values.yaml 配置
2. 部署新的 Helm Chart
3. 验证 ConfigMap 和 Pod 配置
4. 删除旧的 Deployment

示例：

**旧方式**：
```bash
./manager --http-port=8080 --https-port=8443 --leader-elect=true
```

**新方式**：
```yaml
operator:
  server:
    httpPort: 8080
    httpsPort: 8443
  leaderElection:
    enabled: true
```