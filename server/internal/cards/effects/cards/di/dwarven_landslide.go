package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dwarven Landslide — Sorcery {3}{R} (Apocalypse, #5614):
//
//	"Kicker—{2}{R}, Sacrifice a land. (You may pay {2}{R} and sacrifice a land
//	 in addition to any other costs as you cast this spell.)
//	 Destroy target land. If this spell was kicked, destroy another target
//	 land."
//
// The kicker is one optional cost with a mana half and a sacrifice half,
// paid together on casting. The second target exists only when kicked and must
// differ from the first. The sacrificed land is gone before targets are
// checked on resolution, and each target is checked on its own (CR 608.2b).
//
// No simplification.
func init() {
	land := func() *game.TargetSpec { return effects.TargetPermanent("target land", effects.Land()) }
	kicker := effects.KickerSacrifice("a land", effects.Land())
	kicker.ManaCost = "{2}{R}"
	kicker.Label = "Kicker—{2}{R}, Sacrifice a land"
	effects.Register(effects.Spec{
		OracleID:      "5231a7d6-b6cf-4fa8-8da8-f0ecf023e054",
		Name:          "Dwarven Landslide",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.WhenPaid(kicker, effects.Clauses(land(), effects.Distinct(effects.TargetPermanent("another target land", effects.Land()))))},
		Targets:       land(),
		Purpose:       game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for i := 0; i < 2; i++ {
				if i == 1 && !ctx.WasKicked() {
					break
				}
				if t, ok := ctx.ClauseTarget(i); ok && t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			ctx.Game.DestroyPermanentsForEffect(ids)
			return nil
		},
	})
}
