package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gnarlwood Dryad — Creature — Dryad Horror {G}, 1/1 (Duskmourn Commander):
//
//	"Deathtouch
//	 Delirium — This creature gets +2/+2 as long as there are four or more
//	 card types among cards in your graveyard."
//
// Delirium is an ability word (CR 207.2c). The bonus is Spineseeker
// Centipede's: a layer 7c modification on the Dryad alone, gated on the
// controller's own graveyard (drcDeliriumMet) and re-read on every recompute.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e823639f-1308-47f6-9b51-8fd9a682489a",
		Name:            "Gnarlwood Dryad",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Static: []game.StaticAbility{{
			Label:     "Delirium — Gets +2/+2 as long as there are four or more card types among cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: whileYouHave(drcDeliriumMet, selfCreature),
			Apply:     pumpBy(2, 2),
		}},
	})
}
