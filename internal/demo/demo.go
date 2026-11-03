// Package demo is a Backend with fake data, for developing the frontend and
// taking screenshots without a Bilibili account. Run with BLH_DEMO=1.
//
// BLH_DEMO_STATE picks where it starts: "unauth" (default "idle") or "live".
package demo

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"sync"
	"time"

	"github.com/IcyBlue17/bil-live-helper/app"
)

// Backend is the fake backend.
type Backend struct {
	mu       sync.Mutex
	loggedIn bool
	room     app.RoomInfo
	settings app.RoomSettings
	stream   *app.StreamInfo
}

var areas = []app.AreaGroup{
	{ID: 9, Name: "虚拟主播", Areas: []app.Area{{371, "虚拟日常"}, {697, "虚拟声优"}, {745, "虚拟gamer"}}},
	{ID: 6, Name: "单机游戏", Areas: []app.Area{{235, "其他单机"}, {236, "主机游戏"}, {216, "我的世界"}}},
	{ID: 2, Name: "网游", Areas: []app.Area{{86, "英雄联盟"}, {87, "守望先锋"}, {80, "绝地求生"}}},
	{ID: 1, Name: "娱乐", Areas: []app.Area{{21, "视频唱见"}, {207, "舞见"}, {145, "视频聊天"}}},
	{ID: 11, Name: "知识", Areas: []app.Area{{376, "人文社科"}, {372, "科技科普"}, {377, "职场·技能"}}},
}

var user = app.User{UID: 3493118, Name: "迷子不迷路", Face: ""}

// New returns a fake backend.
func New() *Backend {
	b := &Backend{
		loggedIn: os.Getenv("BLH_DEMO_STATE") != "unauth",
		settings: app.RoomSettings{Title: "深夜摇滚练习室 · 一起来听歌", AreaID: 21},
		room: app.RoomInfo{
			RoomID: 22637261, UID: user.UID, Title: "深夜摇滚练习室 · 一起来听歌",
			Description: "每晚十点不见不散", AreaID: 21, AreaName: "视频唱见",
			ParentAreaID: 1, ParentAreaName: "娱乐", Status: app.LiveOffline, Attention: 12873,
		},
	}
	if os.Getenv("BLH_DEMO_STATE") == "live" {
		b.goLive(time.Now().Add(-83 * time.Minute))
	}
	return b
}

func (b *Backend) goLive(since time.Time) {
	b.room.Status = app.LiveOn
	b.room.LiveSince = &since
	b.room.Online = 1342
	b.stream = &app.StreamInfo{
		Addr: "rtmp://live-push.bilivideo.com/live-bvc/",
		Code: "?streamname=live_3493118_7716342&key=8f2c0e1b9a7d4c35b6e1f0a2d3c4b5e6&schedule=rtmp&pflag=1",
	}
}

func (b *Backend) Auth(ctx context.Context) (app.AuthState, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.loggedIn {
		return app.AuthState{}, nil
	}
	u := user
	return app.AuthState{LoggedIn: true, User: &u}, nil
}

func (b *Backend) Login(ctx context.Context, update func(app.QRUpdate) error) error {
	code := &app.QRCode{URL: "https://passport.bilibili.com/h5-app/passport/login/scan?qrcode_key=demo", SVG: fakeQR}
	steps := []app.QRUpdate{
		{Stage: app.QRWaiting, Code: code, Message: "请使用哔哩哔哩 App 扫码"},
		{Stage: app.QRScanned, Message: "已扫码，请在手机上确认"},
		{Stage: app.QRDone, Message: "登录成功"},
	}
	delay := 4 * time.Second
	if os.Getenv("BLH_DEMO_HOLD") == "1" { // keep the QR on screen for screenshots
		delay = time.Hour
	}
	for i, s := range steps {
		if i > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
		if err := update(s); err != nil {
			return err
		}
	}
	b.mu.Lock()
	b.loggedIn = true
	b.mu.Unlock()
	u := user
	_ = app.AuthChanged.Broadcast(app.AuthState{LoggedIn: true, User: &u})
	app.Logf(app.LogOK, "登录成功："+user.Name)
	return nil
}

func (b *Backend) Logout() error {
	b.mu.Lock()
	b.loggedIn = false
	b.mu.Unlock()
	_ = app.AuthChanged.Broadcast(app.AuthState{})
	app.Logf(app.LogInfo, "已退出登录")
	return nil
}

func (b *Backend) Room(ctx context.Context) (app.RoomInfo, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.room.Status == app.LiveOn {
		b.room.Online += int64(rand.IntN(40) - 15)
	}
	return b.room, nil
}

func (b *Backend) Areas(ctx context.Context) ([]app.AreaGroup, error) { return areas, nil }

func (b *Backend) Settings() (app.RoomSettings, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.settings, nil
}

func (b *Backend) SaveSettings(ctx context.Context, s app.RoomSettings) (app.RoomInfo, error) {
	if n := len([]rune(s.Title)); n == 0 || n > 40 {
		return app.RoomInfo{}, errors.New("标题需为 1–40 个字符")
	}
	b.mu.Lock()
	b.settings = s
	b.room.Title = s.Title
	for _, g := range areas {
		for _, a := range g.Areas {
			if a.ID == s.AreaID {
				b.room.AreaID, b.room.AreaName = a.ID, a.Name
				b.room.ParentAreaID, b.room.ParentAreaName = g.ID, g.Name
			}
		}
	}
	room := b.room
	b.mu.Unlock()
	_ = app.RoomChanged.Broadcast(room)
	app.Logf(app.LogOK, "直播间信息已更新")
	return room, nil
}

