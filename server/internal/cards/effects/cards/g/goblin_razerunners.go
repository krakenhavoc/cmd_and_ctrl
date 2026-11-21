package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Razerunners — Creature — Goblin Warrior {2}{R}{R}, 3/4:
//
//	"{1}{R}, Sacrifice a land: Put a +1/+1 counter on this creature.
//	 At the beginning of your end step, you may have this creature deal
//	 damage equal to the number of +1/+1 counters on it to target player
//	 or planeswalker."
//
// The activated ability is Arcane Spyglass's sacrificed land as a cost
// and a counter on this. The end-step trigger targets a player or
// planeswalker (Balefire Liege's clause), and the "you may" is asked as it
// resolves (CR 608.2d); the damage is read from the counters as the
// answer comes in, and from the creature's last-known information if it
// has left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7b3e05e1-31b4-46af-b3d0-f2467f8551e9",
		Name:         "Goblin Razerunners",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, Sacrifice a land: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.B08SacrificeALand()),
			Effect:  effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtYourEndStep("Goblin Razerunners — you may have it deal damage equal to its +1/+1 counters to target player or planeswalker",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Goblin Razerunners — deal damage equal to the number of +1/+1 counters on it to the target?",
						Then:     goblinRazerunnersDamage,
					}.Apply(effects.NewContext(g, item))
				}), effects.TargetPlayerOrPlaneswalker()),
		},
	})
}

// goblinRazerunnersDamage is the "you may" answered: a yes has the
// Razerunners deal damage equal to its +1/+1 counters to the target if it
// is still legal.
var goblinRazerunnersDamage = effects.MayYes("confirm/s82-rpr-b25-goblin-razerunners-damage",
	func(ctx *effects.Context, _ game.ConfirmAnswer) error {
		n := effects.CountersOnThisAsItResolves(game.CounterPlusOne)(ctx)
		if n <= 0 {
			return nil
		}
		return ctx.Game.DamageInstanceForEffect(func() error {
			for _, t := range ctx.LegalTargets() {
				if err := (effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		})
	})
