package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ogre-Head Helm — Artifact Creature — Equipment Ogre {1}{R}, 2/2:
//
//	"Equipped creature gets +2/+2.
//	 Whenever this creature or equipped creature deals combat damage to
//	 a player, you may sacrifice it. If you do, discard your hand, then
//	 draw three cards.
//	 Reconfigure {3}"
//
// "It" is the Helm, whichever creature dealt the damage: the trigger is
// the Helm's, and "sacrifice it" names the permanent with the ability.
// The question is asked as the trigger resolves (MayChoice), and the
// discard and the draw hang off the sacrifice actually happening
// (SacrificePermanent.Then), so a Helm that already left, or came back
// as a new object, does nothing. Reconfigure is reconfigure.go (#2639).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b73c297a-1c39-450f-99c3-c8daa31c5f3f",
		Name:         "Ogre-Head Helm",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisOrEquippedCreatureDealsCombatDamageToAPlayer,
				"Ogre-Head Helm — you may sacrifice it to discard your hand and draw three", ogreHeadHelmMaySacrifice),
		},
		Activated: effects.Reconfigure("{3}"),
	})
}

func ogreHeadHelmMaySacrifice(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if !effects.OnBattlefield(g, ctx.Source()) {
		return nil
	}
	return effects.MayChoice{
		Question: "Ogre-Head Helm — sacrifice it, discard your hand, then draw three cards?",
		YesLabel: "Sacrifice",
		NoLabel:  "Decline",
		Then:     ogreHeadHelmAnswer,
	}.Apply(ctx)
}

// ogreHeadHelmWheel is "discard your hand, then draw three cards" after the
// sacrifice: the draw waits for a prompt a discard asks (a replacement offer such as Library of Leng's; CR 608.2c,
// #4155).
var ogreHeadHelmWheel = effects.RegisterThenSteps("then-steps/ogre-head-helm-wheel",
	effects.Instruction(func(ctx *effects.Context) error {
		_, err := effects.DiscardWholeHand(ctx.Game, ctx.Controller())
		return err
	}),
	effects.Instruction(func(ctx *effects.Context) error { return ctx.Game.DrawNForEffect(ctx.Controller(), 3) }),
)

// ogreHeadHelmAnswer is the "yes": sacrifice the Helm, then discard the hand
// and draw three. Keyed (#4291).
var ogreHeadHelmAnswer = effects.MayYes("confirm/ogre-head-helm-sacrifice", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	return effects.SacrificePermanent{Target: ctx.Source(), ThenTail: ctx.IfItWentThen(ogreHeadHelmWheel)}.Apply(ctx)
})
