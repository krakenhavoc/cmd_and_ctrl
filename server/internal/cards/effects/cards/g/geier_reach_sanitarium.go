package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Geier Reach Sanitarium — Legendary Land (EDHREC rank 441):
//
//	"{T}: Add {C}.
//	 {2}, {T}: Each player draws a card, then discards a card."
//
// A symmetrical loot on a land: the discard deck's engine, and the
// reanimator deck's way to bin a fatty at instant speed. "Each
// player" is the table, controller first: every player draws and
// then chooses their own discard through the same discard prompt
// Faithless Looting uses, so nobody picks for anyone else.
//
// The loop draws and queues a prompt per seat, which lands on the
// printed order by itself: a draw resolves where it is written and a
// discard only QUEUES, so every seat has drawn before any seat's
// prompt can be answered, and the table then waits on all of them
// (#651).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7b9fafe7-d26a-4ed5-b4c4-ce13763770b5",
		Name:         "Geier Reach Sanitarium",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label: "{2}, {T}: Each player draws a card, then discards a card.",
			// ADR 0126 §6: every other player loots too, which no field says.
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1, Discards: 1},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, p := range effects.TablePlayers(ctx) {
					if err := (effects.DrawCards{Player: p, N: 1}).Apply(ctx); err != nil {
						return err
					}
					g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
						Player: p,
						Source: item.SourceCardID,
						N:      1,
					})
				}
				return nil
			},
		}},
	})
}
