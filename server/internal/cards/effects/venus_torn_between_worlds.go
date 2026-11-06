package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Venus, Torn Between Worlds — {4}{G} Legendary Creature — Mutant Frog
// Turtle 5/5 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Whenever Venus is dealt damage, put that many +1/+1 counters on her.
//	 Whenever a creature you control with a counter on it deals combat
//	 damage to a player, you may pay {U}. If you do, draw a card."
//
// The first trigger is the one Strength of Will grants, printed: one
// trigger per batch of damage with the batch's total as "that many", and
// the counters go on only while Venus is still the same permanent (lethal
// damage kills her first). The second is Ray Fillet, Wave Warrior's
// condition (any kind of counter, CR 122.1) with Mentor of the Meek's
// "you may pay": one trigger per creature that connects.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7b3d364d-6581-43de-a80a-d5b07f5d21ed",
		Name:         "Venus, Torn Between Worlds",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return b35SelfWasDealtDamage(ev, source)
				},
				Key:        "Venus — put that many +1/+1 counters on her",
				AtBatchEnd: ThatMuchDamage(),
				Effect:     strengthOfWillCounters,
			},
			On(game.EventDealDamage, yourCreatureWithACounterDealtCombatDamageToAPlayer,
				"Venus — you may pay {U} to draw a card", func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{U}",
						Question: "Venus, Torn Between Worlds — pay {U} to draw a card?",
						Then:     mayPayDrawACard,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// yourCreatureWithACounterDealtCombatDamageToAPlayer is "whenever a
// creature you control with a counter on it deals combat damage to a
// player", the dealer's counters read as the damage is dealt.
func yourCreatureWithACounterDealtCombatDamageToAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dealer, ok := g.LookupCardForEffect(ev.Source)
	return ok && combatDamageToPlayerBy(ev, source.Controller, g) && b41HasACounter()(g, source.Controller, dealer)
}
