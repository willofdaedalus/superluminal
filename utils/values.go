package utils

// headers and their values
// consider transferring less bytes over the connection
// const (
// 	HdrInfoMsg = "sprlmnl+inf:"
// 	HdrErrMsg  = "sprlmnl+err:"
// 	HdrAckMsg  = "sprlmnl+ack:"
// 	HdrResMsg  = "sprlmnl+res:"
// )

// headers
const (
	HdrUnknownVal = -1
	HdrInfoVal    = 100
	HdrAckVal     = 300
	HdrErrVal     = 400
	HdrResVal     = 500
	HdrOut        = 600
)

// ack messages
const (
	AckSelfReport = iota + 301
)

// err messages
const (
	ErrResendPass = iota + 401
	ErrServerFull
	ErrWrongPassphrase
)

// res messages
const (
	RespSelfReport = iota + 501
)
