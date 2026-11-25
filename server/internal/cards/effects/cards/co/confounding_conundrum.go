package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Confounding Conundrum — Enchantment {1}{U} (Zendikar Rising, S82 OLD
// #5508):
//
//	"When this enchantment enters, draw a card.
//	 Whenever a land an opponent controls enters, if that player had
//	 another land enter the battlefield under their control this turn, they
//	 return a land they control to its owner's hand."
//
// The trigger is an intervening "if" (CR 603.4): the entry tally says the
// entering land's controller has had two lands enter this turn (this one
// and another). That player, not the enchantment's controller, picks the
// land they return.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f86b9a01-62fc-4120-8914-5054ed48d37e",
		Name:         "Confounding Conundrum",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Confounding Conundrum — draw a card", effects.Do(effects.DrawCards{N: 1})),
			effects.On(game.EventETB, effects.S82OldB089OpponentLandWithAnotherLand,
				"Confounding Conundrum — that player returns a land they control to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					ev := item.Trigger.Event
					who := ev.Actor
					if c, ok := g.LookupCardForEffect(ev.CardID); ok {
						who = c.Controller
					}
					if who == uuid.Nil {
						return nil
					}
					return effects.ChoosePermanents{
						Player:     who,
						Question:   "Confounding Conundrum — return a land you control to its owner's hand",
						Candidates: effects.S82OldB089LandCandidates,
						Keyed:      effects.BouncedChosen,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
