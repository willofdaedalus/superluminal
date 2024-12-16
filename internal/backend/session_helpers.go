package backend

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"
	"willofdaedalus/superluminal/internal/payload/auth"
	"willofdaedalus/superluminal/internal/payload/base"
	"willofdaedalus/superluminal/internal/payload/common"
	"willofdaedalus/superluminal/internal/payload/info"
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
	for try := 0; try < maxAuthChances; try++ {
		log.Println("try no", try)

		tempCtx, cancel := context.WithTimeout(ctx, clientKickTimeout)
		newPayload := utils.PrependLength(authPayload)
		err := utils.TryWriteCtx(tempCtx, conn, newPayload)
		cancel()
		if err != nil {
			if errors.Is(err, utils.ErrCtxTimeOut) {
				log.Println("write operation timed out...")
				continue
			}
			// let handleNewConn handle the error; send it upstream
			return "", err
		}

		clientResp, err := s.readFromClient(ctx, conn)
		if err != nil {
			// if errors.Is(err, utils.ErrCtxTimeOut) {
			// 	continue
			// }
			return "", err
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

		authResp, ok := authPayload.Content.(*base.Payload_Auth).Auth.AuthType.(*auth.Authentication_Response)
		if !ok {
			if authPayload.Content.(*base.Payload_Info).Info.InfoType == info.Info_INFO_SHUTDOWN {
				return "", utils.ErrClientEarlyExit
			}
			return "", fmt.Errorf("received wrong response")
		}

		if utils.CheckPassphrase(s.hash, authResp.Response.GetPassphrase()) {
			return authResp.Response.GetUsername(), nil
		}
	}

	return "", utils.ErrFailedServerAuth
}

// generate a random passphrase
func (s *session) regenPassLoop(ctx context.Context) {
	ticker := time.NewTicker(s.passRegenTime)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.pass, s.hash, _ = genPassAndHash(debugPassCount)
			fmt.Println(s.pass)
		case <-ctx.Done():
			return
		}
	}
}

// writeToClient provides a way to synchronize writes across the server
func (s *session) writeToClient(ctx context.Context, conn net.Conn, data []byte) error {
	s.tracker.IncrementWrite()
	defer s.tracker.DecrementWrite()

	return utils.TryWriteCtx(ctx, conn, data)
}

// readFromClient provides a way to synchronize reads across the server
func (s *session) readFromClient(ctx context.Context, conn net.Conn) ([]byte, error) {
	s.tracker.IncrementRead()
	defer s.tracker.DecrementRead()

	return utils.TryReadCtx(ctx, conn)
}
