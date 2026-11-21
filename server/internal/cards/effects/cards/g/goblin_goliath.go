package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Goliath — Creature — Goblin Mutant {4}{R}{R}, 5/4:
//
//	"When this creature enters, create a number of 1/1 red Goblin
//	 creature tokens equal to the number of opponents you have.
//	 {3}{R}, {T}: If a source you control would deal damage to an
//	 opponent this turn, it deals double that damage to that player
//	 instead."
//
// The tokens count your opponents as the trigger resolves: the players
// still in the game other than you. The activated ability is ADR 0108
// §3's multiplier, players only — damage to an opponent's creature is
// dealt as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "131069a6-8f30-4caf-8934-3588837b5f7f",
		Name:         "Goblin Goliath",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Goblin Goliath — create a Goblin for each opponent", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				n := len(ctx.Opponents())
				if n == 0 {
					return nil
				}
				return effects.CreateToken{Controller: ctx.Controller(), Template: effects.RedGoblinToken(), N: n}.Apply(ctx)
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{R}, {T}: If a source you control would deal damage to an opponent this turn, it deals double that damage to that player instead.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{3}{R}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.MultiplyDamage{Factor: 2, Sources: game.DamageSourcesYours,
					Recipients: game.DamageRecipientsOpponents,
					Label:      "Goblin Goliath — your sources deal double damage to opponents"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
