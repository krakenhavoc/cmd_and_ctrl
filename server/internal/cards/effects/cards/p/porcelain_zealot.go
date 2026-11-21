package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Porcelain Zealot — Creature — Phyrexian Soldier {3}{W}, 2/3
// (Phyrexia: All Will Be One, #5479):
//
//	"At the beginning of combat on your turn, target creature you control
//	 gets +1/+1 until end of turn. If that creature has toxic, instead it
//	 gets +2/+2 until end of turn."
//
// The size is read as the trigger resolves. A target that left in
// response gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c11bda3-903f-4b7c-9e5e-f6c5c91977cf",
		Name:         "Porcelain Zealot",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.AtBeginningOfYourCombat("Porcelain Zealot — target creature you control gets +1/+1, or +2/+2 if it has toxic",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						n := 1
						if effects.S82OneB05HasToxic(g, id) {
							n = 2
						}
						return effects.BoostUntilEOT{Target: id, Power: n, Toughness: n, Label: "Porcelain Zealot — pump"}.Apply(ctx)
					}),
				effects.TargetCreature("target creature you control", effects.YouControl())),
				effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1))),
		},
	})
}
