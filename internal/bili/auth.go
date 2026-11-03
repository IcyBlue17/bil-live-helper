package bili

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/IcyBlue17/bil-live-helper/app"
)

var (
	errNoWBIKey  = errors.New("未获取到 WBI 密钥")
	errNotLogged = errors.New("未登录，请先扫码登录")
)

// navUser calls x/web-interface/nav and reports the login state.
func (b *Backend) navUser(ctx context.Context) (app.AuthState, error) {
	if !b.loggedIn() {
		return app.AuthState{}, nil
	}
	var d struct {
		IsLogin bool   `json:"isLogin"`
		Mid     int64  `json:"mid"`
		Uname   string `json:"uname"`
		Face    string `json:"face"`
	}
	err := b.cli.getJSON(ctx, b.cli.mainAPI+"/x/web-interface/nav", nil, nil, &d)
	if err != nil {
		if ae, ok := err.(*apiError); ok && ae.Code == -101 {
			return app.AuthState{}, nil
		}
		// Network error: trust the stored login (upstream does the same).
		if !isAPIErr(err) {
			cfg := b.cfg.get()
			if cfg.UID > 0 {
				return app.AuthState{LoggedIn: true, User: &app.User{UID: cfg.UID, Name: storedUname(b)}}, nil
			}
		}
		return app.AuthState{}, err
	}
	if !d.IsLogin {
		return app.AuthState{}, nil
	}
	return app.AuthState{LoggedIn: true, User: &app.User{UID: d.Mid, Name: d.Uname, Face: d.Face}}, nil
}

func isAPIErr(err error) bool {
	_, ok := err.(*apiError)
	return ok
}

// storedUname keeps the last known name from a nav call, for offline display.
func storedUname(b *Backend) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.uname
}

// Login runs the QR login flow, porting auth.py login_with_qr.
func (b *Backend) login(ctx context.Context, update func(app.QRUpdate) error) error {
	// Generate the QR code.
	var gen struct {
		URL       string `json:"url"`
		QRCodeKey string `json:"qrcode_key"`
	}
	r, _, err := b.cli.getRaw(ctx, b.cli.passportAPI+"/x/passport-login/web/qrcode/generate", nil, b.cli.liveHeaders(), false)
	if err != nil {
		return err
	}
	if err := r.err(); err != nil {
		return err
	}
	if err := json.Unmarshal(r.Data, &gen); err != nil {
		return err
	}
	svg, err := qrSVG(gen.URL)
	if err != nil {
		return err
	}
	if err := update(app.QRUpdate{
		Stage:   app.QRWaiting,
		Code:    &app.QRCode{URL: gen.URL, SVG: svg},
		Message: "请使用哔哩哔哩 App 扫码登录",
	}); err != nil {
		return err
	}

	scannedSent := false
	pollURL := b.cli.passportAPI + "/x/passport-login/web/qrcode/poll"
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
		r, cookies, err := b.cli.getRaw(ctx, pollURL,
			url.Values{"qrcode_key": {gen.QRCodeKey}}, b.cli.liveHeaders(), false)
		if err != nil {
			return err
		}
		var d struct {
			Code         int64  `json:"code"`
			Message      string `json:"message"`
			RefreshToken string `json:"refresh_token"`
			URL          string `json:"url"`
		}
		if err := json.Unmarshal(r.Data, &d); err != nil {
			return err
		}
		switch d.Code {
		case 0: // confirmed
			return b.finishLogin(cookies, d.RefreshToken, update)
		case 86090: // scanned, waiting for confirmation
			if !scannedSent {
				scannedSent = true
				if err := update(app.QRUpdate{Stage: app.QRScanned, Message: "已扫码，请在手机上确认登录"}); err != nil {
					return err
				}
			}
		case 86038: // expired
			_ = update(app.QRUpdate{Stage: app.QRExpired, Message: "二维码已过期，请重新扫码"})
			return nil
		default:
			// 86101 waiting, anything else: keep polling.
		}
	}
}

// finishLogin stores the cookies, fetches the room id and broadcasts.
func (b *Backend) finishLogin(cookies []*http.Cookie, refreshToken string, update func(app.QRUpdate) error) error {
	jar := map[string]string{}
	for _, c := range cookies {
		jar[c.Name] = c.Value
	}
	csrf := jar["bili_jct"]
	uid, _ := strconv.ParseInt(jar["DedeUserID"], 10, 64)

	if err := b.cfg.update(func(d *configData) {
		d.Cookies = jar
		d.CSRF = csrf
		d.UID = uid
		if refreshToken != "" {
			d.RefreshToken = refreshToken
		}
	}); err != nil {
		return err
	}

	// Fetch the room id for this uid (room_id_by_uid).
	var room struct {
		RoomID int64 `json:"room_id"`
	}
	if err := b.cli.getJSON(context.Background(), b.cli.liveAPI+"/room/v2/Room/room_id_by_uid",
		url.Values{"uid": {strconv.FormatInt(uid, 10)}}, nil, &room); err == nil && room.RoomID > 0 {
		_ = b.cfg.update(func(d *configData) { d.RoomID = room.RoomID })
	}

	if err := update(app.QRUpdate{Stage: app.QRDone, Message: "登录成功"}); err != nil {
		return err
	}
	state, _ := b.navUser(context.Background())
	if state.User == nil && uid > 0 {
		state = app.AuthState{LoggedIn: true, User: &app.User{UID: uid}}
	}
	_ = app.AuthChanged.Broadcast(state)
	name := ""
	if state.User != nil {
		name = state.User.Name
		b.mu.Lock()
		b.uname = name
		b.mu.Unlock()
	}
	app.Logf(app.LogOK, "登录成功："+name)
	return nil
}

func (b *Backend) logout() error {
	if err := b.cfg.update(func(d *configData) {
		d.Cookies = nil
		d.CSRF = ""
		d.RefreshToken = ""
		d.UID = 0
	}); err != nil {
		return err
	}
	b.mu.Lock()
	b.uname = ""
	b.mu.Unlock()
	_ = app.AuthChanged.Broadcast(app.AuthState{})
	app.Logf(app.LogInfo, "已退出登录")
	return nil
}

func (b *Backend) loggedIn() bool {
	cfg := b.cfg.get()
	return cfg.UID > 0 && len(cfg.Cookies) > 0
}

// mergeSetCookies stores cookies from a response into the jar.
func mergeSetCookies(jar map[string]string, cookies []*http.Cookie) {
	for _, c := range cookies {
		jar[c.Name] = c.Value
	}
}

// cookieHeader builds the Cookie header value (unused small helper kept
// for tests).
func cookieHeader(cookies map[string]string) string {
	var sb strings.Builder
	for k, v := range cookies {
		if sb.Len() > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(v)
	}
	return sb.String()
}
