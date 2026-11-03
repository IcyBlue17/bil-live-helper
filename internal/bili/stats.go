package bili

import (
	"time"

	"github.com/IcyBlue17/bil-live-helper/app"
)

// statsTracker accumulates the per-session FeedEvent.Stats (live messages
// only, history excluded). At most one Stats push per second; online count
// sampled once a minute, keeping at most 240 samples.
type statsTracker struct {
	since    time.Time
	danmaku  int
	senders  map[int64]struct{}
	enters   int
	follows  int
	likes    int
	gifts    int
	giftVal  int64
	scs      int
	guards   int
	peak     int64
	samples  []app.OnlineSample
	lastSend time.Time
	lastSamp time.Time
	dirty    bool
}

func newStatsTracker() *statsTracker {
	return &statsTracker{since: time.Now(), senders: map[int64]struct{}{}}
}

// record counts one live message and marks the stats dirty.
func (s *statsTracker) record(m *app.Message) {
	switch m.Kind {
	case app.KindDanmaku, app.KindSelf:
		s.danmaku++
		s.senders[m.UID] = struct{}{}
	case app.KindEnter:
		s.enters++
	case app.KindFollow:
		s.follows++
	case app.KindLike:
		s.likes++
	case app.KindGift:
		s.gifts++
		if m.Gift != nil {
			s.giftVal += m.Gift.Price
		}
	case app.KindSuperChat:
		s.scs++
		if m.SuperChat != nil {
			s.giftVal += int64(m.SuperChat.Price) * 1000
		}
	case app.KindGuard:
		s.guards++
		if m.Gift != nil {
			s.giftVal += m.Gift.Price
		}
	}
	s.dirty = true
}

// online updates the peak and adds a minute sample.
func (s *statsTracker) online(n int64, now time.Time) {
	if n > s.peak {
		s.peak = n
	}
	if now.Sub(s.lastSamp) >= time.Minute {
		s.lastSamp = now
		s.samples = append(s.samples, app.OnlineSample{Time: now, Online: n})
		if len(s.samples) > 240 {
			s.samples = s.samples[len(s.samples)-240:]
		}
		s.dirty = true
	}
}

// snapshot returns the stats for a FeedEvent, nil when nothing changed or
// throttled (at most one per second).
func (s *statsTracker) snapshot(now time.Time) *app.SessionStats {
	if !s.dirty || now.Sub(s.lastSend) < time.Second {
		return nil
	}
	s.lastSend = now
	s.dirty = false
	return &app.SessionStats{
		Since:      s.since,
		Danmaku:    s.danmaku,
		Senders:    len(s.senders),
		Enters:     s.enters,
		Follows:    s.follows,
		Likes:      s.likes,
		Gifts:      s.gifts,
		GiftValue:  s.giftVal,
		SuperChats: s.scs,
		Guards:     s.guards,
		PeakOnline: s.peak,
		Samples:    append([]app.OnlineSample(nil), s.samples...),
	}
}
