package game

import (
	"fmt"

	"github.com/google/uuid"
)

// sacrifice_all_cost.go — #2097: "As an additional cost to cast this
// spell, sacrifice all creatures you control" (Soulblast). See
// AdditionalCost.SacrificeAll for the shape; this file is the one place
// that decides WHICH permanents the clause takes, so the cast path, the
// protocol view and the legal enumerator cannot disagree about it.

// ErrSacrificeAllMismatch is returned when a cast names sacrifice_ids
// for a "sacrifice all" clause that are not exactly the permanents the
// clause takes. The caster does not choose them (CR 601.2b announces no
// choice here), so a list that differs is a client bug, never a payment.
var ErrSacrificeAllMismatch = fmt.Errorf("%w: a \"sacrifice all\" cost takes every matching permanent you control", ErrInvalidParam)

// sacrificeAllCandidatesLocked is every permanent `playerID` controls
// that the clause matches, in SacrificePaymentOrderForEffect's order.
//
// The walk is SpecCandidatesForEffect's, the one every sacrifice cost
// uses, narrowed to the caster's own permanents (CR 701.21a: you can't
// sacrifice what you don't control). It reads g.Battlefield only, so a
// phased-out permanent, held in g.PhasedOut and treated by CR 702.26b
// as though it does not exist, is never among them. Nothing filters
// indestructible or "can't be destroyed": sacrifice is not destruction.
//
// Nil when the caster controls none — which is a payment, not a failure.
// Caller must hold g.mu.
func (g *Game) sacrificeAllCandidatesLocked(playerID uuid.UUID, spec *TargetSpec) []uuid.UUID {
	if spec == nil {
		return nil
	}
	var ids []uuid.UUID
	for _, id := range g.specCandidatesLocked(playerID, spec).Cards {
		if c := findBattlefieldCard(g, id); c != nil && c.Controller == playerID {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.SacrificePaymentOrderForEffect(ids, uuid.Nil)
}

// SacrificeAllCandidatesForEffect is sacrificeAllCandidatesLocked for a
// reader outside the package: the protocol view, which lists what the
// cost will take, and the legal enumerator, which names it on the move.
// Caller must hold g.mu (read or write).
func (g *Game) SacrificeAllCandidatesForEffect(playerID uuid.UUID, spec *TargetSpec) []uuid.UUID {
	return g.sacrificeAllCandidatesLocked(playerID, spec)
}

// sacrificeAllPaymentLocked settles the sacrifice_ids of a cast whose
// mandatory additional cost sacrifices all (CR 601.2b, 601.2h). The set
// is read once, here, before anything is paid, and from then on it is
// the payment: what the validator checks, what the auto-tapper leaves
// alone, what is sacrificed as one exit and what PaidCost records.
//
// `named` is what the announcement sent. Empty is "the engine fills it",
// which is how the client and the MCP seat cast. A non-empty list must
// be exactly the set, in any order — the legal enumerator names it so a
// policy can price what the cast gives up, and a parked cast resumed
// from a CR 903.9 answer arrives with the list it was filled with. The
// result is always in the engine's own order, so two announcements of
// the same cast pay identically.
//
// Caller must hold g.mu.
func (g *Game) sacrificeAllPaymentLocked(playerID uuid.UUID, cost *AdditionalCost, named []uuid.UUID) ([]uuid.UUID, error) {
	all := g.sacrificeAllCandidatesLocked(playerID, cost.Sacrifice)
	if len(named) == 0 {
		return all, nil
	}
	if len(named) != len(all) {
		return nil, ErrSacrificeAllMismatch
	}
	want := make(map[uuid.UUID]bool, len(all))
	for _, id := range all {
		want[id] = true
	}
	for _, id := range named {
		if !want[id] {
			return nil, ErrSacrificeAllMismatch
		}
		delete(want, id)
	}
	return all, nil
}
