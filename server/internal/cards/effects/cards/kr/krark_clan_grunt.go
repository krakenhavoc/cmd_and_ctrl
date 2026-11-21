package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krark-Clan Grunt — Creature — Goblin Warrior {2}{R}, 2/2 (Mirrodin, #5606):
//
//	"Sacrifice an artifact: This creature gets +1/+0 and gains first
//	 strike until end of turn."
//
// Megatog's row at +1/+0 and first strike: an artifact sacrificed as
// the cost, and the boost and the grant both on the source as it
// resolves (mrdB5SacrificeAnArtifactFor).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e7be5ab3-a11f-4786-9b0b-1daa94f59440",
		Name:         "Krark-Clan Grunt",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.MrdB5SacrificeAnArtifactFor("Krark-Clan Grunt", 1, 0, "first strike")},
	})
}
