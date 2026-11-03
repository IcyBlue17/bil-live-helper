package app

import (
	"time"

	"github.com/egoist/mygo"
)

// RoomChanged is broadcast whenever the backend learns new room info:
// after a refresh, a settings change, a start or a stop, and periodically
// while live.
var RoomChanged = mygo.NewEvent[RoomInfo]("room:changed")

// AuthChanged is broadcast after a login or logout.
var AuthChanged = mygo.NewEvent[AuthState]("auth:changed")

// Log is broadcast for each line of the activity log.
var Log = mygo.NewEvent[LogEntry]("log")

// Logf broadcasts a log entry.
func Logf(level LogLevel, msg string) {
	_ = Log.Broadcast(LogEntry{Time: time.Now(), Level: level, Message: msg})
}

// UpdateChanged is broadcast when the updater's state changes: after a
// check, or an answer in the update window.
var UpdateChanged = mygo.NewEvent[UpdateStatus]("update:changed")
