package common

// BusinessCode 业务状态码类型
type BusinessCode int

// 通用状态码 (0-999)
const (
	// 成功状态码
	CodeSuccess BusinessCode = 0 // 操作成功

	// 通用错误状态码 (100-199)
	CodeBadRequest    BusinessCode = 100 // 请求参数错误
	CodeUnauthorized  BusinessCode = 101 // 未授权访问
	CodeForbidden     BusinessCode = 102 // 禁止访问
	CodeNotFound      BusinessCode = 103 // 资源不存在
	CodeConflict      BusinessCode = 104 // 资源冲突
	CodeInternalError BusinessCode = 105 // 服务器内部错误
	CodeServiceBusy   BusinessCode = 106 // 服务繁忙
	CodeRateLimited   BusinessCode = 107 // 请求频率限制
	CodeTimeout       BusinessCode = 108 // 请求超时
	CodeUnsupported   BusinessCode = 109 // 不支持的操作
)

// 认证相关状态码 (1000-1099)
const (
	CodeAuthInvalidCredentials BusinessCode = 1000 // 账号或密码错误
	CodeAuthAccountNotFound    BusinessCode = 1001 // 账号不存在
	CodeAuthAccountExists      BusinessCode = 1002 // 账号已存在
	CodeAuthPasswordWeak       BusinessCode = 1003 // 密码强度不够
	CodeAuthTokenInvalid       BusinessCode = 1004 // Token无效
	CodeAuthTokenExpired       BusinessCode = 1005 // Token已过期
	CodeAuthPermissionDenied   BusinessCode = 1006 // 权限不足
	CodeAuthLoginRequired      BusinessCode = 1007 // 需要登录
	CodeAuthLogoutFailed       BusinessCode = 1008 // 登出失败
	CodeAuthSessionExpired     BusinessCode = 1009 // 会话已过期
)

// 会话管理状态码 (2000-2099)
const (
	CodeSessionCreateFailed    BusinessCode = 2000 // 会话创建失败
	CodeSessionNotFound        BusinessCode = 2001 // 会话不存在
	CodeSessionClosed          BusinessCode = 2002 // 会话已关闭
	CodeSessionInactive        BusinessCode = 2003 // 会话不活跃
	CodeSessionInTransfer      BusinessCode = 2004 // 会话转移中
	CodeSessionReconnectFailed BusinessCode = 2005 // 会话重连失败
	CodeSessionCloseFailed     BusinessCode = 2006 // 会话关闭失败
	CodeSessionTransferFailed  BusinessCode = 2007 // 会话转移失败
	CodeSessionInvalidState    BusinessCode = 2008 // 会话状态无效
	CodeSessionLimit           BusinessCode = 2009 // 会话数量限制
)

// 消息处理状态码 (3000-3099)
const (
	CodeMessageSendFailed    BusinessCode = 3000 // 消息发送失败
	CodeMessageNotFound      BusinessCode = 3001 // 消息不存在
	CodeMessageInvalidType   BusinessCode = 3002 // 消息类型无效
	CodeMessageTooLarge      BusinessCode = 3003 // 消息内容过大
	CodeMessageEmpty         BusinessCode = 3004 // 消息内容为空
	CodeMessageReadFailed    BusinessCode = 3005 // 消息标记已读失败
	CodeMessageDuplicate     BusinessCode = 3006 // 重复消息
	CodeMessageRateLimited   BusinessCode = 3007 // 消息发送频率限制
	CodeMessageFilterBlocked BusinessCode = 3008 // 消息被过滤拦截
	CodeMessageHistoryFailed BusinessCode = 3009 // 获取消息历史失败
)

// 客服管理状态码 (4000-4099)
const (
	CodeAgentNotFound         BusinessCode = 4000 // 客服不存在
	CodeAgentOffline          BusinessCode = 4001 // 客服离线
	CodeAgentBusy             BusinessCode = 4002 // 客服繁忙
	CodeAgentStatusInvalid    BusinessCode = 4003 // 客服状态无效
	CodeAgentUpdateFailed     BusinessCode = 4004 // 客服状态更新失败
	CodeAgentAssignFailed     BusinessCode = 4005 // 客服分配失败
	CodeAgentNoAvailable      BusinessCode = 4006 // 无可用客服
	CodeAgentSessionLimit     BusinessCode = 4007 // 客服会话数量限制
	CodeAgentPermissionDenied BusinessCode = 4008 // 客服权限不足
	CodeAgentAlreadyOnline    BusinessCode = 4009 // 客服已在线
)

