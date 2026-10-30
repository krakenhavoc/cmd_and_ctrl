package game

import "github.com/google/uuid"

// finality_counters.go — CR 122.1h (#4575):
//
//	122.1h One or more finality counters on a permanent create a single
//	       replacement effect that stops the permanent from going to the
//	       graveyard. That effect is "If this permanent would be put into
//	       a graveyard from the battlefield, exile it instead."
//
// One engine-owned replacement (builtin_replacements.go), because the
// effect is the counter's, printed in the rules, and outlives whatever put
// the counter on: Grim Reaper, Lethal Legionnaire returns another card with
// one, and a Grim Repriser that comes back from the graveyard carries its
// own. Before this the rule lived only in the two catalog cards that print
// a finality counter, as self-replacements, so a counter put on a permanent
// by anything else did nothing. "Single" is the registration, as it is for
// the shield counter (shield_counters.go): one built-in however many
// counters, so two counters on a permanent never ask a CR 616 question of
// each other.
//
// What it replaces is the MOVE, however the permanent came to be moved:
// destroy, sacrifice, lethal damage, zero toughness, the legend rule, a
// wipe's leg. A permanent exiled instead never died, so it fires no dies
// trigger and no "put into a graveyard" trigger (CR 700.4, CR 614.6). A
// token with a counter is exiled too and then ceases to exist (CR 111.7).
//
// It is not Preemptive: the permanent's controller orders it against any
// other replacement that applies to the same move (CR 616.1), a Rest in
// Peace or a Liesa that also exiles it, or a regeneration shield. A
// redirect to exile ends the destruction (ADR 0013, the 2026-10-29
// amendment on #4526), so a shield or umbra armor ordered after it no
// longer applies, and one ordered before it cancels the move and the
// counter stays put.

// finalityCounterLabel names the built-in in a CR 616 prompt, and is how
// the trigger-order check (trigger_independence.go) knows it.
const finalityCounterLabel = "Finality counter"

// finalityCounterReplacement is CR 122.1h's replacement effect.
var finalityCounterReplacement = ReplacementEffect{
	Watches: []EventKind{EventZoneMove},
	AppliesTo: func(ev *ReplacementEvent, g *Game, _ *Card) bool {
		return g.finalityCounterExilesLocked(ev)
	},
	Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
		ev.NewZone = ZoneExile
		ev.NewZoneOwner = uuid.Nil
		return nil
	},
	Controller: func(ev *ReplacementEvent, g *Game, _ *Card) uuid.UUID {
		return g.controllerOfBattlefieldCardLocked(ev.CardID)
	},
	Label: finalityCounterLabel,
}

// finalityCounterExilesLocked is the replacement's AppliesTo: a permanent
// with a finality counter on it that would be put into a graveyard from
// the battlefield. The destination is read off the event, so a move that
// another replacement has already sent to exile (or anywhere else) is no
// longer one this applies to (CR 616.1f).
//
// Caller must hold g.mu.
func (g *Game) finalityCounterExilesLocked(ev *ReplacementEvent) bool {
	if ev == nil || ev.Kind != RepEventMove || ev.OldZone != ZoneBattlefield || ev.NewZone != ZoneGraveyard {
		return false
	}
	c := findBattlefieldCard(g, ev.CardID)
	return c != nil && c.Counters[CounterFinality] > 0
}

// anyFinalityCounterLocked reports whether any permanent has a finality
// counter on it: the only board on which CR 122.1h's effect can apply.
//
// Caller must hold g.mu.
func (g *Game) anyFinalityCounterLocked() bool {
	if g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Counters[CounterFinality] > 0 {
			return true
		}
	}
	return false
}
