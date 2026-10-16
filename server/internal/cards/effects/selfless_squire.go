package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Selfless Squire — Creature — Human Soldier {3}{W}, 1/1 (EDHREC rank
// 2499):
//
//	"Flash
//	 When this creature enters, prevent all damage that would be dealt to
//	 you this turn.
//	 Whenever damage that would be dealt to you is prevented, put that
//	 many +1/+1 counters on this creature."
//
// The enters trigger is Riot Control's turn-long shield on you (ADR 0108
// §7). The second ability is a CR 615.13 trigger (#2047): it triggers
// once each time a prevention effect — the Squire's own shield or any
// other, protection included — prevents damage that would be dealt to
// you, with the total that effect prevented from one simultaneous
// instance of damage, so three attackers stopped by the shield are one
// trigger for all of it. Damage that can't be prevented is dealt, and
// gives no counters. The counters go on the Squire only while it is
// still the permanent that triggered (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7d079c78-b2a9-444f-ab54-c61c041baa6d",
		Name:            "Selfless Squire",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Selfless Squire — prevent all damage that would be dealt to you this turn",
				func(g *game.Game, item *game.StackItem) error {
					return PreventDamageFromSource{Protect: ShieldYou}.Apply(NewContext(g, item))
				}),
			WheneverDamageToYouIsPrevented("Selfless Squire — put that many +1/+1 counters on it",
				selflessSquireCounters),
		},
	})
}

// selflessSquireCounters is "put that many +1/+1 counters on this
// creature": the damage the triggering application prevented.
func selflessSquireCounters(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.Amount <= 0 || !sourceIsStillThisPermanent(g, item) {
		return nil
	}
	return CounterOnThis{Kind: game.CounterPlusOne, N: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
}
