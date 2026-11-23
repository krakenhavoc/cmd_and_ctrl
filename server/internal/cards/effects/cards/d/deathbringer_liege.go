package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deathbringer Liege — Creature — Horror {2}{W/B}{W/B}{W/B}, 3/4
// (Shadowmoor):
//
//	"Other white creatures you control get +1/+1.
//	 Other black creatures you control get +1/+1.
//	 Whenever you cast a white spell, you may tap target creature.
//	 Whenever you cast a black spell, you may destroy target creature if
//	 it's tapped."
//
// Glen Elendra Liege's two colour anthems and Balefire Liege's cast
// triggers: a white-black spell triggers both, a colourless one neither.
// Each target is chosen as its trigger goes on the stack and the "may" is
// asked as it resolves (CR 603.5); a target that has left is skipped (CR
// 608.2b). The destroy reads the target as the ability resolves: only a
// creature that is tapped then is destroyed. A spell that is both white and
// black triggers both, and the one put on the stack last resolves first, so
// a creature the white trigger taps can still be destroyed by the black one
// if the black trigger resolves after it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "caf89756-fb78-4e58-a730-001393ffa0a8",
		Name:         "Deathbringer Liege",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(effects.G2OtherCreaturesYouControlOfColor("W"), 1, 1),
			effects.B16Anthem(effects.G2OtherCreaturesYouControlOfColor("B"), 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(effects.WheneverYouCast(effects.OfColor("W"), "Deathbringer Liege — you may tap target creature", effects.TapTheFirstLegalTarget),
					"Deathbringer Liege — tap target creature?"),
				effects.TargetCreature("target creature")),
			effects.Targeting(
				effects.Optional(effects.WheneverYouCast(effects.OfColor("B"), "Deathbringer Liege — you may destroy target creature if it's tapped",
					func(g *game.Game, item *game.StackItem) error {
						for _, t := range effects.NewContext(g, item).LegalTargets() {
							if c, ok := g.LookupCardForEffect(t.ID); ok && c.Tapped {
								return effects.DestroyFirstLegalCardTarget(g, item)
							}
						}
						return nil
					}), "Deathbringer Liege — destroy target creature if it's tapped?"),
				effects.TargetCreature("target creature")),
		},
	})
}