func (b *Backend) Start(ctx context.Context, update func(app.StartUpdate) error) (app.StreamInfo, error) {
	if err := update(app.StartUpdate{Stage: app.StartRequesting, Message: "正在开播…"}); err != nil {
		return app.StreamInfo{}, err
	}
	time.Sleep(900 * time.Millisecond)
	if os.Getenv("BLH_DEMO_FACE") == "1" {
		face := &app.QRUpdate{Stage: app.QRWaiting, Code: &app.QRCode{URL: "https://www.bilibili.com/blackboard/live/face-auth-middle.html", SVG: fakeQR}, Message: "请使用哔哩哔哩 App 扫码完成人脸认证"}
		if err := update(app.StartUpdate{Stage: app.StartFaceAuth, Face: face, Message: "需要人脸认证"}); err != nil {
			return app.StreamInfo{}, err
		}
		<-ctx.Done()
		return app.StreamInfo{}, ctx.Err()
	}
	b.mu.Lock()
	b.goLive(time.Now())
	s, room := *b.stream, b.room
	b.mu.Unlock()
	_ = app.RoomChanged.Broadcast(room)
	app.Logf(app.LogOK, "开播成功")
	return s, update(app.StartUpdate{Stage: app.StartLive, Stream: &s, Message: "开播成功"})
}

func (b *Backend) Stop(ctx context.Context) error {
	b.mu.Lock()
	b.room.Status, b.room.LiveSince, b.room.Online = app.LiveOffline, nil, 0
	room := b.room
	b.mu.Unlock()
	_ = app.RoomChanged.Broadcast(room)
	app.Logf(app.LogOK, "下播成功")
	return nil
}

func (b *Backend) Stream() (*app.StreamInfo, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.stream, nil
}

var (
	names = []string{"晚风", "橘子汽水", "不睡觉的猫", "Lumi", "海盐芝士", "星野", "阿卡林", "小蘑菇", "三分糖", "Kaze"}
	lines = []string{"晚上好！", "这首好好听", "主播今天状态好棒", "来了来了", "求一首《春日影》", "哈哈哈哈哈", "弹幕护体", "打卡~", "声音好温柔", "下一首唱什么"}
	gifts = []app.Gift{{"小心心", 1, 0}, {"辣条", 10, 1000}, {"牛哇牛哇", 1, 100}, {"舰长", 1, 198000}}
)

func (b *Backend) Watch(ctx context.Context, send func(app.FeedEvent) error) error {
	st := func(s app.ConnState) error { return send(app.FeedEvent{State: &s}) }
	if err := st(app.ConnConnecting); err != nil {
		return err
	}
	now := time.Now()
	for i := range 8 {
		m := fakeMessage(i, now.Add(time.Duration(i-8)*37*time.Second))
		m.Historical = true
		if err := send(app.FeedEvent{Message: &m}); err != nil {
			return err
		}
	}
	if err := st(app.ConnConnected); err != nil {
		return err
	}
	since := time.Now().Add(-83 * time.Minute)
	for i := 8; ; i++ {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Duration(700+rand.IntN(1600)) * time.Millisecond):
		}
		m := fakeMessage(i, time.Now())
		ev := app.FeedEvent{Message: &m, Popularity: 18000 + int64(rand.IntN(900)), Online: 1300 + int64(rand.IntN(90)), Watched: 9421 + int64(i), Likes: 25310 + int64(i*3)}
		if i%3 == 0 {
			ev.Stats = fakeStats(since, i)
		}
		if err := send(ev); err != nil {
			return nil
		}
	}
}

func fakeMessage(i int, t time.Time) app.Message {
	n := names[i%len(names)]
	m := app.Message{ID: fmt.Sprint(i), Kind: app.KindDanmaku, Time: t, UID: int64(1000 + i%len(names)), Name: n, Text: lines[(i*7)%len(lines)], UserLevel: 5 + i%30}
	switch {
	case i%17 == 3:
		m.Kind, m.Text = app.KindSuperChat, "主播唱得太好了！下次可以唱《春日影》吗"
		m.SuperChat = &app.SuperChat{Price: 30, Duration: 60, Color: "#EDF5FF"}
		m.Medal = &app.Medal{Name: "迷子", Level: 12}
	case i%19 == 7:
		g := app.Gift{Name: "舰长", Num: 1, Price: 198000}
		m.Kind, m.Gift, m.Guard, m.Text = app.KindGuard, &g, app.GuardJianzha, ""
	case i%7 == 6:
		m.Kind, m.Text = app.KindFollow, ""
	case i%10 == 9:
		m.Kind, m.Text = app.KindLike, ""
	case i%15 == 2:
		e := emoticons[1].Items[i%3]
		m.Emoticon, m.Text = &e, e.Name
	case i%6 == 2:
		m.Text = "哈哈哈[dog][花]"
		m.Emots = map[string]app.Emoticon{"[dog]": emoticons[0].Items[0], "[花]": emoticons[0].Items[1]}
	case i%11 == 5:
		g := gifts[i%len(gifts)]
		m.Kind, m.Gift, m.Text = app.KindGift, &g, ""
		if g.Name == "舰长" {
			m.Guard = app.GuardJianzha
		}
	case i%9 == 4:
		m.Kind, m.Text = app.KindEnter, ""
	case i%4 == 1:
		m.Medal = &app.Medal{Name: "迷子", Level: 3 + i%20}
		m.Guard = app.GuardLevel([]int{0, 3, 0, 2}[i%4])
	case i%13 == 0:
		m.Admin = true
	}
	return m
}

func (b *Backend) SendDanmaku(ctx context.Context, text string) error {
	if text == "" {
		return errors.New("弹幕不能为空")
	}
	return nil
}

func (b *Backend) ConfigPath() string { return "(demo)" }
