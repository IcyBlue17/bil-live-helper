// Package app is the contract between the Go backend and the web frontend.
// Every exported type here crosses the bridge as JSON, and `mygo generate`
// turns it into a TypeScript interface in src/mygo.ts.
package app

import "time"

// Version is the app's version, shown in the about panel.
const Version = "0.1.0"

// User is the logged-in Bilibili account.
type User struct {
	UID  int64  `json:"uid"`
	Name string `json:"name"`
	// Face is the avatar URL (load it with referrerpolicy="no-referrer").
	Face string `json:"face"`
}

// AuthState tells whether the app holds a valid login.
type AuthState struct {
	LoggedIn bool  `json:"loggedIn"`
	User     *User `json:"user"`
}

// QRStage is a step of a QR code flow (login or face verification).
type QRStage string

const (
	QRWaiting QRStage = "waiting" // shown, not scanned yet
	QRScanned QRStage = "scanned" // scanned, waiting for confirmation on the phone
	QRExpired QRStage = "expired" // the code expired; the flow ends
	QRDone    QRStage = "done"    // confirmed; the flow ends
)

// QRCode is a QR code to show the user.
type QRCode struct {
	// URL is what the code encodes.
	URL string `json:"url"`
	// SVG is the code rendered as a standalone <svg> document, black on transparent.
	SVG string `json:"svg"`
}

// QRUpdate is one step of a QR flow, streamed through a channel.
// The first update of a flow always carries Code.
type QRUpdate struct {
	Stage   QRStage `json:"stage"`
	Code    *QRCode `json:"code"`
	Message string  `json:"message"`
}

// LiveStatus is the state of the live room.
type LiveStatus string

const (
	LiveOffline LiveStatus = "offline" // 未开播
	LiveOn      LiveStatus = "live"    // 直播中
	LiveRound   LiveStatus = "round"   // 轮播中
)

// RoomInfo describes the user's live room.
type RoomInfo struct {
	RoomID         int64      `json:"roomId"`
	UID            int64      `json:"uid"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Cover          string     `json:"cover"`
	AreaID         int64      `json:"areaId"`
	AreaName       string     `json:"areaName"`
	ParentAreaID   int64      `json:"parentAreaId"`
	ParentAreaName string     `json:"parentAreaName"`
	Status         LiveStatus `json:"status"`
	Online         int64      `json:"online"`
	Attention      int64      `json:"attention"`
	// LiveSince is when the current broadcast started, null when offline.
	LiveSince *time.Time `json:"liveSince"`
}

// Area is a sub area (分区) a room can broadcast in.
type Area struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// AreaGroup is a parent area with its sub areas.
type AreaGroup struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Areas []Area `json:"areas"`
}

// StreamInfo is the RTMP address and key to paste into OBS.
type StreamInfo struct {
	Addr string `json:"addr"`
	Code string `json:"code"`
	// Changed is true when Code differs from the previous broadcast's,
	// so the user must update OBS.
	Changed bool `json:"changed"`
}

// RoomSettings are the saved title and area used for the next broadcast.
type RoomSettings struct {
	Title  string `json:"title"`
	AreaID int64  `json:"areaId"`
}

// StartStage is a step of starting a broadcast.
type StartStage string

const (
	StartRequesting StartStage = "requesting" // calling startLive
	StartFaceAuth   StartStage = "faceAuth"   // face verification needed; Face carries the QR
	StartLive       StartStage = "live"       // broadcast started; Stream is set
)

// StartUpdate is one step of Live.Start, streamed through a channel.
type StartUpdate struct {
	Stage   StartStage  `json:"stage"`
	Face    *QRUpdate   `json:"face"`
	Stream  *StreamInfo `json:"stream"`
	Message string      `json:"message"`
}

// GuardLevel is a captain rank (大航海).
type GuardLevel int

const (
	GuardNone    GuardLevel = 0
	GuardZongdu  GuardLevel = 1 // 总督
	GuardTidu    GuardLevel = 2 // 提督
	GuardJianzha GuardLevel = 3 // 舰长
)

// Medal is a fan medal (粉丝勋章).
type Medal struct {
	Name  string `json:"name"`
	Level int    `json:"level"`
}

// MessageKind is the kind of a danmaku feed item.
type MessageKind string

const (
	KindDanmaku MessageKind = "danmaku" // a viewer's comment
	KindSelf    MessageKind = "self"    // a comment the user sent
	KindGift    MessageKind = "gift"    // a gift or guard purchase
	KindEnter   MessageKind = "enter"   // a viewer entered (欢迎…)
	KindNotice  MessageKind = "notice"  // a room or platform notice
	KindSystem  MessageKind = "system"  // the app's own status lines
)

// Message is one item of the danmaku feed.
type Message struct {
	ID         string      `json:"id"`
	Kind       MessageKind `json:"kind"`
	Time       time.Time   `json:"time"`
	UID        int64       `json:"uid"`
	Name       string      `json:"name"`
	Text       string      `json:"text"`
	Admin      bool        `json:"admin"`
	Guard      GuardLevel  `json:"guard"`
	UserLevel  int         `json:"userLevel"`
	Medal      *Medal      `json:"medal"`
	Gift       *Gift       `json:"gift"`
	Historical bool        `json:"historical"` // from the history API, not live
}

// Gift is the gift of a KindGift message.
type Gift struct {
	Name string `json:"name"`
	Num  int    `json:"num"`
	// Price is the total value in gold coins (1000 = ¥1); 0 for free gifts.
	Price int64 `json:"price"`
}

// ConnState is the state of the danmaku connection.
type ConnState string

const (
	ConnConnecting   ConnState = "connecting"
	ConnConnected    ConnState = "connected"
	ConnReconnecting ConnState = "reconnecting"
	ConnClosed       ConnState = "closed"
)

// FeedEvent is an item of the Danmaku.Watch channel: a message, or a change
// of connection state.
type FeedEvent struct {
	Message *Message   `json:"message"`
	State   *ConnState `json:"state"`
	// Popularity is the room's popularity from heartbeats, 0 when unknown.
	Popularity int64 `json:"popularity"`
}

// LogLevel is the severity of a log entry.
type LogLevel string

const (
	LogInfo  LogLevel = "info"
	LogOK    LogLevel = "ok"
	LogWarn  LogLevel = "warn"
	LogError LogLevel = "error"
)

// LogEntry is a line of the activity log.
type LogEntry struct {
	Time    time.Time `json:"time"`
	Level   LogLevel  `json:"level"`
	Message string    `json:"message"`
}

// AppInfo describes the build.
type AppInfo struct {
	Version string `json:"version"`
	// Demo is true when the app runs with fake data (BLH_DEMO=1).
	Demo bool `json:"demo"`
	// ConfigPath is where the login and settings are stored.
	ConfigPath string `json:"configPath"`
}
