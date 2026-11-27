package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyshroud Sentinel — Creature — Elf {2}{G}, 1/1 (Nemesis, S82 OLD, #5618):
//
//	"When this creature enters, you may search your library for up to three
//	 cards named Skyshroud Sentinel, reveal them, put them into your hand, then
//	 shuffle."
//
// Nesting Wurm's search under this card's name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4f437f42-1ac5-4681-9b16-4eae9f49bb22",
		Name:         "Skyshroud Sentinel",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Skyshroud Sentinel — search for up to three cards named Skyshroud Sentinel",
				effects.S82OldB677SearchForUpToThreeNamed("Skyshroud Sentinel")),
		},
	})
}
