package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Electrostatic Infantry — Creature — Dwarf Wizard {1}{R}, 1/2 (Dominaria
// United, #5487):
//
//	"Trample
//	 Whenever you cast an instant or sorcery spell, put a +1/+1 counter on
//	 this creature."
//
// A cast trigger over instants and sorceries; the counter goes on the
// permanent that triggered and is lost if that is a new object by the
// time the ability resolves (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ba8883fc-db24-4c3d-ae36-05085089c9cc",
		Name:            "Electrostatic Infantry",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
				"Electrostatic Infantry — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
