package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Backwoods Survivalists — Creature — Human Warrior {3}{G}, 4/3 (Eldritch
// Moon, #5541):
//
//	"Delirium — This creature gets +1/+1 and has trample as long as there
//	 are four or more card types among cards in your graveyard."
//
// Wildfire Wickerfolk's two delirium statics without its haste: a layer
// 7c pump and a layer-6 keyword grant, both read off the controller's
// graveyard on every recompute (drcDeliriumMet).
//
// No simplification.
func init() {
	const delirium = " as long as there are four or more card types among cards in your graveyard."
	trample := KeywordGrant(whileYouHave(drcDeliriumMet, selfCreature), "trample")
	trample.Label = "Delirium — Has trample" + delirium
	Register(Spec{
		OracleID:     "d2aecca6-fe03-4686-9daa-68f1f6c73ab4",
		Name:         "Backwoods Survivalists",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Delirium — Gets +1/+1" + delirium,
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: whileYouHave(drcDeliriumMet, selfCreature),
			Apply:     pumpBy(1, 1),
		}, trample},
	})
}
