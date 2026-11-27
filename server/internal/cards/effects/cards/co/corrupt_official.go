package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Corrupt Official — Creature — Human Minion {4}{B}, 3/1 (Mercadian Masques,
// S82 OLD b682, #5619):
//
//	"{2}{B}: Regenerate this creature.
//	 Whenever this creature becomes blocked, defending player discards a card
//	 at random."
//
// Gorilla Chieftain's regeneration row, and Slate Street Ruffian's trigger on
// the once-per-attacker "becomes blocked" event (CR 509.3c), whose actor is
// the defending player, with a random discard (CR 701.9b) in place of their
// choice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c152b677-5e75-43da-be1d-876b7b83e2f4",
		Name:         "Corrupt Official",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.RegenerateThisRow("{2}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.Self, "Corrupt Official — defending player discards a card at random",
				func(g *game.Game, item *game.StackItem) error {
					defender := item.Trigger.Event.Actor
					if !effects.DefendingPlayerStillIn(g, defender) {
						return nil
					}
					return effects.DiscardCards{Player: defender, N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
