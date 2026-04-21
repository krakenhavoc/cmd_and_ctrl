package game

import "github.com/google/uuid"

// mana.go is the per-player mana economy: a `ManaPool` slice of
// individual `ManaToken` entries (one per produced unit), with the
// helpers the cost validator + auto-tapper consume. The pool is a
// SLICE, not a multiset, so:
//   - Insertion order is observable (S15+ pain-source preview, S17+
//     filter-land sub-payment routing).
//   - Each token can carry its own restrictions independently of
//     other same-color tokens (Cavern of Souls' tribe restriction
//     when it lands, snow-source tagging if S17 enforces it).
//
// The pool is per-player. Pool-wide invariants (CR 106.4 "empty at
// end of step") are enforced from the step-advance path, not on the
// pool struct itself.
//
// Added in S15 sub-PR 2.

// ManaToken is one unit of mana sitting in a player's pool. Color
// is one of "W"/"U"/"B"/"R"/"G"/"C". Source is the permanent that
// produced it (uuid.Nil when produced by a non-card affordance like
// the admin "give mana" debug action that may land later).
// Restrictions is a list of opaque tags the cost validator can
// honour ("cast_creature_only", "spend_on_X_type"). Empty for
// vanilla basics + Sol Ring; populated by S22's Cavern-of-Souls-
// style cards.
type ManaToken struct {
	Color        string
	Source       uuid.UUID
	Restrictions []string
}

// ManaPool is a player's current mana pool. Order matters — see the
// file-level comment. nil and len-0 are equivalent; helpers preserve
// nil-when-empty so JSON encoding stays sparse.
type ManaPool []ManaToken

// AddMana appends one or more tokens to the pool. Empty tokens
// (zero Color) are dropped on the floor — defensive.
func (p *ManaPool) AddMana(tokens ...ManaToken) {
	for _, t := range tokens {
		if t.Color == "" {
			continue
		}
		*p = append(*p, t)
	}
}

// EmptyPool drops every token. Used by the step-advance hook
// (CR 106.4: pools empty at end of step / phase). Returns the count
// of tokens cleared so callers can emit one event per drop if they
// want (the engine emits a single EventManaPoolEmptied per affected
// player today).
func (p *ManaPool) EmptyPool() int {
	n := len(*p)
	*p = nil
	return n
}

// CanPay reports whether the pool currently holds enough mana to
// cover `cost`. Dry-run; does not mutate the pool.
//
// Algorithm: greedy, restriction-first. For each ColorRequirement
// in cost.Required, find the most-restricted matching token (i.e.
// the token whose set of "still-spendable-on" possibilities is
// smallest) — by S15 every token is unrestricted, so this collapses
// to "find any token of an allowed color." Then satisfy generic
// from whatever's left, taking colorless first to preserve colored
// for later costs.
//
// `xValue` is the announce-time X value the caller picked; the
// pool needs cost.Generic + cost.XSlots*xValue generic-tradeable
// mana. Pass 0 for spells without X.
func (p ManaPool) CanPay(cost ParsedCost, xValue int) bool {
	_, ok := p.attemptSpend(cost, xValue)
	return ok
}

// SpendMana attempts to deduct `cost` from the pool. On success,
// the pool is mutated and the method returns true. On failure, the
// pool is left unchanged. Caller usually calls CanPay first for a
// dry-run, then SpendMana to commit; this two-step shape keeps the
// "warn-and-proceed" sandbox path (S15 sub-PR 3) cheap.
func (p *ManaPool) SpendMana(cost ParsedCost, xValue int) bool {
	remaining, ok := p.attemptSpend(cost, xValue)
	if !ok {
		return false
	}
	*p = remaining
	return true
}

