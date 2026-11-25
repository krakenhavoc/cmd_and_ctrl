package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dismissive Pyromancer — Creature — Human Wizard {1}{R}, 2/2 (Core Set
// 2019, #5527):
//
//	"{R}, {T}, Discard a card: Draw a card.
//	 {2}{R}, {T}, Sacrifice this creature: It deals 4 damage to target
//	 creature."
//
// Rummaging Goblin's loot with a mana cost, and Barbarian Lunatic's
// sacrifice-for-damage with a tap: the discard and the sacrifice are costs
// (CR 118.8, 602.2b), so "it" is the Pyromancer's last-known information
// (CR 113.7a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f040310d-d6ac-4d77-9f76-ca325eddf306",
		Name:         "Dismissive Pyromancer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}, {T}, Discard a card: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost(), effects.DiscardACard()),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
			{
				Label:   "{2}{R}, {T}, Sacrifice this creature: It deals 4 damage to target creature.",
				Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.TapCost(), effects.SacrificeThis()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.DealDamageToTheTarget(4),
			},
		},
	})
}
