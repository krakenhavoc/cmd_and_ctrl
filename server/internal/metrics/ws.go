package metrics

import (
	"slices"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// The WebSocket event metrics (ADR 0123 §3, "HTTP, WebSocket and
// runtime"). The hub records them. cmdctrl_ws_connections, the live
// count, is a collector (tables.go).

// Roles of a WebSocket connection: the role label of
// cmdctrl_ws_connections, cmdctrl_ws_connects_total and
// cmdctrl_ws_disconnects_total.
const (
	// RoleSeat is a connection bound to a seat: a player, or an admin
	// playing as one.
	RoleSeat = "seat"
	// RoleSpectator is a read-only connection with no seat.
	RoleSpectator = "spectator"
	// RoleAdmin is a writable connection with no seat: the admin's
	// omniscient view.
	RoleAdmin = "admin"
)

var roleLabels = []string{RoleSeat, RoleSpectator, RoleAdmin}

// WSRole is the role of a connection with these binding bits. A seat
// wins over the rest: an admin bound to a seat is playing it.
func WSRole(seated, readOnly bool) string {
	switch {
	case seated:
		return RoleSeat
	case readOnly:
		return RoleSpectator
	default:
		return RoleAdmin
	}
}

// Reasons of cmdctrl_ws_upgrade_rejections_total: each way the hub
// refuses an upgrade (ws.Hub.ServeWS).
const (
	// RejectShuttingDown: the hub is shutting down (503).
	RejectShuttingDown = "shutting_down"
	// RejectBadRequest: the authorizer refused a malformed request
	// (400: a bad or missing game or player id).
	RejectBadRequest = "bad_request"
	// RejectUnauthorized: no session, or one that does not validate
	// (401).
	RejectUnauthorized = "unauthorized"
	// RejectForbidden: a valid session that may not bind to this game
	// or seat (403).
	RejectForbidden = "forbidden"
	// RejectRejected: the authorizer refused with any other status.
	RejectRejected = "rejected"
	// RejectNoManager: the hub has no room manager (503).
	RejectNoManager = "no_manager"
	// RejectGameNotFound: the game is not live in this process (404).
	RejectGameNotFound = "game_not_found"
	// RejectPlayerNotInGame: the bound player is not a seat in the
	// game (403).
	RejectPlayerNotInGame = "player_not_in_game"
	// RejectUpgradeFailed: the WebSocket handshake itself failed, a
	// refused Origin included.
	RejectUpgradeFailed = "upgrade_failed"
)

var rejectLabels = []string{
	RejectShuttingDown, RejectBadRequest, RejectUnauthorized, RejectForbidden, RejectRejected,
	RejectNoManager, RejectGameNotFound, RejectPlayerNotInGame, RejectUpgradeFailed,
}

// Directions of cmdctrl_ws_frames_total.
const (
	FrameIn  = "in"
	FrameOut = "out"
)

var frameDirectionLabels = []string{FrameIn, FrameOut}

// FrameOther is the type label of a frame whose kind is not a protocol
// frame kind: an inbound frame that is not JSON, or names a kind the
// protocol does not have. The kind is client-chosen text.
const FrameOther = "other"

// frameTypeLabels is the closed set of the type label: protocol.Kind's
// values (internal/protocol/protocol.go), spelled here because this
// package imports nothing from this module, plus FrameOther. A new
// kind counts as other until it is added here.
var frameTypeLabels = []string{
	"ping", "pong", "error", "action", "snapshot", "chat",
	"legal_moves_request", "legal_moves", "ack", FrameOther,
}

var (
	wsConnects = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cmdctrl_ws_connects_total",
		Help: "WebSocket connections admitted, by role.",
	}, []string{"role"})

	wsDisconnects = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cmdctrl_ws_disconnects_total",
		Help: "WebSocket connections closed, by role.",
	}, []string{"role"})

	wsRejections = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cmdctrl_ws_upgrade_rejections_total",
		Help: "WebSocket upgrades refused, by reason.",
	}, []string{"reason"})

	wsFrames = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cmdctrl_ws_frames_total",
		Help: "WebSocket frames, by direction (in: read from a client; out: queued to one) and protocol frame type.",
	}, []string{"direction", "type"})

	wsBroadcast = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "cmdctrl_ws_broadcast_seconds",
		Help: "Time to filter, marshal and queue one state snapshot to every connection on a game.",
		// 100 µs to 1 s: a handful of viewers is well under a
		// millisecond, and a slow one is the thing to see.
		Buckets: []float64{.0001, .00025, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1},
	})
)

func init() {
	for _, r := range roleLabels {
		wsConnects.WithLabelValues(r)
		wsDisconnects.WithLabelValues(r)
	}
	for _, r := range rejectLabels {
		wsRejections.WithLabelValues(r)
	}
	for _, d := range frameDirectionLabels {
		for _, t := range frameTypeLabels {
			wsFrames.WithLabelValues(d, t)
		}
	}
}

// WSConnected counts one admitted connection with role (a Role*
// constant, from WSRole).
func WSConnected(role string) { wsConnects.WithLabelValues(role).Inc() }

// WSDisconnected counts one closed connection with role.
func WSDisconnected(role string) { wsDisconnects.WithLabelValues(role).Inc() }

// WSUpgradeRejected counts one refused upgrade. reason is a Reject*
// constant; anything else counts as RejectRejected.
func WSUpgradeRejected(reason string) {
	if !slices.Contains(rejectLabels, reason) {
		reason = RejectRejected
	}
	wsRejections.WithLabelValues(reason).Inc()
}

// WSFrame counts one frame of protocol kind kind in direction (FrameIn
// or FrameOut). A kind outside the protocol's set counts as
// FrameOther.
func WSFrame(direction, kind string) {
	if !slices.Contains(frameTypeLabels, kind) {
		kind = FrameOther
	}
	wsFrames.WithLabelValues(direction, kind).Inc()
}

// WSBroadcast observes one snapshot broadcast's duration.
func WSBroadcast(d time.Duration) { wsBroadcast.Observe(d.Seconds()) }
