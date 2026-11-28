package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Katabatic Winds — Enchantment {2}{G} (Visions, S82 older sets batch 791,
// #5631):
//
//	"Phasing (This phases in or out before you untap during each of your untap
//	 steps. While it's phased out, it's treated as though it doesn't exist.)
//	 Creatures with flying can't attack or block, and their activated abilities
//	 with {T} in their costs can't be activated."
//
// Phasing is a printed keyword (CR 702.26a), so the Winds is out every other
// turn of its controller and binds nothing then. While in, every creature with
// flying, whoever controls it, can't attack or block (layer 6, judged on the
// flying the layers leave, CR 508.1c, 509.1b), and none of its activated
// abilities with {T} in the cost can be activated, mana abilities included
// (CR 602.5). The tap bit rides the activation query (game.ActivationAbility.
// TapCost, #5631), so an ability without {T} in its cost is unaffected.
//
// No simplification.
func init() {
	const label = "Creatures with flying can't attack or block, and their activated abilities with {T} in their costs can't be activated."
	effects.Register(effects.Spec{
		OracleID:        "70446c56-07de-4d7f-94ae-4da88fba41cf",
		Name:            "Katabatic Winds",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordPhasing},
		Static:          []game.StaticAbility{effects.Old791FlyingCreaturesCantAttackOrBlock(label)},
		ActivationRestrictions: []game.ActivationRestriction{
			effects.Old791FlyingCreaturesCantActivateTapAbilities(label),
		},
	})
}
