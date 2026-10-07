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
// The activated ability (#2527, ADR 0113's 2026-10-07 amendment):
// "Discard X cards" is the variable-count discard cost,
// effects.DiscardX, whose count IS the X announced with the activation
// (CR 602.2b) -- the discard twin of SacrificeX. The activator names
// that many cards in hand, they are discarded as the cost is paid
// (so a discard payoff triggers above the ability), and the effect
// reads the same number with ctx.X(). X may be zero (a legal, empty
// payment, CR 107.3a); XMatters tells the legal-move enumerator not to
// offer that no-op, as it does for any other card whose whole effect
// is X.
//
// The effect exiles the top X cards of the TARGET OPPONENT's library
// face up, and the activator may play lands and cast spells from among
// them "without paying their mana costs" for as long as they stay in
// exile -- no "this turn", the printed text has none (compare Urza,
// whose free play is until end of turn). That is ExileTopWithPermission
// with Free (a {0} cost in place of the printed one) and WhileExiled.
// A card is exiled "this way" only if it actually went to exile, so a
// commander that goes to the command zone instead (CR 903.9) gets no
// permission; the primitive stamps the grant from its continuation.
// Fewer than X cards in the library exile what is there. If the target
// is gone on resolution nothing is exiled, but the discards were paid.
func init() {
	Register(Spec{
		OracleID:     "928d977e-cff0-4e0e-83bb-16d73a754f35",
		Name:         "Gix, Yawgmoth Praetor",
		Completeness: CompletenessFull,
		XMatters:     true,
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
				return game.NewTriggeredItem(source, gixPayOneLifeLabel)
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
		Activated: []ActivatedAbility{{
			Label:   gixExileLabel,
			Cost:    Plus(ManaCost("{4}{B}{B}{B}"), DiscardX("X cards")),
			Targets: TargetPlayer("target opponent", Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := ctx.X()
				if x <= 0 {
					return nil
				}
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					return ExileTopWithPermission{
						From:        t.ID,
						GrantTo:     item.Controller,
						N:           x,
						Free:        true,
						WhileExiled: true,
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}

// gixExileLabel is the activated ability as printed.
const gixExileLabel = "{4}{B}{B}{B}, Discard X cards: Exile the top X cards of target opponent's library. You may play lands and cast spells from among cards exiled this way without paying their mana costs."

// gixPayOneLifeLabel is the stack label.
const gixPayOneLifeLabel = "Gix, Yawgmoth Praetor — its controller may pay 1 life to draw a card"

// gixLifePayment is the printed 1.
const gixLifePayment = 1
