package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goblin Assault — {2}{R} Enchantment (Shards of Alara, S82 OLD b516):
//
//	"At the beginning of your upkeep, create a 1/1 red Goblin creature token
//	 with haste.
//	 Goblin creatures attack each combat if able."
//
// The attack requirement covers every Goblin on the table, its controller's
// or not, read through the layered view (a changeling counts, CR 205.3m); it
// is a requirement (CR 508.1d), so a Goblin that can't attack is not forced.
//
// No simplification.
func init() {
	attack := effects.AttacksEachCombatWhere(effects.TribeFilter{Tribes: []string{"Goblin"}}.Matches)
	attack.Label = "Goblin creatures attack each combat if able."
	effects.Register(effects.Spec{
		OracleID:     "d1255776-b88b-4f05-9b71-ff89a2eba453",
		Name:         "Goblin Assault",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Goblin Assault — create a 1/1 red Goblin creature token with haste",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 red Goblin with haste"), N: 1})),
		},
		Static: []game.StaticAbility{attack},
	})
}
