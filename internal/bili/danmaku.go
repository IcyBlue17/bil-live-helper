package bili

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/coder/websocket"

	"github.com/IcyBlue17/bil-live-helper/app"
)

const heartbeatInterval = 30 * time.Second

var defaultDanmakuHosts = []danmakuHost{
	{Host: "broadcastlv.chat.bilibili.com", Port: 2243, WSSPort: 443, WSPort: 2244},
}

type danmakuHost struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	WSSPort int    `json:"wss_port"`
	WSPort  int    `json:"ws_port"`
}

// buvid returns the stored buvid3 cookie, fetching it from the spi
// frontpage endpoint when missing.
func (b *Backend) buvid(ctx context.Context) string {
	if v := b.cfg.get().Cookies["buvid3"]; v != "" {
		return v
	}
	var d struct {
		B3 string `json:"b_3"`
	}
	if err := b.cli.getJSON(ctx, b.cli.mainAPI+"/x/frontend/finger/spi", nil,
		b.cli.liveHeaders(), &d); err != nil || d.B3 == "" {
		return ""
	}
	_ = b.cfg.update(func(cd *configData) {
		if cd.Cookies == nil {
			cd.Cookies = map[string]string{}
		}
		cd.Cookies["buvid3"] = d.B3
	})
	return d.B3
}

// danmakuInfo fetches the danmaku server list + token (WBI signed).
func (b *Backend) danmakuInfo(ctx context.Context, roomID int64) ([]danmakuHost, string, error) {
	params, err := b.wbi.sign(ctx, url.Values{
		"id":   {strconv.FormatInt(roomID, 10)},
		"type": {"0"},
	})
	if err != nil {
		return nil, "", err
	}
	var d struct {
		HostList []danmakuHost `json:"host_list"`
		Token    string        `json:"token"`
	}
	err = b.cli.getJSON(ctx, b.cli.liveAPI+"/xlive/web-room/v1/index/getDanmuInfo",
		params, b.cli.liveHeaders(), &d)
	if err != nil {
		if ae, ok := err.(*apiError); ok && ae.Code == -352 {
			b.wbi.reset() // WBI key rejected; refresh next time
		}
		return nil, "", err
	}
	if len(d.HostList) == 0 {
		return nil, "", errors.New("弹幕服务器列表为空")
	}
	return d.HostList, d.Token, nil
}

// fetchHistory loads recent danmaku from dM/gethistory.
func (b *Backend) fetchHistory(ctx context.Context, roomID int64, send func(app.FeedEvent) error) error {
	var d struct {
		Room []struct {
			Nickname string `json:"nickname"`
			Text     string `json:"text"`
			UID      int64  `json:"uid"`
			Timeline string `json:"timeline"`
		} `json:"room"`
	}
	err := b.cli.getJSON(ctx, b.cli.liveAPI+"/xlive/web-room/v1/dM/gethistory",
		url.Values{"roomid": {strconv.FormatInt(roomID, 10)}}, b.cli.liveHeaders(), &d)
	if err != nil {
		return nil // history is best-effort
	}
	uid := b.cfg.get().UID
	for _, e := range d.Room {
		m := app.Message{
			ID:         nextID(),
			Kind:       app.KindDanmaku,
			Time:       time.Now(),
			UID:        e.UID,
			Name:       e.Nickname,
			Text:       e.Text,
			Historical: true,
		}
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", e.Timeline, shanghai); err == nil {
			m.Time = t
		}
		if uid > 0 && e.UID == uid {
			m.Kind = app.KindSelf
		}
		if err := send(app.FeedEvent{Message: &m}); err != nil {
			return err
		}
	}
	return nil
}

