package game

import "github.com/google/uuid"

// untap_replacements.go — "If a <permanent> would untap during your untap
// step, <do this> instead" (CR 502.3, CR 614.1a; Edge of Malacol, ADR 0216
// slice 4, #4870). The replacement is a fact about the untap step's one
// untap loop, not a general replacement event: it is read where the step
// untaps each permanent of its set (untapStepPermanentLocked), so a
// permanent the step does not untap (a restriction, an opt-out) is never
// replaced, and an untap by an effect or the sandbox is not "during the
// untap step" and is left alone.

// UntapReplacement is one printed untap-step replacement. A source - a
// permanent or a face-up plane - declares it on effects.Spec.UntapReplacements.
type UntapReplacement struct {
	// Label is the clause as printed.
	Label string
	// AppliesTo decides whether the untap of c during activePlayer's untap
	// step is replaced. Read-only; it runs with g.mu held.
	AppliesTo func(g *Game, source *Card, c Card, activePlayer uuid.UUID) bool
	// Instead is what happens in place of the untap. It runs with g.mu
	// held and the permanent still tapped.
	Instead func(g *Game, source *Card, c Card)
}

// CatalogUntapReplacements returns a catalog entry's untap replacements,
// keyed by its ability key; wired by the effects package.
var CatalogUntapReplacements func(oracleID string) []UntapReplacement

// untapStepPermanentLocked untaps one permanent of the untap step's set, or
// does the replacing "instead" of a source that applies to it. The first
// matching replacement applies (the permanent's controller would order
// several, which no card makes possible yet).
//
// Caller must hold g.mu in write mode.
func (g *Game) untapStepPermanentLocked(id uuid.UUID) {
	if CatalogUntapReplacements == nil || g.Battlefield == nil || g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) {
		g.untapPermanentByIDLocked(id)
		return
	}
	active := g.Seats[g.Turn.ActiveSeat].ID
	var c *Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			c = &g.Battlefield.Cards[i]
			break
		}
	}
	if c == nil {
		return
	}
	var instead func()
	try := func(src *Card) {
		if instead != nil {
			return
		}
		key := catalogAbilityKeyOf(src)
		if key == "" {
			return
		}
		for _, r := range CatalogUntapReplacements(key) {
			if r.AppliesTo != nil && r.Instead != nil && r.AppliesTo(g, src, *c, active) {
				card := *c
				s := src
				instead = func() { r.Instead(g, s, card) }
				return
			}
		}
	}
	for i := range g.Battlefield.Cards {
		try(&g.Battlefield.Cards[i])
	}
	g.forEachPlanarSourceLocked(func(src *Card) bool {
		try(src)
		return true
	})
	if instead != nil {
		instead()
		return
	}
	g.untapPermanentByIDLocked(id)
}
