# k8s-http-fake-operator

Kubernetes Operator for creating HTTP test stub services.

## 功能特性

- 创建和管理 HTTP 测试桩服务
- 支持静态响应和脚本响应
- 支持 URL 精确匹配、通配符匹配和正则表达式匹配
- 支持 TLS 配置
- 支持多响应规则和计数器
- 自动生成服务地址和状态管理

## 快速开始

### 安装 Operator

```bash
# 克隆仓库
git clone https://github.com/gfhw/k8s-http-fake-operator.git
cd k8s-http-fake-operator

# 安装 CRD
kubectl apply -f config/crd/bases/httpteststub.example.com_httpteststubs.yaml

# 部署 Operator
kubectl apply -f config/manager/manager.yaml
```

### 创建 HTTPTestStub 实例

```yaml
apiVersion: httpteststub.example.com/v1
kind: HTTPTestStub
metadata:
  name: example-stub
  namespace: default
spec:
  baseInfo:
    protocol: http
    port: 8080
  request:
    method: GET
    url:
      type: exact
      pattern: /api/test
  response:
    type: static
    static:
      status: 200
      headers:
        Content-Type: application/json
      body:
        message: "Hello, World!"
```

### 访问服务

```bash
# 获取服务地址
kubectl get httpteststub example-stub -o jsonpath='{.status.address}'

# 测试服务
curl http://<service-address>/api/test
```

## 配置参考

### HTTPTestStub 规范

| 字段 | 描述 | 示例值 |
|------|------|--------|
| `baseInfo.protocol` | 服务协议 | `http` 或 `https` |
| `baseInfo.port` | 服务端口 | `8080` |
| `request.method` | HTTP 方法 | `GET`, `POST`, `PUT`, `DELETE` |
| `request.url.type` | URL 匹配类型 | `exact` (精确), `pattern` (通配符), `regex` (正则) |
| `response.type` | 响应类型 | `static` (静态), `script` (脚本) |
| `response.static.status` | 静态响应状态码 | `200`, `404`, `500` |
| `response.static.body` | 响应体 | 可以是 JSON、字符串等 |
| `stubs` | 多个 stub 配置 | 用于定义复杂的请求-响应规则 |

## 示例

### 1. 基本的静态响应

```yaml
apiVersion: httpteststub.example.com/v1
kind: HTTPTestStub
metadata:
  name: static-response
  namespace: default
spec:
  baseInfo:
    protocol: http
    port: 8080
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

### 2. 带正则表达式匹配

```yaml
apiVersion: httpteststub.example.com/v1
kind: HTTPTestStub
metadata:
  name: regex-match
  namespace: default
spec:
  baseInfo:
    protocol: http
    port: 8080
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
```

## 开发

### 构建 Operator

```bash
make build
```

### 运行测试

```bash
make test
```

### 本地运行

```bash
make run
```

## 贡献

欢迎提交 Issue 和 Pull Request！

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

