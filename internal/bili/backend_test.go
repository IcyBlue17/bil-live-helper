package bili

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/IcyBlue17/bil-live-helper/app"
)

func testBackend(t *testing.T) *Backend {
	t.Helper()
	b := New()
	b.cfg.path = filepath.Join(t.TempDir(), "config.json")
	return b
}

// qrServer emulates the passport QR endpoints.
type qrServer struct {
	srv    *httptest.Server
	polls  *atomic.Int32
	stages []int64 // code returned at each poll
}

func newQRServer(t *testing.T, stages []int64) *qrServer {
	q := &qrServer{polls: &atomic.Int32{}, stages: stages}
	mux := http.NewServeMux()
	mux.HandleFunc("/x/passport-login/web/qrcode/generate", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"code": 0, "message": "0",
			"data": map[string]any{"url": "https://example.com/scan?k=KEY", "qrcode_key": "KEY"},
		})
	})
	mux.HandleFunc("/x/passport-login/web/qrcode/poll", func(w http.ResponseWriter, r *http.Request) {
		n := q.polls.Add(1)
		code := int64(86101)
		if int(n) <= len(q.stages) {
			code = q.stages[n-1]
		} else {
			code = q.stages[len(q.stages)-1]
		}
		data := map[string]any{"code": code, "message": "m"}
		if code == 0 {
			http.SetCookie(w, &http.Cookie{Name: "SESSDATA", Value: "sess"})
			http.SetCookie(w, &http.Cookie{Name: "bili_jct", Value: "csrf-tok"})
			http.SetCookie(w, &http.Cookie{Name: "DedeUserID", Value: "3493118"})
			data["refresh_token"] = "rt-1"
		}
		json.NewEncoder(w).Encode(map[string]any{"code": 0, "message": "0", "data": data})
	})
	mux.HandleFunc("/room/v2/Room/room_id_by_uid", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"code": 0, "message": "0",
			"data": map[string]any{"room_id": 22637261}})
	})
	mux.HandleFunc("/x/web-interface/nav", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"code": 0, "message": "0",
			"data": map[string]any{"isLogin": true, "mid": 3493118, "uname": "迷子", "face": ""}})
	})
	q.srv = httptest.NewServer(mux)
	t.Cleanup(q.srv.Close)
	return q
}

func pointAll(b *Backend, base string) {
	b.cli.liveAPI = base
	b.cli.passportAPI = base
	b.cli.mainAPI = base
	b.cli.wwwURL = base
}

