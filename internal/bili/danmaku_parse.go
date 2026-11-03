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
	case "LIKE_INFO_V3_CLICK":
		return parseLikeClick(jsMap(cmd["data"]))
	case "SUPER_CHAT_MESSAGE":
		return parseSuperChat(jsMap(cmd["data"]))
	case "WARNING", "CUT_OFF":
		return parseNotice(name, cmd)
		// NOTICE_MSG is not shown: it carries the site-wide banners about other
		// rooms ("…投喂…，前往TA的房间去围观吧"), not news of this room, whose
		// gifts and guards arrive as SEND_GIFT / GUARD_BUY.
	}
	return nil
}

// counterCmd extracts the numeric counter fields of an event command
// (WATCHED_CHANGE / ONLINE_RANK_COUNT / LIKE_INFO_V3_UPDATE). Returns
// watched, online, likes — 0 when the command doesn't carry them.
func counterCmd(cmd map[string]any) (watched, online, likes int64) {
	name := jsStr(cmd["cmd"])
	if i := strings.IndexByte(name, ':'); i >= 0 {
		name = name[:i]
	}
	d := jsMap(cmd["data"])
	switch name {
	case "WATCHED_CHANGE":
		watched = jsInt(d["num"])
	case "ONLINE_RANK_COUNT":
		online = jsInt(d["count"])
	case "LIKE_INFO_V3_UPDATE":
		likes = jsInt(d["click_count"])
	}
	return
}

func parseDanmaku(info []any, ownUID int64) *app.Message {
	if len(info) < 3 {
		return nil
	}
	modeInfo := jsArr(at(info, 0))
	ts := jsInt(at(modeInfo, 4)) // ms
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

	// info[0][13]: sticker danmaku payload (JSON string) or "{}".
	if s := jsStr(at(modeInfo, 13)); s != "" && s != "{}" {
		var st struct {
			EmoticonUnique string `json:"emoticon_unique"`
			URL            string `json:"url"`
			Width          int    `json:"width"`
			Height         int    `json:"height"`
		}
		if json.Unmarshal([]byte(s), &st) == nil && st.EmoticonUnique != "" {
			m.Emoticon = &app.Emoticon{
				Token:  st.EmoticonUnique,
				Name:   m.Text,
				URL:    st.URL,
				Width:  st.Width,
				Height: st.Height,
			}
		}
	}

	// info[0][15]: user face + extra (sticker / inline emotes).
	if extra := jsMap(at(modeInfo, 15)); extra != nil {
		m.Face = jsStr(jsMap(jsMap(extra["user"])["base"])["face"])
		var extraJSON struct {
			EmoticonUnique string `json:"emoticon_unique"`
			DMType         int    `json:"dm_type"`
			Emots          map[string]struct {
				Emoji          string `json:"emoji"`
				EmoticonUnique string `json:"emoticon_unique"`
				URL            string `json:"url"`
				Width          int    `json:"width"`
				Height         int    `json:"height"`
			} `json:"emots"`
		}
		if s := jsStr(extra["extra"]); s != "" && json.Unmarshal([]byte(s), &extraJSON) == nil {
			if extraJSON.EmoticonUnique != "" && extraJSON.DMType == 1 && m.Emoticon == nil {
				m.Emoticon = &app.Emoticon{
					Token: extraJSON.EmoticonUnique,
					Name:  m.Text,
				}
			}
			if len(extraJSON.Emots) > 0 {
				m.Emots = map[string]app.Emoticon{}
				for k, e := range extraJSON.Emots {
					m.Emots[k] = app.Emoticon{
						Token:  e.EmoticonUnique,
						Name:   e.Emoji,
						URL:    e.URL,
						Width:  e.Width,
						Height: e.Height,
					}
				}
			}
		}
	}
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
		Face:  jsStr(d["face"]),
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
	guardName := jsStr(d["gift_name"])
	if guardName == "" {
		guardName = jsStr(d["role_name"])
	}
	if guardName == "" {
		guardName = map[app.GuardLevel]string{
			app.GuardZongdu:  "总督",
			app.GuardTidu:    "提督",
			app.GuardJianzha: "舰长",
		}[level]
	}
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
		Kind:  app.KindGuard,
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
	var kind app.MessageKind
	switch jsInt(d["msg_type"]) {
	case 1:
		kind = app.KindEnter
	case 2:
		kind = app.KindFollow
	default:
		return nil // share etc. ignored
	}
	ts := jsInt(d["timestamp"])
	t := time.Now()
	if ts > 0 {
		t = time.Unix(ts, 0)
	}
	m := &app.Message{
		ID:   nextID(),
		Kind: kind,
		Time: t,
		UID:  jsInt(d["uid"]),
		Name: jsStr(d["uname"]),
	}
	if fm := jsMap(d["fans_medal"]); fm != nil {
		if n := jsStr(fm["medal_name"]); n != "" {
			m.Medal = &app.Medal{Name: n, Level: int(jsInt(fm["medal_level"]))}
		}
		m.Guard = app.GuardLevel(jsInt(fm["guard_level"]))
	}
	return m
}

func parseLikeClick(d map[string]any) *app.Message {
	if d == nil {
		return nil
	}
	m := &app.Message{
		ID:   nextID(),
		Kind: app.KindLike,
		Time: time.Now(),
		UID:  jsInt(d["uid"]),
		Name: jsStr(d["uname"]),
		Text: jsStr(d["like_text"]),
	}
	if fm := jsMap(d["fans_medal"]); fm != nil {
		if n := jsStr(fm["medal_name"]); n != "" {
			m.Medal = &app.Medal{Name: n, Level: int(jsInt(fm["medal_level"]))}
		}
	}
	return m
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
	ui := jsMap(d["user_info"])
	m := &app.Message{
		ID:    nextID(),
		Kind:  app.KindSuperChat,
		Time:  t,
		UID:   jsInt(d["uid"]),
		Name:  jsStr(ui["uname"]),
		Face:  jsStr(ui["face"]),
		Text:  jsStr(d["message"]),
		Guard: app.GuardLevel(jsInt(ui["guard_level"])),
		SuperChat: &app.SuperChat{
			Price:    int(jsInt(d["price"])),
			Duration: int(jsInt(d["time"])),
			Color:    jsStr(d["background_color"]),
		},
	}
	if mi := jsMap(d["medal_info"]); mi != nil {
		if n := jsStr(mi["medal_name"]); n != "" {
			m.Medal = &app.Medal{Name: n, Level: int(jsInt(mi["medal_level"]))}
		}
	}
	return m
}

func parseNotice(name string, cmd map[string]any) *app.Message {
	text := jsStr(cmd["msg"])
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
