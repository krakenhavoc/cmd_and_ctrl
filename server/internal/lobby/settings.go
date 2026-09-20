package lobby

// settings.go: the lobby's half of the table-settings surface
// (ADR 0075 §2.3).
//
// Two surfaces change one struct. PATCH /games/{id}/settings is this
// one, and it works in the lobby AND on a live game — the split in the
// ADR's prose ("lobby phase: PATCH; mid-game: the action") is about
// where a client naturally is, not about what the server accepts. A
// host sitting on the lobby page of a running table should not have to
// open the board to turn undos back on.
//
// The mutation itself is the engine's (game.UpdateSettings). What this
// file adds is what the engine deliberately does not have: the room, so
// the change lands in the replay stream and reaches every connected
// client, and the lobby's lock, so it is serialised against a join or a
// start happening at the same moment.
//
// It does NOT decide who may call it. That is CanManageTable, asked at
// the HTTP edge before this is reached (host.go).

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// UpdateSettings applies patch to the game's table settings and
// broadcasts the result. actor is the player who asked, uuid.Nil for
// the server admin; it is recorded on the EventSettingsChanged the
// engine emits, which is what lets the log name them.
//
// Errors are the engine's verbatim — game.ErrInvalidSetting for a
// value out of range, game.ErrStartingLifeLocked for a starting-life
// change after Start — so the HTTP layer can map them to 400 and 422
// without this layer inventing its own vocabulary.
//
// Deliberately routed through applyLocked (ApplyExternal) rather than
// Apply: a settings change is not a play and mints no undo entry
// (ADR 0075 §2.3). An undo must not be able to take back the change
// that was meant to stop it.
func (l *Lobby) UpdateSettings(id, actor uuid.UUID, patch game.SettingsPatch) (game.TableSettings, error) {
	// Registered before the lock defer so it runs after l.mu is
	// released — see applyLocked.
	var broadcast func()
	defer func() {
		if broadcast != nil {
			broadcast()
		}
	}()
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.games[id]
	if !ok {
		return game.TableSettings{}, ErrGameNotFound
	}

	var err error
	if broadcast, err = l.applyLocked(id, entry, func() error {
		return entry.room.Game.UpdateSettings(actor, patch)
	}); err != nil {
		return game.TableSettings{}, err
	}
	// Read back rather than reconstruct from the patch: the answer is
	// the whole struct, including the fields the patch left alone, and
	// including any the engine is entitled to adjust on its way
	// through.
	return entry.room.Game.TableSettingsSnapshot(), nil
}
