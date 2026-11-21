package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Null Brooch — Artifact {4}:
//
//	"{2}, {T}, Discard your hand: Counter target noncreature spell."
//
// A reusable Negate whose price is the hand. The CR 602 owner of
// #1600's "Discard your hand" clause: the hand goes when the
// activation is announced (CR 602.2b, 601.2h), so a countered or
// fizzled activation does not give it back, and a discard payoff
// triggers above the ability and resolves first. An empty hand pays
// it, which is how the card is played — empty the hand, then hold
// the Brooch up.
//
// The target is the ordinary "noncreature spell" clause, re-checked
// on resolution (CR 608.2b): a spell that left the stack in response
// is not countered and the hand is still gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f885041-3e57-4a69-84f2-fd207ff9f31b",
		Name:         "Null Brooch",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Discard your hand: Counter target noncreature spell.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.DiscardYourHand()),
			Targets: effects.TargetSpell("target noncreature spell", effects.Noncreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CounterTheTargetSpell(item, effects.NewContext(g, item))
			},
		}},
	})
}
