package ws

import (
	"fmt"
	"io"
)

// DataFrame represents data frames of WebSocket protocol
type DataFrame struct {
	fin        bool
	opCode     OpCode
	mask       bool
	payloadLen int
	maskingKey [4]byte
	rawPayload []byte
	payload    []byte
}

// OpCode is opcode of WebSocket protocol
type OpCode int

const (
	// OpCodeContinuation represents opcode %x0
	OpCodeContinuation OpCode = 0x0
	// OpCodeText represents opcode %x1
	OpCodeText = 0x1
	// OpCodeBinary represents opcode %x2
	OpCodeBinary = 0x2
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
	n, err := r.Read(buf)
	if err != nil {
		return nil, err
	}
	if n != 2 {
		return nil, fmt.Errorf("invalid frame format %#b, want len %d, got len %d", buf, 2, n)
	}

	df.fin = buf[0]>>7 == 1
	df.opCode = OpCode(buf[0] & 0b00001111)
	df.mask = buf[1]>>7 == 1
	df.payloadLen = int(buf[1] & 0b01111111)

	if df.payloadLen >= 126 {
		return nil, fmt.Errorf("extended payload length is not supported yet")
	}

	mk := make([]byte, 4)
	n, err = r.Read(mk)
	if err != nil {
		return nil, err
	}
	if n != 4 {
		return nil, fmt.Errorf("invalid frame format %#b, want len %d, got len %d", mk, 4, n)
	}
	copy(df.maskingKey[:], mk)

	encoded := make([]byte, df.payloadLen)
	n, err = r.Read(encoded)
	if err != nil {
		return nil, err
	}
	if n != df.payloadLen {
		return nil, fmt.Errorf("invalid frame format %#b, want len %d, got len %d", encoded, df.payloadLen, n)
	}
	df.rawPayload = encoded

	return df, nil
}

// NewDataFrameFromMessage build DataFrame from message to send
func NewDataFrameFromMessage(msg string, mask bool) (*DataFrame, error) {
	df := &DataFrame{
		payload:    []byte(msg),
		mask:       mask,
		fin:        true,
		opCode:     OpCodeText,
		payloadLen: len(msg),
	}
	return df, nil
}

// Message returns payload value
func (d *DataFrame) Message() string {
	if d.payload != nil {
		return string(d.payload)
	}

	decoded := make([]byte, d.payloadLen)
	for i := 0; i < d.payloadLen; i++ {
		decoded[i] = d.rawPayload[i] ^ d.maskingKey[i%4]
	}
	d.payload = decoded

	return string(d.payload)
}

// Frame build DataFrame binary representation
func (d *DataFrame) Frame() []byte {
	res := make([]byte, 2)
	fin := 0
	if d.fin {
		fin = 1
	}
	res[0] = byte((fin << 7) | int(d.opCode))
	res[1] = byte(d.payloadLen)
	if d.mask {
		res[1] = res[1] | 0b10000000
	}
	res = append(res, ([]byte)(d.payload)...)
	return res
}
