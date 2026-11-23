package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mesa Cavalier — Creature — Human Knight {2}{W}, 2/1 (Dominaria United, S82):
//
//	"Flying
//	 When this creature enters, you gain 2 life."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "898763f8-a333-43f5-9f66-52e2adbe5758",
		Name:            "Mesa Cavalier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Mesa Cavalier — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
