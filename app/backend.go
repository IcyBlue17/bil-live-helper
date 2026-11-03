package app

import "context"

// Backend is what the bound services delegate to. internal/demo implements
// it with fake data; the real implementation talks to Bilibili.
//
// Methods are called from many goroutines at once and must be safe for that.
// Backends report notable actions with the Log event and room changes with
// RoomChanged, from any goroutine.
type Backend interface {
	// Auth reports the stored login, checking it against the server.
	Auth(ctx context.Context) (AuthState, error)
	// Login runs a QR login, calling update for each step, until it ends
	// (done, expired) or ctx is canceled. On success it stores the login.
	Login(ctx context.Context, update func(QRUpdate) error) error
	// Logout forgets the stored login.
	Logout() error

	// Room fetches the user's room from the server.
	Room(ctx context.Context) (RoomInfo, error)
	// Areas returns the area tree, cached after the first fetch.
	Areas(ctx context.Context) ([]AreaGroup, error)
	// Settings returns the saved title and area for the next broadcast.
	Settings() (RoomSettings, error)
	// SaveSettings validates, saves and applies (via the room update API)
	// the title and area. Titles are at most 40 characters.
	SaveSettings(ctx context.Context, s RoomSettings) (RoomInfo, error)

	// Start starts a broadcast with the saved area, running face
	// verification when the server asks for it, calling update per step.
	Start(ctx context.Context, update func(StartUpdate) error) (StreamInfo, error)
	// Stop ends the broadcast.
	Stop(ctx context.Context) error
	// Stream returns the stored stream info, nil when there is none.
	Stream() (*StreamInfo, error)

	// Watch connects to the room's danmaku server and sends history then
	// live events to send, reconnecting on failure, until ctx is canceled.
	Watch(ctx context.Context, send func(FeedEvent) error) error
	// SendDanmaku posts a comment to the room.
	SendDanmaku(ctx context.Context, text string) error

	// ConfigPath is where the backend stores its data.
	ConfigPath() string
}
