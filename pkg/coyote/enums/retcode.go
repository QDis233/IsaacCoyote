package enums

type RetCode int

const (
	RetCodeSuccess              RetCode = 200 // 成功
	RetCodeDisconnect           RetCode = 209 // 对方客户端已断开
	RetCodeQRCodeNoClientID     RetCode = 210 // 二维码中没有有效的 clientID
	RetCodeWaitAppIDTimeout     RetCode = 211 // socket 连接上了，但服务器迟迟不下发 app 端的 id 来绑定
	RetCodeClientIDAlreadyUsed  RetCode = 400 // 此 id 已被其他客户端绑定关系
	RetCodeTargetClientNotFound RetCode = 401 // 要绑定的目标客户端不存在
	RetCodeNotBound             RetCode = 402 // 收信方和寄信方不是绑定关系
	RetCodeInvalidMessageFormat RetCode = 403 // 发送的内容不是标准 json 对象
	RetCodeReceiverOffline      RetCode = 404 // 未找到收信人（离线）
	RetCodeMessageTooLong       RetCode = 405 // 下发的 message 长度大于 1950
	RetCodeInternalError        RetCode = 500 // 服务器内部异常
)

func (r RetCode) String() string {
	switch r {
	case RetCodeSuccess:
		return "200"
	case RetCodeDisconnect:
		return "209"
	case RetCodeQRCodeNoClientID:
		return "210"
	case RetCodeWaitAppIDTimeout:
		return "211"
	case RetCodeClientIDAlreadyUsed:
		return "400"
	case RetCodeTargetClientNotFound:
		return "401"
	case RetCodeNotBound:
		return "402"
	case RetCodeInvalidMessageFormat:
		return "403"
	case RetCodeReceiverOffline:
		return "404"
	case RetCodeMessageTooLong:
		return "405"
	case RetCodeInternalError:
		return "500"
	default:
		return ""
	}
}
