package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mutalith Vortex Beast — Creature — Mutant Beast {4}{U}{R}, 6/6 (Warhammer
// 40,000 Commander, S82 GAP batch 194, #5948):
//
//	"Trample
//	 Warp Vortex — When this creature enters, flip a coin for each opponent you
//	 have. For each flip you win, draw a card. For each flip you lose, this
//	 creature deals 3 damage to that player."
//
// One called flip per opponent still in the game (CR 705): the controller
// calls, and the win or loss of each coin is its own. The continuation is keyed
// data carrying the opponents in order, so a table waiting on the call is a
// restore point. Each win draws a card; each loss is 3 damage from the Beast,
// as it last existed if it has left, to the opponent that flip was for.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7e6ca948-a66d-46a9-9174-ac7b75d7b348",
		Name:            "Mutalith Vortex Beast",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Mutalith Vortex Beast — flip a coin for each opponent; draw for each win, 3 damage to that player for each loss", effects.GapB194WarpVortex),
		},
	})
}
