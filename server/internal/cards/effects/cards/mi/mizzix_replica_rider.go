package cardsmi

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mizzix, Replica Rider — Legendary Creature — Goblin Wizard {4}{R}, 4/5
// (Jumpstart 2022, S82 GAP2 b05 #5948, #6088):
//
//	"Flying
//	 Whenever you cast a spell from anywhere other than your hand, you may
//	 pay {1}{U/R}. If you do, copy that spell and you may choose new
//	 targets for the copy. If the copy is a permanent spell, it gains haste
//	 and 'At the beginning of your end step, sacrifice this permanent.'
//	 (A copy of a permanent spell becomes a token.)"
//
// The trigger watches its controller's casts whose spell left a zone other
// than their hand (B25CastFromNotHand). On resolution the controller may
// pay {1}{U/R} (MayPay, a registered branch); if they do, the spell is
// copied from last-known information (CR 608.2h: "copy that spell" names
// it, it does not target it), with the CR 707.10c offer of new targets. A
// copy of a permanent spell gains haste and this card's bundle "At the
// beginning of your end step, sacrifice this permanent." as it is put on
// the stack, for as long as it is that object and then the token it
// becomes (CR 400.7a, 707.10f; game.CopyLastKnownSpellGainingForEffect).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "24a7e1d2-0bc9-40a9-a606-f0017bde3dab",
		Name:            "Mizzix, Replica Rider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Grants: []effects.AbilityGrant{{
			Key: mizzixEndStepSacrificeGrant,
			Triggered: []game.TriggeredAbility{
				effects.AtYourEndStep("Mizzix, Replica Rider — sacrifice this permanent", effects.SacrificeThisPermanentAtYourEndStep),
			},
			Text: "At the beginning of your end step, sacrifice this permanent.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B25CastFromNotHand(ev, source.Controller)
			}, "Mizzix, Replica Rider — you may pay {1}{U/R} to copy that spell", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if ctx.Trigger().Event.CardID == uuid.Nil {
					return nil
				}
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{1}{U/R}",
					Question: "Mizzix, Replica Rider — pay {1}{U/R} to copy that spell?",
					Then:     mizzixPaid,
				}.Apply(ctx)
			}),
		},
	})
}

// mizzixEndStepSacrificeGrant is the bundle a copy of a permanent spell
// gains. An on-disk key.
const mizzixEndStepSacrificeGrant = "mizzix-replica-rider/end-step-sacrifice"

// mizzixPaid copies the spell once {1}{U/R} is paid. An on-disk key.
var mizzixPaid = effects.PayUnlessBranches("pay-unless/mizzix-replica-rider-copy",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		spell := ctx.Trigger().Event.CardID
		var keywords, grants []string
		if c, ok := ctx.Game.LookupCardForEffect(spell); ok && c.IsPermanent() {
			keywords, grants = []string{"haste"}, []string{mizzixEndStepSacrificeGrant}
		}
		err := ctx.Game.CopyLastKnownSpellGainingForEffect(ctx.Source(), spell, ctx.Controller(), true, keywords, grants,
			"Mizzix, Replica Rider — the copy gains haste and \"At the beginning of your end step, sacrifice this permanent.\"")
		if errors.Is(err, game.ErrCardNotFound) {
			return nil
		}
		return err
	}, nil)
