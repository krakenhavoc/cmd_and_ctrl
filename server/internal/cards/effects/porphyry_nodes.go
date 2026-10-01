package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Porphyry Nodes — Enchantment {W}:
//
//	"At the beginning of your upkeep, destroy the creature with the
//	 least power. It can't be regenerated. If two or more creatures are
//	 tied for least power, you choose one of them.
//	 When there are no creatures on the battlefield, sacrifice this
//	 enchantment."
//
// ADR 0107 PR 1 (#1858). Drop of Honey's text in white: the same
// least-power upkeep (destroyCreatureWithLeastPower) and the same CR
// 603.8 state trigger over every creature on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7cfeb732-8a35-4a1f-a417-11db4c1499fd",
		Name:         "Porphyry Nodes",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Porphyry Nodes — destroy the creature with the least power", destroyCreatureWithLeastPower),
			WhenThereAreNo(QueryType("creature"), "Porphyry Nodes — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
	})
}
