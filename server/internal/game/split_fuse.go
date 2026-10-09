package game

// split_fuse.go — a fused split spell's catalog definition (CR 702.102,
// ADR 0103 Decision 3).
//
// The catalog registers a split card's halves as two entries, the left
// under the bare oracle ID and the right under "<oracle>#1" (ADR 0034
// §5). A fused spell is both halves at once, so it has a key of its own
// (FusedCatalogKey) and this file builds its definition from the two
// halves, where catalogDef answers for that key:
//
//   - its target clauses are the left half's followed by the right
//     half's, one announcement in ADR 0065's clause order;
//   - it resolves the left half's instructions, then the right half's
//     (CR 702.102d), each half reading only its own targets, numbered
//     as its own clause list numbers them.
//
// A half the catalog does not know makes the whole fused spell unknown
// (nil), so it resolves by hand like any uncatalogued spell rather than
// doing half of what it says.

// fusedCatalogDef builds the definition of a fused cast of the card
// whose left half is keyed `base`.
func fusedCatalogDef(base string) *CardDef {
	if CatalogLookup == nil || base == "" {
		return nil
	}
	left := CatalogLookup(base)
	right := CatalogLookup(base + "#1")
	if left == nil || right == nil {
		return nil
	}
	nLeft := left.Targets.ClauseCount()
	def := &CardDef{
		Targets: chainTargetSpecs(left.Targets, right.Targets),
		// ADR 0126 §6: a fused spell does what both halves do.
		Purpose: left.Purpose.plus(right.Purpose, nLeft),
	}
	if def.Targets != nil {
		def.TargetMode = def.Targets.Mode
	}
	leftResolve, rightResolve := left.Resolve, right.Resolve
	leftSpec, rightSpec := left.Targets, right.Targets
	def.Resolve = func(g *Game, item *StackItem) error {
		if leftResolve != nil {
			if err := leftResolve(g, fusedHalfItem(item, leftSpec, 0, nLeft)); err != nil {
				return err
			}
		}
		if rightResolve != nil {
			return rightResolve(g, fusedHalfItem(item, rightSpec, nLeft, -1))
		}
		return nil
	}
	return def
}

// fusedHalfItem is the stack item one half of a fused spell resolves
// as: the same item, carrying only the targets announced under that
// half's clauses, renumbered to the half's own clause list, and that
// half's own clause list as its announced spec — so ClauseTarget(0)
// in the right half's OnResolve is the right half's first clause, and
// the CR 608.2b re-check judges each target against the clause it was
// chosen for. `hi` < 0 is "to the end".
func fusedHalfItem(item *StackItem, spec *TargetSpec, lo, hi int) *StackItem {
	sub := *item
	sub.Targets = nil
	for _, t := range item.Targets {
		if t.Kind != TargetPlayer && t.Kind != TargetCard {
			continue
		}
		if t.Slot < lo || (hi >= 0 && t.Slot >= hi) {
			continue
		}
		t.Slot -= lo
		sub.Targets = append(sub.Targets, t)
	}
	// The half's own clause list. A half with no clause carries no
	// target either, so nothing ever asks it for one.
	sub.targetSpec = spec
	return &sub
}

// chainTargetSpecs is `a` followed by `b` as one flat clause list (ADR
// 0065 §1). Either may be nil. The catalog's own specs are never
// mutated: the head is copied and the tail is a new slice.
func chainTargetSpecs(a, b *TargetSpec) *TargetSpec {
	switch {
	case a == nil && b == nil:
		return nil
	case a == nil:
		out := *b
		out.Rest = append([]TargetClause(nil), b.Rest...)
		return &out
	}
	out := *a
	out.Rest = append([]TargetClause(nil), a.Rest...)
	if b != nil {
		out.Then(b)
	}
	return &out
}

// FusedHalvesDeclareExtras reports whether either half of the split
// card keyed `base` declares something a fused announcement cannot
// carry: modes, an alternative cost, a mandatory or optional
// additional cost, or a tap cost. None of the 22 fuse cards prints
// one; the fused cast refuses such a card rather than price it wrong.
func FusedHalvesDeclareExtras(base string) bool {
	if CatalogLookup == nil {
		return false
	}
	for _, key := range []string{base, base + "#1"} {
		d := CatalogLookup(key)
		if d == nil {
			continue
		}
		if d.Modes != nil || len(d.AlternativeCosts) > 0 || d.AdditionalCost != nil ||
			len(d.OptionalCosts) > 0 || d.TapCost != nil {
			return true
		}
	}
	return false
}

// fusedCastAllowed is CR 702.102a's gate on a fused announcement: a
// split card with fuse, cast from its owner's hand, claiming no
// alternative cost (a fused spell pays both halves' mana costs,
// CR 702.102c), and whose halves declare nothing a fused announcement
// cannot carry.
func fusedCastAllowed(c Card, zone ZoneKind, params CastSpellParams, grant *CastPermission) error {
	if !HasFuse(c) || zone != ZoneHand {
		return ErrInvalidFace
	}
	if params.Face != 0 || params.AlternativeCost != "" || grant != nil {
		return ErrInvalidParam
	}
	if FusedHalvesDeclareExtras(c.OracleID) {
		return ErrInvalidParam
	}
	return nil
}

// FusedSpell is `c` as a fused split spell (CR 702.102b): a copy with
// both halves' combined characteristics and the fused catalog key. For
// the view and the bot, which price and target a fused cast without
// making one.
func FusedSpell(c Card) Card {
	c.materialiseFused()
	return c
}
