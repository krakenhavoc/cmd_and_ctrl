package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silkbind Faerie — Creature — Faerie Rogue {2}{W/U}, 1/3 (Shadowmoor, S82 OLD
// b531, #5591):
//
//	"Flying
//	 {1}{W/U}, {Q}: Tap target creature. ({Q} is the untap symbol.)"
//
// {Q} is AbilityCost.UntapSelf (CR 107.6, Order of Whiteclay's cost): the Faerie
// must be tapped to pay it, so it is also the Faerie's only way to use the
// ability once it is untapped. The tap is Arashin Sunshield's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fbf4ce62-26f1-4886-952a-efdc1f4946b6",
		Name:            "Silkbind Faerie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W/U}, {Q}: Tap target creature.",
			Cost:    effects.OldB531UntapCost("{1}{W/U}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
