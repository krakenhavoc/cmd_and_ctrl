package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nesting Dovehawk — Creature — Bird {3}{W}, 2/2:
//
//	"Flying
//	 At the beginning of combat on your turn, populate. (Create a
//	 token that's a copy of a creature token you control.)
//	 Whenever a creature token you control enters, put a +1/+1
//	 counter on this creature."
//
// Flying rides PrintedKeywords. The combat trigger is Populate; the
// token it makes enters under your control and so feeds the second
// trigger, a separate ETB watcher for any creature token you control
// (the Dovehawk is not a token, so "another" is moot). The counter
// lands on the Dovehawk as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fe8fc442-ed17-40b2-8624-69f2eed3f9be",
		Name:            "Nesting Dovehawk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Nesting Dovehawk — populate", effects.Do(effects.Populate{})),
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature() && effects.IsToken(c)
			}, "Nesting Dovehawk — +1/+1 counter", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
