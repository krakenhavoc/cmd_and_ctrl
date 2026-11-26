package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hubris — Instant {1}{U} (Journey into Nyx, S82 OLD, #5558):
//
//	"Return target creature and all Auras attached to it to their owners'
//	 hands."
//
// The creature and every Aura attached to it as the spell resolves leave the
// battlefield as one simultaneous move to their owners' hands (CR 400.7), so
// an Aura is never left to fall off first and go to the graveyard. An Aura is
// read on the layered view, so a Bestow creature that is an Aura counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "95fbd573-c04c-408a-a8c3-f7e1ad5ddac0",
		Name:         "Hubris",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			ids := []uuid.UUID{t.ID}
			for _, a := range ctx.Game.BattlefieldCardsForEffect() {
				if a.IsAttachedTo(t.ID) && a.IsAura() {
					ids = append(ids, a.InstanceID)
				}
			}
			return ctx.Game.BounceCardsToHandKeyedForEffect(ids, game.KeyedTail{})
		},
	})
}
