package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Questing Druid // Seek the Beast — Creature — Human Druid {1}{G}, 1/1 //
// Instant — Adventure {1}{R} (Wilds of Eldraine, #4471):
//
//	Questing Druid — "Whenever you cast a spell that's white, blue,
//	 black, or red, put a +1/+1 counter on this creature."
//	Seek the Beast — "Exile the top two cards of your library. Until your
//	 next end step, you may play those cards."
//
// The trigger reads the spell's colours as it is cast (a multicolour spell
// that is green and red counts once: it is one trigger per spell). The
// Adventure is Opera Love Song's exile with a window that ends as your
// next end step begins (game.UntilYourNextEndStep).
//
// No simplification.
const questingDruidOracleID = "555df32f-f8e9-49af-bc4a-1194ea325a97"

func init() {
	effects.Register(effects.Spec{
		OracleID:     questingDruidOracleID,
		Name:         "Questing Druid",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCast(effects.Or(effects.OfColor("W"), effects.OfColor("U"), effects.OfColor("B"), effects.OfColor("R"))),
				"Questing Druid — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     questingDruidOracleID + "#1",
		Name:         "Seek the Beast",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ExileTopNUntilYourNextEndStep(ctx, 2)
		},
	})
}
