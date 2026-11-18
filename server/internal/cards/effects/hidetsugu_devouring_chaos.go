package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hidetsugu, Devouring Chaos — {3}{B} Legendary Creature — Ogre Demon 4/4
// (Kamigawa: Neon Dynasty, S82 NEO batch 20):
//
//	"{B}, Sacrifice a creature: Scry 2.
//	 {2}{R}, {T}: Exile the top card of your library. You may play that card
//	 this turn. When you exile a nonland card this way, Hidetsugu deals
//	 damage equal to the exiled card's mana value to any target."
//
// The second ability exiles one card with the impulse permission (any
// player may be the activator only as its controller, so "you" is the
// controller). The reflexive trigger (CR 603.12) happens only when the
// exiled card is a nonland card; its target is chosen as it goes on the
// stack, and the amount is the exiled card's mana value, read when it was
// exiled (CR 608.2h). The damage's source is Hidetsugu.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d56f24a1-04dc-4769-9941-05bd99508e40",
		Name:         "Hidetsugu, Devouring Chaos",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{B}, Sacrifice a creature: Scry 2.",
				Purpose: game.Purpose{Scry: 2, Answers: game.AnswerSacOutlet},
				Cost:    Plus(ManaCost("{B}"), SacrificeACreature()),
				Effect:  Do(Scry{N: 2}),
			},
			{
				Label:   "{2}{R}, {T}: Exile the top card of your library. You may play that card this turn. When you exile a nonland card this way, Hidetsugu deals damage equal to the exiled card's mana value to any target.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{2}{R}"), TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return b12ImpulseExileForTurnThen(g, item, 1, hidetsuguDevouringExiled)
				},
			},
		},
	})
}

var hidetsuguDevouringDamageBody = game.ReflexiveBody("hidetsugu-devouring-chaos/damage", breechesBlastEffect, constTargets(TargetAny))

var hidetsuguDevouringExiled = CardTail("then/hidetsugu-devouring-chaos-damage", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	card, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok || card.IsLand() {
		return nil
	}
	return ReflexiveTrigger{
		Label:  "Hidetsugu, Devouring Chaos — deals damage equal to the exiled card's mana value to any target",
		Cards:  o.IDs,
		Body:   hidetsuguDevouringDamageBody,
		Params: game.EffectParams{Amount: card.ManaValue()},
	}.Apply(ctx)
})
