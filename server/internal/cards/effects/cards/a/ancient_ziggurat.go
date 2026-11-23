package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ancient Ziggurat — Land:
//
//	"{T}: Add one mana of any color. Spend this mana only to cast a
//	 creature spell."
//
// Beastcaller Savant's restricted mana on a land: the mana carries the
// CR 106.12 spend restriction, so it pays for a creature spell and for
// nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "0baabe39-72ae-47bd-a095-cbf7eb8a6361",
		Name:          "Ancient Ziggurat",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.Cmdr2B01TapForCreatureOnlyMana()},
	})
}