func TestLoginExpired(t *testing.T) {
	b := testBackend(t)
	srv := newQRServer(t, []int64{86101, 86090, 86038})
	pointAll(b, srv.srv.URL)

	var got []app.QRStage
	err := b.Login(context.Background(), func(u app.QRUpdate) error {
		got = append(got, u.Stage)
		if u.Stage == app.QRWaiting {
			if u.Code == nil || u.Code.SVG == "" || u.Code.URL == "" {
				t.Fatal("first update must carry the QR code")
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []app.QRStage{app.QRWaiting, app.QRScanned, app.QRExpired}
	if len(got) != len(want) {
		t.Fatalf("stages = %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("stages = %v, want %v", got, want)
		}
	}
}

func TestLoginSuccess(t *testing.T) {
	b := testBackend(t)
	srv := newQRServer(t, []int64{86101, 86090, 0})
	pointAll(b, srv.srv.URL)

	var lastStage app.QRStage
	err := b.Login(context.Background(), func(u app.QRUpdate) error {
		lastStage = u.Stage
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if lastStage != app.QRDone {
		t.Fatalf("stage = %v", lastStage)
	}
	cfg := b.cfg.get()
	if cfg.Cookies["SESSDATA"] != "sess" || cfg.CSRF != "csrf-tok" || cfg.UID != 3493118 {
		t.Fatalf("cfg = %+v", cfg)
	}
	if cfg.RefreshToken != "rt-1" {
		t.Fatalf("refresh = %q", cfg.RefreshToken)
	}
	if cfg.RoomID != 22637261 {
		t.Fatalf("roomID = %d", cfg.RoomID)
	}
	st, err := b.Auth(context.Background())
	if err != nil || !st.LoggedIn || st.User == nil || st.User.Name != "迷子" {
		t.Fatalf("auth = %+v err %v", st, err)
	}
}

// TestStartFaceAuth exercises startLive → 60024 → face poll → retry.
func TestStartFaceAuth(t *testing.T) {
	b := testBackend(t)
	var startCalls, faceCalls atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/xlive/app-blink/v1/liveVersionInfo/getHomePageLiveVersion", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"code": 0, "message": "0",
			"data": map[string]any{"curr_version": "7.41.0.10090", "build": 10090}})
	})
	mux.HandleFunc("/room/v1/Room/startLive", func(w http.ResponseWriter, r *http.Request) {
		n := startCalls.Add(1)
		if n == 1 {
			json.NewEncoder(w).Encode(map[string]any{"code": 60024, "message": "需要人脸",
				"data": map[string]any{"qr": "https://example.com/face-qr"}})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"code": 0, "message": "0",
			"data": map[string]any{"rtmp": map[string]any{"addr": "rtmp://push/live-bvc/", "code": "?streamname=live_x&key=KEY"}}})
	})
	mux.HandleFunc("/xlive/app-blink/v1/preLive/IsUserIdentifiedByFaceAuth", func(w http.ResponseWriter, r *http.Request) {
		faceCalls.Add(1)
		json.NewEncoder(w).Encode(map[string]any{"code": 0, "message": "0",
			"data": map[string]any{"is_identified": true}})
	})
	mux.HandleFunc("/room/v1/Room/get_info", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"code": 0, "message": "0",
			"data": map[string]any{"room_id": 22637261, "uid": 1, "title": "t", "live_status": 1,
				"area_id": 21, "area_name": "视频唱见", "live_time": "2024-01-01 20:00:00"}})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	pointAll(b, srv.URL)

	b.cfg.update(func(d *configData) {
		d.UID = 1
		d.RoomID = 22637261
		d.CSRF = "csrf"
		d.Cookies = map[string]string{"bili_jct": "csrf"}
		d.AreaID = 21
		d.Title = "t"
	})

	var stages []app.StartStage
	stream, err := b.Start(context.Background(), func(u app.StartUpdate) error {
		stages = append(stages, u.Stage)
		if u.Stage == app.StartFaceAuth && u.Face == nil {
			t.Error("faceAuth update must carry Face")
		}
		if u.Stage == app.StartFaceAuth && u.Face != nil && u.Face.Stage == app.QRWaiting && u.Face.Code == nil {
			t.Error("face QR missing")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if stream.Addr != "rtmp://push/live-bvc/" || stream.Code == "" {
		t.Fatalf("stream = %+v", stream)
	}
	if startCalls.Load() != 2 {
		t.Fatalf("startLive calls = %d", startCalls.Load())
	}
	if faceCalls.Load() < 1 {
		t.Fatalf("face poll calls = %d", faceCalls.Load())
	}
	if len(stages) < 3 || stages[0] != app.StartRequesting || stages[len(stages)-1] != app.StartLive {
		t.Fatalf("stages = %v", stages)
	}
}

func TestStop(t *testing.T) {
	b := testBackend(t)
	var stopped atomic.Bool
	mux := http.NewServeMux()
	mux.HandleFunc("/room/v1/Room/stopLive", func(w http.ResponseWriter, r *http.Request) {
		stopped.Store(true)
		json.NewEncoder(w).Encode(map[string]any{"code": 0, "message": "0", "data": map[string]any{}})
	})
	mux.HandleFunc("/room/v1/Room/get_info", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"code": 0, "message": "0",
			"data": map[string]any{"room_id": 1, "live_status": 0}})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	pointAll(b, srv.URL)
	b.cfg.update(func(d *configData) {
		d.UID, d.RoomID, d.CSRF = 1, 1, "c"
		d.Cookies = map[string]string{"bili_jct": "c"}
	})
	if err := b.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !stopped.Load() {
		t.Fatal("stopLive not called")
	}
}

func TestConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	c := newConfig()
	c.path = path
	if err := c.update(func(d *configData) {
		d.UID = 42
		d.RoomID = 999
		d.CSRF = "csrf"
		d.RefreshToken = "rt"
		d.Cookies = map[string]string{"SESSDATA": "s", "bili_jct": "csrf"}
		d.Title = "测试标题"
		d.AreaID = 21
		d.RTMPAddr = "rtmp://a"
		d.RTMPCode = "?key=b"
		d.Areas = []app.AreaGroup{{ID: 1, Name: "娱乐", Areas: []app.Area{{ID: 21, Name: "视频唱见"}}}}
	}); err != nil {
		t.Fatal(err)
	}
	// File mode must be 0600.
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Fatalf("perm = %o", fi.Mode().Perm())
	}

	c2 := newConfig()
	c2.path = path
	d := c2.get()
	if d.UID != 42 || d.RoomID != 999 || d.CSRF != "csrf" || d.RefreshToken != "rt" {
		t.Fatalf("data = %+v", d)
	}
	if d.Cookies["SESSDATA"] != "s" || d.Title != "测试标题" || d.AreaID != 21 {
		t.Fatalf("data = %+v", d)
	}
	if d.RTMPAddr != "rtmp://a" || d.RTMPCode != "?key=b" {
		t.Fatalf("rtmp = %+v", d)
	}
	if len(d.Areas) != 1 || d.Areas[0].Areas[0].ID != 21 {
		t.Fatalf("areas = %+v", d.Areas)
	}
}

func TestCookieHeader(t *testing.T) {
	got := cookieHeader(map[string]string{"a": "1"})
	if got != "a=1" {
		t.Fatalf("cookieHeader = %q", got)
	}
}

func TestDanmakuSendValidation(t *testing.T) {
	b := testBackend(t)
	if err := b.SendDanmaku(context.Background(), ""); err == nil {
		t.Fatal("empty text should fail")
	}
	long := make([]rune, 41)
	for i := range long {
		long[i] = '字'
	}
	if err := b.SendDanmaku(context.Background(), string(long)); err == nil {
		t.Fatal(">40 runes should fail")
	}
}

func TestNavLoggedOut(t *testing.T) {
	b := testBackend(t)
	st, err := b.Auth(context.Background())
	if err != nil || st.LoggedIn {
		t.Fatalf("st = %+v err %v", st, err)
	}
}
