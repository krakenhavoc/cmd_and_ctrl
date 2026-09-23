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
	// CodeInsufficientMana — the S15 strict-mode cost gate rejected
	// a cast_spell because the caster's mana pool can't cover the
	// effective cost. The frame carries the missing-symbols slice
	// in ErrorPayload.Missing and the card_id of the rejected cast
	// in ErrorPayload.CardID; the client's "Override strict mode
	// for this cast" toast re-fires the action with `force_cast:
	// true` to bypass the gate. Added in S15 sub-PR 3.
	CodeInsufficientMana = "insufficient_mana"
	// CodeIllegalBlock — a declare_blocker named a pair the engine's
	// block-legality check refuses (CR 509.1b): a "can't block" or
	// "can't be blocked" restriction, flying, or landwalk. The message
	// is a player-facing sentence built server-side; the frame also
	// carries the refusal's token in ErrorPayload.Reason and the
	// blocker in ErrorPayload.CardID. ADR 0045 addendum Decision 8
	// (#705).
	CodeIllegalBlock = "illegal_block"
	// CodeAttackTaxUnpaid — a declare_attacker or declare_attackers
	// was refused because the attacking player cannot pay the
	// CR 508.1a attack tax the declaration owes (Propaganda, Ghostly
	// Prison, Sphere of Safety). The frame carries the whole price in
	// ErrorPayload.Reason — a cost string like "{2}{2}" — and the
	// symbols the pool and the tapper together could not cover in
	// ErrorPayload.Missing.
	//
	// Nothing was declared: no creature is tapped and no attack was
	// announced. The player's move is to attack with fewer creatures
	// or to make more mana; unlike insufficient_mana there is NO
	// override, because a tax waived is the card played as a blank.
	// ADR 0080 (#1063).
	CodeAttackTaxUnpaid = "attack_tax_unpaid"
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
// Code is one of the well-known protocol.Code* constants and is the
// stable handle clients pattern-match on; Message is human-readable
// debug copy. Some codes carry additional structured detail in the
// optional fields below — the wire stays small for the common case
// (omitempty everywhere) but the client can render richer
// affordances without parsing the message string.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	// Missing populates `code: "insufficient_mana"` (S15 sub-PR 3).
	// Each entry is a brace-formatted symbol the caster's pool
	// can't cover ("{R}", "{1}", "{W/U}"). The client renders an
	// "Override strict mode for this cast" toast that re-fires the
	// action with `force_cast: true`. Omitted for other codes.
	Missing []string `json:"missing,omitempty"`
	// CardID populates the structured-error frames that reference
	// a specific card the user was acting on (insufficient_mana
	// carries the card being cast, illegal_block the blocker). Lets the client correlate the
	// toast with its cast UI without keeping an in-flight map of
	// "what was the last action's instance_id?". Empty for
	// generic errors. Added in S15 sub-PR 3.
	CardID string `json:"card_id,omitempty"`
	// Reason populates `code: "illegal_block"`: the stable snake_case
	// token naming why the block was refused ("cant_block",
	// "cant_be_blocked", "flying", "landwalk"), the value of
	// game.BlockReason. Stable once shipped, like the restriction
	// tokens. CardID carries the refused blocker on the same frame.
	// Omitted for other codes. ADR 0045 addendum Decision 8 (#705).
	//
	// It also populates `code: "attack_tax_unpaid"`, where it is the
	// declaration's whole CR 508.1a price as a cost string ("{2}{2}"
	// for two attackers into Propaganda) rather than a token. One
	// field, two codes, because both answer "why" in the shape their
	// own code documents. ADR 0080 (#1063).
	Reason string `json:"reason,omitempty"`
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
	// Kind classifies the line so a client can render or suppress it
	// without parsing the text. Empty means ChatKindSay — every
	// human-typed message, and the only value handleChat ever
	// stamps. The bot kinds are server-originated (S31 sub-PR 8); a
	// client that tries to send one has it overwritten, because
	// handleChat re-stamps every authoritative field.
	Kind string `json:"kind,omitempty"`
	// Reason carries the deciding policy's Decision.Reason on a bot
	// line. Split out of Text rather than appended to it, because
	// the two halves have different rules: the announcement is
	// mandatory disclosure and every client must show it, while the
	// reasoning is a debug surface behind the "show bot reasoning"
	// setting.
	Reason string `json:"reason,omitempty"`
}

// Chat kinds. ChatKindSay is the zero value and covers everything a
// human types. The two bot kinds exist so a client can treat
// disclosure and debug narration differently — ADR 0033 §8.
const (
	// ChatKindSay is an ordinary message from a seat or a spectator.
	ChatKindSay = "say"
	// ChatKindBotImprovisation is a bot disclosing that it applied an
	// effect by hand because the catalog could not execute it.
	// Clients MUST show these: an unannounced improvisation is a bot
	// cheating.
	ChatKindBotImprovisation = "bot_improvisation"
	// ChatKindBotReasoning is a bot narrating why it chose a move.
	// Hidden unless the viewer turns on "show bot reasoning". It can
	// disclose cards in the bot's OWN hand — which is a disadvantage
	// the bot accepts, not an information leak: a policy never sees
	// another seat's hidden state, so it has none to spill.
	ChatKindBotReasoning = "bot_reasoning"
)

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
	// Annotation tags the replay line this payload produced with an
	// out-of-band note about what caused it. Set only on the replay /
	// crash-dump path (ws.Room.captureLocked); the WebSocket snapshot
	// frame is built separately and never carries one, so no client
	// ever sees this field.
	//
	// Its whole job is greppability. The replay log is a stream of
	// full snapshots, so "what happened here" normally has to be
	// diffed back out of two consecutive lines; an annotation says it
	// outright, and `grep bot_improvisation replays/*.jsonl` finds
	// every bot improvisation in a game that went wrong. Added in S31
	// sub-PR 8 (ADR 0033 §8).
	Annotation *ReplayAnnotation `json:"annotation,omitempty"`
}

// Replay annotation tags. These are grep handles — keep them stable.
const (
	// ReplayTagBotImprovisation marks the single commit in which a
	// bot applied a bundle of sandbox verbs by hand because the
	// catalog could not execute the effect its line needed.
	ReplayTagBotImprovisation = "bot_improvisation"
)

// ReplayAnnotation is the note attached to one replay line.
//
// Everything in it is public information — it repeats what the table
// was told in chat — so it is safe inside an artifact a player can
// download. Nothing here may carry hidden state.
type ReplayAnnotation struct {
	// Tag is the grep handle, one of the ReplayTag* constants.
	Tag string `json:"tag"`
	// Seat is the acting seat's player ID, when there is one.
	Seat string `json:"seat,omitempty"`
	// SeatName is that seat's display name, so a replay reads
	// without cross-referencing IDs.
	SeatName string `json:"seat_name,omitempty"`
	// Card names the card the improvisation was played for.
	Card string `json:"card,omitempty"`
	// Effect is the intended effect, in the words the table was given.
	Effect string `json:"effect,omitempty"`
	// Text is the exact chat line that was broadcast, verbatim, so
	// the replay records what the table was actually told rather
	// than a reconstruction of it.
	Text string `json:"text,omitempty"`
	// Reason is the policy's Decision.Reason. Always recorded here;
	// shown in chat only to players who asked for it.
	Reason string `json:"reason,omitempty"`
	// Steps are the wire action types the bundle applied, in order.
	Steps []string `json:"steps,omitempty"`
}
