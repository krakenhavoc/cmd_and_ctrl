package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kessig Dire Swine — Creature — Boar Horror {4}{G}{G}, 6/6
// (Shadows over Innistrad, #5543):
//
//	"Delirium — This creature has trample as long as there are four or
//	 more card types among cards in your graveyard."
//
// Delirium is an ability word (CR 207.2c): a layer 6 grant on this
// creature alone, gated on its controller's own graveyard (drcDeliriumMet)
// and re-read on every recompute, as on Deathcap Cultivator.
//
// No simplification.
func init() {
	const delirium = " as long as there are four or more card types among cards in your graveyard."
	trample := KeywordGrant(whileYouHave(drcDeliriumMet, selfCreature), "trample")
	trample.Label = "Delirium — Has trample" + delirium
	Register(Spec{
		OracleID:     "61c10a4c-b1d5-4777-a813-d231980be2a7",
		Name:         "Kessig Dire Swine",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{trample},
	})
}
