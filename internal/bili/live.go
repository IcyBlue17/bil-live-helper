package bili

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/IcyBlue17/bil-live-helper/app"
)

var shanghai = time.FixedZone("Asia/Shanghai", 8*3600)

// room fetches room/v1/Room/get_info and maps it to app.RoomInfo.
func (b *Backend) room(ctx context.Context) (app.RoomInfo, error) {
	cfg := b.cfg.get()
	if cfg.RoomID <= 0 {
		return app.RoomInfo{}, errors.New("房间ID未知，请重新登录")
	}
	var d struct {
		RoomID         int64  `json:"room_id"`
		UID            int64  `json:"uid"`
		Title          string `json:"title"`
		Description    string `json:"description"`
		Cover          string `json:"cover"`
		AreaID         int64  `json:"area_id"`
		AreaName       string `json:"area_name"`
		ParentAreaID   int64  `json:"parent_area_id"`
		ParentAreaName string `json:"parent_area_name"`
		LiveStatus     int    `json:"live_status"`
		Online         int64  `json:"online"`
		Attention      int64  `json:"attention"`
		LiveTime       string `json:"live_time"`
	}
	err := b.cli.postForm(ctx, b.cli.liveAPI+"/room/v1/Room/get_info",
		url.Values{"room_id": {strconv.FormatInt(cfg.RoomID, 10)}}, b.cli.liveHeaders(), &d)
	if err != nil {
		return app.RoomInfo{}, err
	}
	info := app.RoomInfo{
		RoomID:         d.RoomID,
		UID:            d.UID,
		Title:          d.Title,
		Description:    strings.NewReplacer("<p>", "", "</p>", "").Replace(d.Description),
		Cover:          d.Cover,
		AreaID:         d.AreaID,
		AreaName:       d.AreaName,
		ParentAreaID:   d.ParentAreaID,
		ParentAreaName: d.ParentAreaName,
		Online:         d.Online,
		Attention:      d.Attention,
	}
	switch d.LiveStatus {
	case 1:
		info.Status = app.LiveOn
	case 2:
		info.Status = app.LiveRound
	default:
		info.Status = app.LiveOffline
	}
	if info.Status == app.LiveOn && d.LiveTime != "" && d.LiveTime != "0000-00-00 00:00:00" {
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", d.LiveTime, shanghai); err == nil {
			info.LiveSince = &t
		}
	}
	// Persist the current title/area for the next broadcast.
	_ = b.cfg.update(func(cd *configData) {
		cd.Title = info.Title
		cd.AreaID = info.AreaID
	})
	b.watchLive(info)
	return info, nil
}

// areas fetches room/v1/Area/getList, with in-memory + config caching.
func (b *Backend) areas(ctx context.Context) ([]app.AreaGroup, error) {
	b.mu.Lock()
	if b.areaCache != nil {
		defer b.mu.Unlock()
		return b.areaCache, nil
	}
	b.mu.Unlock()

	type rawGroup struct {
		ID   json.Number `json:"id"`
		Name string      `json:"name"`
		List []struct {
			ID   json.Number `json:"id"`
			Name string      `json:"name"`
		} `json:"list"`
	}
	var raw []rawGroup
	if err := b.cli.getJSON(ctx, b.cli.liveAPI+"/room/v1/Area/getList", nil, b.cli.liveHeaders(), &raw); err != nil {
		// Fall back to the stored list.
		if stored := b.cfg.get().Areas; len(stored) > 0 {
			return stored, nil
		}
		return nil, err
	}
	groups := make([]app.AreaGroup, 0, len(raw))
	for _, g := range raw {
		gid, _ := g.ID.Int64()
		ag := app.AreaGroup{ID: gid, Name: g.Name}
		for _, a := range g.List {
			id, _ := a.ID.Int64()
			ag.Areas = append(ag.Areas, app.Area{ID: id, Name: a.Name})
		}
		groups = append(groups, ag)
	}
	b.mu.Lock()
	b.areaCache = groups
	b.mu.Unlock()
	_ = b.cfg.update(func(d *configData) { d.Areas = groups })
	return groups, nil
}

