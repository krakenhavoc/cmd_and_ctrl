package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brazen Bluff — Instant {R}{R} (Star Trek, #4973):
//
//	"Choose target noncreature spell an opponent controls. That player
//	 may exile it. If they don't, you may cast a noncreature spell from
//	 your hand without paying its mana cost."
//
// The spell's controller is asked as Brazen Bluff resolves (a
// MayChoice addressed to them). On a yes the spell is exiled from the
// stack. On a no the caster is offered the free cast of a noncreature
// spell from their hand during the same resolution (CR 608.2g), judged
// on the face being cast; declining casts nothing. A target gone by
// resolution makes the whole spell do nothing (CR 608.2b): nobody is
// asked and nothing is cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ac20f38-adc4-40a9-95c9-bbdc689b928b",
		Name:         "Brazen Bluff",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target noncreature spell an opponent controls", effects.Noncreature(), effects.OpponentControls()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			// The first still-legal target card: here, the spell.
			spell, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			owner := ctx.Game.StackItemForEffect(spell)
			if owner == nil {
				return nil
			}
			return effects.MayChoice{
				Player:   owner.Controller,
				Question: "Brazen Bluff — exile your spell? If you don't, its caster may cast a noncreature spell from their hand for free.",
				YesLabel: "Exile it",
				NoLabel:  "Don't exile it",
				Then:     brazenBluffAnswer,
				Carry:    []uuid.UUID{spell},
			}.Apply(ctx)
		},
	})
}

// brazenBluffFreeCast is the "you may cast a noncreature spell from your
// hand without paying its mana cost" offer. Its key is an on-disk
// identity: never renamed, never reused.
var brazenBluffFreeCast = effects.RegisterFreeHandCastOffer("choose-cards/brazen-bluff-cast", "Brazen Bluff")

// brazenBluffAnswer is the spell controller's reply. An on-disk
// identity: never renamed, never reused.
var brazenBluffAnswer = effects.MayYesNo("brazen-bluff/exile-or-free-cast",
	func(ctx *effects.Context, r game.ConfirmAnswer) error {
		if len(r.Carry) == 0 || ctx.Game.StackItemForEffect(r.Carry[0]) == nil {
			return nil
		}
		return ctx.Game.ExileSpellForEffect(r.Carry[0])
	},
	func(ctx *effects.Context, _ game.ConfirmAnswer) error {
		effects.OfferFreeHandCast(ctx, ctx.Controller(), brazenBluffFreeCast, "a noncreature spell",
			effects.FreeCastSpell{Filter: game.PermissionFilter{NoncreatureOnly: true}})
		return nil
	})
