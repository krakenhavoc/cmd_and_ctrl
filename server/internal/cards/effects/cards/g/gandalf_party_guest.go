package cardsg

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gandalf, Party Guest — Legendary Creature — Avatar Wizard {1}{U}{R}{W},
// 3/4 (The Hobbit Commander):
//
//	"At the beginning of combat on your turn, you may cast an instant or
//	 sorcery spell with mana value X or less from your hand without paying
//	 its mana cost, where X is twice the number of legendary Wizards you
//	 control."
//
// Glamdring's offer with a different number: the beginning-of-combat trigger
// counts the legendary Wizards its controller controls as it resolves (Gandalf
// himself is one while he is on the battlefield; a Gandalf that left in
// response counts the Wizards that remain), doubles it, and asks for a
// card in hand with an instant or sorcery spell of at most that mana value,
// judged on the face being cast (#4509). The cast is the
// engine's own, so its targets, modes and X are chosen as it goes on the
// stack, and declining leaves the card in hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "11e30c21-41a8-4f2f-b37a-db10656b7cad",
		Name:         "Gandalf, Party Guest",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Gandalf, Party Guest — you may cast an instant or sorcery spell from your hand with mana value at most twice your legendary Wizards without paying its mana cost",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					wizards := effects.PermanentsControlledByMatching(ctx.Game, ctx.Controller(), effects.And(effects.Legendary(), effects.Subtype("Wizard")))
					x := 2 * len(wizards)
					effects.OfferFreeHandCast(ctx, ctx.Controller(), gandalfPartyGuestFreeCast,
						"an instant or sorcery spell with mana value "+strconv.Itoa(x)+" or less",
						effects.FreeInstantOrSorcery().WithManaValueAtMost(x))
					return nil
				}))),
		},
	})
}

var gandalfPartyGuestFreeCast = effects.RegisterFreeHandCastOffer("choose-cards/gandalf-party-guest-cast", "Gandalf, Party Guest")
