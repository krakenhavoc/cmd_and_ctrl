package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leyline Surge — Enchantment, the BACK FACE of Invasion of Shandalar
// (face 1):
//
//	"At the beginning of your upkeep, you may put a permanent card from
//	 your hand onto the battlefield."
//
// Registered under "<oracle_id>#1". Any permanent card, so an Aura asks
// what it enchants as it enters (CR 303.4f) and a nonpermanent card is
// not offered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     invasionOfShandalarOracleID + "#1",
		Name:         "Leyline Surge",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Leyline Surge — you may put a permanent card from your hand onto the battlefield",
				Do(PutFromHandOntoBattlefield{
					Optional: true,
					Label:    "Leyline Surge — you may put a permanent card from your hand onto the battlefield",
				})),
		},
	})
}
