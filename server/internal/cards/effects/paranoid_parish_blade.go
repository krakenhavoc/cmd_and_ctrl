package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Paranoid Parish-Blade — Creature — Human Soldier {2}{W}, 3/2
// (Shadows over Innistrad, #5543):
//
//	"Delirium — This creature gets +1/+0 and has first strike as long as
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
	firststrike := KeywordGrant(whileYouHave(drcDeliriumMet, selfCreature), "first strike")
	firststrike.Label = "Delirium — Has first strike" + delirium
	Register(Spec{
		OracleID:     "d1c2b124-5290-4e66-9f6b-4db4d662a74c",
		Name:         "Paranoid Parish-Blade",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{pump, firststrike},
	})
}
