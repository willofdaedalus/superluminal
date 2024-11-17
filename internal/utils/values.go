package utils

// headers and their values
// consider transferring less bytes over the connection
// const (
//
//	HdrInfoMsg = "sprlmnl+inf:"
//	HdrErrMsg  = "sprlmnl+err:"
//	HdrAckMsg  = "sprlmnl+ack:"
//	HdrResMsg  = "sprlmnl+res:"
//
// )
const (
	MaxPayloadSize = 4096 // 4kb
)

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
	AckClientShutdown
)

// info messages
const (
	InfoServerShutdown = iota + 101
	InfoClientShutdown
)

// err messages
const (
	ErrResendPass = iota + 401
	ErrWrongPassphrase
	ErrServerKick
)

// res messages
const (
	RespSelfReport = iota + 501
	RespClientShutdown
	RespNewPass
)
