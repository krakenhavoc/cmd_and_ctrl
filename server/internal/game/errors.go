// Package game holds the authoritative in-memory game state for a
// single cmd_and_ctrl Commander table: players, zones, cards, and the
// turn/phase/step state machine. It is server-side only; the client
// never imports this package — the client sees a serialised view via
// the v0 protocol (docs/protocol.md). Rules enforcement is deliberately
// absent at S02; mutations are accepted at face value and players
// resolve rules manually until the S13+ B→C rules graft track begins.
package game

import "errors"

var (
	// ErrGameNotInLobby is returned when an operation valid only in the
	// lobby state (like AddPlayer) is attempted on an active or ended game.
	ErrGameNotInLobby = errors.New("game: not in lobby state")

	// ErrGameAlreadyStarted is returned when Start is called on a game
	// that is not in the lobby state.
	ErrGameAlreadyStarted = errors.New("game: already started")

	// ErrGameNotActive is returned when an operation valid only during an
	// active game is attempted in lobby or ended state.
	ErrGameNotActive = errors.New("game: not in active state")

	// ErrGameFull is returned when AddPlayer would exceed MaxPlayers.
	ErrGameFull = errors.New("game: full")

	// ErrNotEnoughPlayers is returned when Start is called with fewer
	// than MinPlayers seats filled.
	ErrNotEnoughPlayers = errors.New("game: not enough players")

	// ErrEmptyName is returned when a player is added without a name.
	ErrEmptyName = errors.New("game: player name cannot be empty")

	// ErrEmptyDeck is returned when a player is added without a deck.
	ErrEmptyDeck = errors.New("game: deck cannot be empty")

	// ErrCardNotFound is returned when a zone operation references a
	// card instance that does not exist in the zone.
	ErrCardNotFound = errors.New("game: card instance not found in zone")

	// ErrZoneEmpty is returned from Top/Bottom/PopTop on an empty zone.
	ErrZoneEmpty = errors.New("game: zone is empty")

	// ErrPlayerNotFound is returned when a mutation references a player
	// ID that is not seated at the table.
	ErrPlayerNotFound = errors.New("game: player not found")

	// ErrZoneNotFound is returned when a ZoneRef cannot be resolved to
	// a concrete Zone on the game (unknown kind or unknown owner).
	ErrZoneNotFound = errors.New("game: zone not found")

	// ErrInvalidParam is returned when an action's parameters are
	// rejected for reasons not covered by a more specific error (e.g.
	// negative hand size on Mulligan).
	ErrInvalidParam = errors.New("game: invalid parameter")

	// ErrPlayerEliminated is returned when a mutation targets a player
	// whose Eliminated flag is set, or when an eliminated player tries
	// to concede a second time.
	ErrPlayerEliminated = errors.New("game: player is eliminated")

	// ErrWrongStep is returned when a step-gated mutation is attempted
	// in a step that doesn't permit it (e.g. declare_attacker outside
	// the declare_attackers step). Added in S08.
	ErrWrongStep = errors.New("game: action not legal in current step")

	// ErrNotACreature is returned when a card-targeting mutation
	// requires a creature but the supplied card isn't one (e.g.
	// declaring a land as an attacker). Added in S08.
	ErrNotACreature = errors.New("game: card is not a creature")

	// ErrCardCallerMismatch is returned when a seated player issues a
	// card-instance-scoped action (tap, move_card, add_counter,
	// set_battlefield_position, declare_attacker, declare_blocker) on
	// a card whose Controller is a different seat. Admin / spectator
	// callers (Caller == uuid.Nil) bypass this check. Added in S08.5.
	ErrCardCallerMismatch = errors.New("game: caller does not control this card")
)
