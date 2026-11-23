package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// MechtitanLeft is the leaves-the-battlefield match for a delayed trigger
// that watches one object (the trigger's Params.Object): it fires when that
// permanent, in the same incarnation, leaves. Mechtitan Core and Ugin, the
// Ineffable share it.
func MechtitanLeft(ev game.Event, _ *game.DelayedTrigger, g *game.Game, p game.EffectParams) []uuid.UUID {
	if p.Object.ID == uuid.Nil || ev.Kind != game.EventLTB || ev.CardID != p.Object.ID {
		return nil
	}
	ref, ok := g.PermanentRefForEffect(ev.CardID)
	if !ok || ref.Epoch != p.Object.Epoch {
		return nil
	}
	return []uuid.UUID{ev.CardID}
}
