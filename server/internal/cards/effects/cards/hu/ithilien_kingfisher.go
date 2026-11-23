package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ithilien Kingfisher — Creature — Bird {2}{U}, 2/1 (The Hobbit):
//
//	"Flying
//	 When this creature dies, draw a card."
//
// A plain dies trigger (CR 700.4): it goes on the stack from the
// graveyard and its controller, as the creature last controlled it,
// draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a2dd89a1-9b3e-4bd1-9b48-36516d63726a",
		Name:            "Ithilien Kingfisher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Ithilien Kingfisher — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
