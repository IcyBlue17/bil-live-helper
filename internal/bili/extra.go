package bili

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"github.com/tidwall/gjson"

	"github.com/IcyBlue17/bil-live-helper/app"
)

// ===== 公告 (manage.md 更新直播间公告) =====

func (b *Backend) announcement(ctx context.Context) (string, error) {
	cfg := b.cfg.get()
	if cfg.UID <= 0 {
		return "", errNotLogged
	}
	var d struct {
		RoomNews struct {
			Content string `json:"content"`
		} `json:"room_news"`
	}
	if err := b.cli.getJSON(ctx, b.cli.liveAPI+"/live_user/v1/Master/info",
		url.Values{"uid": {strconv.FormatInt(cfg.UID, 10)}}, b.cli.liveHeaders(), &d); err != nil {
		return "", err
	}
	return d.RoomNews.Content, nil
}

func (b *Backend) setAnnouncement(ctx context.Context, text string) error {
	if len([]rune(text)) > 60 {
		return errors.New("公告最多 60 个字符")
	}
	cfg := b.cfg.get()
	if cfg.RoomID <= 0 || cfg.CSRF == "" {
		return errNotLogged
	}
	form := url.Values{
		"room_id":    {strconv.FormatInt(cfg.RoomID, 10)},
		"uid":        {strconv.FormatInt(cfg.UID, 10)},
		"content":    {text},
		"csrf_token": {cfg.CSRF},
		"csrf":       {cfg.CSRF},
	}
	if err := b.cli.postForm(ctx, b.cli.liveAPI+"/xlive/app-blink/v1/index/updateRoomNews",
		form, b.cli.liveHeaders(), nil); err != nil {
		return err
	}
	app.Logf(app.LogOK, "公告已更新")
	return nil
}

// ===== 榜单 (user.md getOnlineGoldRank / guard.md topListNew) =====

func (b *Backend) rank(ctx context.Context, kind app.RankKind) ([]app.RankUser, error) {
	cfg := b.cfg.get()
	if cfg.RoomID <= 0 {
		return nil, errors.New("房间ID未知，请重新登录")
	}
	ruid := cfg.UID
	if ruid <= 0 {
		// Fall back to the room owner's uid if known.
		if info, err := b.room(ctx); err == nil {
			ruid = info.UID
		}
	}
	switch kind {
	case app.RankGuard:
		return b.guardRank(ctx, cfg.RoomID, ruid)
	default:
		return b.onlineRank(ctx, cfg.RoomID, ruid)
	}
}

func (b *Backend) onlineRank(ctx context.Context, roomID, ruid int64) ([]app.RankUser, error) {
	var raw json.RawMessage
	err := b.cli.getJSON(ctx,
		b.cli.liveAPI+"/xlive/general-interface/v1/rank/getOnlineGoldRank",
		url.Values{
			"roomId":   {strconv.FormatInt(roomID, 10)},
			"ruid":     {strconv.FormatInt(ruid, 10)},
			"page":     {"1"},
			"pageSize": {"50"},
		}, b.cli.liveHeaders(), &raw)
	if err != nil {
		return nil, err
	}
	var out []app.RankUser
	for _, it := range gjson.GetBytes(raw, "OnlineRankItem").Array() {
		u := app.RankUser{
			Rank:  int(it.Get("userRank").Int()),
			UID:   it.Get("uid").Int(),
			Name:  it.Get("name").String(),
			Face:  it.Get("face").String(),
			Score: it.Get("score").Int(),
			Guard: app.GuardLevel(it.Get("guard_level").Int()),
		}
		if m := it.Get("medalInfo"); m.Get("medalName").String() != "" {
			u.Medal = &app.Medal{Name: m.Get("medalName").String(), Level: int(m.Get("level").Int())}
		}
		out = append(out, u)
	}
	return out, nil
}

