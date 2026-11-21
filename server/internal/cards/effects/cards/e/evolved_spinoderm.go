package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Evolved Spinoderm — Creature — Phyrexian Beast {2}{G}{G}, 5/5:
//
//	"This creature enters with four oil counters on it.
//	 This creature has trample as long as it has two or fewer oil
//	 counters on it. Otherwise, it has hexproof.
//	 At the beginning of your upkeep, remove an oil counter from this
//	 creature. Then if it has no oil counters on it, sacrifice it."
//
// The two statics are complementary layer-6 grants gated on the live oil
// count: trample at 0-2 counters, hexproof at 3 or more. The upkeep trigger
// reads "no oil counters" after removing one, so it is sacrificed at the
// fourth upkeep.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "482ec736-7d4a-49c6-a757-e134123329ca",
		Name:         "Evolved Spinoderm",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.S82OneB12EntersWithOil(4, "Evolved Spinoderm enters with four oil counters on it")},
		Static: []game.StaticAbility{
			effects.S82OneB12GrantsWhileOil("trample", "This creature has trample as long as it has two or fewer oil counters on it.",
				func(oil int) bool { return oil <= 2 }),
			effects.S82OneB12GrantsWhileOil("hexproof", "Otherwise, this creature has hexproof (three or more oil counters).",
				func(oil int) bool { return oil > 2 }),
		},
		Triggered: []game.TriggeredAbility{
			effects.S82OneB12OilUpkeep("Evolved Spinoderm — remove an oil counter; if none, sacrifice it",
				effects.SacrificeThisIfStillOnBattlefield),
		},
	})
}
