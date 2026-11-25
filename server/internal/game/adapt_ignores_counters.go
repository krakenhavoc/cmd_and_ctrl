package game

import "github.com/google/uuid"

// adapt_ignores_counters.go — "The next time target creature adapts
// this turn, it adapts as though it had no +1/+1 counters on it"
// (Biomancer's Familiar, #6239).
//
//	701.46a "Adapt N" means "If this permanent has no +1/+1 counters on
//	        it, put N +1/+1 counters on it."
//
// A ModAdaptIgnoresCounters record pinned to the creature (CR 400.7)
// until cleanup (CR 514.2). The next adapt of that object spends it
// (TakeAdaptIgnoresCountersForEffect) and skips the "no +1/+1 counters"
// check, so a creature with two counters adapting 2 gets two more. The
// adapt after that is an ordinary one. An adapt spends it whatever it
// does, since the creature still adapts. Not an ability of the creature,
// so losing all abilities does not end it.

// ModAdaptIgnoresCounters is that one-shot record. Read at one gate,
// TakeAdaptIgnoresCountersForEffect, which every adapt calls (cards/
// effects AdaptN). Reads nothing. An on-disk identity: never renamed,
// never reused.
const ModAdaptIgnoresCounters ModKind = "adaptIgnoresCounters"

// AdaptIgnoresCountersThisTurnForEffect registers the record on the
// permanent `cardID`. Registers nothing for a card that is not on the
// battlefield and returns false for it.
//
// Caller must hold g.mu (write).
func (g *Game) AdaptIgnoresCountersThisTurnForEffect(sourceID, cardID uuid.UUID, label string) bool {
	affected := g.PinnedObjectsLocked(cardID)
	if len(affected) == 0 {
		return false
	}
	return g.appendScopedEffectLocked(sourceID, affected, ScopeNone, uuid.Nil,
		[]Mod{{Kind: ModAdaptIgnoresCounters}}, g.PinnedTo(g.UntilEndOfTurnDuration(), cardID),
		label, timeNowUnixNano())
}

// TakeAdaptIgnoresCountersForEffect is the adapt's gate: reports
// whether a live ModAdaptIgnoresCounters record covers the permanent
// `cardID`, and drops the first one that does ("the next time"). False
// for a card that is not on the battlefield.
//
// Caller must hold g.mu (write).
func (g *Game) TakeAdaptIgnoresCountersForEffect(cardID uuid.UUID) bool {
	c := findBattlefieldCard(g, cardID)
	if c == nil {
		return false
	}
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		if e.Scope == ScopeNone && scopedEffectHasMod(e, ModAdaptIgnoresCounters) && affectedPredicate(e.Affected)(c, g, nil) {
			seq := e.Seq
			g.dropScopedEffectsLocked(func(_ int, x ScopedEffect) bool { return x.Seq == seq })
			return true
		}
	}
	return false
}
