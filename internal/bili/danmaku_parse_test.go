package bili

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/IcyBlue17/bil-live-helper/app"
)

func loadCmd(t *testing.T, name string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	var cmd map[string]any
	if err := json.Unmarshal(raw, &cmd); err != nil {
		t.Fatal(err)
	}
	return cmd
}

func TestParseDanmaku(t *testing.T) {
	m := parseCommand(loadCmd(t, "danmu_msg.json"), 999)
	if m == nil {
		t.Fatal("nil message")
	}
	if m.Kind != app.KindDanmaku {
		t.Fatalf("kind = %q", m.Kind)
	}
	if m.Text != "主播今天状态好棒" || m.Name != "晚风" || m.UID != 1234567 {
		t.Fatalf("msg = %+v", m)
	}
	if !m.Admin {
		t.Fatal("expected admin")
	}
	if m.Guard != app.GuardJianzha {
		t.Fatalf("guard = %d", m.Guard)
	}
	if m.UserLevel != 40 {
		t.Fatalf("userLevel = %d", m.UserLevel)
	}
	if m.Medal == nil || m.Medal.Name != "迷子" || m.Medal.Level != 12 {
		t.Fatalf("medal = %+v", m.Medal)
	}
	if want := time.UnixMilli(1730000000000); !m.Time.Equal(want) {
		t.Fatalf("time = %v, want %v", m.Time, want)
	}
}

func TestParseDanmakuSelf(t *testing.T) {
	m := parseCommand(loadCmd(t, "danmu_msg.json"), 1234567)
	if m == nil || m.Kind != app.KindSelf {
		t.Fatalf("kind = %v", m)
	}
}

func TestParseGift(t *testing.T) {
	m := parseCommand(loadCmd(t, "send_gift.json"), 0)
	if m == nil {
		t.Fatal("nil message")
	}
	if m.Kind != app.KindGift || m.Gift == nil {
		t.Fatalf("kind/gift = %+v", m)
	}
	if m.Gift.Name != "辣条" || m.Gift.Num != 10 || m.Gift.Price != 10000 {
		t.Fatalf("gift = %+v", m.Gift)
	}
	if m.Name != "橘子汽水" || m.UID != 7654321 {
		t.Fatalf("msg = %+v", m)
	}
}

func TestParseInteract(t *testing.T) {
	m := parseCommand(loadCmd(t, "interact_word.json"), 0)
	if m == nil {
		t.Fatal("nil message")
	}
	if m.Kind != app.KindEnter || m.Name != "不睡觉的猫" || m.UID != 9988776 {
		t.Fatalf("msg = %+v", m)
	}
}

func TestParseInteractNonEnter(t *testing.T) {
	cmd := loadCmd(t, "interact_word.json")
	cmd["data"].(map[string]any)["msg_type"] = 2.0 // follow → ignored
	if m := parseCommand(cmd, 0); m != nil {
		t.Fatalf("expected nil, got %+v", m)
	}
}

func TestParseSuperChat(t *testing.T) {
	m := parseCommand(loadCmd(t, "super_chat.json"), 0)
	if m == nil {
		t.Fatal("nil message")
	}
	if m.Kind != app.KindGift || m.Gift == nil {
		t.Fatalf("kind/gift = %+v", m)
	}
	if m.Gift.Name != "醒目留言" || m.Gift.Price != 30000 {
		t.Fatalf("gift = %+v", m.Gift)
	}
	if m.Text != "点一首春日影" || m.Name != "星野" {
		t.Fatalf("msg = %+v", m)
	}
}

func TestParseGuardBuy(t *testing.T) {
	cmd := map[string]any{
		"cmd": "GUARD_BUY",
		"data": map[string]any{
			"uid":         555,
			"username":    "Lumi",
			"guard_level": 3.0,
			"price":       198000.0,
			"num":         1.0,
		},
	}
	m := parseCommand(cmd, 0)
	if m == nil || m.Kind != app.KindGift || m.Guard != app.GuardJianzha {
		t.Fatalf("msg = %+v", m)
	}
	if m.Gift.Name != "舰长" || m.Gift.Price != 198000 {
		t.Fatalf("gift = %+v", m.Gift)
	}
}

func TestParseNotice(t *testing.T) {
	m := parseCommand(map[string]any{"cmd": "CUT_OFF", "msg": "直播已被切断"}, 0)
	if m == nil || m.Kind != app.KindNotice || m.Text != "直播已被切断" {
		t.Fatalf("msg = %+v", m)
	}
}

func TestParseUnknown(t *testing.T) {
	if m := parseCommand(map[string]any{"cmd": "WIDGET_BANNER"}, 0); m != nil {
		t.Fatalf("expected nil, got %+v", m)
	}
}
