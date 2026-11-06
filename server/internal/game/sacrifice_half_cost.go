package game

import (
	"fmt"

	"github.com/google/uuid"
)

// sacrifice_half_cost.go — "As an additional cost to cast this spell,
// sacrifice half the lands you control, rounded up" (Tectonic Split,
// #5075). See AdditionalCost.SacrificeHalfRoundedUp. The caster chooses
// the permanents; this file is the one place that says how many, so the
// cast path, the protocol view and the legal enumerator cannot disagree.

// ErrSacrificeHalfCount is returned when a cast names a number of
// sacrifice_ids other than half the matching permanents its caster
// controls, rounded up.
var ErrSacrificeHalfCount = fmt.Errorf("%w: a \"sacrifice half, rounded up\" cost takes exactly that many", ErrInvalidParam)

// sacrificeHalfCountLocked is how many permanents a "sacrifice half …,
// rounded up" clause takes from `playerID`: half of the permanents they
// control that the clause matches, rounded up. Every one they control
// counts, one that can't be sacrificed included ("half the lands you
// control" counts lands, not sacrificeable lands), so such a permanent
// can make the cost unpayable rather than cheaper. A phased-out
// permanent is not on the battlefield list and does not count (CR
// 702.26b). Caller must hold g.mu.
func (g *Game) sacrificeHalfCountLocked(playerID uuid.UUID, spec *TargetSpec) int {
	if spec == nil {
		return 0
	}
	n := 0
	for _, id := range g.specCandidatesLocked(playerID, spec).Cards {
		if c := findBattlefieldCard(g, id); c != nil && c.Controller == playerID {
			n++
		}
	}
	return (n + 1) / 2
}

// SacrificeHalfCountForEffect is sacrificeHalfCountLocked for a reader
// outside the package: the protocol view, which tells the picker how
// many to take, and the legal enumerator, which names that many on the
// move. Caller must hold g.mu (read or write).
func (g *Game) SacrificeHalfCountForEffect(playerID uuid.UUID, spec *TargetSpec) int {
	return g.sacrificeHalfCountLocked(playerID, spec)
}
