package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arctic Aven — Creature — Bird Wizard {2}{U}, 2/1 (Magic 2013, #5568):
//
//	"Flying
//	 This creature gets +1/+1 as long as you control a Plains.
//	 {W}: This creature gains lifelink until end of turn. (Damage dealt by
//	 this creature also causes you to gain that much life.)"
//
// Kird Ape's land-type pump (any permanent you control with the Plains
// subtype counts) and Heron of Hope's lifelink for mana, which can be
// activated any number of times, the turn it arrives included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "616cb1be-074a-46c2-b8f1-6b9f15b7bc23",
		Name:            "Arctic Aven",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 as long as you control a Plains.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfWhileYouControlA(effects.Subtype("Plains")),
			Apply:     effects.PumpBy(1, 1),
		}},
		Activated: []effects.ActivatedAbility{effects.ConB1ThisGainsRow("{W}", "lifelink", "Arctic Aven", game.AnswerCombatGrant)},
	})
}
