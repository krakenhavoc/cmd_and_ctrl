package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tibor and Lumia — Legendary Creature — Human Wizard {2}{U}{R}, 3/3
// (Guildpact):
//
//	"Whenever you cast a blue spell, target creature gains flying until end
//	 of turn.
//	 Whenever you cast a red spell, Tibor and Lumia deals 1 damage to each
//	 creature without flying."
//
// Two triggers, one per printed ability; a blue and red spell triggers both.
// The colour of the spell is read through the layers as it is cast (a
// multicolour spell is both). The damage is dealt by Tibor and Lumia to every
// creature without flying as the ability resolves, itself included, in one
// damage instance.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ee6d6f2f-68d5-466a-86c6-20fdb360e629",
		Name:         "Tibor and Lumia",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverYouCast(effects.OfColor("U"), "Tibor and Lumia — target creature gains flying until end of turn",
					effects.NeoB07TargetGetsUntilEOT("Tibor and Lumia — flying until end of turn", 0, 0, "flying")),
				effects.TargetCreature("target creature"),
			),
			effects.TriggerWithPurpose(
				effects.WheneverYouCast(effects.OfColor("R"), "Tibor and Lumia — 1 damage to each creature without flying",
					func(g *game.Game, item *game.StackItem) error {
						return effects.DamageEachMatching(effects.NewContext(g, item), effects.And(effects.Creature(), effects.WithoutKeyword("flying")), 1)
					}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}}),
		},
	})
}