// WebSocket连接状态码 (5000-5099)
const (
	CodeWSConnectFailed    BusinessCode = 5000 // WebSocket连接失败
	CodeWSInvalidParams    BusinessCode = 5001 // WebSocket参数无效
	CodeWSSessionRequired  BusinessCode = 5002 // 需要会话ID
	CodeWSUserRequired     BusinessCode = 5003 // 需要用户ID
	CodeWSTypeInvalid      BusinessCode = 5004 // 用户类型无效
	CodeWSAlreadyConnected BusinessCode = 5005 // 已经连接
	CodeWSDisconnected     BusinessCode = 5006 // 连接已断开
	CodeWSMessageInvalid   BusinessCode = 5007 // WebSocket消息格式无效
	CodeWSHeartbeatTimeout BusinessCode = 5008 // 心跳超时
	CodeWSConnectionLimit  BusinessCode = 5009 // 连接数量限制
)

// 业务规则状态码 (6000-6099)
const (
	CodeValidationFailed     BusinessCode = 6000 // 数据验证失败
	CodeParamRequired        BusinessCode = 6001 // 必需参数缺失
	CodeParamInvalid         BusinessCode = 6002 // 参数格式无效
	CodeParamOutOfRange      BusinessCode = 6003 // 参数超出范围
	CodeDataFormatError      BusinessCode = 6004 // 数据格式错误
	CodeFileTypeNotSupported BusinessCode = 6005 // 文件类型不支持
	CodeFileSizeExceeded     BusinessCode = 6006 // 文件大小超限
	CodeOperationNotAllowed  BusinessCode = 6007 // 操作不被允许
	CodeResourceInUse        BusinessCode = 6008 // 资源正在使用中
	CodeQuotaExceeded        BusinessCode = 6009 // 配额超限
)

// 外部服务状态码 (7000-7099)
const (
	CodeDatabaseError    BusinessCode = 7000 // 数据库错误
	CodeRedisError       BusinessCode = 7001 // Redis错误
	CodeCacheError       BusinessCode = 7002 // 缓存错误
	CodeNetworkError     BusinessCode = 7003 // 网络错误
	CodeThirdPartyError  BusinessCode = 7004 // 第三方服务错误
	CodeConfigError      BusinessCode = 7005 // 配置错误
	CodeStorageError     BusinessCode = 7006 // 存储错误
	CodeQueueError       BusinessCode = 7007 // 队列错误
	CodeLockError        BusinessCode = 7008 // 锁错误
	CodeTransactionError BusinessCode = 7009 // 事务错误
)

