package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Commissar Severina Raine — Legendary Creature — Human Soldier {1}{W}{B}, 2/2:
//
//	"Leading from the Front — Whenever Commissar Severina Raine attacks,
//	 each opponent loses X life, where X is the number of other attacking
//	 creatures.
//	 Summary Execution — {2}, Sacrifice another creature: You gain 2 life
//	 and draw a card."
//
// X is counted as the trigger resolves, so a creature removed from combat in
// the meantime is not counted. The sacrifice is a cost (another creature, never
// Severina herself).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c00074ec-5b31-44d3-b69b-e2bf6f6a2477",
		Name:         "Commissar Severina Raine",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Commissar Severina Raine — each opponent loses X life, X the number of other attacking creatures",
				func(g *game.Game, item *game.StackItem) error {
					n := len(effects.S82OtherAttackers(g, item.SourceCardID))
					if n <= 0 {
						return nil
					}
					return effects.EachOpponentLosesLifeNow(g, item, n)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice another creature: You gain 2 life and draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect:  effects.Do(effects.GainLife{Amount: 2}, effects.DrawCards{N: 1}),
		}},
	})
}
