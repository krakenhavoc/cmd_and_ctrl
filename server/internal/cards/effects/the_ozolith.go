package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Ozolith — Legendary Artifact {1} (EDHREC rank ~300):
//
//	"Whenever a creature you control leaves the battlefield, if it
//	 had counters on it, put those counters on The Ozolith.
//	 At the beginning of combat on your turn, if The Ozolith has
//	 counters on it, you may move all counters from The Ozolith onto
//	 target creature."
//
// A counters deck's insurance policy: a creature that would otherwise
// take its +1/+1 counters (or its stun counters, its shield counters,
// anything) to the graveyard with it instead banks them here, and
// combat hands them to whichever creature needs them next.
//
// # The first ability is the seam this card proves (#1218)
//
// "If it had counters on it" asks about the DEPARTING creature at the
// moment it left — and CR 400.7's own cleanup (MoveCard, zone.go)
// zeroes Card.Counters as part of ending that object, before this
// trigger (or any trigger) ever sees the event. Characteristic — the
// CR 603.10 LKI snapshot every leaves-the-battlefield trigger is
// judged on — deliberately excludes Counters too (characteristic.go:
// "belongs to other engine subsystems"). So neither of the two
// existing answers to "what was this permanent, a beat ago" carried
// the one fact this card needs, and it is a BYSTANDER's trigger: The
// Ozolith is not the permanent that left, so it cannot read its own
// sourceLKI for this either.
//
// game.Game.lastKnownCounters (triggers.go's snapshotLKILocked,
// alongside lastKnownBattlefield) and LastKnownCountersForEffect are
// the fix — a third LKI map, kept on the GAME and not on the card for
// the same reason the other two are: the departing Card value is
// about to be overwritten out from under whatever holds a copy of it.
// See internal/game/lki_counters_test.go for the engine-level proof
// and the tests below for this card's.
//
// # Both abilities double correctly under Doubling Season
//
// The COLLECTION (creature → Ozolith) and the DISTRIBUTION (Ozolith →
// target) each go through AddCounterForEffect, the one counter-writer
// that rides the CR 614 replacement pipeline (doubling_season.go) —
// each is a "put", so each doubles independently, which is the
// printed ruling. The MATCHING removal from whichever permanent just
// gave its counters away — Ozolith zeroing itself in the second
// ability — also goes through AddCounterForEffect with a negative
// delta: a removal is not a "put" and doubling it is observationally
// a no-op here (every kind present is drained to zero either way,
// AddCounterByForEffect's own delete-at-<=0 floor), so accepting the
// same writer for both directions costs nothing and needs no second
// counter primitive.
//
// # Intervening-if, checked once
//
// Both clauses' "if" is evaluated at trigger time only (AppliesTo),
// the posture every intervening-if card in this catalog takes (see
// Land Tax, batch 03) — not re-checked the instant before resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1946ded1-5f53-409f-b0a6-5433bb0357d2",
		Name:         "The Ozolith",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					left, ok := g.LookupCardForEffect(ev.CardID)
					if !ok || !left.IsCreature() || left.Controller != source.Controller {
						return false
					}
					return len(g.LastKnownCountersForEffect(ev.CardID)) > 0
				},
				Key:    "The Ozolith — put those counters on it",
				Effect: ozolithCollectCounters,
			},
			{
				Watches: []game.EventKind{game.EventStepBegan},
				AppliesTo: AllOf(
					StepBegan(game.StepBeginCombat, true),
					func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						here, ok := g.LookupCardForEffect(source.InstanceID)
						return ok && len(here.Counters) > 0
					},
				),
				Targets: TargetCreature("target creature"),
				Key:     "The Ozolith — move its counters onto that creature",
				Effect:  ozolithMoveCounters,
				OptionalPrompt: &game.TriggerOptionalPrompt{
					Question: "The Ozolith — move all its counters onto target creature?",
				},
			},
		},
	})
}

// ozolithCollectCounters is the first ability's resolution: every
// counter the departed creature had, kind by kind, onto The Ozolith.
//
// "Those counters" is read at RESOLUTION off the departed object's
// last-known information (ctx.TriggeringPermanent, #1379, CR 608.2h)
// rather than frozen into a closure as the ability triggered (ADR 0041
// P9, #1497). It is the same reading: battlefieldExitLocked writes the
// CR 603.10 counters AppliesTo checks (lastKnownCounters) and the
// #1379 record read here in consecutive calls, each a copy of the same
// Card.Counters, and the record is keyed by the object's epoch, so a
// card that has come back since is never read in its place.
func ozolithCollectCounters(g *game.Game, item *game.StackItem) error {
	// #1432: an Ozolith that left and came back is not "it".
	if sourceIsNewObject(g, item) {
		return nil
	}
	left, ok := NewContext(g, item).TriggeringPermanent()
	if !ok {
		return nil
	}
	for kind, n := range left.Counters {
		if n <= 0 {
			continue
		}
		if err := g.AddCounterForEffect(item.SourceCardID, kind, n); err != nil {
			return err
		}
	}
	return nil
}

// ozolithMoveCounters is the second ability's resolution: every kind
// of counter The Ozolith is carrying right now moves onto whatever
// item.Targets[0] names, kind by kind so a mixed pile (some +1/+1,
// some stun) arrives exactly as printed. A target gone in response
// (CR 608.2b) makes this a no-op rather than an error.
func ozolithMoveCounters(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	target := item.Targets[0].ID
	if _, ok := g.LookupCardForEffect(target); !ok || sourceIsNewObject(g, item) { // #1432
		return nil
	}
	source, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		return nil
	}
	counts := make(map[string]int, len(source.Counters))
	for kind, n := range source.Counters {
		counts[kind] = n
	}
	for kind, n := range counts {
		if n <= 0 {
			continue
		}
		if err := g.AddCounterForEffect(target, kind, n); err != nil {
			return err
		}
		if err := g.AddCounterForEffect(item.SourceCardID, kind, -n); err != nil {
			return err
		}
	}
	return nil
}
