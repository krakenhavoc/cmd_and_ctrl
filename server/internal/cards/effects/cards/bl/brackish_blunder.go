package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brackish Blunder — Instant {1}{U} (The Lost Caverns of Ixalan, #4470):
//
//	"Return target creature to its owner's hand. If it was tapped, create
//	 a Map token."
//
// The tapped state is read from the creature as it resolves, before it
// moves (CR 608.2h: it is the creature's last state on the battlefield
// that "was tapped" asks about). The spell's controller creates the Map,
// not the creature's. A target that has become illegal makes the whole
// spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4ad34f85-a185-458a-a7a1-eb60f7efce5b",
		Name:         "Brackish Blunder",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			tapped := 0
			if c, ok := ctx.Game.LookupCardForEffect(t.ID); ok && c.Tapped {
				tapped = 1
			}
			return effects.BounceToHand{Target: t.ID, ThenTail: ctx.TailN(brackishBlunderMap, tapped)}.Apply(ctx)
		},
	})
}

// brackishBlunderMap is "if it was tapped, create a Map token": the state
// was read before the creature moved, and rides the tail as its N, so the
// Map follows the return whatever the return did (CR 608.2c). An on-disk key:
// never renamed, never reused.
var brackishBlunderMap = effects.CardTail("then/brackish-blunder-map", func(ctx *effects.Context, o game.TailOutcome) error {
	if o.Tail.N != 1 {
		return nil
	}
	return effects.CreateToken{Template: effects.MapToken(), N: 1}.Apply(ctx)
})
