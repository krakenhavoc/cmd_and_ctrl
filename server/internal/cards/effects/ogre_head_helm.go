package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "b73c297a-1c39-450f-99c3-c8daa31c5f3f",
		Name:         "Ogre-Head Helm",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 2)},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, ThisOrEquippedCreatureDealsCombatDamageToAPlayer,
				"Ogre-Head Helm — you may sacrifice it to discard your hand and draw three", ogreHeadHelmMaySacrifice),
		},
		Activated: Reconfigure("{3}"),
	})
}

func ogreHeadHelmMaySacrifice(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !onBattlefield(g, ctx.Source()) {
		return nil
	}
	return MayChoice{
		Question: "Ogre-Head Helm — sacrifice it, discard your hand, then draw three cards?",
		YesLabel: "Sacrifice",
		NoLabel:  "Decline",
		OnYes: func(ctx *Context) error {
			return SacrificePermanent{
				Target: ctx.Source(),
				Then: func(ctx *Context, sacrificed bool) error {
					if !sacrificed {
						return nil
					}
					if _, err := discardWholeHand(ctx.Game, ctx.Controller()); err != nil {
						return err
					}
					return ctx.Game.DrawNForEffect(ctx.Controller(), 3)
				},
			}.Apply(ctx)
		},
	}.Apply(ctx)
}
