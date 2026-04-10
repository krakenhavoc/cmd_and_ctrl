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
	KindPing  Kind = "ping"
	KindPong  Kind = "pong"
	KindError Kind = "error"
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
