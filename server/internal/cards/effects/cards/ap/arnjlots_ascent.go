package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arnjlot's Ascent — Enchantment {1}{U}{U} (Ice Age, S82 OLD batch 839, #5636):
//
//	"Cumulative upkeep {U}
//	 {1}: Target creature gains flying until end of turn."
//
// Updraft's flying grant as an activated ability: a layer 6 keyword grant that
// ends at cleanup (CR 514.2), on any creature, any controller's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f2cedb38-12b0-4489-a62b-9b6a3776d695",
		Name:         "Arnjlot's Ascent",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.CumulativeUpkeep("Arnjlot's Ascent — cumulative upkeep {U}", "{U}")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: Target creature gains flying until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    effects.ManaCost("{1}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("flying", "Arnjlot's Ascent — flying until end of turn"),
		}},
	})
}
