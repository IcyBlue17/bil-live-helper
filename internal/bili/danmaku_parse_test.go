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

func TestParseInteractFollow(t *testing.T) {
	cmd := loadCmd(t, "interact_word.json")
	cmd["data"].(map[string]any)["msg_type"] = 2.0 // follow
	m := parseCommand(cmd, 0)
	if m == nil || m.Kind != app.KindFollow || m.UID != 9988776 {
		t.Fatalf("msg = %+v", m)
	}
}

func TestParseInteractIgnored(t *testing.T) {
	cmd := loadCmd(t, "interact_word.json")
	cmd["data"].(map[string]any)["msg_type"] = 6.0 // share → ignored
	if m := parseCommand(cmd, 0); m != nil {
		t.Fatalf("expected nil, got %+v", m)
	}
}

func TestParseSuperChat(t *testing.T) {
	m := parseCommand(loadCmd(t, "super_chat.json"), 0)
	if m == nil {
		t.Fatal("nil message")
	}
	if m.Kind != app.KindSuperChat || m.SuperChat == nil {
		t.Fatalf("kind/sc = %+v", m)
	}
	if m.SuperChat.Price != 30 {
		t.Fatalf("sc = %+v", m.SuperChat)
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
	if m == nil || m.Kind != app.KindGuard || m.Guard != app.GuardJianzha {
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

// The site-wide gift banners (NOTICE_MSG) are about other rooms and stay out
// of the feed.
func TestParseNoticeMsgHidden(t *testing.T) {
	cmd := map[string]any{
		"cmd":         "NOTICE_MSG",
		"msg_type":    2,
		"real_roomid": 12345,
		"msg_common":  "<%某用户%>投喂<%某主播%>1个王牌将士，点击前往TA的房间去围观吧！",
	}
	if m := parseCommand(cmd, 0); m != nil {
		t.Fatalf("expected nil, got %+v", m)
	}
}

func TestParseUnknown(t *testing.T) {
	if m := parseCommand(map[string]any{"cmd": "WIDGET_BANNER"}, 0); m != nil {
		t.Fatalf("expected nil, got %+v", m)
	}
}

func TestParseDanmakuSticker(t *testing.T) {
	m := parseCommand(loadCmd(t, "danmu_sticker.json"), 0)
	if m == nil {
		t.Fatal("nil message")
	}
	if m.Face != "https://i0.hdslb.com/face.jpg" {
		t.Fatalf("face = %q", m.Face)
	}
	if m.Emoticon == nil {
		t.Fatal("expected Emoticon")
	}
	if m.Emoticon.Token != "official_124" || m.Emoticon.URL != "https://i0.hdslb.com/bfs/live/emo.png" {
		t.Fatalf("emoticon = %+v", m.Emoticon)
	}
	if m.Emoticon.Width != 100 || m.Emoticon.Height != 100 {
		t.Fatalf("emoticon size = %+v", m.Emoticon)
	}
}

func TestParseDanmakuEmots(t *testing.T) {
	m := parseCommand(loadCmd(t, "danmu_emots.json"), 0)
	if m == nil {
		t.Fatal("nil message")
	}
	if m.Emoticon != nil {
		t.Fatalf("not a sticker: %+v", m.Emoticon)
	}
	if len(m.Emots) != 2 {
		t.Fatalf("emots = %+v", m.Emots)
	}
	e, ok := m.Emots["[doge]"]
	if !ok || e.URL != "https://i0.hdslb.com/e2.png" || e.Width != 20 {
		t.Fatalf("emots[doge] = %+v", e)
	}
}

func TestParseLikeClick(t *testing.T) {
	m := parseCommand(loadCmd(t, "like_info_click.json"), 0)
	if m == nil || m.Kind != app.KindLike {
		t.Fatalf("msg = %+v", m)
	}
	if m.UID != 777888 || m.Name != "点赞人" || m.Text != "为主播点赞了" {
		t.Fatalf("msg = %+v", m)
	}
	if m.Medal == nil || m.Medal.Name != "迷子" || m.Medal.Level != 5 {
		t.Fatalf("medal = %+v", m.Medal)
	}
}

func TestCounterCmd(t *testing.T) {
	w, o, l := counterCmd(map[string]any{
		"cmd":  "WATCHED_CHANGE",
		"data": map[string]any{"num": 12345.0, "text_large": "1.2万"},
	})
	if w != 12345 || o != 0 || l != 0 {
		t.Fatalf("watched = %d/%d/%d", w, o, l)
	}
	w, o, l = counterCmd(map[string]any{
		"cmd":  "ONLINE_RANK_COUNT",
		"data": map[string]any{"count": 678.0},
	})
	if w != 0 || o != 678 || l != 0 {
		t.Fatalf("online = %d/%d/%d", w, o, l)
	}
	w, o, l = counterCmd(map[string]any{
		"cmd":  "LIKE_INFO_V3_UPDATE:abc",
		"data": map[string]any{"click_count": 999.0},
	})
	if w != 0 || o != 0 || l != 999 {
		t.Fatalf("likes = %d/%d/%d", w, o, l)
	}
	w, o, l = counterCmd(map[string]any{"cmd": "DANMU_MSG", "info": []any{}})
	if w != 0 || o != 0 || l != 0 {
		t.Fatalf("danmaku = %d/%d/%d", w, o, l)
	}
}
