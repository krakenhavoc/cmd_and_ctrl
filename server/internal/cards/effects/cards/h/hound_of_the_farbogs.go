package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hound of the Farbogs — Creature — Zombie Dog {4}{B}, 5/3
// (Shadows over Innistrad, #5543):
//
//	"Delirium — This creature has menace as long as there are four or
//	 more card types among cards in your graveyard. (A creature with
//	 menace can't be blocked except by two or more creatures.)"
//
// Delirium is an ability word (CR 207.2c): a layer 6 grant on this
// creature alone, gated on its controller's own graveyard (drcDeliriumMet)
// and re-read on every recompute, as on Deathcap Cultivator.
//
// No simplification.
func init() {
	const delirium = " as long as there are four or more card types among cards in your graveyard."
	menace := effects.KeywordGrant(effects.WhileYouHave(effects.DrcDeliriumMet, effects.SelfCreature), "menace")
	menace.Label = "Delirium — Has menace" + delirium
	effects.Register(effects.Spec{
		OracleID:     "7be2bc4e-3986-4f4f-bd5b-0a9f59b2f865",
		Name:         "Hound of the Farbogs",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{menace},
	})
}
