package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
// @Description 统一API响应格式
type Response struct {
	Code    BusinessCode `json:"code" example:"0"`       // 业务状态码，0表示成功
	Message string       `json:"message" example:"操作成功"` // 响应消息
	Data    interface{}  `json:"data"`                   // 响应数据
}

// SuccessResponse 成功响应结构
// @Description 成功响应格式
type SuccessResponse struct {
	Code    BusinessCode `json:"code" example:"0"`       // 业务状态码，0表示成功
	Message string       `json:"message" example:"操作成功"` // 响应消息
	Data    interface{}  `json:"data"`                   // 响应数据
}

// ErrorResponse 错误响应结构
// @Description 错误响应格式
type ErrorResponse struct {
	Code    BusinessCode `json:"code" example:"100"`     // 业务错误状态码
	Message string       `json:"message" example:"参数错误"` // 错误消息
	Data    interface{}  `json:"data"`                   // 错误详情数据
}

// 已废弃：使用 code.go 中的 BusinessCode 常量
// 保留这些常量是为了向后兼容
const (
	OldCodeSuccess = 0   // 成功 - 已废弃，使用 CodeSuccess
	OldCodeError   = 400 // 错误 - 已废弃，使用对应的业务状态码
)

// SuccessResponseFunc 成功响应函数
// @Summary 成功响应
// @Description 返回成功响应
// @Tags 通用响应
// @Accept json
// @Produce json
// @Param message query string true "成功消息"
// @Param data body interface{} false "响应数据"
// @Success 200 {object} SuccessResponse
// @Router /success [get]
func SuccessResponseFunc(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, SuccessResponse{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// BusinessResponse 业务状态码响应函数
// @Summary 业务状态码响应
// @Description 根据业务状态码返回相应的HTTP响应
// @Tags 通用响应
// @Accept json
// @Produce json
// @Param code body BusinessCode true "业务状态码"
// @Param data body interface{} false "响应数据"
func BusinessResponse(c *gin.Context, code BusinessCode, data interface{}) {
	message := GetMessage(code)
	httpStatus := GetHTTPStatus(code)

	if IsSuccess(code) {
		c.JSON(httpStatus, SuccessResponse{
			Code:    code,
			Message: message,
			Data:    data,
		})
	} else {
		c.JSON(httpStatus, ErrorResponse{
			Code:    code,
			Message: message,
			Data:    data,
		})
	}
}

// BusinessResponseWithMessage 带自定义消息的业务状态码响应函数
// @Summary 带自定义消息的业务状态码响应
// @Description 根据业务状态码返回相应的HTTP响应，可自定义消息
// @Tags 通用响应
// @Accept json
// @Produce json
// @Param code body BusinessCode true "业务状态码"
// @Param message query string true "自定义消息"
// @Param data body interface{} false "响应数据"
func BusinessResponseWithMessage(c *gin.Context, code BusinessCode, message string, data interface{}) {
	if message == "" {
		message = GetMessage(code)
	}
	httpStatus := GetHTTPStatus(code)

	if IsSuccess(code) {
		c.JSON(httpStatus, SuccessResponse{
			Code:    code,
			Message: message,
			Data:    data,
		})
	} else {
		c.JSON(httpStatus, ErrorResponse{
			Code:    code,
			Message: message,
			Data:    data,
		})
	}
}

// ErrorResponseFunc 错误响应函数 (保留向后兼容)
// @Summary 错误响应
// @Description 返回错误响应
// @Tags 通用响应
// @Accept json
// @Produce json
// @Param httpCode path int true "HTTP状态码"
// @Param message query string true "错误消息"
// @Param data body interface{} false "错误详情"
// @Success 400 {object} ErrorResponse
// @Success 401 {object} ErrorResponse
// @Success 403 {object} ErrorResponse
// @Success 404 {object} ErrorResponse
// @Success 500 {object} ErrorResponse
// @Router /error [get]
// Deprecated: 使用 BusinessResponse 或 BusinessResponseWithMessage 替代
func ErrorResponseFunc(c *gin.Context, httpCode int, message string, data interface{}) {
	// 根据 HTTP 状态码映射到对应的业务状态码
	var businessCode BusinessCode
	switch httpCode {
	case http.StatusBadRequest:
		businessCode = CodeBadRequest
	case http.StatusUnauthorized:
		businessCode = CodeUnauthorized
	case http.StatusForbidden:
		businessCode = CodeForbidden
	case http.StatusNotFound:
		businessCode = CodeNotFound
	case http.StatusConflict:
		businessCode = CodeConflict
	case http.StatusInternalServerError:
		businessCode = CodeInternalError
	default:
		businessCode = CodeInternalError
	}

	c.JSON(httpCode, ErrorResponse{
		Code:    businessCode,
		Message: message,
		Data:    data,
	})
}

// 便捷响应函数

// BadRequest 400错误响应 (保留向后兼容)
// Deprecated: 使用 BusinessResponse(c, CodeBadRequest, data) 或 BusinessResponseWithMessage 替代
func BadRequest(c *gin.Context, message string, data interface{}) {
	if message == "" {
		message = GetMessage(CodeBadRequest)
	}
	BusinessResponseWithMessage(c, CodeBadRequest, message, data)
}

// BadRequestWithCode 带业务状态码的错误请求响应
func BadRequestWithCode(c *gin.Context, code BusinessCode, data interface{}) {
	BusinessResponse(c, code, data)
}

// Unauthorized 401错误响应 (保留向后兼容)
// Deprecated: 使用 BusinessResponse(c, CodeUnauthorized, data) 或 BusinessResponseWithMessage 替代
func Unauthorized(c *gin.Context, message string, data interface{}) {
	if message == "" {
		message = GetMessage(CodeUnauthorized)
	}
	BusinessResponseWithMessage(c, CodeUnauthorized, message, data)
}

// UnauthorizedWithCode 带业务状态码的未授权响应
func UnauthorizedWithCode(c *gin.Context, code BusinessCode, data interface{}) {
	BusinessResponse(c, code, data)
}

// Forbidden 403错误响应 (保留向后兼容)
// Deprecated: 使用 BusinessResponse(c, CodeForbidden, data) 或 BusinessResponseWithMessage 替代
func Forbidden(c *gin.Context, message string, data interface{}) {
	if message == "" {
		message = GetMessage(CodeForbidden)
	}
	BusinessResponseWithMessage(c, CodeForbidden, message, data)
}

// ForbiddenWithCode 带业务状态码的禁止访问响应
func ForbiddenWithCode(c *gin.Context, code BusinessCode, data interface{}) {
	BusinessResponse(c, code, data)
}

// NotFound 404错误响应 (保留向后兼容)
// Deprecated: 使用 BusinessResponse(c, CodeNotFound, data) 或 BusinessResponseWithMessage 替代
func NotFound(c *gin.Context, message string, data interface{}) {
	if message == "" {
		message = GetMessage(CodeNotFound)
	}
	BusinessResponseWithMessage(c, CodeNotFound, message, data)
}

// NotFoundWithCode 带业务状态码的资源不存在响应
func NotFoundWithCode(c *gin.Context, code BusinessCode, data interface{}) {
	BusinessResponse(c, code, data)
}

// Conflict 409错误响应 (保留向后兼容)
// Deprecated: 使用 BusinessResponse(c, CodeConflict, data) 或 BusinessResponseWithMessage 替代
func Conflict(c *gin.Context, message string, data interface{}) {
	if message == "" {
		message = GetMessage(CodeConflict)
	}
	BusinessResponseWithMessage(c, CodeConflict, message, data)
}

// ConflictWithCode 带业务状态码的资源冲突响应
func ConflictWithCode(c *gin.Context, code BusinessCode, data interface{}) {
	BusinessResponse(c, code, data)
}

// InternalError 500错误响应 (保留向后兼容)
// Deprecated: 使用 BusinessResponse(c, CodeInternalError, data) 或 BusinessResponseWithMessage 替代
func InternalError(c *gin.Context, message string, data interface{}) {
	if message == "" {
		message = GetMessage(CodeInternalError)
	}
	BusinessResponseWithMessage(c, CodeInternalError, message, data)
}

// InternalErrorWithCode 带业务状态码的内部错误响应
func InternalErrorWithCode(c *gin.Context, code BusinessCode, data interface{}) {
	BusinessResponse(c, code, data)
}

// 注意：推荐使用 BusinessResponse 和 BusinessResponseWithMessage 作为主要函数

// GetResponse 获取响应结构 (保留向后兼容)
// Deprecated: 使用 GetBusinessResponse 替代
func GetResponse(code int, message string, data interface{}) *Response {
	return &Response{
		Code:    BusinessCode(code),
		Message: message,
		Data:    data,
	}
}

// GetBusinessResponse 获取业务响应结构
func GetBusinessResponse(code BusinessCode, data interface{}) *Response {
	return &Response{
		Code:    code,
		Message: GetMessage(code),
		Data:    data,
	}
}

// GetBusinessResponseWithMessage 获取带自定义消息的业务响应结构
func GetBusinessResponseWithMessage(code BusinessCode, message string, data interface{}) *Response {
	if message == "" {
		message = GetMessage(code)
	}
	return &Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// 客服系统专用的便捷响应函数

// AuthResponse 认证相关响应
func AuthResponse(c *gin.Context, code BusinessCode, data interface{}) {
	BusinessResponse(c, code, data)
}

// SessionResponse 会话相关响应
func SessionResponse(c *gin.Context, code BusinessCode, data interface{}) {
	BusinessResponse(c, code, data)
}

// MessageResponse 消息相关响应
func MessageResponse(c *gin.Context, code BusinessCode, data interface{}) {
	BusinessResponse(c, code, data)
}

// AgentResponse 客服相关响应
func AgentResponse(c *gin.Context, code BusinessCode, data interface{}) {
	BusinessResponse(c, code, data)
}

// WSResponse WebSocket相关响应
func WSResponse(c *gin.Context, code BusinessCode, data interface{}) {
	BusinessResponse(c, code, data)
}

// Success 成功响应的快捷方法
func Success(c *gin.Context, data interface{}) {
	BusinessResponse(c, CodeSuccess, data)
}

// SuccessWithMessage 带消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	BusinessResponseWithMessage(c, CodeSuccess, message, data)
}
