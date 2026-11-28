package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Woolly Spider — Creature — Spider {1}{G}{G}, 2/3 (Ice Age, S82 OLD, #5636):
//
//	"Reach
//	 Whenever this creature blocks a creature with flying, this creature gets
//	 +0/+2 until end of turn."
//
// Frostweb Spider's trigger (one per flying attacker blocked, CR 509.3b) with a
// pump instead of the counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "811c6f52-03cb-49a9-a3c7-7f4ae40e7155",
		Name:            "Woolly Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, effects.S82OldB565BlocksAFlyer, "Woolly Spider — gets +0/+2 until end of turn",
				effects.ThisGetsUntilEndOfTurn(0, 2, "Woolly Spider — +0/+2")),
		},
	})
}
