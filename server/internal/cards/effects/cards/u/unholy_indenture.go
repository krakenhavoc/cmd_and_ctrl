package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unholy Indenture — Enchantment — Aura {2}{B} (Core Set 2020, #5518):
//
//	"Enchant creature
//	 When enchanted creature dies, return that card to the battlefield under
//	 your control with a +1/+1 counter on it."
//
// Minion's Return with the counter put on as the creature enters (so it is
// there before any enters trigger sees it). "Your control" is the Aura's
// controller; the card comes back as a new object (CR 400.7) and a token or a
// card that has left the graveyard returns nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "df8fe368-00eb-42fa-89d6-794fc61f8f8c",
		Name:         "Unholy Indenture",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Unholy Indenture — return that card to the battlefield under your control with a +1/+1 counter on it",
				effects.OldB155ReturnDiedCardWithCounter),
		},
	})
}
