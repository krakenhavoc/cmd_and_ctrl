package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wrath of Marit Lage — Enchantment {3}{U}{U} (Eighth Edition, S82 older sets batch 619, #5607):
//
//	"When this enchantment enters, tap all red creatures.
//	 Red creatures don't untap during their controllers' untap steps."
//
// Back to Basics' untap lock with a red-creature test, and a tap-all at entry
// (Naya Charm's walk: snapshot the untapped red creatures, then tap each). Red
// is read through the layered view, so a creature another effect made red is
// locked down.
//
// No simplification.
func init() {
	red := effects.And(effects.Creature(), effects.OfColor("R"))
	effects.Register(effects.Spec{
		OracleID:              "740cafc6-9521-403a-ab5f-11582bf55af0",
		Name:                  "Wrath of Marit Lage",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringTheirControllersUntapSteps(red)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Wrath of Marit Lage — tap all red creatures", func(g *game.Game, item *game.StackItem) error {
				return effects.S82OldB619TapAllMatching(effects.NewContext(g, item), red)
			}),
		},
	})
}
