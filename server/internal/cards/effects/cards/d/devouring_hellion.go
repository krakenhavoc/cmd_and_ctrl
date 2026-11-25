package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Devouring Hellion — Creature — Hellion {2}{R}, 2/2 (War of the Spark):
//
//	"As this creature enters, you may sacrifice any number of creatures and/or
//	 planeswalkers. If you do, it enters with twice that many +1/+1 counters on
//	 it."
//
// Devour 2 over creatures and planeswalkers, written out (it is not the keyword,
// the card prints no devour): the sacrifice is a CR 614.12 as-enters choice made
// before the Hellion is on the battlefield, so it is never a candidate, the dies
// triggers of what was sacrificed fire, and the counters ride the entry event so
// doublers apply. Declining enters it with none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f6620c3-2c15-467c-92ca-fdf9042a63f7",
		Name:         "Devouring Hellion",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.DevourOver("Devouring Hellion", 2, "creatures and/or planeswalkers",
				func(c game.Card) bool { return c.IsCreature() || c.IsPlaneswalker() }),
		},
	})
}
