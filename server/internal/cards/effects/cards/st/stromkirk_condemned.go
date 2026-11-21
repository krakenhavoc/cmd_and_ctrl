package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stromkirk Condemned — Creature — Vampire Horror {B}{B}, 2/2:
//
//	"Discard a card: Vampires you control get +1/+1 until end of turn.
//	 Activate only once each turn."
//
// The discard is a cost, paid whether or not the ability resolves; the
// once-per-turn limit is spent at
// the announce (OncePerTurnActivation). The affected set is fixed on
// resolution (CR 611.2c).
//
// No simplification.
func init() {
	const label = "Discard a card: Vampires you control get +1/+1 until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "67b87332-44e1-4948-9269-5fc8f6107d0e",
		Name:         "Stromkirk Condemned",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Cost:      effects.DiscardACard(),
			Condition: effects.OncePerTurnActivation(label),
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl(), effects.OfCreatureType("Vampire")), Power: 1, Toughness: 1,
					Label: "Stromkirk Condemned — Vampires you control get +1/+1",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
