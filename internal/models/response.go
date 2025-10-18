package models

// Response 是标准 API 响应格式。
// Data 和 Meta 字段可选，根据具体接口需求决定是否包含。
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *PageMeta   `json:"meta,omitempty"`
}

// PageMeta 包含分页相关的信息。
type PageMeta struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}
