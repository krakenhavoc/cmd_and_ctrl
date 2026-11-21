package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moorland Drifter — Creature — Spirit {1}{W}, 2/2
// (Shadows over Innistrad, #5543):
//
//	"Delirium — This creature has flying as long as there are four or
//	 more card types among cards in your graveyard."
//
// Delirium is an ability word (CR 207.2c): a layer 6 grant on this
// creature alone, gated on its controller's own graveyard (drcDeliriumMet)
// and re-read on every recompute, as on Deathcap Cultivator.
//
// No simplification.
func init() {
	const delirium = " as long as there are four or more card types among cards in your graveyard."
	flying := effects.KeywordGrant(effects.WhileYouHave(effects.DrcDeliriumMet, effects.SelfCreature), "flying")
	flying.Label = "Delirium — Has flying" + delirium
	effects.Register(effects.Spec{
		OracleID:     "f5fa9138-9640-4586-bc81-caa6b8eae8b6",
		Name:         "Moorland Drifter",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{flying},
	})
}