func (b *Backend) guardRank(ctx context.Context, roomID, ruid int64) ([]app.RankUser, error) {
	var raw json.RawMessage
	err := b.cli.getJSON(ctx,
		b.cli.liveAPI+"/xlive/app-room/v2/guardTab/topListNew",
		url.Values{
			"roomid":    {strconv.FormatInt(roomID, 10)},
			"ruid":      {strconv.FormatInt(ruid, 10)},
			"page":      {"1"},
			"page_size": {"30"},
		}, b.cli.liveHeaders(), &raw)
	if err != nil {
		return nil, err
	}
	var out []app.RankUser
	push := func(it gjson.Result) {
		u := app.RankUser{
			Rank:  int(it.Get("rank").Int()),
			UID:   it.Get("uinfo.uid").Int(),
			Name:  it.Get("uinfo.base.name").String(),
			Face:  it.Get("uinfo.base.face").String(),
			Guard: app.GuardLevel(it.Get("uinfo.guard.level").Int()),
		}
		if u.Guard == 0 {
			u.Guard = app.GuardLevel(it.Get("uinfo.medal.guard_level").Int())
		}
		if m := it.Get("uinfo.medal"); m.Get("name").String() != "" {
			u.Medal = &app.Medal{Name: m.Get("name").String(), Level: int(m.Get("level").Int())}
		}
		out = append(out, u)
	}
	// top3 first, then list — dedupe by uid.
	seen := map[int64]bool{}
	for _, it := range gjson.GetBytes(raw, "top3").Array() {
		if !seen[it.Get("uinfo.uid").Int()] {
			seen[it.Get("uinfo.uid").Int()] = true
			push(it)
		}
	}
	for _, it := range gjson.GetBytes(raw, "list").Array() {
		if !seen[it.Get("uinfo.uid").Int()] {
			seen[it.Get("uinfo.uid").Int()] = true
			push(it)
		}
	}
	return out, nil
}

// ===== 表情 (emoticons.md + danmaku.md msg/send dm_type=1) =====

func (b *Backend) emoticons(ctx context.Context) ([]app.EmoticonPackage, error) {
	cfg := b.cfg.get()
	if cfg.RoomID <= 0 {
		return nil, errors.New("房间ID未知，请重新登录")
	}
	var raw json.RawMessage
	err := b.cli.getJSON(ctx,
		b.cli.liveAPI+"/xlive/web-ucenter/v2/emoticon/GetEmoticons",
		url.Values{
			"platform": {"pc"},
			"room_id":  {strconv.FormatInt(cfg.RoomID, 10)},
		}, b.cli.liveHeaders(), &raw)
	if err != nil {
		return nil, err
	}
	var out []app.EmoticonPackage
	for _, p := range gjson.GetBytes(raw, "data").Array() {
		pkg := app.EmoticonPackage{
			ID:    p.Get("pkg_id").Int(),
			Name:  p.Get("pkg_name").String(),
			Cover: p.Get("current_cover").String(),
		}
		for _, e := range p.Get("emoticons").Array() {
			pkg.Items = append(pkg.Items, app.Emoticon{
				Token:  e.Get("emoticon_unique").String(),
				Name:   e.Get("emoji").String(),
				URL:    e.Get("url").String(),
				Width:  int(e.Get("width").Int()),
				Height: int(e.Get("height").Int()),
			})
		}
		out = append(out, pkg)
	}
	return out, nil
}

func (b *Backend) sendEmoticon(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("表情为空")
	}
	cfg := b.cfg.get()
	if cfg.RoomID <= 0 || cfg.CSRF == "" {
		return errNotLogged
	}
	return b.sendMsg(ctx, url.Values{
		"msg":     {token},
		"dm_type": {"1"},
	})
}

// ===== 观众卡片 / 禁言 =====

