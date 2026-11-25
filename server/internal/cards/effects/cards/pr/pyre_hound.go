package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyre Hound — Creature — Elemental Dog {3}{R}, 2/3 (Shadows over
// Innistrad, #5543):
//
//	"Trample
//	 Whenever you cast an instant or sorcery spell, put a +1/+1 counter on
//	 this creature."
//
// Electrostatic Infantry's trigger: a cast of yours, judged on the spell's
// own types, and one counter on the source if it is still the permanent
// the ability came from.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0a6f45e9-5c7a-4e2c-bf8a-fc5fe0ae0793",
		Name:            "Pyre Hound",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
				"Pyre Hound — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
