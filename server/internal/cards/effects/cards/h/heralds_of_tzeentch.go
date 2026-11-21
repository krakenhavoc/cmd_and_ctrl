package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heralds of Tzeentch — Creature — Demon {4}{U}, 3/3 (Warhammer 40,000
// Commander, S82 GAP batch 195, #5948):
//
//	"Flying
//	 Cascade (When you cast this spell, exile cards from the top of your library
//	 until you exile a nonland card that costs less. You may cast it without
//	 paying its mana cost. Put the exiled cards on the bottom in a random
//	 order.)"
//
// Flying and the keyword's own cast trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c4ecfc13-a335-42bb-92db-133998b2af5d",
		Name:            "Heralds of Tzeentch",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "cascade"},
		Triggered:       []game.TriggeredAbility{effects.Cascade()},
	})
}
