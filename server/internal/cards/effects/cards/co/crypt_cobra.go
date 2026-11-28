package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crypt Cobra — Creature — Snake {3}{B}, 3/3 (Mirage, S82 OLD batch 810,
// #5633):
//
//	"Whenever this creature attacks and isn't blocked, defending player gets a
//	 poison counter."
//
// Keeper of Tresserhorn's "attacks and isn't blocked" trigger (#3531): the
// defending player is the one the block declaration fired it for; a player who
// has left the game gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "456bbc0e-a648-481f-b37e-9996029556bd",
		Name:         "Crypt Cobra",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenAttacksAndIsNotBlockedEffect("Crypt Cobra — defending player gets a poison counter",
				func(g *game.Game, item *game.StackItem, defender uuid.UUID) error {
					if !effects.DefendingPlayerStillIn(g, defender) {
						return nil
					}
					return g.AddPlayerCounterByForEffect(item.SourceCardID, defender, game.CounterPoison, 1)
				}),
		},
	})
}
