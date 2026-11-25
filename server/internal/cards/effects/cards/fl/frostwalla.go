package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frostwalla — Snow Creature — Lizard {2}{G}, 2/2 (Modern Horizons):
//
//	"{S}: This creature gets +2/+2 until end of turn. Activate only once each
//	 turn. ({S} can be paid with one mana from a snow source.)"
//
// Basking Rootwalla's once-a-turn pump (OncePerTurnActivation, CR 602.5b) at
// a cost of {S}, which only mana from a snow source can pay (CR 107.4h), so
// the activation is refused with ordinary mana in the pool. The pump is an
// until-end-of-turn boost of the Frostwalla itself.
//
// No simplification.
func init() {
	const label = "{S}: This creature gets +2/+2 until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "90772e8f-9b39-4cdd-864f-5e2c917f645e",
		Name:         "Frostwalla",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.ManaCost("{S}"),
			Condition: effects.OncePerTurnActivation(label),
			Effect:    effects.ThisGetsUntilEndOfTurn(2, 2, "Frostwalla — +2/+2 until end of turn"),
		}},
	})
}
