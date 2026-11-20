package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thraben Foulbloods — Creature — Zombie Dog {2}{B}, 3/2 (Eldritch Moon,
// #5541):
//
//	"Delirium — This creature gets +1/+1 and has menace as long as there
//	 are four or more card types among cards in your graveyard."
//
// Backwoods Survivalists' two delirium statics with menace for trample.
//
// No simplification.
func init() {
	const delirium = " as long as there are four or more card types among cards in your graveyard."
	menace := KeywordGrant(whileYouHave(drcDeliriumMet, selfCreature), "menace")
	menace.Label = "Delirium — Has menace" + delirium
	Register(Spec{
		OracleID:     "08adacab-96c3-4532-8ab3-0c8064fa4150",
		Name:         "Thraben Foulbloods",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Delirium — Gets +1/+1" + delirium,
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: whileYouHave(drcDeliriumMet, selfCreature),
			Apply:     pumpBy(1, 1),
		}, menace},
	})
}
