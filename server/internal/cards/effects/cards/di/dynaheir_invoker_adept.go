package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dynaheir, Invoker Adept — Legendary Creature — Human Wizard {1}{U}{R}{W},
// 4/4 (Commander Legends: Battle for Baldur's Gate, S82 GAP2 b05 #5948,
// #6094):
//
//	"Haste
//	 You may activate abilities of other creatures you control as though
//	 those creatures had haste.
//	 {T}: When you next activate an ability that isn't a mana ability this
//	 turn by spending four or more mana to activate it, copy that ability.
//	 You may choose new targets for the copy."
//
// The static is ActivateAsThoughHasteOthers (CR 302.6, 609.4): its
// controller may pay {T} and {Q} costs of their other summoning-sick
// creatures; those creatures gain no haste, and Dynaheir's own abilities
// are not reached by it (it has printed haste for that). The {T} ability
// sets up a CR 603.7 delayed trigger on the controller's next activation
// this turn whose ability went on the stack (a mana ability never does, CR
// 605.3b) with four or more mana spent on it, read from the activation's
// payment record (ReadsYourActivations, so a payment made on paper is
// asked for); it fires once, copies that ability while it is still on
// the stack, and offers new targets (CR 707.10c, Pit Automaton's body).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                    "a968524a-3d2d-4c0c-8f66-bfbfb9ff3c36",
		Name:                        "Dynaheir, Invoker Adept",
		Completeness:                effects.CompletenessFull,
		PrintedKeywords:             []string{"haste"},
		ActivateAsThoughHasteOthers: true,
		ReadsManaSpent:              game.ReadsYourActivations,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: When you next activate an ability that isn't a mana ability this turn by spending four or more mana to activate it, copy that ability. You may choose new targets for the copy.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				return effects.DelayedOnEvent{
					Label:     "Dynaheir, Invoker Adept — copy that ability",
					On:        []game.EventKind{game.EventActivateAbility},
					Condition: dynaheirCondition,
					Body:      effects.CopyTheAbilityBody,
				}.Apply(ctx)
			})),
		}},
	})
}

// dynaheirCondition is "you next activate an ability that isn't a mana
// ability by spending four or more mana": an activation by the trigger's
// controller whose stack item records four or more mana paid. An on-disk
// key.
var dynaheirCondition = game.DelayedCondition("dynaheir/next-activation-four-mana",
	func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
		return ev.Kind == game.EventActivateAbility && ev.Actor == dt.Controller &&
			g.StackItemForEffect(ev.StackItemID) != nil &&
			g.StackItemPaidForEffect(ev.StackItemID).ManaSpentCount() >= 4
	})
