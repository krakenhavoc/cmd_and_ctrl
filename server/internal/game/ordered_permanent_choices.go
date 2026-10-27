package game

import "github.com/google/uuid"

// CanEffectCauseSacrificeForEffect applies the effect's protection gate.
// Costs do not use this query. Caller must hold g.mu.
func (g *Game) CanEffectCauseSacrificeForEffect(playerID uuid.UUID) bool {
	p := g.playerByIDLocked(playerID)
	return p != nil && !p.Eliminated && !g.effectSacrificeBlockedLocked(playerID)
}

// EffectSacrificeCandidatesForEffect returns controlled permanents eligible
// for this effect's instruction. It does not target. Caller must hold g.mu.
func (g *Game) EffectSacrificeCandidatesForEffect(playerID uuid.UUID, spec *TargetSpec) []uuid.UUID {
	if !g.CanEffectCauseSacrificeForEffect(playerID) {
		return nil
	}
	return g.sacrificeCandidatesLocked(playerID, spec)
}

// EarlierPermanentChoicesForEffect returns earlier public choices, in the
// supplied seat order. Copies prevent a consumer from mutating the prompt.
//
// Both collectors carry them as data: a keyed ordered collection (#4290)
// on its state, and a keyed per-object any-player leg (#4264) on its
// offer. (The closure collector, ChoosePermanentsInOrderThenForEffect, and
// the frame it kept them on are gone: #4361 S7b.)
func (c *PendingChoice) EarlierPermanentChoicesForEffect() []SeatCards {
	if c == nil {
		return nil
	}
	if c.ordered != nil {
		out := make([]SeatCards, len(c.ordered.Earlier))
		for i, p := range c.ordered.Earlier {
			out[i] = SeatCards{Seat: p.Seat, Cards: copyUUIDs(p.Cards)}
		}
		return out
	}
	if c.AnyPlayerOffer == nil || len(c.AnyPlayerOffer.Picks) == 0 {
		return nil
	}
	out := make([]SeatCards, len(c.AnyPlayerOffer.Picks))
	for i, p := range c.AnyPlayerOffer.Picks {
		out[i] = SeatCards{Seat: p.Seat, Cards: copyUUIDs(p.Cards)}
	}
	return out
}

// suggestUnchosenOwnLocked is the chooser's own candidates that no
// earlier seat chose, in candidate order, cut to `limit` (ADR 0187 §2,
// owner decision 6). "Own" is read off the permanent's controller now.
// Caller must hold g.mu.
func (g *Game) suggestUnchosenOwnLocked(chooser uuid.UUID, cards []uuid.UUID, limit int, earlier PromptedPicks) []uuid.UUID {
	taken := make(map[uuid.UUID]bool)
	for _, e := range earlier {
		for _, id := range e.Cards {
			taken[id] = true
		}
	}
	var out []uuid.UUID
	for _, id := range cards {
		if len(out) >= limit {
			break
		}
		if c := findBattlefieldCard(g, id); c != nil && c.Controller == chooser && !taken[id] {
			out = append(out, id)
		}
	}
	return out
}