// attemptSpend is the shared core of CanPay + SpendMana. Returns
// (remaining-pool, true) on success, (nil, false) on failure.
// Operates on a copy so the caller's slice is never aliased.
func (p ManaPool) attemptSpend(cost ParsedCost, xValue int) (ManaPool, bool) {
	// Copy so we can mark spends without disturbing the caller.
	work := make(ManaPool, len(p))
	copy(work, p)
	used := make([]bool, len(work))

	// Step 1 — colored requirements first. Each requirement names
	// a set of allowed colors (monocolored = one entry, hybrid =
	// two). Pick the first available token whose color matches.
	// Phyrexian / numeric-alt requirements pay the color half in
	// S15 (life self-pay is S17).
	for _, req := range cost.Required {
		idx := -1
		for i, tok := range work {
			if used[i] {
				continue
			}
			if matchColor(tok.Color, req.Options) {
				idx = i
				break
			}
		}
		if idx < 0 {
			return nil, false
		}
		used[idx] = true
	}

	// Step 2 — generic requirement. Total = explicit Generic +
	// XSlots * xValue. Take colorless first (Sol Ring output, snow
	// generics) to preserve colored for the next cast; then drop
	// colored mana onto the rest. Both loops walk in pool insertion
	// order so the auto-tapper preview rendered alongside is stable.
	need := cost.Generic + cost.XSlots*xValue
	if need > 0 {
		for _, color := range []string{"C", "W", "U", "B", "R", "G"} {
			if need == 0 {
				break
			}
			for i, tok := range work {
				if need == 0 {
					break
				}
				if used[i] {
					continue
				}
				if tok.Color != color {
					continue
				}
				used[i] = true
				need--
			}
		}
		if need > 0 {
			return nil, false
		}
	}

	// Build remaining slice from un-used tokens, preserving order.
	out := make(ManaPool, 0, len(work))
	for i, tok := range work {
		if used[i] {
			continue
		}
		out = append(out, tok)
	}
	if len(out) == 0 {
		return nil, true
	}
	return out, true
}

// matchColor reports whether `color` is in `options`. Empty options
// means "any color" — never produced by ParseCost today, but cheap
// to handle defensively for future requirement shapes.
func matchColor(color string, options []string) bool {
	if len(options) == 0 {
		return true
	}
	for _, o := range options {
		if o == color {
			return true
		}
	}
	return false
}

// Missing returns the list of cost symbols `pool` cannot cover, in
// the order they appear in the cost. Returns nil when the cost is
// fully payable (CanPay would return true). Used by the strict-mode
// gate (S15 sub-PR 3) to populate the structured error frame the
// client renders into the override-toast affordance.
func (p ManaPool) Missing(cost ParsedCost, xValue int) []string {
	if p.CanPay(cost, xValue) {
		return nil
	}
	work := make(ManaPool, len(p))
	copy(work, p)
	used := make([]bool, len(work))
	var missing []string

	for _, req := range cost.Required {
		idx := -1
		for i, tok := range work {
			if used[i] {
				continue
			}
			if matchColor(tok.Color, req.Options) {
				idx = i
				break
			}
		}
		if idx < 0 {
			missing = append(missing, "{"+formatRequirement(req)+"}")
			continue
		}
		used[idx] = true
	}

	need := cost.Generic + cost.XSlots*xValue
	for _, color := range []string{"C", "W", "U", "B", "R", "G"} {
		if need == 0 {
			break
		}
		for i, tok := range work {
			if need == 0 {
				break
			}
			if used[i] || tok.Color != color {
				continue
			}
			used[i] = true
			need--
		}
	}
	for need > 0 {
		missing = append(missing, "{1}")
		need--
	}
	return missing
}

// emptyAllManaPoolsLocked clears every seated player's mana pool
// and emits one EventManaPoolEmptied per affected player. Called
// from runStepEntryHooksLocked at every step boundary (CR 106.4).
// No-op for players whose pool is already empty so step-cycle
// chatter stays quiet. Caller must hold g.mu.
func (g *Game) emptyAllManaPoolsLocked() {
	for _, p := range g.Seats {
		n := p.ManaPool.EmptyPool()
		if n == 0 {
			continue
		}
		g.EmitEvent(Event{
			Kind:   EventManaPoolEmptied,
			Actor:  p.ID,
			Amount: n,
		})
	}
}

func formatRequirement(req ColorRequirement) string {
	if len(req.Options) == 0 {
		return "?"
	}
	if len(req.Options) == 1 {
		return req.Options[0]
	}
	out := req.Options[0]
	for _, o := range req.Options[1:] {
		out += "/" + o
	}
	return out
}
