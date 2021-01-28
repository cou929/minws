package minws

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"strings"

	minwshttp "github.com/cou929/minws/http"
	"github.com/cou929/minws/ws"
)

const magicStr = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

// HandShake establishes websocket connection
func HandShake(tcpConn net.Conn) (*ws.Conn, error) {
	httpConn := minwshttp.NewConn(tcpConn)

	res, err := httpConn.ReadRequest()
	if err != nil {
		return nil, fmt.Errorf("failed to readRequest %w", err)
	}
	err = handleHandShake(res, res.Req)
	if err != nil {
		log.Printf("Failed to complete handshake %v", err)
	}
	res.FinishRequest()

	wsConn := ws.NewConn(tcpConn)

	return wsConn, nil
}

func handleHandShake(w minwshttp.ResponseWriter, req *minwshttp.Request) error {
	if err := validateRequest(req); err != nil {
		w.SetStatus(minwshttp.StatusBadRequest)
		w.SetHeader("Connection", "close")
		fmt.Fprintf(w, "%s\n", err)
		return err
	}

	w.SetStatus(minwshttp.StatusSwitchingProtocols)
	w.SetHeader("Upgrade", "websocket")
	w.SetHeader("Connection", "Upgrade")
	swa := calcSecWebsocketAccept(req.Header.Get("Sec-WebSocket-Key"))
	w.SetHeader("Sec-WebSocket-Accept", swa)
	if req.Header.Has("Sec-WebSocket-Protocol") {
		// todo: check protocol to support
		// todo: handle multiple proto case
		w.SetHeader("Sec-WebSocket-Protocol", req.Header.Get("Sec-WebSocket-Protocol"))
	}
	// todo: handle Sec-WebSocket-Extensions
	// todo: handle Sec-WebSocket-Version

	return nil
}

func validateRequest(req *minwshttp.Request) error {
	// HTTP/1.1 Upgrade request
	if req.ProtoMajor != 1 || req.ProtoMinor != 1 {
		return fmt.Errorf("Must be HTTP/1.1")
	}
	if req.Method != "GET" {
		return fmt.Errorf("Must be GET Request")
	}
	if !req.Header.Has("Connection") || req.Header.Get("Connection") != "Upgrade" {
		return fmt.Errorf("Must send header Connection: Upgrade")
	}
	if !req.Header.Has("Upgrade") || !strings.Contains(req.Header.Get("Upgrade"), "websocket") {
		return fmt.Errorf("Must send header Upgrade: websocket")
	}

	// Websocket headers
	if !req.Header.Has("Sec-WebSocket-Key") {
		return fmt.Errorf("Must send header Sec-WebSocket-Key")
	}
	if !req.Header.Has("Sec-WebSocket-Version") {
		return fmt.Errorf("Must send header Sec-WebSocket-Version")
	}

	return nil
}

func calcSecWebsocketAccept(key string) string {
	sha1ed := sha1.Sum(([]byte)(key + magicStr))
	return base64.StdEncoding.EncodeToString(sha1ed[:])
}
