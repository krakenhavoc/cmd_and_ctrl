package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wildfire Wickerfolk — Artifact Creature — Scarecrow {R}{G}, 3/2
// (Duskmourn: House of Horror, #4466):
//
//	"Haste
//	 Delirium — This creature gets +1/+1 and has trample as long as there
//	 are four or more card types among cards in your graveyard."
//
// Delirium is an ability word (CR 207.2c). The bonus is Spineseeker
// Centipede's: a layer 7c modification and a layer 6 grant on the
// Wickerfolk alone (selfCreature), both gated on the controller's
// graveyard (drcDeliriumMet) and re-read on every recompute, so they turn
// on and off together as the graveyard changes.
//
// No simplification.
func init() {
	const delirium = " as long as there are four or more card types among cards in your graveyard."
	pump := game.StaticAbility{
		Label:     "Delirium — Gets +1/+1" + delirium,
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: whileYouHave(drcDeliriumMet, selfCreature),
		Apply:     pumpBy(1, 1),
	}
	trample := KeywordGrant(whileYouHave(drcDeliriumMet, selfCreature), "trample")
	trample.Label = "Delirium — Has trample" + delirium
	Register(Spec{
		OracleID:        "50c7b13d-f217-45de-837c-49f934840368",
		Name:            "Wildfire Wickerfolk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static:          []game.StaticAbility{pump, trample},
	})
}
