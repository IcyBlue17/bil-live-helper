package bili

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/json"
	"testing"

	"github.com/andybalholm/brotli"
)

func TestMakePacket(t *testing.T) {
	p := makePacket(map[string]any{"uid": 1}, opAuth)
	if len(p) < headerLen {
		t.Fatal("packet too short")
	}
	if got := binary.BigEndian.Uint32(p[0:4]); got != uint32(len(p)) {
		t.Fatalf("packLen = %d, want %d", got, len(p))
	}
	if got := binary.BigEndian.Uint16(p[4:6]); got != headerLen {
		t.Fatalf("headerLen = %d", got)
	}
	if got := binary.BigEndian.Uint32(p[8:12]); got != opAuth {
		t.Fatalf("op = %d", got)
	}
	var body map[string]any
	if err := json.Unmarshal(p[headerLen:], &body); err != nil || body["uid"] != float64(1) {
		t.Fatalf("body = %v err %v", body, err)
	}
}

func frameWithVer(body []byte, op uint32, ver uint16) []byte {
	buf := make([]byte, headerLen+len(body))
	binary.BigEndian.PutUint32(buf[0:4], uint32(len(buf)))
	binary.BigEndian.PutUint16(buf[4:6], headerLen)
	binary.BigEndian.PutUint16(buf[6:8], ver)
	binary.BigEndian.PutUint32(buf[8:12], op)
	binary.BigEndian.PutUint32(buf[12:16], 1)
	copy(buf[headerLen:], body)
	return buf
}

func TestParsePacketsMulti(t *testing.T) {
	// Two plain command packets in one frame.
	a := frameWithVer([]byte(`{"cmd":"A"}`), opSendMsgReply, protoNormal)
	b := frameWithVer([]byte(`{"cmd":"B"}`), opSendMsgReply, protoNormal)
	var got []string
	err := parsePackets(append(a, b...), func(h header, body []byte) error {
		got = append(got, string(body))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != `{"cmd":"A"}` || got[1] != `{"cmd":"B"}` {
		t.Fatalf("got %v", got)
	}
}

func TestParsePacketsZlib(t *testing.T) {
	inner := append(frameWithVer([]byte(`{"cmd":"X"}`), opSendMsgReply, protoNormal),
		frameWithVer([]byte(`{"cmd":"Y"}`), opSendMsgReply, protoNormal)...)
	var zb bytes.Buffer
	w := zlib.NewWriter(&zb)
	w.Write(inner)
	w.Close()

	var got []string
	err := parsePackets(frameWithVer(zb.Bytes(), opSendMsgReply, protoDeflate),
		func(h header, body []byte) error {
			got = append(got, string(body))
			return nil
		})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != `{"cmd":"X"}` || got[1] != `{"cmd":"Y"}` {
		t.Fatalf("got %v", got)
	}
}

func TestParsePacketsBrotli(t *testing.T) {
	inner := frameWithVer([]byte(`{"cmd":"B1"}`), opSendMsgReply, protoNormal)
	var bb bytes.Buffer
	w := brotli.NewWriter(&bb)
	w.Write(inner)
	w.Close()

	var got []string
	err := parsePackets(frameWithVer(bb.Bytes(), opSendMsgReply, protoBrotli),
		func(h header, body []byte) error {
			got = append(got, string(body))
			return nil
		})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != `{"cmd":"B1"}` {
		t.Fatalf("got %v", got)
	}
}

func TestHeartbeatReply(t *testing.T) {
	body := []byte{0, 0, 0x1E, 0x75}
	if got := parseHeartbeatReply(body); got != 7797 {
		t.Fatalf("popularity = %d", got)
	}
	if got := parseHeartbeatReply([]byte{1, 2}); got != 0 {
		t.Fatalf("short body = %d", got)
	}
}
