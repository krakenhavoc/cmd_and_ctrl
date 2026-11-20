package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inquisitor's Ox — Creature — Ox {3}{W}, 2/5
// (Shadows over Innistrad, #5543):
//
//	"Delirium — This creature gets +1/+0 and has vigilance as long as
//	 there are four or more card types among cards in your graveyard."
//
// Delirium is an ability word (CR 207.2c): a layer 7c pump and a layer 6
// grant on this creature alone, gated on its controller's own graveyard
// (drcDeliriumMet) and re-read on every recompute, as on Spineseeker
// Centipede.
//
// No simplification.
func init() {
	const delirium = " as long as there are four or more card types among cards in your graveyard."
	pump := game.StaticAbility{
		Label:     "Delirium — Gets +1/+0" + delirium,
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: whileYouHave(drcDeliriumMet, selfCreature),
		Apply:     pumpBy(1, 0),
	}
	vigilance := KeywordGrant(whileYouHave(drcDeliriumMet, selfCreature), "vigilance")
	vigilance.Label = "Delirium — Has vigilance" + delirium
	Register(Spec{
		OracleID:     "224788aa-3876-4844-949e-918e6273151d",
		Name:         "Inquisitor's Ox",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{pump, vigilance},
	})
}
