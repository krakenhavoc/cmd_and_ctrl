package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Necron Deathmark — Artifact Creature — Necron {3}{B}{B}, 5/3 (Warhammer
// 40,000 Commander, S82 GAP batch 188, #5948):
//
//	"Flash
//	 Synaptic Disintegrator — When this creature enters, destroy up to one
//	 target creature and target player mills three cards."
//
// Two target clauses, chosen as the trigger goes on the stack (CR 603.3d):
// up to one creature and exactly one player. Each part checks its own
// target as the ability resolves (CR 608.2b): losing the creature does not
// stop the mill.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "45866598-4ee4-46b6-ade2-ffcdbe98e58f",
		Name:            "Necron Deathmark",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Necron Deathmark — destroy up to one target creature and target player mills three cards",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if t, ok := ctx.ClauseTarget(1); ok {
						// The destroy waits for the mill's CR 614 window (a
						// replacement may stop to ask), so nothing is gated on
						// a read taken before it.
						return effects.MillToZone{Player: t.ID, N: 3, ThenTail: ctx.ItemTail(necronDeathmarkAfterMill)}.Apply(ctx)
					}
					return necronDeathmarkDestroy(ctx)
				}),
				effects.Clauses(
					effects.UpToOneTargetCreature("up to one target creature"),
					effects.TargetPlayer("target player"),
				)),
		},
	})
}

// necronDeathmarkDestroy is "destroy up to one target creature" for the
// creature chosen as the trigger went on the stack, if it is still legal.
func necronDeathmarkDestroy(ctx *effects.Context) error {
	if t, ok := ctx.ClauseTarget(0); ok {
		return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
	}
	return nil
}

// necronDeathmarkAfterMill is the destroy half once the mill has settled.
// An on-disk key: never renamed, never reused.
var necronDeathmarkAfterMill = effects.CardTail("necron-deathmark/after-mill", func(ctx *effects.Context, _ game.TailOutcome) error {
	return necronDeathmarkDestroy(ctx)
})
