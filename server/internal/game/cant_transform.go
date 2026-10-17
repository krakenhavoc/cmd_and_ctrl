package game

import "github.com/google/uuid"

// cant_transform.go: "<these permanents> can't transform" (CR 101.2),
// Immerwolf's "Non-Human Werewolves you control can't transform". ADR
// 0132 amendment 2026-10-16, #4071.
//
// The restriction is a static ability of the RESTRICTING permanent, so
// it is a layer static from that permanent that sets
// Characteristic.CantTransform on each permanent it covers, the
// CantBeSacrificed shape. Its filter is read through the layers, so
// "non-Human Werewolf" is the permanent's face as it is now: a front face
// that is a Human Werewolf can transform, and the back face it turns into
// cannot turn back.
//
// "Can't" beats every instruction that would turn the permanent over in
// place (CR 101.2), and there are two routes:
//
//   - an effect's instruction to transform or convert it
//     (turnOverForEffectLocked: Moonmist, a werewolf's own upkeep
//     trigger, Vildin-Pack Alpha's "you may transform it"; CR 701.28f
//     carries a "can't transform" to a convert);
//   - the daybound / nightbound turn-over (settleDayNightLocked: a
//     Tovolar, the Midnight Scourge you control stays on its back face by
//     day).
//
// What it does not stop is a permanent ENTERING transformed (CR
// 702.145b's first ability, CR 712.14a): that permanent never transformed
// (CR 712.18), so nothing was refused.

// cantTransformLocked reports whether a "can't transform" effect covers
// the permanent `c` right now. The layer cache is refreshed first, since
// the restricting permanent may have arrived since the last pass. False
// for nil.
//
// Caller must hold g.mu in write mode (the refresh writes the cache).
func (g *Game) cantTransformLocked(c *Card) bool {
	if c == nil {
		return false
	}
	g.RecomputeLayersIfStaleLocked()
	return c.Effective().CantTransform
}

// CanTransformForEffect reports whether an instruction to transform the
// battlefield permanent cardID would turn it over now: it is a
// double-faced permanent with a permanent face to turn to (CR 712.9,
// 701.27c-d), it has neither daybound nor nightbound (CR 702.145b,
// 702.145e), and no "can't transform" effect covers it (CR 101.2). A
// card asks this before offering "you may transform it", so a player is
// never asked a question whose "yes" does nothing.
//
// Caller must hold g.mu in write mode.
func (g *Game) CanTransformForEffect(cardID uuid.UUID) bool {
	c := findBattlefieldCard(g, cardID)
	return c != nil && CanTransform(*c) && !hasDayNightKeyword(c) && !g.cantTransformLocked(c)
}
