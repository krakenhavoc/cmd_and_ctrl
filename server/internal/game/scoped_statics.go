package game

// scoped_statics.go is the SWEEP for continuous effects whose lifetime
// is a DURATION rather than a permanent's presence on the battlefield
// (CR 611.2). It began in S32 as the "until end of turn" registry and
// grew the other three CR 611.2 durations in S38 (ADR 0063, #755).
//
// The registry it swept — `ScopedStatic`, a raw `StaticAbility` with a
// duration attached — is gone (ADR 0041 phase 3, tier 3a, #1497). Its
// two closures kept every table holding a Giant Growth or a crewed
// Vehicle off the restore path until cleanup. Every continuous effect a
// resolving spell or ability creates is now a `ScopedEffect`
// (scoped_effects.go): an affected set, operations from a closed
// vocabulary and a `Duration`, all data. The sweep functions keep their
// names because every turn-boundary site calls them.
//
// What is NOT here:
//   - Layer 1 (copy) with a duration (Mirage Mirror, Cytoshape).
//     The bucket exists (ADR 0043 §5) and nothing registers into it.
//   - A separate registry for the REPLACEMENT or BLOCK-RULE twins.
//     Since ADR 0041 tier 3b a replacement a spell creates (Fog, a
//     prevention shield, the Whip's redirect) and a block rule a spell
//     or ability creates (Gingerbrute, Mirri, Weatherlight Duelist)
//     are both ScopedEffect records too, with a non-layer mod the
//     layer pass skips and the replacement gather or the block-rule
//     walk reads (scoped_replacements.go, scoped_block_rules.go). Both
//     are swept here with everything else.

// ClearEndOfTurnScopedStaticsLocked is the CR 514.2 cleanup sweep:
// "until end of turn" effects end during the cleanup step. Called
// from `sweepTurnEndLocked` alongside the damage wipe and the
// impulse-exile sweep — the same "this turn is over" pass. It drops
// effects of every other duration whose time has also run out, because
// it is the same sweep.
//
// Caller must hold g.mu.
func (g *Game) ClearEndOfTurnScopedStaticsLocked() {
	g.sweepScopedStaticsLocked(true)
	// #1197: granted PLAYER abilities ride the same schedule. They
	// are not continuous effects over objects and so are not in the
	// scoped-effect registry, but they carry the same Duration and are ended
	// by the same durationExpiredLocked, and a second schedule would
	// be a second thing to keep in step. See player_statics.go.
	g.sweepPlayerStaticsLocked(true)
}

// ClearExpiredScopedStaticsLocked is the ordinary sweep: it drops
// every scoped static whose duration has run out, WITHOUT treating
// the current moment as the end of a turn. Called when a turn begins
// (the "until your next turn" boundary) and at the top of every
// layer recompute (where a "for as long as" condition can have gone
// false since the last pass).
//
// Caller must hold g.mu.
func (g *Game) ClearExpiredScopedStaticsLocked() {
	g.sweepScopedStaticsLocked(false)
	// #1197, as above: the "until your next turn" boundary is exactly
	// where Teferi's Protection and The One Ring's shield end.
	g.sweepPlayerStaticsLocked(false)
}

// sweepScopedStaticsLocked drops expired entries and bumps the layer
// version when it drops any, so the next recompute rebuilds without
// them. `endOfTurn` is handed straight to `durationExpiredLocked`,
// which is the only code that decides what a duration means.
//
// sweepScopedEffectsLocked allocates a fresh slice rather than
// filtering in place with `s[:0]`: the backing array is shared with every undo snapshot
// `Clone` has taken, so an in-place compaction would rewrite
// history. That is the same trap `cloneCard` exists to avoid on the
// card side.
//
// The sweep is idempotent, which matters because CR 514.3a can give a
// turn a second cleanup step and the hook runs again for it (#661),
// because the cleanup hook runs again after a discard pause drains,
// and because the layer recompute runs it on every pass.
//
// Caller must hold g.mu.
func (g *Game) sweepScopedStaticsLocked(endOfTurn bool) {
	if g.sweepScopedEffectsLocked(endOfTurn) {
		g.layerVersion.Add(1)
	}
}

// scopedContinuousEffectsLocked adapts the scoped-effect registry
// into the `ContinuousEffect` values the layer engine sorts and
// applies (scopedEffectContinuousEffectsLocked, memoised). Order
// within a bucket is decided by timestamp (CR 613.7), not by position.
//
// Caller must hold g.mu in write mode (the recompute pass does).
func (g *Game) scopedContinuousEffectsLocked() []ContinuousEffect {
	return g.scopedEffectContinuousEffectsLocked()
}
