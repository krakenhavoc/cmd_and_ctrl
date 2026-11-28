package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curse of Marit Lage — Enchantment {3}{R}{R} (Ice Age, S82 OLD batch 846, #5636):
//
//	"When this enchantment enters, tap all Islands.
//	 Islands don't untap during their controllers' untap steps."
//
// Wrath of Marit Lage's one-instruction tap of every Island (whoever controls
// it, read through the layers as the trigger resolves) and Rising Waters'
// untap-step restriction, over Islands instead of lands. An Island that
// becomes an Island later is held tapped too, and one that stops being one
// untaps as usual: the restriction reads the type live.
//
// No simplification.
func init() {
	island := effects.HasSubtype("Island")
	effects.Register(effects.Spec{
		OracleID:              "6fde4790-4186-4786-aa1a-b6a53bfadc1c",
		Name:                  "Curse of Marit Lage",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringTheirControllersUntapSteps(island)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Curse of Marit Lage — tap all Islands", func(g *game.Game, item *game.StackItem) error {
				return effects.S82OldB619TapAllMatching(effects.NewContext(g, item), island)
			}),
		},
	})
}
