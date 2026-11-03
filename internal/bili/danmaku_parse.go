package bili

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/IcyBlue17/bil-live-helper/app"
)

var msgCounter atomic.Int64

func nextID() string { return fmt.Sprint(msgCounter.Add(1)) }

// --- JSON helpers for the loosely-typed command payloads ---

func jsArr(v any) []any {
	if a, ok := v.([]any); ok {
		return a
	}
	return nil
}

func jsMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}

func jsStr(v any) string {
	s, _ := v.(string)
	return s
}

func jsNum(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case json.Number:
		f, _ := n.Float64()
		return f
	}
	return 0
}

func jsInt(v any) int64 { return int64(jsNum(v)) }

func at(a []any, i int) any {
	if i < len(a) {
		return a[i]
	}
	return nil
}

// parseCommand turns one decoded command into an app.Message.
// ownUID marks弹幕 sent by the logged-in user (KindSelf).
// Returns nil when the command is ignored.
func parseCommand(cmd map[string]any, ownUID int64) *app.Message {
	name := jsStr(cmd["cmd"])
	if i := strings.IndexByte(name, ':'); i >= 0 {
		name = name[:i]
	}
	switch name {
	case "DANMU_MSG", "DANMU_MSG_MIRROR":
		return parseDanmaku(jsArr(cmd["info"]), ownUID)
	case "SEND_GIFT":
		return parseGift(jsMap(cmd["data"]))
	case "GUARD_BUY", "USER_TOAST_MSG":
		return parseGuard(jsMap(cmd["data"]))
	case "INTERACT_WORD":
		return parseInteract(jsMap(cmd["data"]))
	case "SUPER_CHAT_MESSAGE":
		return parseSuperChat(jsMap(cmd["data"]))
	case "NOTICE_MSG", "WARNING", "CUT_OFF":
		return parseNotice(name, cmd)
	}
	return nil
}

func parseDanmaku(info []any, ownUID int64) *app.Message {
	if len(info) < 3 {
		return nil
	}
	modeInfo := jsArr(at(info, 0))
	ts := jsInt(at(modeInfo, 4)) // ms
	rnd := jsInt(at(modeInfo, 5))
	user := jsArr(at(info, 2))
	uid := jsInt(at(user, 0))

	m := &app.Message{
		ID:   nextID(),
		Kind: app.KindDanmaku,
		Time: time.UnixMilli(ts),
		UID:  uid,
		Name: jsStr(at(user, 1)),
		Text: jsStr(at(info, 1)),
	}
	if ts == 0 {
		m.Time = time.Now()
	}
	if jsInt(at(user, 2)) != 0 {
		m.Admin = true
	}
	if m.ID != "" && rnd != 0 {
		m.ID = fmt.Sprintf("%d-%d", ts, rnd)
	}
	if ownUID > 0 && uid == ownUID {
		m.Kind = app.KindSelf
	}
	if medal := jsArr(at(info, 3)); len(medal) > 1 {
		if name := jsStr(at(medal, 1)); name != "" {
			m.Medal = &app.Medal{Name: name, Level: int(jsInt(at(medal, 0)))}
		}
	}
	if ul := jsArr(at(info, 4)); len(ul) > 0 {
		m.UserLevel = int(jsInt(at(ul, 0)))
	}
	m.Guard = app.GuardLevel(jsInt(at(info, 7)))
	return m
}

func parseGift(d map[string]any) *app.Message {
	if d == nil {
		return nil
	}
	num := int(jsInt(d["num"]))
	var price int64
	if jsStr(d["coin_type"]) == "gold" {
		if total := jsInt(d["total_coin"]); total > 0 {
			price = total
		} else {
			price = jsInt(d["price"]) * int64(num)
		}
	}
	ts := jsInt(d["timestamp"])
	t := time.Now()
	if ts > 0 {
		t = time.Unix(ts, 0)
	}
	return &app.Message{
		ID:    nextID(),
		Kind:  app.KindGift,
		Time:  t,
		UID:   jsInt(d["uid"]),
		Name:  jsStr(d["uname"]),
		Guard: app.GuardLevel(jsInt(d["guard_level"])),
		Gift:  &app.Gift{Name: jsStr(d["giftName"]), Num: num, Price: price},
	}
}

func parseGuard(d map[string]any) *app.Message {
	if d == nil {
		return nil
	}
	level := app.GuardLevel(jsInt(d["guard_level"]))
	name := jsStr(d["username"])
	if name == "" {
		name = jsStr(d["uname"])
	}
	guardName := map[app.GuardLevel]string{
		app.GuardZongdu:  "总督",
		app.GuardTidu:    "提督",
		app.GuardJianzha: "舰长",
	}[level]
	if guardName == "" {
		guardName = "大航海"
	}
	price := jsInt(d["price"])
	num := int(jsInt(d["num"]))
	if num <= 0 {
		num = 1
	}
	return &app.Message{
		ID:    nextID(),
		Kind:  app.KindGift,
		Time:  time.Now(),
		UID:   jsInt(d["uid"]),
		Name:  name,
		Guard: level,
		Gift:  &app.Gift{Name: guardName, Num: num, Price: price},
	}
}

func parseInteract(d map[string]any) *app.Message {
	if d == nil {
		return nil
	}
	if jsInt(d["msg_type"]) != 1 { // 1 = enter, others (follow/share) ignored
		return nil
	}
	ts := jsInt(d["timestamp"])
	t := time.Now()
	if ts > 0 {
		t = time.Unix(ts, 0)
	}
	return &app.Message{
		ID:   nextID(),
		Kind: app.KindEnter,
		Time: t,
		UID:  jsInt(d["uid"]),
		Name: jsStr(d["uname"]),
	}
}

func parseSuperChat(d map[string]any) *app.Message {
	if d == nil {
		return nil
	}
	ts := jsInt(d["start_time"])
	if ts == 0 {
		ts = jsInt(d["ts"])
	}
	t := time.Now()
	if ts > 0 {
		t = time.Unix(ts, 0)
	}
	return &app.Message{
		ID:   nextID(),
		Kind: app.KindGift,
		Time: t,
		UID:  jsInt(d["uid"]),
		Name: jsStr(jsMap(d["user_info"])["uname"]),
		Text: jsStr(d["message"]),
		Gift: &app.Gift{Name: "醒目留言", Num: 1, Price: jsInt(d["price"]) * 1000},
	}
}

func parseNotice(name string, cmd map[string]any) *app.Message {
	var text string
	switch name {
	case "NOTICE_MSG":
		text = jsStr(cmd["msg_common"])
		if text == "" {
			if m := jsMap(cmd["data"]); m != nil {
				text = jsStr(m["msg_common"])
			}
		}
	case "WARNING", "CUT_OFF":
		text = jsStr(cmd["msg"])
	}
	if text == "" {
		return nil
	}
	return &app.Message{
		ID:   nextID(),
		Kind: app.KindNotice,
		Time: time.Now(),
		Text: text,
	}
}

// parseHeartbeatReply reads the popularity from a heartbeat reply body.
func parseHeartbeatReply(body []byte) int64 {
	if len(body) >= 4 {
		return int64(binary.BigEndian.Uint32(body))
	}
	return 0
}
