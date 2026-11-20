package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nikara, Lair Scavenger — Legendary Creature — Human Cleric {2}{B}, 2/2
// (Commander 2020, #5513):
//
//	"Partner with Yannik, Scavenging Sentinel (When this creature enters,
//	 target player may put Yannik into their hand from their library, then
//	 shuffle.)
//	 Menace
//	 Whenever another creature you control leaves the battlefield, if it had
//	 one or more counters on it, you draw a card and you lose 1 life."
//
// Partner with is the CR 702.124j entry search (partner_with.go). The leave
// trigger is Lunarch Veteran's "another creature you control leaves",
// whatever zone it went to, with Ambitious Augmenter's intervening "if" read
// off the creature's last-known counters (CR 603.10a): any kind counts. The
// counters it had cannot change once it has left, so the resolution check
// of CR 603.4 always agrees with the trigger-time one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2a2376b8-6962-4f9e-ba36-fd7de87dc77d",
		Name:            "Nikara, Lair Scavenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			PartnerWith("Nikara, Lair Scavenger", "Yannik, Scavenging Sentinel"),
			On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return anotherCreatureYouControlLeft(ev, source, lki, g) && len(g.LastKnownCountersForEffect(ev.CardID)) > 0
			}, "Nikara, Lair Scavenger — you draw a card and you lose 1 life",
				Do(DrawCards{N: 1}, GainLife{Amount: -1})),
		},
	})
}
