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
// value out of range, game.ErrStartingLifeLocked,
// game.ErrFreeMulliganLocked or game.ErrPlanechaseLocked for a change after Start — so the HTTP layer can map them to 400 and 422
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
	// ADR 0216 owner answer 3: the option is offered where the
	// deployment says so, and only there. This is the door POST /games,
	// PATCH /games/{id}/settings and a scheduled table's opening all go
	// through; the hub refuses set_table_settings the same way.
	if patch.TurnsOnPlanechase() && !l.PlanechaseOffered() {
		return game.TableSettings{}, game.ErrPlanechaseNotOffered
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.games[id]
	if !ok {
		return game.TableSettings{}, ErrGameNotFound
	}

	// ADR 0222 §1: a format is chosen when the table is made. It may
	// still change in the lobby, but not once a seat has a deck, which
	// was checked against the format the table had.
	if patch.Format != nil && patch.Format.Spec().Format != entry.room.Game.TableSettingsSnapshot().EffectiveFormat() {
		for _, seat := range entry.meta.Players {
			if seat.DeckUploaded {
				return game.TableSettings{}, ErrFormatDecksLoaded
			}
		}
	}

	var err error
	if broadcast, err = l.applyLocked(id, entry, func() error {
		return entry.room.Game.UpdateSettings(actor, patch)
	}); err != nil {
		return game.TableSettings{}, err
	}
	syncFormatLocked(entry)
	// ADR 0211 §1: the host's change keeps the table open; the
	// server admin's (actor uuid.Nil) does not.
	if actor != uuid.Nil {
		l.touchLocked(entry)
	}
	// Read back rather than reconstruct from the patch: the answer is
	// the whole struct, including the fields the patch left alone, and
	// including any the engine is entitled to adjust on its way
	// through.
	return entry.room.Game.TableSettingsSnapshot(), nil
}

// PlanechaseOffered reports whether this deployment offers the
// Planechase table option (ADR 0216, owner answer 3).
func (l *Lobby) PlanechaseOffered() bool { return l.mgr.PlanechaseOffered() }

// Settings reads the game's table settings without changing them, for
// GET /games/{id}/settings (#3499). Who may ask is the HTTP edge's
// question (canReadTableSettings), as it is for UpdateSettings.
func (l *Lobby) Settings(id uuid.UUID) (game.TableSettings, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.games[id]
	if !ok {
		return game.TableSettings{}, ErrGameNotFound
	}
	return entry.room.Game.TableSettingsSnapshot(), nil
}

// syncFormatLocked copies the table's format and its seat cap from the
// engine's settings onto the entry's meta (ADR 0222 §1). The engine is
// the authority; the meta is a cache for the table list and the join
// page. A Commander table's meta carries neither field. Callers hold
// l.mu.
func syncFormatLocked(entry *gameEntry) {
	settings := entry.room.Game.TableSettingsSnapshot()
	spec := settings.FormatSpec()
	entry.meta.Format, entry.meta.MaxSeats, entry.meta.BestOf = "", 0, 0
	// ADR 0222 §3: and the match length, for a best-of-three.
	if n := settings.EffectiveBestOf(); n > 1 {
		entry.meta.BestOf = n
	}
	if spec.Format != game.FormatCommander {
		entry.meta.Format = string(spec.Format)
	}
	if spec.MaxSeats != game.MaxPlayers {
		entry.meta.MaxSeats = spec.MaxSeats
	}
}

// Format is the table's format, for the deck routes: a deck is checked
// against the format of the table it is seated at (ADR 0222 §2).
func (l *Lobby) Format(id uuid.UUID) (game.Format, error) {
	s, err := l.Settings(id)
	if err != nil {
		return "", err
	}
	return s.EffectiveFormat(), nil
}
