package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seal from Existence — Enchantment {1}{W}{W} (March of the Machine,
// #5477):
//
//	"Ward {3}
//	 When this enchantment enters, exile target nonland permanent an
//	 opponent controls until this enchantment leaves the battlefield."
//
// Banishing Light's two rows (CR 610.3, exileUntilThisLeavesTriggers)
// plus the shared ward trigger. If the enchantment leaves before the
// trigger resolves nothing is exiled (CR 610.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "72c4c682-df97-4408-997b-e849f54aba39",
		Name:         "Seal from Existence",
		Completeness: CompletenessFull,
		Triggered: append([]game.TriggeredAbility{
			Ward(WardMana("{3}"), "Seal from Existence — ward {3}"),
		}, exileUntilThisLeavesTriggers("Seal from Existence")...),
	})
}
