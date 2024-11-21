package backend

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"willofdaedalus/superluminal/internal/payload/auth"
	"willofdaedalus/superluminal/internal/payload/base"
	"willofdaedalus/superluminal/internal/payload/common"
	"willofdaedalus/superluminal/internal/utils"
)

// genPassAndHash generates, hashes and returns a new pass and hash
func genPassAndHash(count int) (string, string, error) {
	pass, err := utils.GeneratePassphrase(count)
	if err != nil {
		return "", "", err
	}

	hash, err := utils.HashPassphrase(pass)
	if err != nil {
		return "", "", err
	}

	return pass, hash, nil
}

// authenticateClient exchanges authentication with the client by sending an auth
// request to the client upon successful connection
// it tries to read from the client up to a minute and if no activity such as an
// auth response, close the client with a message otherwise for every wrong passphrase
// reset the wait timeout up to 3x
func (s *session) authenticateClient(ctx context.Context, conn net.Conn) (string, error) {
	authReq := base.GenerateAuthReq()
	authPayload, err := base.EncodePayload(common.Header_HEADER_AUTH, authReq)
	if err != nil {
		return "", err
	}

	return s.tryValidateClientPass(ctx, conn, authPayload)
}

func (s *session) tryValidateClientPass(ctx context.Context, conn net.Conn, authPayload []byte) (string, error) {
	for try := 0; try < MaxAuthChances; try++ {
		// send the auth request bytes
		// let TryWriteCtx handle the ctx timeout
		err := utils.TryWriteCtx(ctx, conn, authPayload)
		if err != nil {
			// let handleNewConn handle the error; send it upstream
			return "", err
		}

		clientResp, err := utils.TryReadCtx(ctx, conn)
		if err != nil {
			if errors.Is(err, utils.ErrCtxTimeOut) {
				continue
			}
			// Return other errors immediately
			return "", fmt.Errorf("failed to read client response: %w", err)
		}
		fmt.Println("received", len(clientResp))

		authPayload, err := base.DecodePayload(clientResp)
		if err != nil {
			return "", fmt.Errorf("failed to decode payload %v", err)
		}

		if authPayload.GetHeader() != common.Header_HEADER_AUTH {
			log.Printf("got %v", authPayload)
			return "", utils.ErrInvalidHeader
		}

		// we expect an auth response; anything else is an error
		// TODO; possible scenario where the client cancels and sends a shutdown message
		// might cause this check to fail...
		authResp, _ := authPayload.Content.(*base.Payload_Auth).Auth.AuthType.(*auth.Authentication_Response)
		// if !ok {
		// 	return
		// }

		if utils.CheckPassphrase(s.hash, authResp.Response.GetPassphrase()) {
			return authResp.Response.GetUsername(), nil
		}
		fmt.Println("wrong passphrase")
	}

	return "", utils.ErrFailedServerAuth
}