func (b *Backend) viewer(ctx context.Context, uid int64) (app.ViewerCard, error) {
	cfg := b.cfg.get()
	ruid := cfg.UID
	if ruid <= 0 {
		if info, err := b.room(ctx); err == nil {
			ruid = info.UID
		}
	}
	var d struct {
		UID       int64  `json:"uid"`
		Uname     string `json:"uname"`
		Face      string `json:"face"`
		Desc      string `json:"desc"`
		FollowNum int64  `json:"follow_num"`
		FansMedal *struct {
			MedalName  string `json:"medal_name"`
			MedalLevel int    `json:"medal_level"`
			GuardLevel int    `json:"guard_level"`
		} `json:"fans_medal"`
		WealthInfo struct {
			Level int `json:"level"`
		} `json:"wealth_info"`
		Guard struct {
			Level int `json:"level"`
		} `json:"guard"`
	}
	err := b.cli.getJSON(ctx, b.cli.liveAPI+"/xlive/app-ucenter/v2/card/user",
		url.Values{
			"uid":  {strconv.FormatInt(uid, 10)},
			"ruid": {strconv.FormatInt(ruid, 10)},
		}, b.cli.liveHeaders(), &d)
	if err != nil {
		return app.ViewerCard{}, err
	}
	card := app.ViewerCard{
		UID:       d.UID,
		Name:      d.Uname,
		Face:      d.Face,
		Sign:      d.Desc,
		Level:     d.WealthInfo.Level,
		Guard:     app.GuardLevel(d.Guard.Level),
		Followers: d.FollowNum,
	}
	if card.Guard == 0 && d.FansMedal != nil {
		card.Guard = app.GuardLevel(d.FansMedal.GuardLevel)
	}
	if d.FansMedal != nil && d.FansMedal.MedalName != "" {
		card.Medal = &app.Medal{Name: d.FansMedal.MedalName, Level: d.FansMedal.MedalLevel}
	}
	// Best-effort muted flag.
	if list, err := b.muted(ctx); err == nil {
		for _, m := range list {
			if m.UID == uid {
				card.Muted = true
				break
			}
		}
	}
	return card, nil
}

func (b *Backend) mute(ctx context.Context, uid int64, hours int) error {
	cfg := b.cfg.get()
	if cfg.RoomID <= 0 || cfg.CSRF == "" {
		return errNotLogged
	}
	form := url.Values{
		"room_id":    {strconv.FormatInt(cfg.RoomID, 10)},
		"tuid":       {strconv.FormatInt(uid, 10)},
		"mobile_app": {"web"},
		"hour":       {strconv.Itoa(hours)},
		"csrf_token": {cfg.CSRF},
		"csrf":       {cfg.CSRF},
		"visit_id":   {""},
	}
	if err := b.cli.postForm(ctx, b.cli.liveAPI+"/xlive/web-ucenter/v1/banned/AddSilentUser",
		form, b.cli.liveHeaders(), nil); err != nil {
		return fmt.Errorf("禁言失败: %w", err)
	}
	app.Logf(app.LogOK, fmt.Sprintf("已禁言用户 %d", uid))
	return nil
}

func (b *Backend) unmute(ctx context.Context, id int64) error {
	cfg := b.cfg.get()
	if cfg.RoomID <= 0 || cfg.CSRF == "" {
		return errNotLogged
	}
	form := url.Values{
		"roomid":     {strconv.FormatInt(cfg.RoomID, 10)},
		"id":         {strconv.FormatInt(id, 10)},
		"csrf_token": {cfg.CSRF},
		"csrf":       {cfg.CSRF},
		"visit_id":   {""},
	}
	if err := b.cli.postForm(ctx, b.cli.liveAPI+"/banned_service/v1/Silent/del_room_block_user",
		form, b.cli.liveHeaders(), nil); err != nil {
		return fmt.Errorf("解除禁言失败: %w", err)
	}
	app.Logf(app.LogOK, "已解除禁言")
	return nil
}

