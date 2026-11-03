package bili

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// wbiKeyIndexTable is the mixin index table from utils/constants.py.
var wbiKeyIndexTable = []int{
	46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35,
	27, 43, 5, 49, 33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13,
}

// mixinKeyFromURLs mixes the img/sub key names from the nav API,
// porting danmaku_wbi.py.
func mixinKeyFromURLs(imgURL, subURL string) string {
	keyName := func(u string) string {
		// last path segment, without the extension
		if i := strings.LastIndex(u, "/"); i >= 0 {
			u = u[i+1:]
		}
		if i := strings.Index(u, "."); i >= 0 {
			u = u[:i]
		}
		return u
	}
	return mixinKey(keyName(imgURL) + keyName(subURL))
}

// mixinKey applies the index table and keeps the first 32 chars.
func mixinKey(shuffled string) string {
	var sb strings.Builder
	for _, idx := range wbiKeyIndexTable {
		if idx < len(shuffled) {
			sb.WriteByte(shuffled[idx])
		}
	}
	s := sb.String()
	if len(s) > 32 {
		s = s[:32]
	}
	return s
}

// wbiSign adds wts and w_rid to params, porting WbiSigner.add_wbi_sign.
// Values are filtered of '!()* before signing; the returned params keep the
// original values plus wts/w_rid.
func wbiSign(key string, params url.Values, wts int64) url.Values {
	toSign := url.Values{}
	for k, vs := range params {
		toSign[k] = append([]string(nil), vs...)
	}
	toSign.Set("wts", strconv.FormatInt(wts, 10))

	keys := make([]string, 0, len(toSign))
	for k := range toSign {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteByte('&')
		}
		sb.WriteString(url.QueryEscape(k))
		sb.WriteByte('=')
		v := strings.Map(func(r rune) rune {
			switch r {
			case '!', '\'', '(', ')', '*':
				return -1
			}
			return r
		}, toSign.Get(k))
		sb.WriteString(url.QueryEscape(v))
	}
	sum := md5.Sum([]byte(sb.String() + key))

	out := url.Values{}
	for k, vs := range params {
		out[k] = append([]string(nil), vs...)
	}
	out.Set("wts", strconv.FormatInt(wts, 10))
	out.Set("w_rid", hex.EncodeToString(sum[:]))
	return out
}

// wbiKeyTTL matches WbiSigner.WBI_KEY_TTL.
const wbiKeyTTL = 11*time.Hour + 59*time.Minute + 30*time.Second

// wbiSigner caches the WBI key per client.
type wbiSigner struct {
	c *client

	mu      sync.Mutex
	key     string
	fetched time.Time
}

func newWBISigner(c *client) *wbiSigner { return &wbiSigner{c: c} }

func (w *wbiSigner) reset() {
	w.mu.Lock()
	w.key, w.fetched = "", time.Time{}
	w.mu.Unlock()
}

// key returns the cached key, refreshing it from the nav API when needed.
func (w *wbiSigner) getKey(ctx context.Context) (string, error) {
	w.mu.Lock()
	if w.key != "" && time.Since(w.fetched) < wbiKeyTTL {
		k := w.key
		w.mu.Unlock()
		return k, nil
	}
	w.mu.Unlock()

	var nav struct {
		WbiImg struct {
			ImgURL string `json:"img_url"`
			SubURL string `json:"sub_url"`
		} `json:"wbi_img"`
	}
	if err := w.c.getJSON(ctx, w.c.mainAPI+"/x/web-interface/nav", nil, nil, &nav); err != nil {
		return "", err
	}
	k := mixinKeyFromURLs(nav.WbiImg.ImgURL, nav.WbiImg.SubURL)
	if k == "" {
		return "", errNoWBIKey
	}
	w.mu.Lock()
	w.key, w.fetched = k, time.Now()
	w.mu.Unlock()
	return k, nil
}

// sign returns params plus wts/w_rid, fetching the key when needed.
func (w *wbiSigner) sign(ctx context.Context, params url.Values) (url.Values, error) {
	k, err := w.getKey(ctx)
	if err != nil {
		return nil, err
	}
	return wbiSign(k, params, time.Now().Unix()), nil
}
