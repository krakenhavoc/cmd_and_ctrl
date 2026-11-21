package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Limits of Solidarity — Sorcery {3}{R}
// (Amonkhet, #5535):
//
//	"Gain control of target creature until end of turn. Untap that creature. It gains haste until end of turn.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The threaten body Take for a Ride shares, plus cycling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a2184c37-ee92-40f2-b7d1-97c39ce5f277",
		Name:         "Limits of Solidarity",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || !ctx.IsTargetLegal(t) {
				return nil
			}
			return effects.StdB31Threaten(ctx, t, "Limits of Solidarity")
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
