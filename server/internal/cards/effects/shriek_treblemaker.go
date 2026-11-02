package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shriek, Treblemaker — Legendary Creature — Mutant Villain {2}{B/R}, 2/3
// (Marvel's Spider-Man):
//
//	"At the beginning of your first main phase, you may discard a card.
//	 When you do, target creature can't block this turn.
//	 Sonic Blast — Whenever a creature an opponent controls dies, Shriek
//	 deals 1 damage to that player."
//
// The first ability is a may-pay whose payment is the discard (Bullseye,
// Death Dealer's shape): a controller with an empty hand is not asked, and
// paying creates the CR 603.12 reflexive trigger, which picks its target as
// it goes on the stack (CR 603.3d) and, as it resolves, stops that creature
// blocking for the turn. The reflexive half is mandatory. Sonic Blast reads
// the controller the dying creature had as it left (last-known information,
// CR 608.2h) and hits that player, once per creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5773ae05-44bc-423c-846c-f323dd8a5079",
		Name:         "Shriek, Treblemaker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourPrecombatMain("Shriek, Treblemaker — you may discard a card; when you do, target creature can't block this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MayPay{
						Chooser:  ctx.Controller(),
						Pay:      DiscardPayment(1),
						Question: "Shriek, Treblemaker — discard a card to make a creature unable to block this turn?",
						Then:     shriekPaid,
					}.Apply(ctx)
				}),
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b18OpponentsCreatureDied(ev, source, g)
			}, "Shriek, Treblemaker — Sonic Blast: 1 damage to that player", shriekSonicBlast),
		},
	})
}

// shriekPaid is the discard's "when you do".
var shriekPaid = PayUnlessBranches("pay-unless/shriek-treblemaker-cant-block",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return ReflexiveTrigger{
			Label: "Shriek, Treblemaker — target creature can't block this turn",
			Body:  shriekCantBlockBody,
		}.Apply(ctx)
	}, nil)

var shriekCantBlockBody = game.ReflexiveBody("shriek-treblemaker/cant-block", shriekCantBlock,
	constTargets(func() *game.TargetSpec { return TargetCreature("target creature") }))

func shriekCantBlock(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return RestrictUntilEOT{Target: id, Restrictions: game.CantBlock, Label: "Shriek, Treblemaker — can't block"}.Apply(ctx)
}

// shriekSonicBlast is 1 damage from Shriek to the player the creature that
// died was controlled by.
func shriekSonicBlast(g *game.Game, item *game.StackItem) error {
	ev := item.Trigger.Event
	dead, ok := diedCreature(ev, g)
	if !ok {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: leftUnderControlOf(ev, dead), Amount: 1}.Apply(NewContext(g, item))
}
