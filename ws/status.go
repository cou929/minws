package ws

// StatusCodes.
// See: https://tools.ietf.org/html/rfc6455#section-7.4
const (
	StatusNormalClosure           = 1000
	StatusGoingAway               = 1001
	StatusProtocolError           = 1002
	StatusNotAcceptable           = 1003
	StatusNoStatusCode            = 1005
	StatusAbnormalClosure         = 1006
	StatusInconsistentMessageType = 1007
	StatusPolicyViolation         = 1008
	StatusTooLarge                = 1009
	StatusNoExtension             = 1010 // client only
	StatusUnexpectedCondition     = 1011
	StatusTLSHandShakeFailure     = 1015
)

var statusText = map[int]string{
	StatusNormalClosure:           "Normal Closure",
	StatusGoingAway:               "Going Away",
	StatusProtocolError:           "Protocol Error",
	StatusNotAcceptable:           "Not Acceptable",
	StatusNoStatusCode:            "No Status Code",
	StatusAbnormalClosure:         "Abnormal Closure",
	StatusInconsistentMessageType: "Inconsistent Message Type",
	StatusPolicyViolation:         "Policy Violation",
	StatusTooLarge:                "Too Large",
	StatusNoExtension:             "No Extension",
	StatusUnexpectedCondition:     "Unexpected Condition",
	StatusTLSHandShakeFailure:     "TLS HandShake Failure",
}

// StatusText returns a text for the WebSocket status code
func StatusText(code int) string {
	return statusText[code]
}
