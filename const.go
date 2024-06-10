package ginhelper

// cloudfare headers
const (
	CfRay            = "CF-RAY"
	CfConnectingIP   = "CF-Connecting-IP"
	CfConnectingIPv6 = "CF-Connecting-IPv6"
	CfEwVia          = "CF-EW-Via"
	CfPseudoIPv4     = "CF-Pseudo-IPv4"
	TrueClientIP     = "True-Client-IP" // Enterprise plan only
	XForwardedFor    = "X-Forwarded-For"
	XForwardedProto  = "X-Forwarded-Proto"
	CfIPCountry      = "CF-IPCountry"
	CfVisitor        = "CF-Visitor"
	CdnLoop          = "CDN-Loop"
	CfWorker         = "CF-Worker"
	Connection       = "Connection"
	AcceptEncoding   = "Accept-Encoding"
)
