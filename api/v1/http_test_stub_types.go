package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HTTPTestStubSpec defines the desired state of HTTPTestStub
type HTTPTestStubSpec struct {
	// 基础配置
	BaseInfo BaseInfo `json:"baseInfo"`
	// 请求匹配规则
	Request Request `json:"request"`
	// 响应配置
	Response Response `json:"response"`
	// 脚本配置
	Scripts []Script `json:"scripts,omitempty"`
	// TLS配置
	TLS TLSConfig `json:"tls,omitempty"`
	// 多个stub配置
	Stubs []Stub `json:"stubs,omitempty"`
}

// BaseInfo 基础配置
type BaseInfo struct {
	Protocol string `json:"protocol"` // http or https
	Port     int32  `json:"port"`     // 服务端口
}

// Request 请求匹配规则
type Request struct {
	Method string `json:"method"` // HTTP方法
	URL    URL    `json:"url"`    // URL匹配规则
}

// URL URL匹配规则
type URL struct {
	Type   string `json:"type"`   // 匹配类型: exact, pattern, regex
	Pattern string `json:"pattern,omitempty"` // 通配符模式
	Regex   string `json:"regex,omitempty"`   // 正则表达式
}

// Response 响应配置
type Response struct {
	Type   string  `json:"type"`   // 响应类型: static, script
	Static *Static `json:"static,omitempty"` // 静态响应
	Script *Script `json:"script,omitempty"` // 脚本响应
}

// Static 静态响应配置
type Static struct {
	Status  int               `json:"status"`  // 状态码
	Headers map[string]string `json:"headers"` // 响应头
	Body    interface{}       `json:"body"`    // 响应体
}

// Script 脚本配置
type Script struct {
	Name    string   `json:"name"`    // 脚本名称
	Type    string   `json:"type"`    // 脚本类型: shell, python, javascript
	Path    string   `json:"path"`    // 脚本路径
	Timeout int      `json:"timeout"` // 脚本执行超时时间（秒）
	Input   []string `json:"input"`   // 脚本输入参数
	Content string   `json:"content,omitempty"` // 脚本内容
}

// TLSConfig TLS配置
type TLSConfig struct {
	Enabled    bool   `json:"enabled"`    // 是否启用TLS
	SecretName string `json:"secretName"` // TLS密钥Secret名称
}

// Stub 单个stub配置
type Stub struct {
	Request       Request         `json:"request"`       // 请求匹配规则
	ResponseRules []ResponseRule  `json:"responseRules"` // 响应规则
	Counter       Counter         `json:"counter"`       // 计数器配置
}

// ResponseRule 响应规则
type ResponseRule struct {
	Rule     Rule   `json:"rule"`     // 规则条件
	Response Static `json:"response"` // 响应配置
}

// Rule 规则条件
type Rule struct {
	Type  string `json:"type"`  // 规则类型: range, default
	Start int    `json:"start,omitempty"` // 范围开始
	End   int    `json:"end,omitempty"`   // 范围结束
}

// Counter 计数器配置
type Counter struct {
	Reset       bool `json:"reset"`       // 是否重置
	ResetAfter  int  `json:"resetAfter"`  // 重置阈值
}

// HTTPTestStubStatus defines the observed state of HTTPTestStub
type HTTPTestStubStatus struct {
	// 服务状态
	Status string `json:"status"`
	// 服务地址
	Address string `json:"address"`
	// 已处理请求数
	RequestCount int `json:"requestCount"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// HTTPTestStub is the Schema for the httpteststubs API
type HTTPTestStub struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HTTPTestStubSpec   `json:"spec,omitempty"`
	Status HTTPTestStubStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HTTPTestStubList contains a list of HTTPTestStub
type HTTPTestStubList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HTTPTestStub `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HTTPTestStub{}, &HTTPTestStubList{})
}
