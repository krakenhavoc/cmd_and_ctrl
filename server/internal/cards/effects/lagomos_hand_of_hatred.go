package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "12c98c1e-7e02-473d-b494-ad8a0731fe82",
		Name:         "Lagomos, Hand of Hatred",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Lagomos, Hand of Hatred — create a 2/1 Elemental", func(g *game.Game, item *game.StackItem) error {
				cursor := b25LastEventSeq(g)
				if err := g.CreateTokenForEffect(item.Controller, TokenCard("2/1 red Elemental with trample and haste"), 1); err != nil {
					return err
				}
				tokens := b27TokensCreatedByAfter(g, item.Controller, cursor)
				if len(tokens) == 0 {
					return nil
				}
				return ScheduleDelayedTrigger{
					Label:  "Lagomos, Hand of Hatred — sacrifice the Elemental",
					Cards:  tokens,
					Effect: b33SacrificeListedCards,
				}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{{
			Label:     "{T}: Search your library for a card, put it into your hand, then shuffle. Activate only if five or more creatures died this turn.",
			Cost:      TapCost(),
			Condition: lagomosFiveCreaturesDied,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(game.Card) bool { return true },
					Dest:      game.ZoneHand,
					Limit:     1,
					Shuffle:   true,
					Reason:    "Lagomos, Hand of Hatred — search for a card",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

// lagomosFiveCreaturesDied is Lagomos's tutor condition.
func lagomosFiveCreaturesDied(g *game.Game, _, _ uuid.UUID) bool {
	return g.TurnTally.CreaturesDied >= 5
}
