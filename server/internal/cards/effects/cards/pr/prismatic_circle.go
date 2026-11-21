package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prismatic Circle — Enchantment {2}{W}:
//
//	"Cumulative upkeep {1} (At the beginning of your upkeep, put an age counter on this permanent, then sacrifice it unless you pay its upkeep cost for each age counter on it.)
//	 As this enchantment enters, choose a color.
//	 {1}: The next time a source of your choice of the chosen color would deal damage to you this turn, prevent that damage."
//
// ADR 0107 §6 (#1860): Story Circle's shield (CR 615.8, 609.7a, 615.9)
// for {1}, on an enchantment with cumulative upkeep {1} (CR 702.24).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "022dae2e-7fc3-486e-9168-59652d9ab21c",
		Name:         "Prismatic Circle",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseColorAsEnters(game.ColorForProtection, "Prismatic Circle"),
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Prismatic Circle — cumulative upkeep {1}", "{1}"),
		},
		Activated: []effects.ActivatedAbility{effects.ChosenColorShieldRow(
			"{1}: The next time a source of your choice of the chosen color would deal damage to you this turn, prevent that damage.", effects.ManaCost("{1}"))},
	})
}
