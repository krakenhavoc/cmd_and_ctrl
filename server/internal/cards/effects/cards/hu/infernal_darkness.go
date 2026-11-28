package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infernal Darkness — Enchantment {2}{B}{B} (Ice Age, S82 OLD batch 839, #5636):
//
//	"Cumulative upkeep—Pay {B} and 1 life.
//	 If a land is tapped for mana, it produces {B} instead of any other type."
//
// Cumulative upkeep with a mana-and-life payment (ManaAndLifePayment, #6189),
// both halves multiplied by the age counters. The land half is Contamination's
// replacement on the mana production event for every land tapped for mana; the
// amount of mana is unchanged.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dd821fca-79a0-48a9-b1cf-5a496bc04f65",
		Name:         "Infernal Darkness",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeepPaying("Infernal Darkness — cumulative upkeep—pay {B} and 1 life", effects.ManaAndLifePayment("{B}", 1), "", ""),
		},
		Replacements: []game.ReplacementEffect{
			effects.OldB839LandsProduce("Infernal Darkness: a land tapped for mana produces {B}", "B"),
		},
	})
}
