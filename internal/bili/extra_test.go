package bili

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/IcyBlue17/bil-live-helper/app"
)

func loginForTest(b *Backend) {
	b.cfg.update(func(d *configData) {
		d.UID = 3493118
		d.RoomID = 22637261
		d.CSRF = "csrf"
		d.Cookies = map[string]string{"bili_jct": "csrf", "SESSDATA": "s"}
	})
}

func okEnvelope(data map[string]any) []byte {
	raw, _ := json.Marshal(map[string]any{"code": 0, "message": "0", "data": data})
	return raw
}

// ===== 公告 =====

func TestAnnouncement(t *testing.T) {
	b := testBackend(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/live_user/v1/Master/info", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("uid") != "3493118" {
			t.Errorf("uid = %q", r.URL.Query().Get("uid"))
		}
		w.Write(okEnvelope(map[string]any{
			"room_news": map[string]any{"content": "周六晚上八点歌回", "anchor_name": "迷子"},
		}))
	})
	var gotForm url.Values
	mux.HandleFunc("/xlive/app-blink/v1/index/updateRoomNews", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		gotForm = r.PostForm
		w.Write(okEnvelope(map[string]any{}))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	pointAll(b, srv.URL)
	loginForTest(b)

	got, err := b.Announcement(context.Background())
	if err != nil || got != "周六晚上八点歌回" {
		t.Fatalf("announcement = %q err %v", got, err)
	}

	if err := b.SetAnnouncement(context.Background(), "新的公告"); err != nil {
		t.Fatal(err)
	}
	if gotForm.Get("content") != "新的公告" || gotForm.Get("room_id") != "22637261" {
		t.Fatalf("form = %v", gotForm)
	}
}

func TestSetAnnouncementTooLong(t *testing.T) {
	b := testBackend(t)
	loginForTest(b)
	long := strings.Repeat("字", 61)
	if err := b.SetAnnouncement(context.Background(), long); err == nil {
		t.Fatal("61 chars should fail")
	}
}

// ===== 榜单 =====

func TestOnlineRank(t *testing.T) {
	b := testBackend(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/xlive/general-interface/v1/rank/getOnlineGoldRank", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("roomId") != "22637261" || q.Get("ruid") != "3493118" || q.Get("pageSize") != "50" {
			t.Errorf("query = %v", q)
		}
		w.Write(okEnvelope(map[string]any{
			"OnlineRankItem": []any{
				map[string]any{"userRank": 1, "uid": 11, "name": "榜一", "face": "f1", "score": 9999,
					"guard_level": 3, "medalInfo": map[string]any{"medalName": "迷子", "level": 15}},
				map[string]any{"userRank": 2, "uid": 22, "name": "榜二", "face": "f2", "score": 500},
			},
		}))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	pointAll(b, srv.URL)
	loginForTest(b)

	rows, err := b.Rank(context.Background(), app.RankOnline)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows = %+v", rows)
	}
	if rows[0].Rank != 1 || rows[0].Name != "榜一" || rows[0].Score != 9999 ||
		rows[0].Guard != app.GuardJianzha || rows[0].Medal == nil || rows[0].Medal.Name != "迷子" {
		t.Fatalf("row0 = %+v", rows[0])
	}
}

func TestGuardRank(t *testing.T) {
	b := testBackend(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/xlive/app-room/v2/guardTab/topListNew", func(w http.ResponseWriter, r *http.Request) {
		w.Write(okEnvelope(map[string]any{
			"top3": []any{
				map[string]any{"rank": 1, "uinfo": map[string]any{
					"uid":   99,
					"base":  map[string]any{"name": "总督大人", "face": "f"},
					"guard": map[string]any{"level": 1},
					"medal": map[string]any{"name": "迷子", "level": 28, "guard_level": 1},
				}},
			},
			"list": []any{
				map[string]any{"rank": 1, "uinfo": map[string]any{
					"uid":   99, // duplicate of top3
					"base":  map[string]any{"name": "总督大人", "face": "f"},
					"guard": map[string]any{"level": 1},
					"medal": map[string]any{"name": "迷子", "level": 28, "guard_level": 1},
				}},
				map[string]any{"rank": 4, "uinfo": map[string]any{
					"uid":   55,
					"base":  map[string]any{"name": "舰长甲", "face": ""},
					"guard": map[string]any{"level": 3},
					"medal": map[string]any{"name": "迷子", "level": 10, "guard_level": 3},
				}},
			},
		}))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	pointAll(b, srv.URL)
	loginForTest(b)

	rows, err := b.Rank(context.Background(), app.RankGuard)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows = %+v", rows)
	}
	if rows[0].Name != "总督大人" || rows[0].Guard != app.GuardZongdu {
		t.Fatalf("row0 = %+v", rows[0])
	}
	if rows[1].Name != "舰长甲" || rows[1].Guard != app.GuardJianzha {
		t.Fatalf("row1 = %+v", rows[1])
	}
}

