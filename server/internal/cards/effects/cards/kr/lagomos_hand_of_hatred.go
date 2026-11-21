package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lagomos, Hand of Hatred — Legendary Creature — Human Shaman {1}{B}{R},
// 1/3 (EDHREC rank 2899):
//
//	"At the beginning of combat on your turn, create a 2/1 red
//	 Elemental creature token with trample and haste. Sacrifice it at
//	 the beginning of the next end step.
//	 {T}: Search your library for a card, put it into your hand, then
//	 shuffle. Activate only if five or more creatures died this turn."
//
// A free hasty attacker every turn, and a tutor once the board has
// seen enough bodies hit the graveyard. The Elemental is sacrificed by
// a delayed trigger that names the token it made (Dalkovan
// Encampment's shape); a token someone else has taken by then is not
// yours to sacrifice and stays. The tutor's "five or more creatures died this turn" is its
// activation condition (CR 602.1b, #743), read off the turn tally's
// table-wide CreaturesDied (#586) — anyone's creatures, tokens
// included. The card is not revealed; the printed text does not say
// so.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "12c98c1e-7e02-473d-b494-ad8a0731fe82",
		Name:         "Lagomos, Hand of Hatred",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Lagomos, Hand of Hatred — create a 2/1 Elemental", func(g *game.Game, item *game.StackItem) error {
				cursor := effects.B25LastEventSeq(g)
				if err := g.CreateTokenForEffect(item.Controller, effects.NewContext(g, item).TextToken(effects.TokenCard("2/1 red Elemental with trample and haste")), 1); err != nil {
					return err
				}
				tokens := effects.B27TokensCreatedByAfter(g, item.Controller, cursor)
				if len(tokens) == 0 {
					return nil
				}
				return effects.ScheduleDelayedTrigger{
					Label: "Lagomos, Hand of Hatred — sacrifice the Elemental",
					Cards: tokens,
					Body:  effects.SacrificeListedCardsBody,
				}.Apply(effects.NewContext(g, item))
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}: Search your library for a card, put it into your hand, then shuffle. Activate only if five or more creatures died this turn.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.TapCost(),
			Condition: lagomosFiveCreaturesDied,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:  item.Controller,
					Dest:    game.ZoneHand,
					Limit:   1,
					Shuffle: true,
					Reason:  "Lagomos, Hand of Hatred — search for a card",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// lagomosFiveCreaturesDied is Lagomos's tutor condition.
func lagomosFiveCreaturesDied(g *game.Game, _, _ uuid.UUID) bool {
	return g.TurnTally.CreaturesDied >= 5
}
