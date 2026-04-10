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
)
