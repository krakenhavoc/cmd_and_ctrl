package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dragonloft Idol — Artifact Creature — Gargoyle {4}, 3/3 (Jumpstart,
// #5510):
//
//	"As long as you control a Dragon, this creature gets +1/+1 and has flying
//	 and trample."
//
// Kargan Dragonrider's condition on a layer-7c pump and two layer-6 grants.
// The Idol itself is never a Dragon, and a changeling counts.
//
// No simplification.
func init() {
	whileDragon := effects.SelfWhileYouControlA(effects.Subtype("Dragon"))
	effects.Register(effects.Spec{
		OracleID:     "736ce43b-28cd-4b8b-9936-3e3975a56b6a",
		Name:         "Dragonloft Idol",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "Gets +1/+1 as long as you control a Dragon.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: whileDragon,
				Apply:     effects.PumpBy(1, 1),
			},
			effects.KeywordGrant(whileDragon, "flying"),
			effects.KeywordGrant(whileDragon, "trample"),
		},
	})
}
