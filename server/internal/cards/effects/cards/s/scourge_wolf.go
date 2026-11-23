package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scourge Wolf — Creature — Wolf Horror {R}{R}, 2/2
// (Shadows over Innistrad, #5543):
//
//	"First strike
//	 Delirium — This creature has double strike as long as there are four
//	 or more card types among cards in your graveyard."
//
// Delirium is an ability word (CR 207.2c): a layer 6 grant on this
// creature alone, gated on its controller's own graveyard (drcDeliriumMet)
// and re-read on every recompute, as on Deathcap Cultivator. First strike
// is printed data.
//
// No simplification.
func init() {
	const delirium = " as long as there are four or more card types among cards in your graveyard."
	doublestrike := effects.KeywordGrant(effects.WhileYouHave(effects.DrcDeliriumMet, effects.SelfCreature), "double strike")
	doublestrike.Label = "Delirium — Has double strike" + delirium
	effects.Register(effects.Spec{
		OracleID:        "fd74d0f6-bbe5-47f6-af15-a127489ca7f3",
		Name:            "Scourge Wolf",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Static:          []game.StaticAbility{doublestrike},
	})
}