// watch runs the danmaku feed: history, then the websocket with
// reconnect/failover, until ctx is done or send fails.
func (b *Backend) watch(ctx context.Context, send func(app.FeedEvent) error) error {
	cfg := b.cfg.get()
	roomID := cfg.RoomID
	if roomID <= 0 {
		return errors.New("房间ID未知，请重新登录")
	}

	sendState := func(s app.ConnState) error {
		return send(app.FeedEvent{State: &s})
	}
	if err := sendState(app.ConnConnecting); err != nil {
		return err
	}
	if err := b.fetchHistory(ctx, roomID, send); err != nil {
		return err
	}

	var hosts []danmakuHost
	var token string
	retry := 0
	backoff := time.Second
	first := true
	stats := newStatsTracker()

	for {
		if ctx.Err() != nil {
			return nil
		}
		// (Re)initialize the host list when needed.
		if len(hosts) == 0 {
			var err error
			hosts, token, err = b.danmakuInfo(ctx, roomID)
			if err != nil || len(hosts) == 0 {
				hosts = defaultDanmakuHosts
				token = ""
			}
		}
		host := hosts[retry%len(hosts)]
		wsURL := fmt.Sprintf("wss://%s:%d/sub", host.Host, host.WSSPort)

		err := b.danmakuSession(ctx, wsURL, roomID, token, first, send, sendState, stats)
		first = false
		if ctx.Err() != nil {
			_ = sendState(app.ConnClosed)
			return nil
		}
		if errors.Is(err, errSendFailed) {
			return nil
		}
		if err := sendState(app.ConnReconnecting); err != nil {
			return err
		}
		retry++
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff *= 2
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
		}
	}
}

var errSendFailed = errors.New("bili: feed send failed")

// danmakuSession runs one websocket connection until it breaks.
func (b *Backend) danmakuSession(ctx context.Context, wsURL string, roomID int64, token string, first bool, send func(app.FeedEvent) error, sendState func(app.ConnState) error, stats *statsTracker) error {
	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"User-Agent": {userAgent}},
	})
	if err != nil {
		return err
	}
	defer conn.Close(websocket.StatusInternalError, "")
	conn.SetReadLimit(4 << 20)

	cfg := b.cfg.get()
	auth := map[string]any{
		"uid":      cfg.UID,
		"roomid":   roomID,
		"protover": 3,
		"platform": "web",
		"type":     2,
		"buvid":    b.buvid(ctx),
	}
	if token != "" {
		auth["key"] = token
	}
	if err := conn.Write(ctx, websocket.MessageBinary, makePacket(auth, opAuth)); err != nil {
		return err
	}

	readErr := make(chan error, 1)
	msgs := make(chan []byte, 32)
	go func() {
		defer close(msgs)
		for {
			_, data, err := conn.Read(ctx)
			if err != nil {
				readErr <- err
				return
			}
			msgs <- data
		}
	}()

	heart := time.NewTicker(heartbeatInterval)
	defer heart.Stop()
	// First heartbeat right away (upstream sends one on auth reply too).
	sendHeartbeat := func() {
		_ = conn.Write(context.Background(), websocket.MessageBinary, makePacket("", opHeartbeat))
	}
	defer sendHeartbeat()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-readErr:
			return err
		case <-heart.C:
			if err := conn.Write(ctx, websocket.MessageBinary, makePacket("", opHeartbeat)); err != nil {
				return err
			}
		case data, ok := <-msgs:
			if !ok {
				return errors.New("bili: websocket closed")
			}
			if err := b.handleFrame(data, first, send, sendState, stats); err != nil {
				return err
			}
			first = false
		}
	}
}

