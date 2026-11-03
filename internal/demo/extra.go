package demo

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"time"

	"github.com/IcyBlue17/bil-live-helper/app"
)

var (
	extraMu      sync.Mutex
	announcement = "每晚 22:00 开播，点歌请发「点歌 + 歌名」~"
	prefs        = app.Preferences{Notify: true, NotifyGiftMin: 10000, Blocked: []string{"广告"}}
	muted        = []app.MutedUser{{ID: 1, UID: 88001, Name: "刷屏小号", Admin: "迷子不迷路"}}
)

func (b *Backend) Announcement(ctx context.Context) (string, error) {
	extraMu.Lock()
	defer extraMu.Unlock()
	return announcement, nil
}

func (b *Backend) SetAnnouncement(ctx context.Context, text string) error {
	if len([]rune(text)) > 60 {
		return errors.New("公告最多 60 个字符")
	}
	extraMu.Lock()
	announcement = text
	extraMu.Unlock()
	app.Logf(app.LogOK, "公告已更新")
	return nil
}

func (b *Backend) Rank(ctx context.Context, kind app.RankKind) ([]app.RankUser, error) {
	var rows []app.RankUser
	for i, n := range names {
		r := app.RankUser{Rank: i + 1, UID: int64(1000 + i), Name: n, Medal: &app.Medal{Name: "迷子", Level: 25 - i*2}}
		if kind == app.RankGuard {
			r.Guard = app.GuardJianzha
			if i == 0 {
				r.Guard = app.GuardTidu
			}
		} else {
			r.Score = int64(5200 - i*480)
			if i < 2 {
				r.Guard = app.GuardJianzha
			}
		}
		rows = append(rows, r)
	}
	return rows, nil
}

func emoteURL(text string, bg string) string {
	// a self-contained SVG data URL, so screenshots need no network
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><rect width="64" height="64" rx="14" fill="%s"/><text x="32" y="40" font-size="24" text-anchor="middle" fill="#fff" font-family="sans-serif">%s</text></svg>`, bg, text)
	return "data:image/svg+xml," + strings.NewReplacer("#", "%23", "<", "%3C", ">", "%3E", `"`, "'").Replace(svg)
}

var emoticons = []app.EmoticonPackage{
	{ID: 1, Name: "通用表情", Items: []app.Emoticon{
		{Token: "[dog]", Name: "dog", URL: emoteURL("🐶", "#f59e0b"), Width: 20, Height: 20},
		{Token: "[花]", Name: "花", URL: emoteURL("🌸", "#fb7299"), Width: 20, Height: 20},
		{Token: "[妙]", Name: "妙", URL: emoteURL("妙", "#8b5cf6"), Width: 20, Height: 20},
		{Token: "[哇]", Name: "哇", URL: emoteURL("哇", "#00aeec"), Width: 20, Height: 20},
	}},
	{ID: 2, Name: "房间专属", Items: []app.Emoticon{
		{Token: "room_22637261_1", Name: "晚安", URL: emoteURL("晚安", "#334155"), Width: 120, Height: 120},
		{Token: "room_22637261_2", Name: "好耶", URL: emoteURL("好耶", "#16a34a"), Width: 120, Height: 120},
		{Token: "room_22637261_3", Name: "比心", URL: emoteURL("比心", "#e11d48"), Width: 120, Height: 120},
	}},
}

func (b *Backend) Emoticons(ctx context.Context) ([]app.EmoticonPackage, error) {
	return emoticons, nil
}

func (b *Backend) SendEmoticon(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("表情为空")
	}
	return nil
}

func (b *Backend) Viewer(ctx context.Context, uid int64) (app.ViewerCard, error) {
	i := int(uid) % len(names)
	if i < 0 {
		i = 0
	}
	extraMu.Lock()
	defer extraMu.Unlock()
	isMuted := false
	for _, m := range muted {
		if m.UID == uid {
			isMuted = true
		}
	}
	return app.ViewerCard{
		UID: uid, Name: names[i], Sign: "今天也要好好听歌", Level: 4 + i%3,
		Guard: app.GuardLevel([]int{0, 3, 0, 2}[i%4]), Medal: &app.Medal{Name: "迷子", Level: 3 + i*2},
		Followers: int64(120 + i*37), Muted: isMuted,
	}, nil
}

func (b *Backend) Mute(ctx context.Context, uid int64, hours int) error {
	card, _ := b.Viewer(ctx, uid)
	extraMu.Lock()
	defer extraMu.Unlock()
	m := app.MutedUser{ID: int64(len(muted) + 1), UID: uid, Name: card.Name, Admin: user.Name}
	if hours > 0 {
		t := time.Now().Add(time.Duration(hours) * time.Hour)
		m.Until = &t
	}
	muted = append(muted, m)
	app.Logf(app.LogOK, "已禁言 "+card.Name)
	return nil
}

func (b *Backend) Unmute(ctx context.Context, id int64) error {
	extraMu.Lock()
	defer extraMu.Unlock()
	for i, m := range muted {
		if m.ID == id {
			muted = append(muted[:i], muted[i+1:]...)
			app.Logf(app.LogOK, "已解除禁言 "+m.Name)
			return nil
		}
	}
	return errors.New("未找到该禁言记录")
}

func (b *Backend) Muted(ctx context.Context) ([]app.MutedUser, error) {
	extraMu.Lock()
	defer extraMu.Unlock()
	return append([]app.MutedUser(nil), muted...), nil
}

func (b *Backend) Preferences() (app.Preferences, error) {
	extraMu.Lock()
	defer extraMu.Unlock()
	return prefs, nil
}

func (b *Backend) SavePreferences(p app.Preferences) error {
	extraMu.Lock()
	prefs = p
	extraMu.Unlock()
	app.Logf(app.LogOK, "设置已保存")
	return nil
}

// fakeStats builds plausible session totals for the demo feed.
func fakeStats(since time.Time, n int) *app.SessionStats {
	s := &app.SessionStats{
		Since: since, Danmaku: 312 + n, Senders: 87 + n/3, Enters: 540 + n*2, Follows: 23 + n/10,
		Likes: 1880 + n*5, Gifts: 64 + n/5, GiftValue: 482000 + int64(n)*1200, SuperChats: 3, Guards: 2,
	}
	online := int64(900)
	for t := since; t.Before(time.Now()); t = t.Add(time.Minute) {
		online += int64(rand.IntN(80) - 30)
		if online < 100 {
			online = 100
		}
		s.Samples = append(s.Samples, app.OnlineSample{Time: t, Online: online})
		s.PeakOnline = max(s.PeakOnline, online)
	}
	if len(s.Samples) > 240 {
		s.Samples = s.Samples[len(s.Samples)-240:]
	}
	return s
}
