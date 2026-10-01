package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drop of Honey — Enchantment {G}:
//
//	"At the beginning of your upkeep, destroy the creature with the
//	 least power. It can't be regenerated. If two or more creatures are
//	 tied for least power, you choose one of them.
//	 When there are no creatures on the battlefield, sacrifice this
//	 enchantment."
//
// ADR 0107 PR 1 (#1858). The upkeep trigger reads every creature on the
// battlefield as it resolves, anyone's, and its controller breaks a tie
// (destroyCreatureWithLeastPower). The sacrifice is a CR 603.8 state
// trigger over every creature on the battlefield, so the upkeep that
// destroys the last creature also sends the enchantment away.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "383e9005-5869-4d1d-917d-30e5f214fbd9",
		Name:         "Drop of Honey",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Drop of Honey — destroy the creature with the least power", destroyCreatureWithLeastPower),
			WhenThereAreNo(QueryType("creature"), "Drop of Honey — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
	})
}
