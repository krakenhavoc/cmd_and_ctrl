package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cemetery Puca — Creature — Shapeshifter {1}{U/B}{U/B}, 1/2 (Shadowmoor, #5591):
//
//	"Whenever a creature dies, you may pay {1}. If you do, this creature
//	 becomes a copy of that creature, except it has this ability."
//
// The trigger is the same one on the Puca and on every copy it becomes: the
// "except" clause (CR 707.9a) grants the ability as a registered bundle, so a
// copy of a copy keeps it. The copy has no duration (CR 611.2a) and is made
// from the dead creature's copiable values as it last existed (CR 707.2).
// It fires for the Puca's own death too, but the copy only happens while
// the Puca is still the permanent the trigger came from (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e93ea61-bc16-4d22-a5e5-e6ba0f7ae14f",
		Name:         "Cemetery Puca",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       cemeteryPucaGrant,
			Text:      "Whenever a creature dies, you may pay {1}. If you do, this creature becomes a copy of that creature, except it has this ability.",
			Triggered: []game.TriggeredAbility{cemeteryPucaTrigger()},
		}},
		Triggered: []game.TriggeredAbility{cemeteryPucaTrigger()},
	})
}

const cemeteryPucaGrant = "cemetery-puca/copy-ability"

func cemeteryPucaTrigger() game.TriggeredAbility {
	return effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
		_, ok := effects.DiedCreature(ev, g)
		return ok
	}, "Cemetery Puca — you may pay {1}: this creature becomes a copy of that creature",
		func(g *game.Game, item *game.StackItem) error {
			return effects.MayPay{
				Chooser:  item.Controller,
				Cost:     "{1}",
				Question: "Pay {1}: Cemetery Puca becomes a copy of the creature that died?",
				Then:     cemeteryPucaPaid,
			}.Apply(effects.NewContext(g, item))
		})
}

// cemeteryPucaPaid is the paid half, keyed: the Puca becomes a copy of the
// creature that died, with this ability. An on-disk key: never renamed,
// never reused.
var cemeteryPucaPaid = effects.PayUnlessBranches("pay-unless/cemetery-puca-copy",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		if ctx.Item == nil || ctx.Item.Trigger == nil || !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
			return nil
		}
		dead := ctx.Item.Trigger.Event.CardID
		if dead == uuid.Nil {
			return nil
		}
		// The creature as it last existed on the battlefield (CR 608.2h,
		// CR 707.2): the only place a token that has ceased to exist is
		// still known.
		var values *game.PrintedValues
		if info, ok := ctx.Game.LastKnownPermanentForEffect(dead); ok && info.Copiable != nil {
			v := info.Copiable.Clone()
			values = &v
		}
		return effects.BecomeCopy{
			Targets:  []uuid.UUID{ctx.Source()},
			Of:       dead,
			Values:   values,
			Duration: effects.CopyIndefinite,
			Except:   func(v *game.PrintedValues) { v.GrantAbility(cemeteryPucaGrant) },
			Label:    "Cemetery Puca — becomes a copy of the creature that died, except it has this ability",
		}.Apply(ctx)
	}, nil)
