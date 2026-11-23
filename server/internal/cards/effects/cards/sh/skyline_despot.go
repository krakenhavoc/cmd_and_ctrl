package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyline Despot — Creature — Dragon {5}{R}{R}, 5/5:
//
//	"Flying
//	 When this creature enters, you become the monarch.
//	 At the beginning of your upkeep, if you're the monarch, create a
//	 5/5 red Dragon creature token with flying."
//
// The upkeep trigger has an intervening "if" (CR 603.4): you must be the
// monarch as the step begins, and again as the ability resolves
// (CondMonarch), so a crown stolen in between leaves the token uncreated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c0ce54a1-14a1-45f3-91d0-c57b581afc63",
		Name:            "Skyline Despot",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouBecomeTheMonarch("Skyline Despot"),
			effects.On(game.EventBeginUpkeep,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.ByYou(ev, source, game.Characteristic{}, g) && effects.YoureTheMonarch(g, source.Controller)
				},
				"Skyline Despot — create a 5/5 red Dragon creature token with flying",
				effects.Do(effects.If{
					Cond: effects.CondMonarch{},
					Then: []effects.Applier{effects.CreateToken{Template: effects.TokenCard("5/5 red Dragon with flying"), N: 1}},
				})),
		},
	})
}
