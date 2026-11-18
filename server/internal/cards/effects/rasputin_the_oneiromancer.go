package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "c16fd2b9-9147-43bc-aa08-6d1bddf1d3e8",
		Name:         "Rasputin, the Oneiromancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Rasputin, the Oneiromancer — a dream counter for each opponent; each opponent creates a 1/1 Goblin",
				Do(
					Instruction(func(ctx *Context) error {
						return CounterOnThis{Kind: game.CounterDream, N: len(ctx.Opponents())}.Apply(ctx)
					}),
					Instruction(func(ctx *Context) error {
						for _, opp := range ctx.Opponents() {
							if err := (CreateToken{Controller: opp, Template: TokenCard("1/1 red Goblin"), N: 1}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					}),
				)),
		},
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{
				Tap:            true,
				RemoveCounters: RemoveCountersXFromThis(game.CounterDream, 1).RemoveCounters,
			},
			ProducedForPaid: func(_ *game.Game, _, _ uuid.UUID, paid game.PaidCost) game.ManaOutput {
				return game.ManaOutput{Amount: game.ManaAmount{Options: []string{"C"}, N: paid.CountersRemoved}}
			},
			Label: "{T}, Remove one or more dream counters from Rasputin: Add that much {C}.",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Remove a dream counter from Rasputin: Create a 2/2 white Knight creature token with protection from red.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    Plus(TapCost(), RemoveCountersFromThis(game.CounterDream, 1)),
			Effect:  Do(CreateToken{Template: TokenCard("2/2 white Knight with protection from red"), N: 1}),
		}},
	})
}
