// Package bili is the real app.Backend: it talks to Bilibili's live APIs
// (a port of the Python bilibili_live_tool) and implements QR login,
// room management, broadcast control and the danmaku websocket client.
package bili

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	appKey    = "aae92bc66f3edfab"
	appSecret = "af125a0d5279fd576c1b4418a3e8276d"

	livehimeBuild   = "10082"
	livehimeVersion = "7.40.0.10082"

	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36 Edg/137.0.0.0"
)

// apiError is a non-zero `code` in a Bilibili JSON reply.
type apiError struct {
	Code int64
	Msg  string
}

func (e *apiError) Error() string { return fmt.Sprintf("%s（%d）", e.Msg, e.Code) }

// client wraps the HTTP calls to Bilibili. Base URLs are fields so tests
// can point them at an httptest.Server.
type client struct {
	hc *http.Client

	liveAPI     string // https://api.live.bilibili.com
	passportAPI string // https://passport.bilibili.com
	mainAPI     string // https://api.bilibili.com
	wwwURL      string // https://www.bilibili.com

	// cookies returns the stored cookies to send with requests.
	cookies func() map[string]string
}

func newClient(cookies func() map[string]string) *client {
	return &client{
		hc:          &http.Client{Timeout: 10 * time.Second},
		liveAPI:     "https://api.live.bilibili.com",
		passportAPI: "https://passport.bilibili.com",
		mainAPI:     "https://api.bilibili.com",
		wwwURL:      "https://www.bilibili.com",
		cookies:     cookies,
	}
}

// liveHeaders mirrors live.py _get_headers.
func (c *client) liveHeaders() http.Header {
	h := http.Header{
		"Accept":             {"application/json, text/plain, */*"},
		"Accept-Language":    {"zh-CN,zh;q=0.9,en;q=0.8"},
		"Origin":             {"https://link.bilibili.com"},
		"Referer":            {"https://link.bilibili.com/p/center/index"},
		"Sec-Ch-Ua":          {`"Microsoft Edge";v="137", "Not=A?Brand";v="8", "Chromium";v="137"`},
		"Sec-Ch-Ua-Mobile":   {"?0"},
		"Sec-Ch-Ua-Platform": {`"Windows"`},
		"Sec-Fetch-Dest":     {"empty"},
		"Sec-Fetch-Mode":     {"cors"},
		"Sec-Fetch-Site":     {"same-site"},
		"User-Agent":         {userAgent},
	}
	return h
}

func (c *client) applyHeaders(req *http.Request, h http.Header, withCookies bool) {
	for k, vs := range h {
		req.Header[k] = append([]string(nil), vs...)
	}
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", userAgent)
	}
	if withCookies && c.cookies != nil {
		var sb strings.Builder
		for k, v := range c.cookies() {
			if sb.Len() > 0 {
				sb.WriteString("; ")
			}
			fmt.Fprintf(&sb, "%s=%s", k, v)
		}
		if sb.Len() > 0 {
			req.Header.Set("Cookie", sb.String())
		}
	}
}

// reply is the common {code, message|msg, data} envelope.
type reply struct {
	Code    int64           `json:"code"`
	Message string          `json:"message"`
	Msg     string          `json:"msg"`
	Data    json.RawMessage `json:"data"`
}

func (r *reply) err() error {
	if r.Code == 0 {
		return nil
	}
	msg := r.Message
	if msg == "" {
		msg = r.Msg
	}
	if msg == "" {
		msg = "未知错误"
	}
	return &apiError{Code: r.Code, Msg: msg}
}

// do issues req and decodes the JSON envelope. Non-zero code → *apiError.
// It also returns the Set-Cookie headers of the response (QR login needs them).
func (c *client) do(req *http.Request, out *reply) ([]*http.Cookie, error) {
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	if err := json.Unmarshal(body, out); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return resp.Cookies(), nil
}

// getJSON does GET url, decodes the envelope, and errors on non-zero code.
func (c *client) getJSON(ctx context.Context, url string, params url.Values, h http.Header, data any) error {
	if len(params) > 0 {
		url += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	c.applyHeaders(req, h, true)
	var r reply
	if _, err := c.do(req, &r); err != nil {
		return err
	}
	if err := r.err(); err != nil {
		return err
	}
	if data != nil && len(r.Data) > 0 {
		return json.Unmarshal(r.Data, data)
	}
	return nil
}

// getRaw does GET url and returns the envelope without checking code
// (for endpoints like QR poll whose meaningful code is inside data).
func (c *client) getRaw(ctx context.Context, url string, params url.Values, h http.Header, withCookies bool) (*reply, []*http.Cookie, error) {
	if len(params) > 0 {
		url += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	c.applyHeaders(req, h, withCookies)
	var r reply
	cs, err := c.do(req, &r)
	return &r, cs, err
}

// postForm POSTs a form body, decodes the envelope, errors on non-zero code.
func (c *client) postForm(ctx context.Context, url string, form url.Values, h http.Header, data any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	c.applyHeaders(req, h, true)
	var r reply
	if _, err := c.do(req, &r); err != nil {
		return err
	}
	if err := r.err(); err != nil {
		return err
	}
	if data != nil && len(r.Data) > 0 {
		return json.Unmarshal(r.Data, data)
	}
	return nil
}

// postFormRaw is postForm without the code check; the envelope is returned.
func (c *client) postFormRaw(ctx context.Context, url string, form url.Values, h http.Header) (*reply, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	c.applyHeaders(req, h, true)
	var r reply
	_, err = c.do(req, &r)
	return &r, err
}

// sign adds appkey/ts to params and appends the app-side md5 sign,
// porting utils/crypto.py sign_api_data. Values stay strings; ints must be
// formatted by the caller (strconv.Itoa / FormatInt).
func sign(params url.Values) url.Values {
	return signFixed(params, strconv.FormatInt(time.Now().Unix(), 10))
}

// signFixed is sign with a caller-supplied timestamp, kept separate so
// tests can reproduce known vectors.
func signFixed(params url.Values, ts string) url.Values {
	p := url.Values{}
	for k, vs := range params {
		p[k] = append([]string(nil), vs...)
	}
	p.Set("ts", ts)
	p.Set("appkey", appKey)

	keys := make([]string, 0, len(p))
	for k := range p {
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
		sb.WriteString(url.QueryEscape(p.Get(k)))
	}
	sum := md5.Sum([]byte(sb.String() + appSecret))
	p.Set("sign", hex.EncodeToString(sum[:]))
	return p
}