func (b *Backend) muted(ctx context.Context) ([]app.MutedUser, error) {
	cfg := b.cfg.get()
	if cfg.RoomID <= 0 || cfg.CSRF == "" {
		return nil, errNotLogged
	}
	var out []app.MutedUser
	for page := 1; ; page++ {
		var d struct {
			Data []struct {
				TUID  int64  `json:"tuid"`
				TName string `json:"tname"`
				Name  string `json:"name"`
				Ctime string `json:"ctime"`
				ID    int64  `json:"id"`
			} `json:"data"`
			TotalPage int `json:"total_page"`
		}
		err := b.cli.postForm(ctx, b.cli.liveAPI+"/xlive/web-ucenter/v1/banned/GetSilentUserList",
			url.Values{
				"room_id":    {strconv.FormatInt(cfg.RoomID, 10)},
				"ps":         {strconv.Itoa(page)},
				"csrf_token": {cfg.CSRF},
				"csrf":       {cfg.CSRF},
				"visit_id":   {""},
			}, b.cli.liveHeaders(), &d)
		if err != nil {
			return nil, err
		}
		for _, e := range d.Data {
			m := app.MutedUser{ID: e.ID, UID: e.TUID, Name: e.TName, Admin: e.Name}
			out = append(out, m)
		}
		if page >= d.TotalPage || len(d.Data) == 0 {
			break
		}
	}
	return out, nil
}

// ===== 偏好设置 =====

func (b *Backend) preferences() app.Preferences {
	cfg := b.cfg.get()
	if cfg.Prefs != nil {
		return *cfg.Prefs
	}
	return app.Preferences{Notify: true, NotifyGiftMin: 10000}
}

func (b *Backend) savePreferences(p app.Preferences) error {
	if err := b.cfg.update(func(d *configData) { d.Prefs = &p }); err != nil {
		return err
	}
	app.Logf(app.LogOK, "设置已保存")
	return nil
}

// blocked checks text against the preference blocklist (case-insensitive).
func (b *Backend) blocked(text string) bool {
	p := b.preferences()
	lower := strings.ToLower(text)
	for _, w := range p.Blocked {
		if w != "" && strings.Contains(lower, strings.ToLower(w)) {
			return true
		}
	}
	return false
}

// ===== 桌面通知 =====

var (
	notifyMu      sync.Mutex
	notifyWarned  bool
	lastNotify    time.Time
	notifyBackoff = 3 * time.Second
)

// notify shows a desktop notification for SC / 上舰 / 高价礼物.
func (b *Backend) notify(title, body string) {
	if !b.preferences().Notify {
		return
	}
	if !mygo.NotificationsSupported() {
		return
	}
	notifyMu.Lock()
	if time.Since(lastNotify) < notifyBackoff {
		notifyMu.Unlock()
		return
	}
	lastNotify = time.Now()
	notifyMu.Unlock()
	go func() {
		n := mygo.NewNotification(mygo.NotificationOptions{Title: title, Body: body})
		if err := n.ShowContext(context.Background()); err != nil {
			notifyMu.Lock()
			if !notifyWarned {
				notifyWarned = true
				app.Logf(app.LogWarn, "桌面通知不可用："+err.Error())
			}
			notifyMu.Unlock()
		}
	}()
}

// maybeNotify checks a parsed message against notification prefs.
func (b *Backend) maybeNotify(m *app.Message) {
	p := b.preferences()
	if !p.Notify {
		return
	}
	switch m.Kind {
	case app.KindSuperChat:
		price := 0
		if m.SuperChat != nil {
			price = m.SuperChat.Price
		}
		b.notify(fmt.Sprintf("醒目留言 ¥%d", price), m.Name+"："+m.Text)
	case app.KindGuard:
		name := ""
		if m.Gift != nil {
			name = m.Gift.Name
		}
		b.notify("大航海", fmt.Sprintf("%s 开通了%s", m.Name, name))
	case app.KindGift:
		if m.Gift != nil && p.NotifyGiftMin > 0 && m.Gift.Price >= p.NotifyGiftMin {
			b.notify("礼物", fmt.Sprintf("%s 送出 %s ×%d", m.Name, m.Gift.Name, m.Gift.Num))
		}
	}
}
