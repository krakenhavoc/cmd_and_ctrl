package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aven Trailblazer — Creature — Bird Soldier {2}{W}, 2/* (Conflux,
// #5587):
//
//	"Flying
//	 Domain — Aven Trailblazer's toughness is equal to the number of
//	 basic land types among lands you control."
//
// Flying is a printed keyword. The toughness is a characteristic-
// defining ability (CR 604.3, layer 7a) through the shared DomainPT with
// a zero power factor, which leaves the printed power of 2 alone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "518115c9-0f0f-4249-81ff-e79b27dda1be",
		Name:            "Aven Trailblazer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			effects.DomainPT("Domain — Aven Trailblazer's toughness is equal to the number of basic land types among lands you control.", 0, 1),
		},
	})
}