// ===== 表情 =====

func TestEmoticonsAndSend(t *testing.T) {
	b := testBackend(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/xlive/web-ucenter/v2/emoticon/GetEmoticons", func(w http.ResponseWriter, r *http.Request) {
		w.Write(okEnvelope(map[string]any{
			"data": []any{
				map[string]any{"pkg_id": 2, "pkg_name": "UP主大表情", "current_cover": "cover.png",
					"emoticons": []any{
						map[string]any{"emoticon_unique": "official_124", "emoji": "打call",
							"url": "https://i0.hdslb.com/emo.png", "width": 100, "height": 100},
					}},
			},
		}))
	})
	var gotForm url.Values
	mux.HandleFunc("/msg/send", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		gotForm = r.PostForm
		w.Write(okEnvelope(map[string]any{"mode_info": map[string]any{}}))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	pointAll(b, srv.URL)
	loginForTest(b)

	pkgs, err := b.Emoticons(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 1 || pkgs[0].ID != 2 || pkgs[0].Name != "UP主大表情" {
		t.Fatalf("pkgs = %+v", pkgs)
	}
	it := pkgs[0].Items[0]
	if it.Token != "official_124" || it.URL == "" || it.Width != 100 {
		t.Fatalf("item = %+v", it)
	}

	if err := b.SendEmoticon(context.Background(), "official_124"); err != nil {
		t.Fatal(err)
	}
	if gotForm.Get("dm_type") != "1" || gotForm.Get("msg") != "official_124" {
		t.Fatalf("form = %v", gotForm)
	}
}

// ===== 观众卡片 / 禁言 =====

func TestViewerAndMute(t *testing.T) {
	b := testBackend(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/xlive/app-ucenter/v2/card/user", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("uid") != "777" {
			t.Errorf("uid = %q", r.URL.Query().Get("uid"))
		}
		w.Write(okEnvelope(map[string]any{
			"uid": 777, "uname": "被观察者", "face": "f", "desc": "签名内容", "follow_num": 1200,
			"fans_medal":  map[string]any{"medal_name": "迷子", "medal_level": 8, "guard_level": 3},
			"wealth_info": map[string]any{"level": 42},
			"guard":       map[string]any{"level": 0},
		}))
	})
	mux.HandleFunc("/xlive/web-ucenter/v1/banned/GetSilentUserList", func(w http.ResponseWriter, r *http.Request) {
		w.Write(okEnvelope(map[string]any{
			"total_page": 1,
			"data": []any{
				map[string]any{"tuid": 777, "tname": "被观察者", "name": "管理员A", "ctime": "2026-11-03 12:00:00", "id": 8001},
				map[string]any{"tuid": 888, "tname": "另一人", "name": "管理员B", "ctime": "2026-11-03 12:00:00", "id": 8002},
			},
		}))
	})
	var muteForm url.Values
	mux.HandleFunc("/xlive/web-ucenter/v1/banned/AddSilentUser", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		muteForm = r.PostForm
		w.Write(okEnvelope(map[string]any{}))
	})
	var unmuteForm url.Values
	mux.HandleFunc("/banned_service/v1/Silent/del_room_block_user", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		unmuteForm = r.PostForm
		w.Write(okEnvelope(map[string]any{}))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	pointAll(b, srv.URL)
	loginForTest(b)

	card, err := b.Viewer(context.Background(), 777)
	if err != nil {
		t.Fatal(err)
	}
	if card.Name != "被观察者" || card.Sign != "签名内容" || card.Followers != 1200 ||
		card.Level != 42 || !card.Muted {
		t.Fatalf("card = %+v", card)
	}
	if card.Medal == nil || card.Medal.Name != "迷子" || card.Guard != app.GuardJianzha {
		t.Fatalf("card medal = %+v guard %v", card.Medal, card.Guard)
	}

	muted, err := b.Muted(context.Background())
	if err != nil || len(muted) != 2 {
		t.Fatalf("muted = %+v err %v", muted, err)
	}
	if muted[0].UID != 777 || muted[0].Admin != "管理员A" || muted[0].ID != 8001 {
		t.Fatalf("muted[0] = %+v", muted[0])
	}

	if err := b.Mute(context.Background(), 777, 0); err != nil {
		t.Fatal(err)
	}
	if muteForm.Get("tuid") != "777" || muteForm.Get("hour") != "0" {
		t.Fatalf("mute form = %v", muteForm)
	}

	if err := b.Unmute(context.Background(), 8001); err != nil {
		t.Fatal(err)
	}
	if unmuteForm.Get("id") != "8001" || unmuteForm.Get("roomid") != "22637261" {
		t.Fatalf("unmute form = %v", unmuteForm)
	}
}

