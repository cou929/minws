package ws

// StatusCodes.
// See: https://tools.ietf.org/html/rfc6455#section-7.4
const (
	StatusNormalClosure           = 1000
	StatusGoingAway               = 1001
	StatusProtocolError           = 1002
	StatusUnsupportedData         = 1003
	StatusNoStatusReceived        = 1005
	StatusAbnormalClosure         = 1006
	StatusInvalidFramePayloadData = 1007
	StatusPolicyViolation         = 1008
	StatusMessageTooBig           = 1009
	StatusMissingExtension        = 1010 // client only
	StatusInternalError           = 1011
	StatusServiceRestart          = 1012
	StatusTryAgainLater           = 1013
	StatusBadGateway              = 1014
	StatusTLSHandShake            = 1015
)

// name according to https://developer.mozilla.org/en-US/docs/Web/API/CloseEvent
var statusText = map[int]string{
	StatusNormalClosure:           "Normal Closure",
	StatusGoingAway:               "Going Away",
	StatusProtocolError:           "Protocol Error",
	StatusUnsupportedData:         "Unsupported Data",
	StatusNoStatusReceived:        "No Status Received",
	StatusAbnormalClosure:         "Abnormal Closure",
	StatusInvalidFramePayloadData: "Invalid frame payload data",
	StatusPolicyViolation:         "Policy Violation",
	StatusMessageTooBig:           "Message too big",
	StatusMissingExtension:        "Missing Extension",
	StatusInternalError:           "Internal Error",
	StatusServiceRestart:          "Service Restart",
	StatusTryAgainLater:           "Try Again Later",
	StatusBadGateway:              "Bad Gateway",
	StatusTLSHandShake:            "TLS HandShake",
}

// StatusText returns a text for the WebSocket status code
func StatusText(code int) string {
	return statusText[code]
}
