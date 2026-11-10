package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Baloth Prime — Creature — Beast {3}{G}, 10/10 (Edge of Eternities
// Commander):
//
//	"This creature enters tapped with six stun counters on it. (If a
//	 permanent with a stun counter would become untapped, remove one
//	 from it instead.)
//	 Whenever you sacrifice a land, create a tapped 4/4 green Beast
//	 creature token and untap this creature.
//	 {4}, Sacrifice a land: You gain 2 life."
//
// The entry is Sleep-Cursed Faerie's SelfEntersTappedWithCounters. The
// trigger is one per land sacrificed (the ruling: a Prime that leaves
// with the lands still triggers for each), the token enters tapped
// (CreateTokenAdvanced) and the untap goes through the engine's one
// untap primitive, so a stun counter left on the Prime is removed
// instead. The life ability sacrifices a land as a cost, which is itself
// a sacrifice that triggers the Prime.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dba3dcc1-2177-4a8c-8cf3-7d4fb83c7a16",
		Name:         "Baloth Prime",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedWithCounters(game.CounterStun, 6)},
		Triggered: []game.TriggeredAbility{
			On(game.EventSacrifice, cmdrB9YouSacrificedLand,
				"Baloth Prime — create a tapped 4/4 green Beast token and untap this creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (CreateTokenAdvanced{Spec: Token(TokenCard("4/4 green Beast")).EntersTapped(), N: 1}).Apply(ctx); err != nil {
						return err
					}
					return b31UntapSelf(g, item)
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}, Sacrifice a land: You gain 2 life.",
			Purpose: game.Purpose{LifeGain: 2, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{4}"), b08SacrificeALand()),
			Effect:  Do(GainLife{Amount: 2}),
		}},
	})
}
