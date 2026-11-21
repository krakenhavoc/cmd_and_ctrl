package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Koma, World-Eater — Legendary Creature — Serpent {3}{G}{G}{U}{U}, 8/12
// (Foundations, #4465):
//
//	"This spell can't be countered.
//	 Trample, ward {4}
//	 Whenever Koma deals combat damage to a player, create four 3/3 blue
//	 Serpent creature tokens named Koma's Coil."
//
// Koma, Cosmos Serpent's Spec.CantBeCountered and Koma's Coil token, Kappa
// Cannoneer's ward {4} trigger, and Ant-Man's combat-damage trigger making
// four Coils at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1ebf2142-f171-40cd-8143-15f50084b707",
		Name:            "Koma, World-Eater",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{4}"), "Koma, World-Eater — ward {4}"),
			effects.WheneverThisDealsCombatDamageToAPlayer("Koma, World-Eater — create four Koma's Coil tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 blue Serpent named Koma's Coil"), N: 4})),
		},
	})
}
