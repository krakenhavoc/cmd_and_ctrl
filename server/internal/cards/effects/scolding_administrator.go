package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scolding Administrator — Creature — Dwarf Cleric {W}{B}, 2/2 (Secrets of
// Strixhaven, #4456):
//
//	"Menace
//	 Repartee — Whenever you cast an instant or sorcery spell that targets a
//	 creature, put a +1/+1 counter on this creature.
//	 When this creature dies, if it had counters on it, put those counters on up
//	 to one target creature."
//
// Repartee is Informed Inkwright's trigger (informedInkwrightCast: the cast
// spell's own announced targets); the counter goes on this permanent only
// while it is still the object that triggered (putACounterOnThis). The dies
// trigger is Resourceful Defense's: "if it had counters on it" is an
// intervening-if (CR 603.4) read from last-known information at trigger time
// (g.LastKnownCountersForEffect, CR 603.10a) and the counters it carries are
// read again at resolution off the same record (resourcefulDefensePutThoseCounters),
// every kind as printed. The target is "up to one", so the trigger goes on the
// stack with none chosen and then does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3b8240ad-71d3-4117-8531-6c8e79bffdb9",
		Name:            "Scolding Administrator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, informedInkwrightCast,
				"Scolding Administrator — Repartee: put a +1/+1 counter on this creature",
				putACounterOnThis(game.CounterPlusOne)),
			Targeting(
				On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return cardDied(ev, source) && len(g.LastKnownCountersForEffect(ev.CardID)) > 0
				}, "Scolding Administrator — put the counters it had on up to one target creature",
					resourcefulDefensePutThoseCounters),
				UpToOneTargetCreature("up to one target creature")),
		},
	})
}