// CodeMessage 状态码对应的消息
var CodeMessage = map[BusinessCode]string{
	// 通用状态码
	CodeSuccess:       "操作成功",
	CodeBadRequest:    "请求参数错误",
	CodeUnauthorized:  "未授权访问",
	CodeForbidden:     "禁止访问",
	CodeNotFound:      "资源不存在",
	CodeConflict:      "资源冲突",
	CodeInternalError: "服务器内部错误",
	CodeServiceBusy:   "服务繁忙，请稍后重试",
	CodeRateLimited:   "请求频率过高，请稍后重试",
	CodeTimeout:       "请求超时",
	CodeUnsupported:   "不支持的操作",

	// 认证相关
	CodeAuthInvalidCredentials: "账号或密码错误",
	CodeAuthAccountNotFound:    "账号不存在",
	CodeAuthAccountExists:      "账号已存在",
	CodeAuthPasswordWeak:       "密码强度不够",
	CodeAuthTokenInvalid:       "Token无效",
	CodeAuthTokenExpired:       "Token已过期，请重新登录",
	CodeAuthPermissionDenied:   "权限不足",
	CodeAuthLoginRequired:      "请先登录",
	CodeAuthLogoutFailed:       "登出失败",
	CodeAuthSessionExpired:     "会话已过期，请重新登录",

	// 会话管理
	CodeSessionCreateFailed:    "会话创建失败",
	CodeSessionNotFound:        "会话不存在",
	CodeSessionClosed:          "会话已关闭",
	CodeSessionInactive:        "会话不活跃",
	CodeSessionInTransfer:      "会话转移中",
	CodeSessionReconnectFailed: "会话重连失败",
	CodeSessionCloseFailed:     "会话关闭失败",
	CodeSessionTransferFailed:  "会话转移失败",
	CodeSessionInvalidState:    "会话状态无效",
	CodeSessionLimit:           "会话数量已达上限",

	// 消息处理
	CodeMessageSendFailed:    "消息发送失败",
	CodeMessageNotFound:      "消息不存在",
	CodeMessageInvalidType:   "消息类型无效",
	CodeMessageTooLarge:      "消息内容过大",
	CodeMessageEmpty:         "消息内容不能为空",
	CodeMessageReadFailed:    "消息标记已读失败",
	CodeMessageDuplicate:     "重复消息",
	CodeMessageRateLimited:   "消息发送频率过高",
	CodeMessageFilterBlocked: "消息被安全过滤拦截",
	CodeMessageHistoryFailed: "获取消息历史失败",

	// 客服管理
	CodeAgentNotFound:         "客服不存在",
	CodeAgentOffline:          "客服离线",
	CodeAgentBusy:             "客服繁忙",
	CodeAgentStatusInvalid:    "客服状态无效",
	CodeAgentUpdateFailed:     "客服状态更新失败",
	CodeAgentAssignFailed:     "客服分配失败",
	CodeAgentNoAvailable:      "暂无可用客服，请稍后重试",
	CodeAgentSessionLimit:     "客服会话数量已达上限",
	CodeAgentPermissionDenied: "客服权限不足",
	CodeAgentAlreadyOnline:    "客服已在线",

	// WebSocket连接
	CodeWSConnectFailed:    "WebSocket连接失败",
	CodeWSInvalidParams:    "WebSocket连接参数无效",
	CodeWSSessionRequired:  "缺少会话ID",
	CodeWSUserRequired:     "缺少用户ID",
	CodeWSTypeInvalid:      "用户类型无效",
	CodeWSAlreadyConnected: "WebSocket已连接",
	CodeWSDisconnected:     "WebSocket连接已断开",
	CodeWSMessageInvalid:   "WebSocket消息格式无效",
	CodeWSHeartbeatTimeout: "心跳超时",
	CodeWSConnectionLimit:  "连接数量已达上限",

	// 业务规则
	CodeValidationFailed:     "数据验证失败",
	CodeParamRequired:        "必需参数缺失",
	CodeParamInvalid:         "参数格式无效",
	CodeParamOutOfRange:      "参数超出有效范围",
	CodeDataFormatError:      "数据格式错误",
	CodeFileTypeNotSupported: "文件类型不支持",
	CodeFileSizeExceeded:     "文件大小超出限制",
	CodeOperationNotAllowed:  "操作不被允许",
	CodeResourceInUse:        "资源正在使用中",
	CodeQuotaExceeded:        "配额已超限",

	// 外部服务
	CodeDatabaseError:    "数据库操作失败",
	CodeRedisError:       "缓存服务异常",
	CodeCacheError:       "缓存操作失败",
	CodeNetworkError:     "网络连接异常",
	CodeThirdPartyError:  "第三方服务异常",
	CodeConfigError:      "配置错误",
	CodeStorageError:     "存储服务异常",
	CodeQueueError:       "消息队列异常",
	CodeLockError:        "分布式锁异常",
	CodeTransactionError: "事务操作失败",
}

// GetMessage 获取状态码对应的消息
func GetMessage(code BusinessCode) string {
	if msg, ok := CodeMessage[code]; ok {
		return msg
	}
	return "未知错误"
}

// IsSuccess 判断是否为成功状态码
func IsSuccess(code BusinessCode) bool {
	return code == CodeSuccess
}

// IsClientError 判断是否为客户端错误 (1000-5999)
func IsClientError(code BusinessCode) bool {
	return code >= 100 && code < 7000
}

// IsServerError 判断是否为服务端错误 (7000+)
func IsServerError(code BusinessCode) bool {
	return code >= 7000
}

// GetHTTPStatus 根据业务状态码获取对应的HTTP状态码
func GetHTTPStatus(code BusinessCode) int {
	switch {
	case code == CodeSuccess:
		return 200
	case code == CodeInternalError:
		return 500
	case code >= CodeBadRequest && code <= CodeUnsupported:
		return 400
	case code == CodeConflict ||
		code == CodeAuthAccountExists ||
		code == CodeSessionInTransfer:
		return 409
	case code >= CodeAuthInvalidCredentials && code <= CodeAuthSessionExpired && code != CodeAuthAccountExists:
		return 401
	case code == CodeNotFound ||
		code == CodeSessionNotFound ||
		code == CodeMessageNotFound ||
		code == CodeAgentNotFound:
		return 404
	case code == CodeForbidden ||
		code == CodeAuthPermissionDenied ||
		(code >= CodeAgentBusy && code <= CodeAgentPermissionDenied && code != CodeAgentNotFound):
		return 403
	case code >= CodeValidationFailed && code <= CodeQuotaExceeded:
		return 400
	case code >= CodeServiceBusy && code <= CodeAgentNoAvailable:
		return 503
	case code >= CodeDatabaseError:
		return 500
	default:
		return 500
	}
}
