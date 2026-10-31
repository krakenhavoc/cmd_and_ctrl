package game

// replacement_turn_mark.go — "the first time you would <event> each turn,
// <replacement>" (Reed Richards, Smartest Man; CR 614.1a, 614.5).
//
// A replacement effect applies to the event it replaces and no other, so
// "the first time each turn" is a memory of whether THIS object's effect
// has already been applied since the turn began. It is a mark on the turn,
// written by the replacement's own Replace and read by its own AppliesTo,
// rather than a count of events: the events a given effect cares about
// (draws that are not the exempt one) are not the events the tally already
// counts.
//
// The mark is keyed by ObjectTallyKey(source, epoch, label), so it is per
// OBJECT: a second permanent with the same effect has a mark of its own
// (CR 614.5 — each is its own effect), and one that left and came back is a
// new object with a fresh one (CR 400.7). It lives in the TurnTally, so it
// is flushed when the turn ends, copied for undo, and carried by the
// snapshot under TurnTally.ReplacementsApplied (additive within schema v9).

// ReplacementAppliedThisTurnForEffect reports whether the replacement
// `label` on the object `source` names has already been applied this turn.
// Caller must hold g.mu.
func (g *Game) ReplacementAppliedThisTurnForEffect(source *Card, label string) bool {
	if source == nil {
		return false
	}
	return g.TurnTally.ReplacementsApplied[ObjectTallyKey(source.InstanceID, source.ObjectEpoch, label)] > 0
}

// MarkReplacementAppliedThisTurnForEffect records that the replacement
// `label` on the object `source` names has been applied this turn. Call it
// from the replacement's Replace, which runs once for the event it
// replaces. Caller must hold g.mu.
func (g *Game) MarkReplacementAppliedThisTurnForEffect(source *Card, label string) {
	if source == nil {
		return
	}
	if g.TurnTally.ReplacementsApplied == nil {
		g.TurnTally.ReplacementsApplied = map[string]int{}
	}
	g.TurnTally.ReplacementsApplied[ObjectTallyKey(source.InstanceID, source.ObjectEpoch, label)]++
}
