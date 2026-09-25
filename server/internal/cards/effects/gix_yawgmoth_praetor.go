package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gix, Yawgmoth Praetor — Legendary Creature — Phyrexian Praetor
// {1}{B}{B}, 3/3:
//
//	"Whenever a creature deals combat damage to one of your
//	 opponents, its controller may pay 1 life. If they do, they draw
//	 a card.
//	 {4}{B}{B}{B}, Discard X cards: Exile the top X cards of target
//	 opponent's library. You may play lands and cast spells from
//	 among cards exiled this way without paying their mana costs."
//
// Edric in black, with a price. The trigger watches EVERY creature
// that connects with one of Gix's controller's opponents — including
// creatures Gix's controller does not control — and the whole payoff
// belongs to the damaging creature's CONTROLLER, which the combat
// damage event carries in Actor. The trigger is still Gix's ability,
// so it goes on the stack under Gix's controller and is APNAP-ordered
// with their other triggers; only the question and the card go
// elsewhere.
//
// The "may" is asked as the ability RESOLVES rather than before it
// goes on the stack, because the payment is part of the effect:
// MayChoice is the CR 608.2 yes/no a resolving effect asks, and it
// takes any seat as its chooser (#568, #796). LifeCost is declared so
// the move list prices it (#547) — a bot at 1 life that could not see
// the price would answer "pay" and lose the game. The branch re-checks
// affordability when the answer arrives, since life moves between the
// question and the answer.
//
// The activated ability is NOT implemented, and it is the cost rather
// than the effect that blocks it: "Discard X cards" is a variable
// COUNT, and the engine's X only lives in the mana component of a cost
// (AbilityCost.DiscardCards is a fixed number, and "this ability
// prompts for X" is derived from the mana string, which here is a
// fixed {4}{B}{B}{B}). Shipping it with a fixed discard would be a
// different card; shipping it with none would be a free one. Both
// halves of the seam — a variable-count cost component — are the same
// gap Ruthless Technomancer's "Sacrifice X artifacts" waits on.
func init() {
	Register(Spec{
		OracleID:     "928d977e-cff0-4e0e-83bb-16d73a754f35",
		Name:         "Gix, Yawgmoth Praetor",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The last ability — \"{4}{B}{B}{B}, Discard X cards: Exile the top X cards of target opponent's library\" — isn't implemented. The combat-damage draw works.",
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return creatureDealtCombatDamageToAnOpponentOf(ev, source.Controller, g)
			},
			Key: gixPayOneLifeLabel,
			// Actor is the dealing creature's controller, stamped at
			// emit time and carried on item.Trigger.Event — still valid
			// if the creature has since died to simultaneous combat
			// damage. The fill-in Build keeps the "no controller"
			// suppression; the payment itself is the row's Effect.
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				if ev.Actor == uuid.Nil {
					return nil
				}
				return game.NewTriggeredItem(source, gixPayOneLifeLabel, nil)
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				payer := item.Trigger.Event.Actor
				return MayChoice{
					Player:   payer,
					Question: "Gix, Yawgmoth Praetor — pay 1 life to draw a card?",
					YesLabel: "Pay 1 life",
					NoLabel:  "Decline",
					LifeCost: gixLifePayment,
					OnYes: func(ctx *Context) error {
						return payLifeThenDrawFor(ctx, payer, gixLifePayment, 1)
					},
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

// gixPayOneLifeLabel is the stack label.
const gixPayOneLifeLabel = "Gix, Yawgmoth Praetor — its controller may pay 1 life to draw a card"

// gixLifePayment is the printed 1.
const gixLifePayment = 1
