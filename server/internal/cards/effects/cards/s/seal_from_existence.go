package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "72c4c682-df97-4408-997b-e849f54aba39",
		Name:         "Seal from Existence",
		Completeness: effects.CompletenessFull,
		Triggered: append([]game.TriggeredAbility{
			effects.Ward(effects.WardMana("{3}"), "Seal from Existence — ward {3}"),
		}, effects.ExileUntilThisLeavesTriggers("Seal from Existence")...),
	})
}
