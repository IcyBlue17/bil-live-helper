package bili

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/egoist/mygo"

	"github.com/IcyBlue17/bil-live-helper/app"
)

// configData is the persisted state (config.json, mode 0600).
type configData struct {
	UID          int64             `json:"uid"`
	RoomID       int64             `json:"roomId"`
	CSRF         string            `json:"csrf"`
	RefreshToken string            `json:"refreshToken"`
	Cookies      map[string]string `json:"cookies"`

	Title  string `json:"title"`
	AreaID int64  `json:"areaId"`

	RTMPAddr    string `json:"rtmpAddr"`
	RTMPCode    string `json:"rtmpCode"`
	RTMPCodeOld string `json:"rtmpCodeOld"`

	LiveVersion string `json:"liveVersion"`
	LiveBuild   string `json:"liveBuild"`

	Areas []app.AreaGroup `json:"areas,omitempty"`
}

// config is the mutex-protected store. The path is resolved lazily because
// mygo.App.Path must not be called before the app is ready (and `mygo
// generate` runs main without ever starting it).
type config struct {
	mu   sync.Mutex
	path string
	data configData
	// loaded reports whether we already tried to read the file.
	loaded bool
}

func newConfig() *config {
	return &config{data: configData{
		AreaID:      -1,
		LiveVersion: livehimeVersion,
		LiveBuild:   livehimeBuild,
	}}
}

// dir returns the directory holding config.json. Tests override it by
// setting path directly.
func (c *config) configPath() string {
	if c.path != "" {
		return c.path
	}
	dir, err := mygo.App.Path(mygo.PathUserData)
	if err != nil {
		dir, err = os.UserConfigDir()
		if err != nil {
			dir = "."
		}
		dir = filepath.Join(dir, "bil-live-helper")
	}
	c.path = filepath.Join(dir, "config.json")
	return c.path
}

// load reads the file once; subsequent calls are no-ops.
func (c *config) load() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.loadLocked()
}

func (c *config) loadLocked() {
	if c.loaded {
		return
	}
	c.loaded = true
	raw, err := os.ReadFile(c.configPath())
	if errors.Is(err, os.ErrNotExist) {
		c.importUpstreamLocked()
		return
	}
	if err != nil {
		return
	}
	var d configData
	if json.Unmarshal(raw, &d) == nil {
		if d.AreaID == 0 {
			d.AreaID = -1
		}
		if d.LiveVersion == "" {
			d.LiveVersion = livehimeVersion
		}
		if d.LiveBuild == "" {
			d.LiveBuild = livehimeBuild
		}
		c.data = d
	}
}

// importUpstreamLocked picks up an upstream bilibili_live_tool config.json
// (version 2 format) from the working directory on first run.
func (c *config) importUpstreamLocked() {
	raw, err := os.ReadFile("config.json")
	if err != nil {
		return
	}
	var up struct {
		Version int `json:"version"`
		User    struct {
			UID          int64  `json:"uid"`
			CookiesStr   string `json:"cookies_str"`
			CSRF         string `json:"csrf"`
			RefreshToken string `json:"refresh_token"`
		} `json:"user"`
		Live struct {
			RoomID   int64  `json:"room_id"`
			Title    string `json:"title"`
			AreaID   int64  `json:"area_id"`
			RTMPAddr string `json:"rtmp_addr"`
			RTMPCode string `json:"rtmp_code"`
		} `json:"live"`
	}
	if json.Unmarshal(raw, &up) != nil || up.User.UID <= 0 {
		return
	}
	c.data.UID = up.User.UID
	c.data.CSRF = up.User.CSRF
	c.data.RefreshToken = up.User.RefreshToken
	c.data.RoomID = up.Live.RoomID
	c.data.Title = up.Live.Title
	c.data.AreaID = up.Live.AreaID
	c.data.RTMPAddr = up.Live.RTMPAddr
	c.data.RTMPCode = up.Live.RTMPCode
	c.data.RTMPCodeOld = up.Live.RTMPCode
	if up.User.CookiesStr != "" {
		var cookies map[string]string
		if json.Unmarshal([]byte(up.User.CookiesStr), &cookies) == nil {
			c.data.Cookies = cookies
		}
	}
	app.Logf(app.LogInfo, "已导入旧版配置")
}

// save writes the file atomically (temp + rename) with mode 0600.
func (c *config) save() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.loadLocked()
	return c.saveLocked()
}

func (c *config) saveLocked() error {
	path := c.configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(c.data, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// update applies fn to the data under the lock, then persists.
func (c *config) update(fn func(*configData)) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.loadLocked()
	fn(&c.data)
	return c.saveLocked()
}

// get returns a copy of the data (Cookies map is copied too).
func (c *config) get() configData {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.loadLocked()
	d := c.data
	if c.data.Cookies != nil {
		d.Cookies = make(map[string]string, len(c.data.Cookies))
		for k, v := range c.data.Cookies {
			d.Cookies[k] = v
		}
	}
	if c.data.Areas != nil {
		d.Areas = append([]app.AreaGroup(nil), c.data.Areas...)
	}
	return d
}
