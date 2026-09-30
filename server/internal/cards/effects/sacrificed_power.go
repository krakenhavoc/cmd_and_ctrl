package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sacrificed_power.go — "the sacrificed creature's power" (Greater
// Good, Disciple of Bolas, Jarad, Golgari Lich Lord).
//
// The creature is gone by the time the ability resolves, so the power
// is its last-known information (CR 608.2h): the permanent as it last
// existed on the battlefield. Every battlefield exit records that
// object — post-layer characteristics, counters included — before it
// moves (game/permanent_lki.go, battlefieldExitLocked), so an anthem's
// bonus counts exactly as a +1/+1 counter does. The older read, printed
// power plus counters, missed the anthem.

// departedCreaturePower is the power of the permanent object `cardID`
// most recently was: the live permanent while it is still on the
// battlefield, its last-known information once it has left this turn.
// Clamped at zero, because a negative power draws and drains nothing.
// Zero when the engine has no record of it as a permanent this turn.
//
// Caller holds g.mu in write mode.
func departedCreaturePower(g *game.Game, cardID uuid.UUID) int {
	ref, ok := g.PermanentRefForEffect(cardID)
	if !ok {
		return 0
	}
	info, ok := g.PermanentForEffect(ref)
	if !ok || info.Power < 0 {
		return 0
	}
	return info.Power
}
