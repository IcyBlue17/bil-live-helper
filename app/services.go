package app

import (
	"context"
	"os"

	"github.com/egoist/mygo"
)

// Bind registers the services the frontend calls, delegating to b.
func Bind(b Backend) {
	mygo.Bind(&Auth{b})
	mygo.Bind(&Live{b})
	mygo.Bind(&Danmaku{b})
	mygo.Bind(&System{b})
}

// Auth logs the user in and out.
type Auth struct{ b Backend }

// State reports the stored login, checking it against the server.
func (s *Auth) State(ctx context.Context) (AuthState, error) { return s.b.Auth(ctx) }

// Login runs a QR login, streaming its steps. Closing the channel cancels it.
func (s *Auth) Login(ctx context.Context, updates *mygo.Channel[QRUpdate]) error {
	return s.b.Login(ctx, updates.Send)
}

// Logout forgets the stored login.
func (s *Auth) Logout() error { return s.b.Logout() }

// Live manages the room and the broadcast.
type Live struct{ b Backend }

// Room fetches the user's room.
func (s *Live) Room(ctx context.Context) (RoomInfo, error) { return s.b.Room(ctx) }

// Areas returns the area tree.
func (s *Live) Areas(ctx context.Context) ([]AreaGroup, error) { return s.b.Areas(ctx) }

// Settings returns the saved title and area.
func (s *Live) Settings() (RoomSettings, error) { return s.b.Settings() }

// SaveSettings saves and applies the title and area.
func (s *Live) SaveSettings(ctx context.Context, settings RoomSettings) (RoomInfo, error) {
	return s.b.SaveSettings(ctx, settings)
}

// Start starts a broadcast, streaming its steps (including face
// verification). Closing the channel cancels it.
func (s *Live) Start(ctx context.Context, updates *mygo.Channel[StartUpdate]) (StreamInfo, error) {
	return s.b.Start(ctx, updates.Send)
}

// Stop ends the broadcast.
func (s *Live) Stop(ctx context.Context) error { return s.b.Stop(ctx) }

// Stream returns the stored RTMP address and key, null when there is none.
func (s *Live) Stream() (*StreamInfo, error) { return s.b.Stream() }

// Announcement returns the room announcement.
func (s *Live) Announcement(ctx context.Context) (string, error) { return s.b.Announcement(ctx) }

// SetAnnouncement updates the room announcement.
func (s *Live) SetAnnouncement(ctx context.Context, text string) error {
	return s.b.SetAnnouncement(ctx, text)
}

// Rank returns the 高能榜 or 大航海 list.
func (s *Live) Rank(ctx context.Context, kind RankKind) ([]RankUser, error) {
	return s.b.Rank(ctx, kind)
}

// Danmaku reads and sends the room's comments.
type Danmaku struct{ b Backend }

// Watch streams history, then live comments, gifts and connection states,
// until the channel is closed.
func (s *Danmaku) Watch(ctx context.Context, events *mygo.Channel[FeedEvent]) error {
	return s.b.Watch(ctx, events.Send)
}

// Send posts a comment to the room.
func (s *Danmaku) Send(ctx context.Context, text string) error { return s.b.SendDanmaku(ctx, text) }

// Emoticons returns the emoticon packages usable in the room.
func (s *Danmaku) Emoticons(ctx context.Context) ([]EmoticonPackage, error) {
	return s.b.Emoticons(ctx)
}

// SendEmoticon posts a sticker danmaku.
func (s *Danmaku) SendEmoticon(ctx context.Context, token string) error {
	return s.b.SendEmoticon(ctx, token)
}

// Viewer returns a viewer's card.
func (s *Danmaku) Viewer(ctx context.Context, uid int64) (ViewerCard, error) {
	return s.b.Viewer(ctx, uid)
}

// Mute mutes a viewer for hours (0 = until the broadcast ends).
func (s *Danmaku) Mute(ctx context.Context, uid int64, hours int) error {
	return s.b.Mute(ctx, uid, hours)
}

// Unmute lifts a mute.
func (s *Danmaku) Unmute(ctx context.Context, id int64) error { return s.b.Unmute(ctx, id) }

// Muted lists muted viewers.
func (s *Danmaku) Muted(ctx context.Context) ([]MutedUser, error) { return s.b.Muted(ctx) }

// System has app-level helpers.
type System struct{ b Backend }

// Info describes the build.
func (s *System) Info() AppInfo {
	return AppInfo{Version: Version, Demo: os.Getenv("BLH_DEMO") == "1", ConfigPath: s.b.ConfigPath()}
}

// Copy puts text on the clipboard.
func (s *System) Copy(text string) { mygo.Clipboard.WriteText(text) }

// Preferences returns the saved preferences.
func (s *System) Preferences() (Preferences, error) { return s.b.Preferences() }

// SavePreferences saves and applies the preferences.
func (s *System) SavePreferences(p Preferences) error { return s.b.SavePreferences(p) }

// OpenURL opens a URL in the default browser.
func (s *System) OpenURL(url string) error { return mygo.Shell.OpenExternal(url) }
