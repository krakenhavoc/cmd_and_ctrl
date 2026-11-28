package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Breeding Pit — Enchantment {3}{B} (Fallen Empires, S82 OLD b856, #5637):
//
//	"At the beginning of your upkeep, sacrifice this enchantment unless you pay
//	 {B}{B}.
//	 At the beginning of your end step, create a 0/1 black Thrull creature token."
//
// The shared "sacrifice unless you pay" upkeep trigger (Molting Harpy's), and an
// end-step token (Thrull is the printed 0/1 black creature token). One trigger each.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ddc83b35-cf7c-495b-bd30-e409b622e299",
		Name:         "Breeding Pit",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.UpkeepSacrificeUnlessYouPay("Breeding Pit", "{B}{B}"),
			effects.TriggerWithPurpose(effects.AtYourEndStep("Breeding Pit — create a 0/1 black Thrull creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("0/1 black Thrull"), N: 1})), game.Purpose{Tokens: 1}),
		},
	})
}
