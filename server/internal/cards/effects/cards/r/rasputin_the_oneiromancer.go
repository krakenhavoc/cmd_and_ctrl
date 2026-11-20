package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rasputin, the Oneiromancer — Legendary Creature — Human Wizard {1}{W}{U},
// 4/1 (Dominaria United Commander, S82):
//
//	"When Rasputin enters, put a dream counter on it for each opponent you
//	 have. Each opponent creates a 1/1 red Goblin creature token.
//	 {T}, Remove one or more dream counters from Rasputin: Add that much {C}.
//	 {T}, Remove a dream counter from Rasputin: Create a 2/2 white Knight
//	 creature token with protection from red."
//
// The enter trigger counts your opponents as it resolves, then each
// opponent creates their own Goblin. The first {T} ability is a mana
// ability (CR 605.1a): its cost removes an announced number of counters,
// one at least, and the mana is that many {C}, read off the paid-cost
// record. The second is an ordinary activated ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c16fd2b9-9147-43bc-aa08-6d1bddf1d3e8",
		Name:         "Rasputin, the Oneiromancer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rasputin, the Oneiromancer — a dream counter for each opponent; each opponent creates a 1/1 Goblin",
				effects.Do(
					effects.Instruction(func(ctx *effects.Context) error {
						return effects.CounterOnThis{Kind: game.CounterDream, N: len(ctx.Opponents())}.Apply(ctx)
					}),
					effects.Instruction(func(ctx *effects.Context) error {
						for _, opp := range ctx.Opponents() {
							if err := (effects.CreateToken{Controller: opp, Template: effects.TokenCard("1/1 red Goblin"), N: 1}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					}),
				)),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{
				Tap:            true,
				RemoveCounters: effects.RemoveCountersXFromThis(game.CounterDream, 1).RemoveCounters,
			},
			ProducedForPaid: func(_ *game.Game, _, _ uuid.UUID, paid game.PaidCost) game.ManaOutput {
				return game.ManaOutput{Amount: game.ManaAmount{Options: []string{"C"}, N: paid.CountersRemoved}}
			},
			Label: "{T}, Remove one or more dream counters from Rasputin: Add that much {C}.",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Remove a dream counter from Rasputin: Create a 2/2 white Knight creature token with protection from red.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.TapCost(), effects.RemoveCountersFromThis(game.CounterDream, 1)),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 white Knight with protection from red"), N: 1}),
		}},
	})
}
