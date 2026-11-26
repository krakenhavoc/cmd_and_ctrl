package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vent Sentinel — Creature — Elemental {3}{R}, 2/4 (Conspiracy, S82 OLD batch
// 364, #5557):
//
//	"Defender
//	 {1}{R}, {T}: This creature deals damage to target player or planeswalker
//	 equal to the number of creatures you control with defender."
//
// The amount is counted as the ability resolves (CR 608.2c), through the
// layered view, and includes the Sentinel itself while it is on the
// battlefield; if it has left, the count it last had is not kept, only the
// creatures still there (the damage source is its last-known information,
// CR 608.2h). A target that became illegal means the ability does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "38013607-dd80-4eb9-b0b6-f07e6023c0d0",
		Name:            "Vent Sentinel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, {T}: This creature deals damage to target player or planeswalker equal to the number of creatures you control with defender.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.TapCost()),
			Targets: effects.TargetPlayerOrPlaneswalker(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				n := effects.CountYouControl(ctx, effects.And(effects.Creature(), effects.HasKeyword("defender")))
				legal := ctx.LegalTargets()
				if len(legal) == 0 {
					return nil
				}
				return effects.DealDamage{Source: item.SourceCardID, Target: legal[0].ID, Amount: n}.Apply(ctx)
			},
		}},
	})
}
