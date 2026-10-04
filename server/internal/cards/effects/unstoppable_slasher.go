package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unstoppable Slasher — Creature — Zombie Assassin {2}{B}, 2/3:
//
//	"Deathtouch
//	 Whenever this creature deals combat damage to a player, they lose
//	 half their life, rounded up.
//	 When this creature dies, if it had no counters on it, return it to
//	 the battlefield tapped under its owner's control with two stun
//	 counters on it."
//
// Deathtouch rides PrintedKeywords. The damage trigger is the shared
// WheneverThisDealsCombatDamageToAPlayer with Havoc Festival's
// halving (playerLosesHalfTheirLife: 21 loses 11), aimed at the player
// the damage event names ("they").
//
// The dies trigger is an intervening-if (CR 603.4) on the permanent's
// last-known counters (CR 603.10a, LastKnownCountersForEffect): the
// move out of the battlefield has already cleared them on the card, so
// the record the engine takes as the creature leaves is the answer. A
// creature that died with a +1/+1 counter, a stun counter or any other
// does not come back. It returns tapped under its owner's control with
// the two stun counters entering with it (CR 614.1c), so Hardened
// Scales sees them; a returned Slasher is a new object (CR 400.7) with
// counters, so it will not return a second time.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6662968e-ca9e-454c-be44-24ad10e43d3e",
		Name:            "Unstoppable Slasher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Unstoppable Slasher — that player loses half their life, rounded up",
				triggeringTargetLosesHalfTheirLife),
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return cardDied(ev, source) && len(g.LastKnownCountersForEffect(ev.CardID)) == 0
			}, "Unstoppable Slasher — return it tapped with two stun counters",
				returnThisCreatureFromGraveyard(map[string]int{game.CounterStun: 2}, nil)),
		},
	})
}
