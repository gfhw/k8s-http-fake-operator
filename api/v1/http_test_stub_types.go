package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HTTPTestStubSpec defines the desired state of HTTPTestStub
type HTTPTestStubSpec struct {
	Protocol string    `json:"protocol"`
	Request  Request   `json:"request"`
	Response Response  `json:"response"`
	Scripts  []Script  `json:"scripts,omitempty"`
	TLS      TLSConfig `json:"tls,omitempty"`
	Stubs    []Stub    `json:"stubs,omitempty"`
}

// Request 请求匹配规则
type Request struct {
	Method string `json:"method"`
	URL    URL    `json:"url"`
}

// URL URL匹配规则
type URL struct {
	Type    string `json:"type"`
	Pattern string `json:"pattern,omitempty"`
	Regex   string `json:"regex,omitempty"`
}

// Response 响应配置
type Response struct {
	Type   string  `json:"type"`
	Static *Static `json:"static,omitempty"`
	Script *Script `json:"script,omitempty"`
}

// Static 静态响应配置
type Static struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// Script 脚本配置
type Script struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`              // shell, python, etc.
	Path    string   `json:"path"`              // 脚本文件路径
	Timeout int      `json:"timeout"`           // 超时时间（秒）
	Input   []string `json:"input"`             // 脚本输入
	Content string   `json:"content,omitempty"` // 内联脚本内容
}

// TLSConfig TLS配置
type TLSConfig struct {
	Enabled    bool   `json:"enabled"`
	SecretName string `json:"secretName"`
}

// Stub 单个stub配置
type Stub struct {
	Request       Request        `json:"request"`
	ResponseRules []ResponseRule `json:"responseRules"`
	Counter       Counter        `json:"counter"`
}

// ResponseRule 响应规则
type ResponseRule struct {
	Rule     Rule    `json:"rule"`
	Response *Static `json:"response,omitempty"`
}

// Rule 规则条件
type Rule struct {
	Type  string `json:"type"`
	Start int    `json:"start,omitempty"`
	End   int    `json:"end,omitempty"`
}

// Counter 计数器配置
type Counter struct {
	Reset      bool `json:"reset"`
	ResetAfter int  `json:"resetAfter"`
}

// HTTPTestStubStatus defines the observed state of HTTPTestStub
type HTTPTestStubStatus struct {
	Status       string `json:"status"`
	RequestCount int    `json:"requestCount"`
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
