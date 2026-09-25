package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nekusar, the Mindrazer — Legendary Creature — Zombie Wizard
// {2}{U}{B}{R}, 2/4 (EDHREC rank 3742):
//
//	"At the beginning of each player's draw step, that player draws
//	 an additional card.
//	 Whenever an opponent draws a card, Nekusar deals 1 damage to
//	 that player."
//
// The wheel commander. Two abilities, each a card already in the
// catalog:
//
//   - The draw-step half is Howling Mine's without the untapped
//     clause: EventBeginDrawStep fires once per draw step with the
//     active player in Actor, after the turn-based draw (CR 504.1
//     does not use the stack), which is what "an ADDITIONAL card"
//     means. EACH player's draw step, so it fires for every seat and
//     the drawer is the event's Actor, captured in Build by value.
//   - The ping is Fate Unraveler's: EventDrawCard fires once per
//     card with the drawer in Actor, so a wheel is one trigger per
//     card, and Nekusar is the damage source, so a Fog-class shield
//     stops it. An opponent's additional draw from the first half
//     fires the second, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8a5e3c8e-8e22-49b9-8ee5-4a36361f0da6",
		Name:         "Nekusar, the Mindrazer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventBeginDrawStep},
				AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return b35DrawStepBegan(ev)
				},
				Key: "Nekusar, the Mindrazer — that player draws an additional card",
				Effect: func(g *game.Game, item *game.StackItem) error {
					drawer := item.Trigger.Event.Actor
					if g.PlayerByIDForEffect(drawer) == nil {
						return nil
					}
					return DrawCards{Player: drawer, N: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Watches: []game.EventKind{game.EventDrawCard},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return b35OpponentDrewACard(ev, source)
				},
				Key: "Nekusar, the Mindrazer — 1 damage to the player who drew",
				Effect: func(g *game.Game, item *game.StackItem) error {
					victim := item.Trigger.Event.Actor
					if g.PlayerByIDForEffect(victim) == nil {
						return nil
					}
					return DealDamage{Source: item.SourceCardID, Target: victim, Amount: 1}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
