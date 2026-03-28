# 容器持续运行和镜像构建改进

## 概述

为了确保 k8s-http-fake-operator 容器能够持续运行，并且提供便捷的镜像构建方式，我们进行了以下改进：

## 1. 容器持续运行机制

### 问题
容器中的进程如果崩溃，容器会退出，导致服务中断。

### 解决方案

#### 1.1 无限循环启动脚本
在 `charts/k8s-http-fake-operator/templates/configmap.yaml` 中添加了无限循环：

```bash
# Infinite loop to keep the container running and restart the manager if it crashes
while true; do
    echo "Starting manager process..."
    /manager $ARGS
    
    # If manager exits, wait a bit before restarting
    echo "Manager process exited with code $?"
    echo "Waiting 5 seconds before restarting..."
    sleep 5
    
    echo "Restarting manager..."
done
```

#### 1.2 智能入口点脚本
创建了 `scripts/entrypoint.sh` 作为容器入口点：

- **优先使用 ConfigMap**: 如果 `/config/start.sh` 存在，使用 ConfigMap 中的配置
- **默认配置回退**: 如果没有 ConfigMap，使用默认配置启动
- **自动重启**: manager 崩溃后自动重启，间隔 5 秒
- **详细日志**: 记录启动和重启过程，便于调试

#### 1.3 更新 Dockerfile
```dockerfile
# Copy the entrypoint script
COPY scripts/entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

# Run the entrypoint script which will keep the container running
CMD ["/entrypoint.sh"]
```

## 2. 镜像构建系统

### 创建的文件

#### 2.1 构建脚本
- `build/build-image.sh`: Linux/macOS Bash 构建脚本
- `build/build-image.ps1`: Windows PowerShell 构建脚本

#### 2.2 文档
- `build/README.md`: 构建脚本使用说明
- `scripts/README.md`: 脚本功能说明

#### 2.3 Makefile
更新了根目录的 `Makefile`，提供便捷的构建命令。

## 3. 使用方法

### 3.1 构建镜像

#### Windows PowerShell
```powershell
# 基本构建
.\build\build-image.ps1

# 自定义名称和标签
.\build\build-image.ps1 -ImageName my-operator -ImageTag v1.0.0

# 不使用缓存
.\build\build-image.ps1 -NoCache

# 查看帮助
.\build\build-image.ps1 -Help
```

#### Linux/macOS Bash
```bash
# 基本构建
./build/build-image.sh

# 自定义名称和标签
./build/build-image.sh --name my-operator --tag v1.0.0

# 不使用缓存
./build/build-image.sh --no-cache

# 查看帮助
./build/build-image.sh --help
```

#### Makefile
```bash
# 查看所有命令
make help

# 构建镜像
make build

# 不使用缓存构建
make build-no-cache

# 清理镜像
make clean

# 测试运行
make test

# 进入容器
make shell

# 查看日志
make logs

# 清理并重新构建
make rebuild
```

### 3.2 环境变量配置

```bash
export IMAGE_NAME=my-operator
export IMAGE_TAG=v1.0.0
export NO_CACHE=true
./build/build-image.sh
```

### 3.3 自定义配置

#### 使用 ConfigMap
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: custom-config
data:
  start.sh: |
    #!/bin/sh
    ARGS="--http-port=9090 --https-port=9443"
    while true; do
        /manager $ARGS
        sleep 5
    done
```

#### 在 values.yaml 中配置
```yaml
operator:
  server:
    httpPort: 9090
    httpsPort: 9443
```

## 4. 容器生命周期管理

### 4.1 启动流程
1. 容器启动，执行 `/entrypoint.sh`
2. 检查 `/config/start.sh` 是否存在
3. 如果存在，使用 ConfigMap 配置
4. 如果不存在，使用默认配置
5. 启动 manager 进程
6. 监控进程状态
7. 如果崩溃，等待 5 秒后重启

### 4.2 健康检查
容器配置了健康检查端点：
- `/healthz`: 存活探针
- `/readyz`: 就绪探针

### 4.3 日志管理
所有启动和重启事件都会记录到日志中：
```
Starting k8s-http-fake-operator...
Starting manager process...
Manager process exited with code 1
Waiting 5 seconds before restarting...
Restarting manager...
```

## 5. 部署流程

### 5.1 本地开发
```bash
# 1. 构建镜像
make build

# 2. 测试运行
make test

# 3. 推送到仓库
docker push my-registry/k8s-http-fake-operator:latest
```

### 5.2 生产部署
```bash
# 1. 更新 values.yaml
image:
  repository: my-registry/k8s-http-fake-operator
  tag: v1.0.0

# 2. 部署
helm install k8s-http-fake-operator ./charts/k8s-http-fake-operator

# 3. 验证
kubectl get pods
kubectl logs -f deployment/k8s-http-fake-operator
```

## 6. 故障排除

### 6.1 容器不断重启
这是正常行为，说明 manager 崩溃了。检查日志：
```bash
kubectl logs -f deployment/k8s-http-fake-operator
```

### 6.2 ConfigMap 不生效
验证 ConfigMap 挂载：
```bash
kubectl exec -it <pod-name> -- cat /config/start.sh
```

### 6.3 构建失败
检查：
- Docker 是否运行
- Dockerfile 是否存在
- 网络连接是否正常
- 磁盘空间是否充足

### 6.4 权限问题
确保脚本有执行权限：
```bash
chmod +x build/build-image.sh
chmod +x scripts/entrypoint.sh
```

## 7. 优势

### 7.1 可靠性
- 自动重启机制确保服务持续可用
- 详细的日志便于问题诊断
- 默认配置回退保证基本功能

### 7.2 易用性
- 一键构建脚本
- 跨平台支持（Windows/Linux/macOS）
- Makefile 提供便捷命令

### 7.3 灵活性
- 支持自定义配置
- 环境变量配置
- ConfigMap 动态配置

### 7.4 可维护性
- 清晰的文档
- 标准化的构建流程
- 易于扩展和修改

## 8. 总结

通过这些改进，我们实现了：

1. ✅ **容器持续运行**: 无限循环 + 自动重启机制
2. ✅ **便捷构建**: 跨平台构建脚本 + Makefile
3. ✅ **灵活配置**: ConfigMap + 环境变量
4. ✅ **完善文档**: 详细的使用说明和故障排除指南

现在你可以轻松构建镜像并部署到 Kubernetes 集群，容器会自动保持运行状态！