package bili

import (
	"context"
	"sync"

	"github.com/IcyBlue17/bil-live-helper/app"
)

// Backend implements app.Backend against the real Bilibili APIs.
// It is constructed cheaply: no I/O happens until the first call, so
// `mygo generate` (which runs main without a window) stays side-effect free.
type Backend struct {
	cli *client
	cfg *config
	wbi *wbiSigner

	mu        sync.Mutex
	areaCache []app.AreaGroup
	stream    *app.StreamInfo
	uname     string
	liveStop  chan struct{}
}

// New returns the real backend.
func New() *Backend {
	cfg := newConfig()
	b := &Backend{cfg: cfg}
	b.cli = newClient(func() map[string]string { return cfg.get().Cookies })
	b.wbi = newWBISigner(b.cli)
	return b
}

func (b *Backend) Auth(ctx context.Context) (app.AuthState, error) {
	return b.navUser(ctx)
}

func (b *Backend) Login(ctx context.Context, update func(app.QRUpdate) error) error {
	return b.login(ctx, update)
}

func (b *Backend) Logout() error { return b.logout() }

func (b *Backend) Room(ctx context.Context) (app.RoomInfo, error) { return b.room(ctx) }

func (b *Backend) Areas(ctx context.Context) ([]app.AreaGroup, error) { return b.areas(ctx) }

func (b *Backend) Settings() (app.RoomSettings, error) {
	cfg := b.cfg.get()
	return app.RoomSettings{Title: cfg.Title, AreaID: cfg.AreaID}, nil
}

func (b *Backend) SaveSettings(ctx context.Context, s app.RoomSettings) (app.RoomInfo, error) {
	return b.saveSettings(ctx, s)
}

func (b *Backend) Start(ctx context.Context, update func(app.StartUpdate) error) (app.StreamInfo, error) {
	return b.start(ctx, update)
}

func (b *Backend) Stop(ctx context.Context) error { return b.stop(ctx) }

func (b *Backend) Stream() (*app.StreamInfo, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.stream != nil {
		return b.stream, nil
	}
	cfg := b.cfg.get()
	if cfg.RTMPAddr == "" || cfg.RTMPCode == "" {
		return nil, nil
	}
	return &app.StreamInfo{Addr: cfg.RTMPAddr, Code: cfg.RTMPCode, Changed: false}, nil
}

func (b *Backend) Watch(ctx context.Context, send func(app.FeedEvent) error) error {
	return b.watch(ctx, send)
}

func (b *Backend) SendDanmaku(ctx context.Context, text string) error {
	return b.sendDanmaku(ctx, text)
}

func (b *Backend) Announcement(ctx context.Context) (string, error) {
	return b.announcement(ctx)
}

func (b *Backend) SetAnnouncement(ctx context.Context, text string) error {
	return b.setAnnouncement(ctx, text)
}

func (b *Backend) Rank(ctx context.Context, kind app.RankKind) ([]app.RankUser, error) {
	return b.rank(ctx, kind)
}

func (b *Backend) Emoticons(ctx context.Context) ([]app.EmoticonPackage, error) {
	return b.emoticons(ctx)
}

func (b *Backend) SendEmoticon(ctx context.Context, token string) error {
	return b.sendEmoticon(ctx, token)
}

func (b *Backend) Viewer(ctx context.Context, uid int64) (app.ViewerCard, error) {
	return b.viewer(ctx, uid)
}

func (b *Backend) Mute(ctx context.Context, uid int64, hours int) error {
	return b.mute(ctx, uid, hours)
}

func (b *Backend) Unmute(ctx context.Context, id int64) error {
	return b.unmute(ctx, id)
}

func (b *Backend) Muted(ctx context.Context) ([]app.MutedUser, error) {
	return b.muted(ctx)
}

func (b *Backend) Preferences() (app.Preferences, error) {
	return b.preferences(), nil
}

func (b *Backend) SavePreferences(p app.Preferences) error {
	return b.savePreferences(p)
}

// ConfigPath resolves the config path lazily; it may be called before the
// app is ready (e.g. by mygo generate), so errors fall back to ".".
func (b *Backend) ConfigPath() string { return b.cfg.configPath() }
