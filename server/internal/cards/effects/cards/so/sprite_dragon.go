package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sprite Dragon — Creature — Faerie Dragon {U}{R}, 1/1 (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"Flying, haste
//	 Whenever you cast a noncreature spell, put a +1/+1 counter on this
//	 creature."
//
// Tempest Angler's trigger: it fires as the spell becomes cast and resolves
// before the spell does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a9d8ab76-70a4-475e-b87e-4737c090553a",
		Name:            "Sprite Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Sprite Dragon — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