// ===== 偏好设置 / 屏蔽词 =====

func TestPreferencesRoundTrip(t *testing.T) {
	b := testBackend(t)

	p, err := b.Preferences()
	if err != nil || !p.Notify || p.NotifyGiftMin != 10000 {
		t.Fatalf("defaults = %+v", p)
	}

	want := app.Preferences{Notify: false, NotifyGiftMin: 2000, Blocked: []string{"广告", "Spam"}}
	if err := b.SavePreferences(want); err != nil {
		t.Fatal(err)
	}

	// A fresh backend over the same config file must see the saved prefs.
	b2 := New()
	b2.cfg.path = b.cfg.path
	got, err := b2.Preferences()
	if err != nil {
		t.Fatal(err)
	}
	if got.Notify != false || got.NotifyGiftMin != 2000 || len(got.Blocked) != 2 {
		t.Fatalf("saved = %+v", got)
	}

	if !b2.blocked("这是个SPAM弹幕") || !b2.blocked("看到广告了") {
		t.Fatal("blocked filter failed")
	}
	if b2.blocked("正常弹幕") {
		t.Fatal("normal danmaku blocked")
	}
}

// ===== 会话统计 =====

func TestStatsTracker(t *testing.T) {
	s := newStatsTracker()
	now := time.Now()

	s.record(&app.Message{Kind: app.KindDanmaku, UID: 1})
	s.record(&app.Message{Kind: app.KindDanmaku, UID: 1})
	s.record(&app.Message{Kind: app.KindDanmaku, UID: 2})
	s.record(&app.Message{Kind: app.KindEnter})
	s.record(&app.Message{Kind: app.KindFollow})
	s.record(&app.Message{Kind: app.KindLike})
	s.record(&app.Message{Kind: app.KindGift, Gift: &app.Gift{Price: 19980}})
	s.record(&app.Message{Kind: app.KindSuperChat, SuperChat: &app.SuperChat{Price: 50}})
	s.record(&app.Message{Kind: app.KindGuard, Gift: &app.Gift{Price: 138000}})

	snap := s.snapshot(now.Add(2 * time.Second))
	if snap == nil {
		t.Fatal("expected snapshot")
	}
	if snap.Danmaku != 3 || snap.Senders != 2 || snap.Enters != 1 || snap.Follows != 1 ||
		snap.Likes != 1 || snap.Gifts != 1 || snap.SuperChats != 1 || snap.Guards != 1 {
		t.Fatalf("snap = %+v", snap)
	}
	wantVal := int64(19980 + 50*1000 + 138000)
	if snap.GiftValue != wantVal {
		t.Fatalf("giftValue = %d want %d", snap.GiftValue, wantVal)
	}

	// Throttled: second snapshot within a second is nil even when dirty.
	s.record(&app.Message{Kind: app.KindDanmaku, UID: 3})
	if s.snapshot(now.Add(2500*time.Millisecond)) != nil {
		t.Fatal("expected throttled nil")
	}
	// Nothing new → nil.
	if got := s.snapshot(now.Add(4 * time.Second)); got == nil {
		t.Fatal("expected dirty snapshot after throttle window")
	}
	if s.snapshot(now.Add(5*time.Second)) != nil {
		t.Fatal("expected nil when clean")
	}
}

func TestStatsOnlineSamples(t *testing.T) {
	s := newStatsTracker()
	base := time.Now()
	for i := 0; i < 241; i++ {
		s.online(int64(100+i), base.Add(time.Duration(i)*time.Minute))
	}
	if s.peak != 340 {
		t.Fatalf("peak = %d", s.peak)
	}
	snap := s.snapshot(base.Add(300 * time.Minute))
	if snap == nil {
		t.Fatal("expected snapshot")
	}
	if len(snap.Samples) != 240 {
		t.Fatalf("samples = %d", len(snap.Samples))
	}
	// Oldest sample is trimmed; the kept tail ends with the newest.
	last := snap.Samples[len(snap.Samples)-1]
	if last.Online != 340 {
		t.Fatalf("last sample = %+v", last)
	}
}
