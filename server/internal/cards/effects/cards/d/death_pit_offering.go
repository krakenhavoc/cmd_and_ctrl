package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Death Pit Offering — Enchantment {2}{B}{B} (Eighth Edition, S82 older sets batch 619, #5607):
//
//	"When this enchantment enters, sacrifice all creatures you control.
//	 Creatures you control get +2/+2."
//
// Yukora's sacrifice of everything you control matching a predicate, and Gaea's
// Anthem's static pump. The creatures are sacrificed at once as the trigger
// resolves, so the enchantment's own anthem does not save any of them.
//
// No simplification.
func init() {
	mine := effects.And(effects.Creature(), effects.YouControl())
	effects.Register(effects.Spec{
		OracleID:     "d445032b-b243-4203-9c88-a0ac357fb9b5",
		Name:         "Death Pit Offering",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Death Pit Offering — sacrifice all creatures you control", func(g *game.Game, item *game.StackItem) error {
				return effects.B08SacrificeAllMatching(effects.NewContext(g, item), mine)
			}),
		},
		Static: []game.StaticAbility{effects.Labelled(effects.B16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.Controller == source.Controller
		}, 2, 2), "Creatures you control get +2/+2.")},
	})
}
