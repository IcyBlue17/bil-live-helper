package bili

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/andybalholm/brotli"
)

// Bilibili danmaku websocket protocol (core/danmaku_protocol.py).
// Header: packLen(4) + headerLen(2) + ver(2) + op(4) + seq(4), big-endian.
const headerLen = 16

const (
	protoNormal    = 0
	protoHeartbeat = 1
	protoDeflate   = 2
	protoBrotli    = 3
)

const (
	opHandshake      = 0
	opHandshakeReply = 1
	opHeartbeat      = 2
	opHeartbeatReply = 3
	opSendMsg        = 4
	opSendMsgReply   = 5
	opDisconnect     = 6
	opAuth           = 7
	opAuthReply      = 8
	opRaw            = 9
)

type header struct {
	packLen   uint32
	headerLen uint16
	ver       uint16
	op        uint32
	seq       uint32
}

// makePacket builds a full packet; body may be map/struct (JSON), string or bytes.
func makePacket(body any, op uint32) []byte {
	var b []byte
	switch v := body.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		b, _ = json.Marshal(body)
	}
	buf := make([]byte, headerLen+len(b))
	binary.BigEndian.PutUint32(buf[0:4], uint32(len(buf)))
	binary.BigEndian.PutUint16(buf[4:6], headerLen)
	binary.BigEndian.PutUint16(buf[6:8], protoHeartbeat) // ver=1 like upstream
	binary.BigEndian.PutUint32(buf[8:12], op)
	binary.BigEndian.PutUint32(buf[12:16], 1)
	copy(buf[headerLen:], b)
	return buf
}

var errShortPacket = errors.New("bili: short packet")

// parsePackets splits a websocket frame into (header, body) pairs.
// Bodies of compressed packets are decompressed and recursed into, so every
// returned packet has an uncompressed body.
func parsePackets(data []byte, fn func(h header, body []byte) error) error {
	for off := 0; off < len(data); {
		if len(data)-off < headerLen {
			return errShortPacket
		}
		h := header{
			packLen:   binary.BigEndian.Uint32(data[off:]),
			headerLen: binary.BigEndian.Uint16(data[off+4:]),
			ver:       binary.BigEndian.Uint16(data[off+6:]),
			op:        binary.BigEndian.Uint32(data[off+8:]),
			seq:       binary.BigEndian.Uint32(data[off+12:]),
		}
		if h.packLen == 0 || int(h.packLen) > len(data)-off {
			return fmt.Errorf("bili: bad packLen %d at offset %d", h.packLen, off)
		}
		body := data[off+int(h.headerLen) : off+int(h.packLen)]

		switch h.op {
		case opSendMsgReply:
			switch h.ver {
			case protoDeflate:
				raw, err := inflateZlib(body)
				if err != nil {
					return err
				}
				if err := parsePackets(raw, fn); err != nil {
					return err
				}
			case protoBrotli:
				raw, err := inflateBrotli(body)
				if err != nil {
					return err
				}
				if err := parsePackets(raw, fn); err != nil {
					return err
				}
			default:
				if err := fn(h, body); err != nil {
					return err
				}
			}
		default:
			if err := fn(h, body); err != nil {
				return err
			}
		}
		off += int(h.packLen)
	}
	return nil
}

func inflateZlib(b []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

func inflateBrotli(b []byte) ([]byte, error) {
	return io.ReadAll(brotli.NewReader(bytes.NewReader(b)))
}
