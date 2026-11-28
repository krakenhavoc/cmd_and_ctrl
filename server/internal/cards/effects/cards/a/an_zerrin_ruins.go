package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// An-Zerrin Ruins — Enchantment {2}{R}{R} (Homelands, S82 OLD b827, #5635):
//
//	"As this enchantment enters, choose a creature type.
//	 Creatures of the chosen type don't untap during their controllers' untap
//	 steps."
//
// Cover of Darkness' stored "as this enters" choice (Card.NamedTribe), read by
// Back to Basics' untap-step restriction: every creature of the chosen type,
// whoever controls it and read through the layers (a changeling counts), stays
// tapped in its controller's untap step. Before a type is chosen it holds
// nothing back.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f805b100-49b8-4d23-a768-e8de8bb0daa3",
		Name:         "An-Zerrin Ruins",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseCreatureTypeAsEnters("An-Zerrin Ruins — choose a creature type"),
		UntapStepRestrictions: []game.UntapStepRestriction{{
			Label:     "Creatures of the chosen type don't untap during their controllers' untap steps.",
			Restricts: effects.TribeFilter{Chosen: true}.Matches,
		}},
	})
}
