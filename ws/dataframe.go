package ws

import (
	"encoding/binary"
	"fmt"
	"io"
)

// DataFrame represents data frames of WebSocket protocol
type DataFrame struct {
	fin        bool
	OpCode     OpCode
	mask       bool
	payloadLen int
	maskingKey [4]byte
	rawPayload []byte
	payload    []byte
}

// OpCode is opcode of WebSocket protocol
type OpCode int

// see: https://tools.ietf.org/html/rfc6455#section-5.2
const (
	OpCodeContinuation OpCode = 0x0
	OpCodeText                = 0x1
	OpCodeBinary              = 0x2
	OpCodeClose               = 0x8
	OpCodePing                = 0x9
	OpCodePong                = 0xA
)

/*
    0                   1                   2                   3
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
   +-+-+-+-+-------+-+-------------+-------------------------------+
   |F|R|R|R| opcode|M| Payload len |    Extended payload length    |
   |I|S|S|S|  (4)  |A|     (7)     |             (16/64)           |
   |N|V|V|V|       |S|             |   (if payload len==126/127)   |
   | |1|2|3|       |K|             |                               |
   +-+-+-+-+-------+-+-------------+ - - - - - - - - - - - - - - - +
   |     Extended payload length continued, if payload len == 127  |
   + - - - - - - - - - - - - - - - +-------------------------------+
   |                               |Masking-key, if MASK set to 1  |
   +-------------------------------+-------------------------------+
   | Masking-key (continued)       |          Payload Data         |
   +-------------------------------- - - - - - - - - - - - - - - - +
   :                     Payload Data continued ...                :
   + - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - +
   |                     Payload Data continued ...                |
   +---------------------------------------------------------------+
*/

// NewDataFrameFromReader read request and build DataFrame
func NewDataFrameFromReader(r io.Reader) (*DataFrame, error) {
	df := &DataFrame{}

	buf := make([]byte, 2)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, fmt.Errorf("Failed to read first 16 bit %w", err)
	}

	df.fin = buf[0]>>7 == 1
	df.OpCode = OpCode(buf[0] & 0b00001111)
	df.mask = buf[1]>>7 == 1
	leadingPayloadLen := int(buf[1] & 0b01111111)
	payloadLen, err := readExtendedPayloadLen(r, leadingPayloadLen)
	if err != nil {
		return nil, err
	}
	df.payloadLen = payloadLen

	mk := make([]byte, 4)
	if _, err := io.ReadFull(r, mk); err != nil {
		return nil, fmt.Errorf("Failed to read masking key %w", err)
	}
	copy(df.maskingKey[:], mk)

	encoded := make([]byte, df.payloadLen)
	if _, err := io.ReadFull(r, encoded); err != nil {
		return nil, fmt.Errorf("Failed to read payload %w", err)
	}
	df.rawPayload = encoded

	return df, nil
}

// NewDataFrameFromTextMessage build DataFrame from text message to send
func NewDataFrameFromTextMessage(msg string, mask bool) (*DataFrame, error) {
	df := &DataFrame{
		payload:    []byte(msg),
		mask:       mask,
		fin:        true,
		OpCode:     OpCodeText,
		payloadLen: len(msg),
	}
	return df, nil
}

// NewDataFrameFromBinaryMessage build DataFrame from binary message to send
func NewDataFrameFromBinaryMessage(msg []byte, mask bool) (*DataFrame, error) {
	df := &DataFrame{
		payload:    msg,
		mask:       mask,
		fin:        true,
		OpCode:     OpCodeBinary,
		payloadLen: len(msg),
	}
	return df, nil
}

// Message returns payload value
func (d *DataFrame) Message() []byte {
	if d.payload != nil {
		return d.payload
	}

	decoded := make([]byte, d.payloadLen)
	for i := 0; i < d.payloadLen; i++ {
		decoded[i] = d.rawPayload[i] ^ d.maskingKey[i%4]
	}
	d.payload = decoded

	return d.payload
}

// Frame build DataFrame binary representation
func (d *DataFrame) Frame() []byte {
	res := make([]byte, 2)
	fin := 0
	if d.fin {
		fin = 1
	}
	res[0] = byte((fin << 7) | int(d.OpCode))
	res[1] = byte(d.payloadLen)
	if d.mask {
		res[1] = res[1] | 0b10000000
	}
	res = append(res, ([]byte)(d.payload)...)
	return res
}

func readExtendedPayloadLen(r io.Reader, leadingPayloadLen int) (int, error) {
	if leadingPayloadLen < 126 {
		return leadingPayloadLen, nil
	}

	var res int
	switch leadingPayloadLen {
	case 126:
		bl := 2
		buf := make([]byte, bl)
		if _, err := io.ReadFull(r, buf); err != nil {
			return 0, fmt.Errorf("Failed to read extended payload length %w", err)
		}
		l := binary.BigEndian.Uint16(buf)
		res = int(l)
	case 127:
		bl := 8
		buf := make([]byte, bl)
		if _, err := io.ReadFull(r, buf); err != nil {
			return 0, fmt.Errorf("Failed to read extended payload length %w", err)
		}
		l := binary.BigEndian.Uint64(buf)
		res = int(l)
	}

	return res, nil
}
