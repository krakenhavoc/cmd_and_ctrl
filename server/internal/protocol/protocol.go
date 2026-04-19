// Package protocol defines the v0 wire format between the game server and
// its clients. Every frame on the WebSocket is one of these structs,
// serialised to JSON. The canonical spec is docs/protocol.md at the repo
// root.
package protocol

import "encoding/json"

// Version is the protocol version this server speaks. Frames with a
// different v field are rejected.
const Version = 0

// Kind is the discriminator for a frame.
type Kind string

const (
	KindPing     Kind = "ping"
	KindPong     Kind = "pong"
	KindError    Kind = "error"
	KindAction   Kind = "action"
	KindSnapshot Kind = "snapshot"
	KindChat     Kind = "chat"
)

// Error codes for Kind == KindError. Kept deliberately small at v0.
const (
	CodeBadVersion = "bad_version"
	CodeBadJSON    = "bad_json"
	CodeBadRequest = "bad_request"
	CodeInternal   = "internal"
)

// Frame is the envelope around every message. Payload is left as raw JSON
// and decoded by whichever handler owns the Kind.
type Frame struct {
	V       int             `json:"v"`
	Kind    Kind            `json:"kind"`
	ID      string          `json:"id"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// PingPayload is the payload body for a Kind == KindPing frame.
type PingPayload struct {
	Msg string `json:"msg,omitempty"`
}

// PongPayload is the payload body for a Kind == KindPong frame.
type PongPayload struct {
	Msg        string `json:"msg,omitempty"`
	ServerTime string `json:"server_time"`
}

// ErrorPayload is the payload body for a Kind == KindError frame.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ActionPayload is the payload body for a Kind == KindAction frame
// (client → server). It carries a typed action name and an
// action-specific params blob that the server decodes based on Type.
// See the actions package for the full catalog of action types.
type ActionPayload struct {
	Type   string          `json:"type"`
	Player string          `json:"player,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
}

// ChatPayload is the payload body for a Kind == KindChat frame in
// either direction. Client → server: only Text is honoured; the server
// re-stamps AuthorID, AuthorName, and Timestamp from the connection's
// principal before broadcasting. Server → client: all four fields are
// authoritative.
//
// Chat frames bypass Room.Apply: they don't mutate game state, don't
// bump the snapshot seq, and don't appear in crash-recovery dumps.
// Reconnecting clients see an empty chat history; persistent chat
// arrives with the S11 replay log.
type ChatPayload struct {
	AuthorID   string `json:"author_id,omitempty"`
	AuthorName string `json:"author_name"`
	Text       string `json:"text"`
	Timestamp  string `json:"timestamp"`
}

// MaxChatTextLen caps the text size of a single chat message. Messages
// exceeding this length are rejected with a bad_request error.
const MaxChatTextLen = 1000

// SnapshotPayload is the payload body for a Kind == KindSnapshot
// frame (server → client). It carries a full authoritative view of
// the game state. The server emits a snapshot on initial connect and
// after every successful action.
//
// S03 deliberately broadcasts full snapshots rather than incremental
// deltas. Bandwidth is not a concern at four clients; diff-based
// delta frames can be added later without a protocol version bump.
type SnapshotPayload struct {
	// Seq is a monotonically increasing per-game sequence number,
	// incremented before each broadcast. Clients can use it to detect
	// dropped or out-of-order frames.
	Seq  uint64   `json:"seq"`
	Game GameView `json:"game"`
}
