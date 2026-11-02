package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const (
	spinneretCounterLabel = "Spinneret and Spiderling — you attacked with two or more Spiders: a +1/+1 counter on it"
	spinneretDamageLabel  = "Spinneret and Spiderling — it dealt 4 or more damage: exile the top card; you may play it until the end of your next turn"
)

// Spinneret and Spiderling — Legendary Creature — Spider Human Hero {R},
// 1/2 (Marvel's Spider-Man promos):
//
//	"Whenever you attack with two or more Spiders, put a +1/+1 counter
//	 on Spinneret and Spiderling.
//	 Whenever Spinneret and Spiderling deals 4 or more damage, exile the
//	 top card of your library. Until the end of your next turn, you may
//	 play that card."
//
// Both triggers wait for the event batch to settle (AtBatchEnd), because
// the engine opens one event per attacker and one per (source, recipient)
// of damage.
//
//   - "Attack with two or more Spiders" counts the EventAttack events of
//     one declaration whose attacker is a Spider you control, read
//     through the layered view (changelings and granted types count), and
//     fires once if there are at least two (ThatManyEventsAtLeast). It
//     does not matter whether Spinneret is one of them. The counter goes
//     on Spinneret only while it is still the same object on the
//     battlefield.
//   - "Deals 4 or more damage" is the owner's reading (2026-11-02): all
//     the damage Spinneret deals at the same time (one combat damage
//     step, or one damage instruction) is summed across every recipient,
//     so 2 to a blocker and 2 to a player is one 4, and 3 does not
//     trigger. Damage dealt at different times (the first-strike step and
//     the regular step, or two separate abilities) is never summed
//     (game.BatchTotalDamageDealt, ThatMuchDamageDealtAtLeast, CR 120.3,
//     CR 510.2). Prevented damage is not dealt, so it does not count.
//     One event is one trigger.
//   - The exile and the permission are Crimson Operative's:
//     ExileTopNUntilYourNextTurn, "play" so a land is not stranded, the
//     permission lasting through the end of your next turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "daca1d5e-349d-4e8a-abc7-427dab43fb0c",
		Name:         "Spinneret and Spiderling",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			withAtBatchEnd(On(game.EventAttack, attackedWithASpider, spinneretCounterLabel,
				putACounterOnThis(game.CounterPlusOne)), ThatManyEventsAtLeast(2)),
			withAtBatchEnd(On(game.EventDealDamage, spinneretDealtDamage, spinneretDamageLabel,
				Do(Instruction(func(ctx *Context) error { return ExileTopNUntilYourNextTurn(ctx, 1) }))),
				ThatMuchDamageDealtAtLeast(4)),
		},
	})
}

// attackedWithASpider is "whenever you attack with … Spiders", read per
// attacker: declared by you, and a Spider creature in the layered view.
func attackedWithASpider(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.HasSubtype("Spider")
}

// spinneretDealtDamage is one slice of the damage Spinneret dealt: a
// damage event with it as the source and something actually dealt.
func spinneretDealtDamage(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventDealDamage && source != nil && ev.Source == source.InstanceID && ev.Amount > 0
}
