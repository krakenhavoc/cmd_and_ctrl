package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sea Gate Colossus — Artifact Creature — Golem Warrior {7}, 7/5 (Zendikar
// Rising, #5508):
//
//	"This spell costs {1} less to cast for each creature in your party.
//	 (Your party consists of up to one each of Cleric, Rogue, Warrior, and
//	 Wizard.)"
//
// Thwart the Grave's reduction: the party is counted for the caster when
// the cost is worked out (CR 700.8a, CR 601.2f), reducing generic mana only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "ef7f6766-e878-40c6-aa9a-da7415cc757a",
		Name:              "Sea Gate Colossus",
		Completeness:      effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{effects.S82OldB097PartyCostReduction()},
	})
}
