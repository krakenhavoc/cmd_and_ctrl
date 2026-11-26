package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kurkesh, Onakke Ancient — Legendary Creature — Ogre Spirit {2}{R}{R}, 4/3
// (Magic 2015, #5556):
//
//	"Whenever you activate an ability of an artifact, if it isn't a mana
//	 ability, you may pay {R}. If you do, copy that ability. You may choose new
//	 targets for the copy."
//
// Rings of Brighthearth's trigger (the activation event, which only a CR 602
// activation emits, so a mana ability never triggers it, and a copy is created
// rather than activated, so it can't loop) narrowed to abilities whose source
// is an artifact and with {R} to pay. The artifact is judged by its effective
// types, or as it last existed if it was sacrificed as a cost. The copy is the
// activation's own stack item, read off the triggering event, still on the
// stack under an open prompt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea7302b3-c9a2-4bc4-81b8-3a8d1d7c3b8b",
		Name:         "Kurkesh, Onakke Ancient",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventActivateAbility, effects.AllOf(effects.ByYou, effects.S82OldB359AbilityOfAnArtifact),
				"Kurkesh, Onakke Ancient — pay {R} to copy that ability", kurkeshCopyActivatedAbility),
		},
	})
}

// kurkeshCopyActivatedAbility is the resolution: offer the {R}, and on payment
// copy the ability the trigger fired on.
func kurkeshCopyActivatedAbility(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if ctx.Trigger().Event.StackItemID == uuid.Nil {
		return nil
	}
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{R}",
		Question: "Kurkesh, Onakke Ancient — pay {R} to copy that ability?",
		Then:     kurkeshPaid,
	}.Apply(ctx)
}

// kurkeshPaid is the paid half, keyed: the ability is the asking trigger's
// event.
var kurkeshPaid = effects.PayUnlessBranches("pay-unless/kurkesh-copy-ability",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CopyAbility{
			ItemID:           ctx.Trigger().Event.StackItemID,
			Controller:       ctx.Controller(),
			ChooseNewTargets: true,
		}.Apply(ctx)
	}, nil)
