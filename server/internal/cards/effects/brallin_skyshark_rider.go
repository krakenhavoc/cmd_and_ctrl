package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brallin, Skyshark Rider — Legendary Creature — Human Shaman {3}{R}, 3/3
// (Commander 2020, #5513):
//
//	"Partner with Shabraz, the Skyshark (When this creature enters, target
//	 player may put Shabraz into their hand from their library, then
//	 shuffle.)
//	 Whenever you discard a card, put a +1/+1 counter on Brallin and it deals
//	 1 damage to each opponent.
//	 {R}: Target Shark gains trample until end of turn."
//
// Partner with is the CR 702.124j entry search (partner_with.go). The
// discard trigger is Archfiend of Ifnir's "whenever you discard a card",
// once per card (the engine emits one EventDiscardCard per card), with
// Brallin as the damage source; a Brallin that has left deals it as it last
// existed (CR 608.2h) and gets no counter. The activation is Air Marshal's
// targeted keyword grant narrowed to a Shark.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "464df46a-079b-47c1-95e7-930160abef2f",
		Name:         "Brallin, Skyshark Rider",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			PartnerWith("Brallin, Skyshark Rider", "Shabraz, the Skyshark"),
			On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return discardedByYou(ev, source)
			}, "Brallin, Skyshark Rider — put a +1/+1 counter on it and it deals 1 damage to each opponent",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}, DamageEachOpponent{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{R}: Target Shark gains trample until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "trample")),
			Cost:    ManaCost("{R}"),
			Targets: TargetCreature("target Shark", Subtype("Shark")),
			Effect:  ebTargetCreatureGainsUntilEOT("trample", "Brallin, Skyshark Rider — trample until end of turn"),
		}},
	})
}
