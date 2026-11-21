package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tuskeri Firewalker — Creature — Human Berserker {2}{R}, 3/2:
//
//	"Boast — {1}: Exile the top card of your library. You may play that card this turn.
//	 (Activate only if this creature attacked this turn and only once each turn.)"
//
// Boast (CR 702.142a) is built with the Boast constructor (boast.go):
// the engine reads the attack record and the activation tally, so the
// card names neither. The activation is spent at the announce, whether
// or not it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "36dda03b-a3c1-4ef4-ae62-da318028a39e",
		Name:         "Tuskeri Firewalker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.BoastAnswering(game.AnswerValue, "{1}: Exile the top card of your library. You may play that card this turn.",
				effects.ManaCost("{1}"),
				func(g *game.Game, item *game.StackItem) error {
					_, err := effects.B12ImpulseExileForTurn(g, item, 1)
					return err
				}),
		},
	})
}
