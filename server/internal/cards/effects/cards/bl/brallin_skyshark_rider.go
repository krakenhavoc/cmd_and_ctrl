package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "464df46a-079b-47c1-95e7-930160abef2f",
		Name:         "Brallin, Skyshark Rider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.PartnerWith("Brallin, Skyshark Rider", "Shabraz, the Skyshark"),
			effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.DiscardedByYou(ev, source)
			}, "Brallin, Skyshark Rider — put a +1/+1 counter on it and it deals 1 damage to each opponent",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}, effects.DamageEachOpponent{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Target Shark gains trample until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "trample")),
			Cost:    effects.ManaCost("{R}"),
			Targets: effects.TargetCreature("target Shark", effects.Subtype("Shark")),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("trample", "Brallin, Skyshark Rider — trample until end of turn"),
		}},
	})
}
