package client

import (
	"log"
	"willofdaedalus/superluminal/internal/payload/base"
	err1 "willofdaedalus/superluminal/internal/payload/error"
	"willofdaedalus/superluminal/internal/utils"
)

func (c *client) handleErrPayload(payload base.Payload_Error) error {
	switch payload.Error.GetCode() {
	case err1.ErrorMessage_ERROR_SERVER_FULL:
		// TODO => add a notification system to send messages like this to the
		// user through bubble tea type  interface
		log.Println(payload.Error.GetDetail())
		c.exitChan <- struct{}{}
		return utils.ErrServerFull
	}

	return utils.ErrUnspecifiedPayload
}
