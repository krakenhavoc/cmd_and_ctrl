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
// Both halves are here. The thresholds are ADR 0071 decision 2; the
// station ABILITY is Station() below, on #758's tap-another cost
// component (AbilityCost.TapOthers), with the tapped creature carried
// on the payment record (#759, ADR 0071 addendum 2026-09-23).
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

// StationLabel is the station ability as the menu shows it. One
// string for every station card, because CR 702.184a defines the
// keyword with one text: the printed reminder says "this Spacecraft"
// or "this Planet", and the rule says "this permanent".
const StationLabel = "Station (Tap another untapped creature you control: Put charge counters equal to its power on this permanent. Activate only as a sorcery.)"

// Station is the keyword ability itself (CR 702.184a):
//
//	"Tap another untapped creature you control: Put a number of charge
//	counters on this permanent equal to the tapped creature's power.
//	Activate only as a sorcery."
//
// Every part of it is an existing shape:
//
//   - The cost is TapAnotherUntapped — #758's TapOthersCost with a
//     count of one, a creature filter and ExcludeSource for the word
//     "another". Not the {T} symbol, so a creature cast this turn may
//     station (CR 302.6), and not a target, so a hexproof one may too.
//   - "Activate only as a sorcery" is ActivatedAbility.SorcerySpeed.
//   - CR 721.4: the ability has no ActiveWhen gate — a station card
//     has it at every counter count, including zero and including
//     past its last threshold.
//
// WHICH power is the one decision, and CR 608.2h makes it rather than
// the payment: the tapped creature's power AS THE ABILITY RESOLVES,
// or as it last existed on the battlefield if it has left (the Edge
// of Eternities release notes say both in as many words). So the
// payment record carries the creature's identity, not a banked number
// — Context.TappedPower reads it live, or the last-known value the
// exit hook wrote. A creature pumped in response stations for more.
//
// A power of zero or less puts nothing on and takes nothing off (the
// release notes: "no charge counters are put onto or removed from"
// the permanent) — a negative AddCounter would REMOVE counters, which
// is the one outcome the clause rules out.
//
// The counters go on through the CR 614 counter window, so Doubling
// Season doubles them, as it does in paper.
//
// A source that has stopped being this permanent by resolution —
// destroyed in response, or bounced and replayed as a new object
// (CR 400.7) — gets nothing: there is no "this permanent" left to put
// counters on, and AddCounterForEffect would otherwise stamp them
// onto a card in a graveyard.
//
// Out of scope, stated in ADR 0071: CR 702.184c modifiers (Tapestry
// Warden reads toughness instead) — a static over this ability, which
// no catalog card has yet.
func Station() ActivatedAbility {
	return ActivatedAbility{
		Label:        StationLabel,
		Cost:         TapAnotherUntapped("another untapped creature you control", Creature()),
		SorcerySpeed: true,
		Effect:       stationEffect,
	}
}

// stationEffect is Station's resolution, package-level so the stack
// item captures nothing (undo restores a cloned game and the effect
// has to resolve against that one).
func stationEffect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	power, ok := ctx.TappedPower()
	if !ok || power <= 0 {
		return nil
	}
	if g.AbilitySourceGoneForEffect(item) {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterCharge, N: power}.Apply(ctx)
}
