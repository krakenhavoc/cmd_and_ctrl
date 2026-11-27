package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arc Mage — Creature — Human Spellshaper {2}{R}, 2/2 (Nemesis, S82 OLD,
// #5618):
//
//	"{2}{R}, {T}, Discard a card: This creature deals 2 damage divided as you
//	 choose among one or two targets."
//
// Skarrgan Hellkite's divided damage (CR 601.2d: each of the one or two targets
// gets at least 1) behind Rummaging Goblin's discard cost (CR 118.8): a player
// with no card to discard cannot activate it, and the {T} waits out summoning
// sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "011bb6cc-f1a1-4790-befc-373acabf63b8",
		Name:         "Arc Mage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, {T}, Discard a card: This creature deals 2 damage divided as you choose among one or two targets.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.TapCost(), effects.DiscardACard()),
			Targets: effects.TargetAny().WithCount(1, 2).Dividing(effects.Divide(2)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DealDividedDamage(effects.NewContext(g, item))
			},
		}},
	})
}
