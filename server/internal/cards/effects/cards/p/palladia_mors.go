package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Palladia-Mors — Legendary Creature — Elder Dragon
// {2}{R}{R}{G}{G}{W}{W}, 7/7 (Legends, #5639):
//
//	"Flying, trample
//	 At the beginning of your upkeep, sacrifice Palladia-Mors unless you
//	 pay {R}{G}{W}."
//
// Flying and trample are printed keywords. The upkeep payment is
// UpkeepSacrificeUnlessYouPay, the Elder Dragons' shared shape: pay
// {R}{G}{W} or sacrifice it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d7121ac7-e425-46cd-b006-1af391d97f87",
		Name:            "Palladia-Mors",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.UpkeepSacrificeUnlessYouPay("Palladia-Mors", "{R}{G}{W}"),
		},
	})
}