func (b *Backend) areaExists(areaID int64) bool {
	for _, g := range b.cfg.get().Areas {
		for _, a := range g.Areas {
			if a.ID == areaID {
				return true
			}
		}
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, g := range b.areaCache {
		for _, a := range g.Areas {
			if a.ID == areaID {
				return true
			}
		}
	}
	return false
}

// saveSettings validates and applies the title/area via Room/update.
func (b *Backend) saveSettings(ctx context.Context, s app.RoomSettings) (app.RoomInfo, error) {
	if n := len([]rune(s.Title)); n == 0 || n > 40 {
		return app.RoomInfo{}, errors.New("标题需为 1–40 个字符")
	}
	if s.AreaID <= 0 || !b.areaExists(s.AreaID) {
		return app.RoomInfo{}, errors.New("无效的分区ID")
	}
	cfg := b.cfg.get()
	if cfg.RoomID <= 0 || cfg.CSRF == "" {
		return app.RoomInfo{}, errNotLogged
	}
	form := url.Values{
		"room_id":     {strconv.FormatInt(cfg.RoomID, 10)},
		"platform":    {"pc_link"},
		"activity_id": {"0"},
		"csrf_token":  {cfg.CSRF},
		"csrf":        {cfg.CSRF},
		"title":       {s.Title},
		"area_id":     {strconv.FormatInt(s.AreaID, 10)},
	}
	if err := b.cli.postForm(ctx, b.cli.liveAPI+"/room/v1/Room/update", sign(form), b.cli.liveHeaders(), nil); err != nil {
		return app.RoomInfo{}, err
	}
	_ = b.cfg.update(func(d *configData) {
		d.Title = s.Title
		d.AreaID = s.AreaID
	})
	app.Logf(app.LogOK, "直播间信息已更新")
	info, err := b.room(ctx)
	if err != nil {
		return app.RoomInfo{}, err
	}
	_ = app.RoomChanged.Broadcast(info)
	return info, nil
}

// updateLiveVersion refreshes the livehime version used by startLive.
func (b *Backend) updateLiveVersion(ctx context.Context) {
	var d struct {
		CurrVersion string      `json:"curr_version"`
		Build       json.Number `json:"build"`
	}
	params := sign(url.Values{"system_version": {"2"}})
	if err := b.cli.getJSON(ctx, b.cli.liveAPI+"/xlive/app-blink/v1/liveVersionInfo/getHomePageLiveVersion",
		params, b.cli.liveHeaders(), &d); err != nil {
		return // keep the constants fallback
	}
	_ = b.cfg.update(func(cd *configData) {
		if d.CurrVersion != "" {
			cd.LiveVersion = d.CurrVersion
		}
		if d.Build.String() != "" {
			cd.LiveBuild = d.Build.String()
		}
	})
}

// startLiveOnce issues the startLive call; returns the stream info or the
// face-auth QR URL.
func (b *Backend) startLiveOnce(ctx context.Context) (addr, code, faceQR string, err error) {
	cfg := b.cfg.get()
	if cfg.RoomID <= 0 || cfg.CSRF == "" {
		return "", "", "", errNotLogged
	}
	if cfg.AreaID <= 0 {
		return "", "", "", errors.New("请先选择直播分区")
	}
	form := url.Values{
		"room_id":    {strconv.FormatInt(cfg.RoomID, 10)},
		"platform":   {"pc_link"},
		"area_v2":    {strconv.FormatInt(cfg.AreaID, 10)},
		"csrf_token": {cfg.CSRF},
		"csrf":       {cfg.CSRF},
		"type":       {"2"},
		"build":      {cfg.LiveBuild},
		"version":    {cfg.LiveVersion},
	}
	r, err := b.cli.postFormRaw(ctx, b.cli.liveAPI+"/room/v1/Room/startLive", sign(form), b.cli.liveHeaders())
	if err != nil {
		return "", "", "", err
	}
	switch r.Code {
	case 60024:
		var d struct {
			QR string `json:"qr"`
		}
		_ = json.Unmarshal(r.Data, &d)
		return "", "", d.QR, nil
	case 60043:
		return "", "", fmt.Sprintf("%s/blackboard/live/face-auth-middle.html?source_event=400&mid=%d", b.cli.wwwURL, cfg.UID), nil
	case 0:
		var d struct {
			RTMP struct {
				Addr string `json:"addr"`
				Code string `json:"code"`
			} `json:"rtmp"`
		}
		if err := json.Unmarshal(r.Data, &d); err != nil {
			return "", "", "", err
		}
		if d.RTMP.Addr == "" || d.RTMP.Code == "" {
			return "", "", "", errors.New("开播成功但未返回推流地址")
		}
		return d.RTMP.Addr, d.RTMP.Code, "", nil
	default:
		msg := r.Msg
		if msg == "" {
			msg = r.Message
		}
		return "", "", "", &apiError{Code: r.Code, Msg: msg}
	}
}

// checkFaceAuth polls IsUserIdentifiedByFaceAuth (≤5 min) like live.py.
func (b *Backend) checkFaceAuth(ctx context.Context) bool {
	cfg := b.cfg.get()
	form := url.Values{
		"room_id":        {strconv.FormatInt(cfg.RoomID, 10)},
		"face_auth_code": {"60024"},
		"csrf_token":     {cfg.CSRF},
		"csrf":           {cfg.CSRF},
		"visit_id":       {""},
	}
	deadline := time.Now().Add(5 * time.Minute)
	for time.Now().Before(deadline) {
		var d struct {
			IsIdentified bool `json:"is_identified"`
		}
		if err := b.cli.postForm(ctx, b.cli.liveAPI+"/xlive/app-blink/v1/preLive/IsUserIdentifiedByFaceAuth",
			form, b.cli.liveHeaders(), &d); err == nil && d.IsIdentified {
			return true
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(time.Second):
		}
	}
	return false
}

// start runs the broadcast start flow with the channel updates.
func (b *Backend) start(ctx context.Context, update func(app.StartUpdate) error) (app.StreamInfo, error) {
	if err := update(app.StartUpdate{Stage: app.StartRequesting, Message: "正在开播…"}); err != nil {
		return app.StreamInfo{}, err
	}
	b.updateLiveVersion(ctx)

	addr, code, faceQR, err := b.startLiveOnce(ctx)
	if err != nil {
		return app.StreamInfo{}, fmt.Errorf("开播失败: %w", err)
	}
	if faceQR != "" {
		svg, _ := qrSVG(faceQR)
		if err := update(app.StartUpdate{
			Stage:   app.StartFaceAuth,
			Face:    &app.QRUpdate{Stage: app.QRWaiting, Code: &app.QRCode{URL: faceQR, SVG: svg}, Message: "请使用哔哩哔哩 App 扫码完成人脸认证"},
			Message: "需要人脸认证",
		}); err != nil {
			return app.StreamInfo{}, err
		}
		if !b.checkFaceAuth(ctx) {
			return app.StreamInfo{}, errors.New("人脸认证未完成或已取消")
		}
		if err := update(app.StartUpdate{Stage: app.StartFaceAuth, Face: &app.QRUpdate{Stage: app.QRDone, Message: "人脸认证完成"}, Message: "人脸认证完成，正在重新开播…"}); err != nil {
			return app.StreamInfo{}, err
		}
		addr, code, faceQR, err = b.startLiveOnce(ctx)
		if err != nil {
			return app.StreamInfo{}, fmt.Errorf("开播失败: %w", err)
		}
		if faceQR != "" {
			return app.StreamInfo{}, errors.New("人脸认证后仍需验证，请稍后再试")
		}
	}

	prev := b.cfg.get().RTMPCode
	stream := app.StreamInfo{Addr: addr, Code: code, Changed: prev != "" && prev != code}
	_ = b.cfg.update(func(d *configData) {
		d.RTMPCodeOld = d.RTMPCode
		d.RTMPAddr = addr
		d.RTMPCode = code
	})
	b.mu.Lock()
	b.stream = &stream
	b.mu.Unlock()

	app.Logf(app.LogOK, "开播成功")
	if info, err := b.room(ctx); err == nil {
		_ = app.RoomChanged.Broadcast(info)
	}
	if err := update(app.StartUpdate{Stage: app.StartLive, Stream: &stream, Message: "开播成功"}); err != nil {
		return app.StreamInfo{}, err
	}
	return stream, nil
}

// stop ends the broadcast via Room/stopLive.
func (b *Backend) stop(ctx context.Context) error {
	cfg := b.cfg.get()
	if cfg.RoomID <= 0 || cfg.CSRF == "" {
		return errNotLogged
	}
	form := url.Values{
		"room_id":    {strconv.FormatInt(cfg.RoomID, 10)},
		"platform":   {"pc_link"},
		"csrf_token": {cfg.CSRF},
		"csrf":       {cfg.CSRF},
	}
	if err := b.cli.postForm(ctx, b.cli.liveAPI+"/room/v1/Room/stopLive", sign(form), b.cli.liveHeaders(), nil); err != nil {
		return fmt.Errorf("下播失败: %w", err)
	}
	app.Logf(app.LogOK, "下播成功")
	if info, err := b.room(ctx); err == nil {
		_ = app.RoomChanged.Broadcast(info)
	}
	return nil
}

// watchLive starts/stops the 30s room refresh goroutine while live.
func (b *Backend) watchLive(info app.RoomInfo) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if info.Status == app.LiveOn {
		if b.liveStop != nil {
			return // already running
		}
		done := make(chan struct{})
		b.liveStop = done
		go b.liveRefreshLoop(done)
	} else {
		if b.liveStop != nil {
			close(b.liveStop)
			b.liveStop = nil
		}
		if info.Status == app.LiveOffline {
			b.stream = nil
		}
	}
}

func (b *Backend) liveRefreshLoop(done chan struct{}) {
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-done:
			return
		case <-t.C:
		}
		info, err := b.room(context.Background())
		if err != nil {
			continue
		}
		_ = app.RoomChanged.Broadcast(info)
	}
}
