package game

import "github.com/google/uuid"

// awaken.go — what a paid awaken cost does to its land (CR 702.113,
// ADR 0135 §3, #2411).
//
//	Awaken N—[cost] (If you cast this spell for [cost], also put N
//	+1/+1 counters on target land you control and it becomes a 0/0
//	Elemental creature with haste. It's still a land.)
//
// CR 702.113a: "If this spell's awaken cost was paid, put N +1/+1
// counters on target land you control. That land becomes a 0/0
// Elemental creature with haste. It's still a land." The cost and the
// extra target clause are catalog data (effects.Awaken: an alternative
// cost whose Targets add "target land you control" after the spell's
// own, so CR 702.113b's "only if that player chose to pay the spell's
// awaken cost" is the existing rewrite). This file is the verb.
//
// It is earthbend's animation (earthbend.go) with three differences,
// each read off the printed text (owner decision 4):
//
//  1. COUNTERS FIRST (CR 608.2c: instructions in the order written).
//     The counters go on while the land is not yet a creature, through
//     the CR 614 placement window, so Doubling Season ("a permanent you
//     control") doubles them and Hardened Scales ("a creature you
//     control") does not. Earthbend animates first because its text
//     does.
//  2. ELEMENTAL. The record adds the Elemental subtype beside the
//     Creature type, haste and base 0/0: one scoped effect, one
//     timestamp (CR 613.7), so the subtype ends with the animation
//     (layers 4, 6 and 7b; the counters are 7c, CR 613.4c).
//  3. NO RETURN TRIGGER.
//
// No stated duration, so it lasts until the game ends (CR 611.2a); the
// record is pinned to the object, so it stops applying when the land
// leaves (CR 400.7) and is plain data in a restore point. It uses only
// existing mod kinds, so a restore point needs no schema bump.

// awakenLabel is the animation record's attribution.
const awakenLabel = "awaken — becomes a 0/0 Elemental creature with haste that's still a land"

// AwakenForEffect puts n +1/+1 counters on land, then makes it a 0/0
// Elemental creature with haste that is still a land (CR 702.113a),
// with no duration (CR 611.2a), pinned to the object (CR 400.7).
//
// `actor` is the spell's controller, who places the counters; `source`
// is the resolving spell, the attribution on the continuous effect.
// `land` is the awaken TARGET, whose legality (a land you control) the
// card's clause and the CR 608.2b re-check have already judged; a land
// that is no longer on the battlefield is a silent no-op.
//
// If the counters' window pauses on a CR 616 ordering prompt (Doubling
// Season beside another counter replacement), the animation waits for
// the resume: it rides the placement's continuation, so it happens
// exactly once, after the counters, on every path. If the land left the
// battlefield (or left and came back, a new object) by then, the
// animation is skipped.
//
// Caller must hold g.mu (it is an effect-time helper).
func (g *Game) AwakenForEffect(actor, source, land uuid.UUID, n int) error {
	if land == uuid.Nil {
		return nil
	}
	c, ok := g.battlefieldCardLocked(land)
	if !ok {
		return nil
	}
	stamp := c.EnteredBattlefieldAt
	animate := func(g *Game, _ int) error {
		now, ok := g.battlefieldCardLocked(land)
		if !ok || now.EnteredBattlefieldAt != stamp {
			return nil
		}
		g.animateLandLocked(source, land, stamp, awakenLabel, "Elemental")
		return nil
	}
	if n <= 0 {
		return animate(g, 0)
	}
	return g.AddCounterByThenForEffect(actor, land, CounterPlusOne, n, animate)
}
