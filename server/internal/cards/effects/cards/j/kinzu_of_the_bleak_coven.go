package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kinzu of the Bleak Coven — Legendary Creature — Phyrexian Vampire
// {4}{B}, 5/4:
//
//	"Flying
//	 Whenever another nontoken creature you control dies, you may pay 2
//	 life and exile it. If you do, create a token that's a copy of that
//	 creature, except it's 1/1 and has toxic 1."
//
// Nightmare Shepherd's exile-then-copy with a life payment in front of
// the exile (MayPay, LifePayment): a controller who can't pay 2 life
// (CR 119.4) is not asked. The "if you do" is the whole "pay 2 life and
// exile it", so a card that is no longer the object that died (CR 400.7)
// is not exiled and makes no token; the 2 life are still paid. The copy
// is the creature's copiable values as it last existed (CR 707.2,
// 608.2h); "1/1 and has toxic 1" is part of the copiable values (CR
// 707.9a), and toxic adds to any toxic the creature already had (CR
// 702.164b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e4f4c41b-2533-497f-96bb-7083ef597268",
		Name:            "Kinzu of the Bleak Coven",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.HofriGhostforgeAnotherNontokenCreatureDied,
				"Kinzu of the Bleak Coven — pay 2 life and exile it; if you do, create a 1/1 toxic 1 token copy",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Pay:      effects.LifePayment(2),
						Question: "Kinzu of the Bleak Coven — pay 2 life and exile that creature to create a 1/1 token copy of it with toxic 1?",
						Then:     kinzuPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// kinzuCopyTail is "create a token that's a copy of that creature, except
// it's 1/1 and has toxic 1", run once the exile happened. An on-disk key.
var kinzuCopyTail = effects.CardTail("kinzu-of-the-bleak-coven/copy", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) {
		return nil
	}
	obj := ctx.Trigger().Object
	if obj == nil {
		return nil
	}
	tmpl, ok := ctx.Game.TokenCopyTemplateForObjectForEffect(obj.Ref())
	if !ok {
		return nil
	}
	tmpl.Power, tmpl.Toughness, tmpl.VariableToughness = 1, 1, false
	tmpl.Keywords = append(append([]string(nil), tmpl.Keywords...), "toxic 1")
	return effects.CreateTokenCopies(ctx, ctx.Controller(), tmpl, 1)
})

// kinzuPaid is the paid branch: exile the creature that died, if it is
// still that object in the graveyard, then copy it.
var kinzuPaid = effects.PayUnlessBranches("pay-unless/kinzu-of-the-bleak-coven-exile",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		dead := ctx.Trigger().Event.CardID
		if !effects.DiedCardStillInGraveyard(ctx, dead) {
			return nil
		}
		return effects.ExileTarget{Target: dead, ThenTail: ctx.ItemTail(kinzuCopyTail)}.Apply(ctx)
	}, nil)
