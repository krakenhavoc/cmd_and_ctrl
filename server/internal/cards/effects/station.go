package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// station.go — the card-facing half of Station (CR 702.184, CR 721),
// ADR 0071 decision 2. A station card (a Spacecraft, or a Planet,
// which is a land with station) banks CHARGE COUNTERS and switches
// printed lines on as it passes their thresholds:
//
//	Station (Tap another untapped creature you control: Put charge
//	counters equal to its power on this. Station only as a sorcery.)
//	7+ | Flying
//
// The threshold half is here. The station ABILITY is not: its cost
// is "tap another untapped creature you control", which AbilityCost
// cannot express — that component is #758, and ADR 0071 names the
// field it should add (AbilityCost.TapOthers) so the Station()
// constructor can be written the day it lands. Until then a station
// card ships with its thresholds working and a caveat naming #758,
// and its charge counters are added by hand from the counter menu.
//
// The gate is an ordinary designation, so nothing here is
// station-specific machinery: it is the same ActiveWhen field a Class
// level and a solved Case use, reading Counters["charge"] instead of
// a field on Card. That is the whole point of ADR 0071.

// AtChargeCounters is the CR 721.2a "{n+}" gate: this ability exists
// while the permanent has n or more charge counters. Read LIVE, so a
// Spacecraft that loses charge counters loses the abilities again.
func AtChargeCounters(n int) game.Designation { return game.ChargeCounters(n) }

// --- Station (CR 721) ------------------------------------------------

// SpacecraftAt is CR 721.2b: "{n+} … it's also a creature with base
// power and toughness [power]/[toughness] in addition to its other
// types".
//
// Two statics from one call, because the type change and the P/T set
// genuinely live in different CR 613 layers (4 and 7b) and a single
// entry could not sort into both. Both carry the same charge-counter
// gate, so they switch on and off together.
//
// "In addition to its other types" is why the type is APPENDED rather
// than assigned: The Seriema at seven charge counters is a Legendary
// Artifact Creature — Spacecraft, not a bare creature. Since ADR 0039
// made layer 4 authoritative, that is a real creature to combat,
// targeting and the state-based actions, not a wire-only label.
//
// Summoning sickness needs nothing here. CR 302.6 asks how long the
// permanent has been under its controller's control, not how long it
// has been a creature, and Card.SummonedThisTurn already records
// exactly that.
func SpacecraftAt(n, power, toughness int) []game.StaticAbility {
	gate := AtChargeCounters(n)
	return []game.StaticAbility{
		{
			Layer:      game.Layer4Type,
			ActiveWhen: gate,
			AppliesTo:  selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, t := range c.Types {
					if t == "Creature" {
						return
					}
				}
				c.Types = append(c.Types, "Creature")
			},
		},
		{
			Layer:      game.Layer7PT,
			SubLayer:   game.SubLayer7B_Set,
			ActiveWhen: gate,
			AppliesTo:  selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power = power
				c.Toughness = toughness
			},
		},
	}
}

// ThresholdKeywords is a station card's "{n+} | Flying" line: the
// permanent has those keyword abilities while it has n or more charge
// counters (CR 721.2a). A layer-6 self-grant with the same gate
// SpacecraftAt uses.
func ThresholdKeywords(n int, keywords ...string) game.StaticAbility {
	kws := append([]string(nil), keywords...)
	return game.StaticAbility{
		Layer:      game.Layer6Ability,
		ActiveWhen: AtChargeCounters(n),
		AppliesTo:  selfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			appendKeywordsTo(c, kws)
		},
	}
}
