package utils

// headers and their values
// consider transferring less bytes over the connection
const (
	HdrInfoMsg = "sprlmnl+inf:"
	HdrErrMsg  = "sprlmnl+err:"
	HdrAckMsg  = "sprlmnl+ack:"
	HdrResMsg  = "sprlmnl+res:"
)

const (
	HdrUnknownVal = iota + 1
	HdrInfoVal
	HdrAckVal
	HdrErrVal
	HdrResVal
)

// ack messages
const (
	AckSelfReport = "self_report"
)
