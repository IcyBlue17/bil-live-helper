package bili

import (
	"net/url"
	"testing"
)

// Vectors computed with the upstream Python algorithm (utils/crypto.py,
// danmaku_wbi.py) — see the task notes.

func TestSign(t *testing.T) {
	// Reproduce sign_api_data with fixed ts/appkey by signing the same
	// query the Python code would produce and comparing the md5.
	//
	// Python:
	//   data={'room_id':'12345','platform':'pc_link','ts':'1700000000','appkey':APP_KEY}
	//   query='appkey=aae92bc66f3edfab&platform=pc_link&room_id=12345&ts=1700000000'
	//   md5(query+APP_SECRET) == '849a242fdc801b93fc9af7d85564bfff'
	p := url.Values{
		"room_id":  {"12345"},
		"platform": {"pc_link"},
	}
	out := sign(p)
	out.Set("ts", "1700000000") // re-sign below with fixed ts
	if got := out.Get("appkey"); got != appKey {
		t.Fatalf("appkey = %q", got)
	}
	if out.Get("sign") == "" {
		t.Fatal("sign missing")
	}

	// Now verify the md5 itself with the known vector by calling the inner
	// routine through signParamsFixed.
	s := signFixed(url.Values{
		"room_id":  {"12345"},
		"platform": {"pc_link"},
	}, "1700000000")
	if got := s.Get("sign"); got != "849a242fdc801b93fc9af7d85564bfff" {
		t.Fatalf("sign = %q, want 849a242fdc801b93fc9af7d85564bfff", got)
	}
	if got := s.Get("ts"); got != "1700000000" {
		t.Fatalf("ts = %q", got)
	}
}

func TestWBIMixin(t *testing.T) {
	// From Python: img='7cd084941338484aae1ad9425b84077c',
	// sub='4932caff0ff746eab6f01bf08b70ac45' → 'ea1db124af3c7062474693fa704f4ff8'
	img := "https://i0.hdslb.com/bfs/wbi/7cd084941338484aae1ad9425b84077c.png"
	sub := "https://i0.hdslb.com/bfs/wbi/4932caff0ff746eab6f01bf08b70ac45.png"
	got := mixinKeyFromURLs(img, sub)
	if got != "ea1db124af3c7062474693fa704f4ff8" {
		t.Fatalf("mixin = %q", got)
	}
}

func TestWBISign(t *testing.T) {
	key := "ea1db124af3c7062474693fa704f4ff8"
	out := wbiSign(key, url.Values{
		"id":   {"22637261"},
		"type": {"0"},
	}, 1700000000)
	if got := out.Get("w_rid"); got != "f83c24c9a207cf5543af92757f9327a9" {
		t.Fatalf("w_rid = %q, want f83c24c9a207cf5543af92757f9327a9", got)
	}
	if got := out.Get("wts"); got != "1700000000" {
		t.Fatalf("wts = %q", got)
	}
}

func TestWBISignFiltersChars(t *testing.T) {
	key := "ea1db124af3c7062474693fa704f4ff8"
	out := wbiSign(key, url.Values{"msg": {"a'b!(c)*d"}}, 1700000000)
	// Signed value strips !'()* but the returned params keep the original.
	if got := out.Get("msg"); got != "a'b!(c)*d" {
		t.Fatalf("msg = %q", got)
	}
	if out.Get("w_rid") == "" {
		t.Fatal("w_rid missing")
	}
}
