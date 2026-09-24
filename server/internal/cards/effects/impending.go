package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// impending.go — impending (CR 702.176), the Duskmourn keyword the
// Overlord cycle prints.
//
//	"Impending N—[cost] (If you cast this spell for its impending
//	 cost, it enters with N time counters and isn't a creature until
//	 the last is removed. At the beginning of your end step, remove a
//	 time counter from it.)"
//
// Like suspend, the interesting part is how little of it is new. It
// is THREE existing shapes fastened together, and the fastening is
// the whole reason this is a constructor rather than three lines a
// card file remembers:
//
//  1. An ALTERNATIVE COST whose rider is entry counters. The two
//     fields escape added in S29 (EntersWithCounterName /
//     EntersWithCounterCount) say exactly what impending needs:
//     counters that land only when THAT cost was paid, through the
//     CR 614 entry pipeline, so Doubling Season doubles them and the
//     permanent's own enters-the-battlefield trigger already sees
//     them.
//  2. A SELF-ONLY LAYER 4 TYPE REWRITE gated on the counters —
//     Arixmethes' slumber shape and The Warring Triad's graveyard
//     shape, with a different gate. "Isn't a creature" is the type
//     and the creature subtypes together (notACreature, CR 205.1b).
//  3. A TRIGGERED ABILITY at the controller's end step that takes one
//     counter off, which is an ordinary AddCounter with a negative
//     delta.
//
// # Why the three ship as one call
//
// Each piece alone is a wrong card, and each is wrong in the
// dangerous direction:
//
//   - The cost without the static is a 5/5 for two mana with no
//     drawback whatsoever.
//   - The cost and the static without the trigger is a permanent that
//     is NEVER a creature again — the countdown has no clock.
//   - The static and the trigger without the cost is a card nobody can
//     cast for the discount, which is merely useless.
//
// So Impending takes the whole Spec and returns it with all three
// attached. A card file writes
//
//	Register(Impending(5, "{1}{B}", Spec{…}))
//
// and cannot ship two of the three.
//
// # The clauses are self-consistent with no special-casing
//
// The Overlord cycle's payoff triggers on "enters or attacks". A
// permanent with time counters on it is not a creature, so it cannot
// be declared as an attacker (CR 508.1a, enforced in
// game.AttackerEligible off the effective type line) and the attack
// half simply never fires while the countdown runs. Nothing has to
// say so.
//
// # What is deliberately NOT here
//
// No `ExileOnLeavingStack`, and no permission of any kind: an
// impending permanent that dies goes to the graveyard like any other
// and can be cast again from wherever it lands, for either price.
// Copying flashback's constructor and swapping the key — the mistake
// the escape doc warns about — would ship a card that exiles itself.

// AltCostKeyImpending is the wire key an impending cast rides on
// `cast_spell` and lands on StackItem.AltCost. Exported for the same
// reason game.AltCostKeyEscape is: the coverage probes and any card
// that wants to branch on "was this impended" read it by name rather
// than by a string literal that can drift.
const AltCostKeyImpending = "impending"

// Impending attaches impending N—cost to a card's Spec: the
// alternative cost, the "isn't a creature while time counters remain"
// static, and the end-step countdown trigger.
//
// It APPENDS to Spec.AlternativeCosts, Spec.Static and
// Spec.Triggered rather than replacing them, so a card declares its
// own printed abilities in the Spec literal as usual and wraps the
// whole thing:
//
//	func init() {
//	    Register(Impending(5, "{1}{B}", Spec{
//	        OracleID: "…",
//	        Name:     "Overlord of the Balemurk",
//	        Triggered: []game.TriggeredAbility{ … the printed trigger … },
//	    }))
//	}
//
// n must be at least 1 — "impending 0" is not a thing any card
// prints, and a zero would put no counters on and hand the player the
// discount with no drawback. It panics rather than shipping that,
// at boot, the way Register's own validations do.
func Impending(n int, cost string, spec Spec) Spec {
	if n < 1 {
		panic("effects.Impending: " + spec.Name + " impends with " + strconv.Itoa(n) +
			" time counters — impending N is at least one (CR 702.176a)")
	}
	spec.AlternativeCosts = append(spec.AlternativeCosts, impendingCost(n, cost))
	spec.Static = append(spec.Static, impendingNotACreature())
	spec.Triggered = append(spec.Triggered, impendingCountdown(spec.Name))
	return spec
}

// impendingCost is the CR 702.176a alternative cost: the cheaper
// price, plus the time counters that are the price's real cost.
//
// The counters hang off the COST and not off the card, which is the
// field's whole point and is observable both ways: an Overlord hard
// cast for {3}{B}{B}, reanimated, or blinked enters as the creature
// it prints, with no counters and no countdown, while one cast for
// {1}{B} enters as an enchantment that takes five turns to wake up.
func impendingCost(n int, cost string) game.AlternativeCost {
	return game.AlternativeCost{
		Key:                    AltCostKeyImpending,
		Label:                  "Impending " + strconv.Itoa(n) + "—" + cost,
		ManaCost:               cost,
		EntersWithCounterName:  game.CounterTime,
		EntersWithCounterCount: n,
	}
}

// impendingNotACreature is "it isn't a creature until the last [time
// counter] is removed" (CR 702.176a) — a continuous self-only layer 4
// rewrite read off the live counter pile, not a flag set at entry.
//
// Read off the counters rather than off "was this impended" on
// purpose. It is the counters the rule names, so Clockspinning,
// Vampire Hexmage or anything else that empties the pile early wakes
// the permanent up the same way the end step does, and a card that
// somehow gained time counters without impending is covered too.
func impendingNotACreature() game.StaticAbility {
	return game.StaticAbility{
		Layer: game.Layer4Type,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID && target.Counters[game.CounterTime] > 0
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			notACreature(c)
		},
	}
}

// impendingCountdown is "at the beginning of your end step, remove a
// time counter from it" (CR 702.176a).
//
// Gated on there being a counter to remove, so a woken permanent
// stops putting an ability on the stack every end step for the rest
// of the game. The gate is invisible in the removal itself — taking
// a counter off an empty pile does nothing either way — and visible
// in the log, in the response window it would otherwise open, and to
// anything that counts triggered abilities.
func impendingCountdown(cardName string) game.TriggeredAbility {
	return On(game.EventBeginEndStep, impendingStillTicking,
		cardName+" — remove a time counter", impendingTick)
}

// impendingStillTicking is the trigger's condition: the controller's
// own end step, and a counter left to take off.
func impendingStillTicking(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Actor == source.Controller && source.Counters[game.CounterTime] > 0
}

// impendingTick is the countdown's resolution. Everything is re-read
// against the game as it is (CR 608.2): the permanent may have left
// the battlefield, or something may have taken the last counter off
// in response, and either way the ability does as much as it can,
// which is nothing.
func impendingTick(g *game.Game, item *game.StackItem) error {
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !onBattlefield(g, item.SourceCardID) || c.Counters[game.CounterTime] <= 0 {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterTime, N: -1}.Apply(NewContext(g, item))
}
