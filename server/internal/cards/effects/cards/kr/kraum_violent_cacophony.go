package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kraum, Violent Cacophony — Legendary Creature — Zombie Horror
// {2}{U}{R}, 2/3 (Outlaws of Thunder Junction, #4468):
//
//	"Flying
//	 Whenever you cast your second spell each turn, put a +1/+1 counter
//	 on Kraum and draw a card."
//
// The trigger is Alphinaud Leveilleur's YouCastYourSecondSpellEachTurn
// (exactly the second spell you cast this turn, any spell). Its effect
// is two footprinted Do steps: the counter goes on Kraum only while it
// is still the permanent that triggered (CounterOnThis, CR 400.7), and
// the draw happens either way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1d764f70-985e-4cf0-9482-f30f73b9baf1",
		Name:            "Kraum, Violent Cacophony",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn,
				"Kraum, Violent Cacophony — put a +1/+1 counter on Kraum and draw a card",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}, effects.DrawCards{N: 1})),
		},
	})
}
