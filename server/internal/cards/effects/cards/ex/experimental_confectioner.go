package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Experimental Confectioner — Creature — Human Peasant {2}{B}, 2/2 (Wilds
// of Eldraine, #4471):
//
//	"When this creature enters, create a Food token.
//	 Whenever you sacrifice a Food, create a 1/1 black Rat creature token
//	 with "This token can't block.""
//
// The entry Food is Cat Collector's. "You sacrifice a Food" is Rapacious
// Guest's sacrifice-event reader (the Food is read while it is still on
// the battlefield), and the Rat is Harried Spearguard's catalog token.
// Each Food sacrificed is one trigger, whether it was paid as a cost or
// sacrificed by an effect.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87f82d15-5b6b-400f-8a15-d45a2a735f7d",
		Name:         "Experimental Confectioner",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Experimental Confectioner — create a Food token", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
			effects.On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B31YouSacrificedAFood(ev, source, g)
			}, "Experimental Confectioner — create a 1/1 black Rat creature token with \"This token can't block.\"",
				effects.Do(effects.CreateToken{Template: effects.CantBlockRatToken(), N: 1})),
		},
	})
}
