package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Converter Beast — Creature — Phyrexian Beast {3}{G} (March of the
// Machine):
//
//	"When this creature enters, incubate 5."
//
// A plain enters trigger on the stack (ADR 0018), so opponents may
// respond to it; the Incubator (CR 701.53a, ADR 0185) is made when it
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7fbb9547-a805-4768-8574-e283e64600a4",
		Name:         "Converter Beast",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Converter Beast — incubate 5", effects.Do(effects.Incubate{N: 5})),
		},
	})
}
