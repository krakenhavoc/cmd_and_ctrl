package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Falkenrath Pit Fighter — Creature — Vampire Warrior {R}, 2/1 (Innistrad:
// Midnight Hunt, #5497):
//
//	"{1}{R}, Discard a card, Sacrifice a Vampire: Draw two cards. Activate
//	 only if an opponent lost life this turn."
//
// Airdrop Condor's sacrifice cost (the Pit Fighter is itself a legal
// Vampire to sacrifice) with Charging Strifeknight's discard cost and an
// activation condition read off the per-turn life-loss tally.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "46d6bc30-9a21-4476-947b-87951cd9c882",
		Name:         "Falkenrath Pit Fighter",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{1}{R}, Discard a card, Sacrifice a Vampire: Draw two cards. Activate only if an opponent lost life this turn.",
			Cost: Plus(ManaCost("{1}{R}"), DiscardACard(),
				game.AbilityCost{SacrificeOther: sacrificeSpec("a Vampire", OfCreatureType("Vampire"))}),
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return opponentLifeLostThisTurn(g, controller) > 0
			},
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Effect:  Do(DrawCards{N: 2}),
		}},
	})
}
