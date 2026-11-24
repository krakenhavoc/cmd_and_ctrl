package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tinybones, Trinket Thief — Legendary Creature — Skeleton Rogue {1}{B}, 1/2
// (Secret Lair Drop):
//
//	"At the beginning of each end step, if an opponent discarded a card this
//	 turn, you draw a card and you lose 1 life.
//	 {4}{B}{B}: Each opponent with no cards in hand loses 10 life."
//
// "Each end step" is any player's. The intervening "if" (CR 603.4) reads the
// per-turn tally of cards each opponent discarded, as the step begins and
// again as the trigger resolves. The activated ability checks each opponent's
// hand as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "008eda1f-b913-4f04-8c29-5e87c270149e",
		Name:         "Tinybones, Trinket Thief",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return tinybonesOpponentDiscarded(g, source.Controller)
			}, "Tinybones, Trinket Thief — draw a card and lose 1 life", func(g *game.Game, item *game.StackItem) error {
				if !tinybonesOpponentDiscarded(g, item.Controller) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				if err := (effects.DrawCards{N: 1}).Apply(ctx); err != nil {
					return err
				}
				return effects.GainLife{Amount: -1}.Apply(ctx)
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{B}{B}: Each opponent with no cards in hand loses 10 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{4}{B}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, opp := range ctx.Opponents() {
					p := g.PlayerByIDForEffect(opp)
					if p == nil || p.Hand == nil || p.Hand.Size() != 0 {
						continue
					}
					if err := g.ChangePlayerLifeForEffect(ctx.Source(), opp, -10); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}

// tinybonesOpponentDiscarded is "an opponent discarded a card this turn".
func tinybonesOpponentDiscarded(g *game.Game, controller uuid.UUID) bool {
	for _, p := range g.Seats {
		if p == nil || p.ID == controller || p.Eliminated {
			continue
		}
		if g.TurnTallyFor(p.ID).CardsDiscarded > 0 {
			return true
		}
	}
	return false
}