// handleFrame decodes one websocket frame and forwards events.
func (b *Backend) handleFrame(data []byte, first bool, send func(app.FeedEvent) error, sendState func(app.ConnState) error, stats *statsTracker) error {
	uid := b.cfg.get().UID
	connectedSent := false
	return parsePackets(data, func(h header, body []byte) error {
		switch h.op {
		case opAuthReply:
			var r struct {
				Code int `json:"code"`
			}
			if err := json.Unmarshal(body, &r); err == nil && r.Code != 0 {
				return fmt.Errorf("弹幕认证失败 code=%d", r.Code)
			}
			if !connectedSent {
				connectedSent = true
				if err := sendState(app.ConnConnected); err != nil {
					return errSendFailed
				}
				sys := &app.Message{
					ID:   nextID(),
					Kind: app.KindSystem,
					Time: time.Now(),
					Text: fmt.Sprintf("已连接到直播间 %d", b.cfg.get().RoomID),
				}
				if err := send(app.FeedEvent{Message: sys}); err != nil {
					return errSendFailed
				}
			}
		case opHeartbeatReply:
			if err := send(app.FeedEvent{Popularity: parseHeartbeatReply(body)}); err != nil {
				return errSendFailed
			}
		case opSendMsgReply:
			var cmd map[string]any
			if err := json.Unmarshal(body, &cmd); err != nil {
				return nil // ignore undecodable commands
			}
			ev := app.FeedEvent{}
			watched, online, likes := counterCmd(cmd)
			ev.Watched, ev.Online, ev.Likes = watched, online, likes
			if online > 0 {
				stats.online(online, time.Now())
			}
			if m := parseCommand(cmd, uid); m != nil {
				if m.Kind == app.KindDanmaku && b.blocked(m.Text) {
					// Still update counters, but hide the message.
				} else {
					stats.record(m)
					ev.Message = m
					b.maybeNotify(m)
				}
			}
			ev.Stats = stats.snapshot(time.Now())
			if ev.Message == nil && ev.Watched == 0 && ev.Online == 0 && ev.Likes == 0 && ev.Stats == nil {
				return nil
			}
			if err := send(ev); err != nil {
				return errSendFailed
			}
		}
		return nil
	})
}

// sendDanmaku posts msg/send like danmaku_panel.send_message.
func (b *Backend) sendDanmaku(ctx context.Context, text string) error {
	if n := len([]rune(text)); n == 0 {
		return errors.New("弹幕不能为空")
	} else if n > 40 {
		return fmt.Errorf("弹幕过长（%d/40 字）", n)
	}
	return b.sendMsg(ctx, url.Values{"msg": {text}})
}

// sendMsg POSTs msg/send; extra carries msg/dm_type overrides.
func (b *Backend) sendMsg(ctx context.Context, extra url.Values) error {
	cfg := b.cfg.get()
	if cfg.RoomID <= 0 || cfg.CSRF == "" {
		return errNotLogged
	}
	params, err := b.wbi.sign(ctx, url.Values{"web_location": {"444.8"}})
	if err != nil {
		params = url.Values{"web_location": {"444.8"}}
	}
	form := url.Values{
		"bubble":     {"0"},
		"color":      {"16777215"},
		"mode":       {"1"},
		"fontsize":   {"25"},
		"rnd":        {strconv.FormatInt(time.Now().Unix(), 10)},
		"roomid":     {strconv.FormatInt(cfg.RoomID, 10)},
		"csrf":       {cfg.CSRF},
		"csrf_token": {cfg.CSRF},
	}
	for k, vs := range extra {
		form[k] = append([]string(nil), vs...)
	}
	u := b.cli.liveAPI + "/msg/send?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	b.cli.applyHeaders(req, b.cli.liveHeaders(), true)
	var r reply
	if _, err := b.cli.do(req, &r); err != nil {
		return err
	}
	if r.Code != 0 {
		msg := r.Message
		if msg == "" {
			msg = r.Msg
		}
		return fmt.Errorf("发送弹幕失败：%s（%d）", danmakuErrMsg(r.Code, msg), r.Code)
	}
	return nil
}

// danmakuErrMsg maps known send errors to friendlier Chinese text.
func danmakuErrMsg(code int64, msg string) string {
	switch code {
	case 10031:
		return "发送太频繁了，请稍后再试"
	case 10030:
		return "弹幕包含被屏蔽的内容"
	case -101:
		return "未登录或登录已过期"
	case -111:
		return "CSRF 校验失败，请重新登录"
	case 1999015:
		return "弹幕包含违禁词"
	}
	if msg != "" {
		return msg
	}
	return "发送失败"
}
