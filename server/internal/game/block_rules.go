package game

// block_rules.go holds the CR 509.1b block restrictions that carry a
// PARAMETER — "can't be blocked except by Walls", "can't be blocked by
// creatures with power 2 or less", "can't be blocked by more than one
// creature" (#750).
//
// Why they are not Restriction bits. restrictions.go's vocabulary is a
// set of bits with nothing attached (ADR 0045 §2). A condition on the
// OTHER creature — its power, its subtypes, its colour, a keyword it
// has — cannot be a bit, and neither can a bound on how many creatures
// are allowed to block. ADR 0045's addendum (Decision 11) puts them in
// a list of RULES instead, derived from their sources when a block is
// checked, the way ADR 0048 §1 derives cost modifiers. The bits stay
// as they are: they are the cheap, common case, and the client reads
// them (ADR 0045 §6).
//
// Everything here is read LIVE, at the moment the block is checked.
// A predicate reads the other creature's effective characteristics
// after layer 7, and a threshold reads its own source's. Restrictions
// are checked only at declaration (ADR 0045 §5, CR 509.1b), so a pump
// after the block is declared changes nothing.
//
// Two halves, two consumers:
//
//   - Pair rules are slot 4 of BlockPairRefusalLocked, so they are
//     refused by DeclareBlocker, withheld by the legal-move
//     enumerator and honoured by the #328 auto-pass signal without
//     any of the three learning a new rule (ADR 0045 §3).
//   - Count rules are bounds on a whole declaration, not pair checks,
//     so they are judged where the declaration is complete — the
//     set-based DeclareBlockers in block_declaration.go, which
//     REFUSES an illegal count instead of storing it (Decisions 12
//     and 13).
//
// Rules come from two registries, both walked on every check: the
// battlefield, through CatalogBlockRules, and Game.TurnScopedBlockRules
// for the until-end-of-turn ones (Gingerbrute's shape), which the
// turn's cleanup sweep empties.
//
// Not here, and deliberately: BlockRule.Limit (Silent Arbiter's
// "no more than one creature can block each combat"), which ADR 0045's
// addendum Decision 12 leaves unbuilt until its first card.
// See docs/decisions/0045-combat-restrictions.md, addendum
// Decisions 11 and 12.

// BlockRule is one printed block restriction with a parameter, as the
// engine reads it. Exactly one of Pair and Count is set; a rule with
// neither is inert.
//
// The `source` argument is the permanent the rule was read from, so a
// threshold that compares against its own source (Champion of
// Lambholt's power) reads it live rather than closing over a value
// captured at catalog-build time.
type BlockRule struct {
	// Pair reports whether this rule REFUSES the pair — true means
	// "these two may not be paired". Reads attacker and blocker after
	// layer 7. Never mutates anything: it is called from inside
	// BlockPairRefusalLocked, whose read-only contract the enumerator
	// and the view both depend on.
	Pair func(g *Game, attacker, blocker, source *Card) bool

	// Count bounds how many creatures may block this attacker. Zero
	// means "no bound" on either end, so a rule that only sets a
	// maximum returns (0, n). The effective minimum over all rules is
	// the largest minimum and the effective maximum is the smallest
	// maximum (Decision 12).
	Count func(g *Game, attacker, source *Card) (min, max int)

	// Reason is the BlockReason a refused pair reports. Empty defaults
	// to BlockReasonCantBeBlockedBy, which is the commonest shape.
	// Ignored for a Count rule: a bound is judged on the declaration,
	// not on a pair, and nothing player-facing is sent for it yet.
	Reason BlockReason

	// Label is the rule's parameter as the card prints it — "Walls",
	// "creatures with power 2 or less" — and it is what the player
	// reads in the illegal_block sentence. Empty falls back to a
	// generic phrase rather than to a rule the client would have to
	// re-derive (ADR 0045 §6).
	Label string
}

// CatalogBlockRules returns the block rules a battlefield permanent
// with the given catalog key imposes, or nil. A nil hook is "no card
// declares one", which is where the catalog stands today.
//
// The card side is one card PR away and deliberately not here: it is
// the `effects.Spec.BlockRules` field, its line in `effects.buildDef`,
// the `CardDef.BlockRules` slot carddef.go projects, and the first
// cards (Prowler's Helm, Hungering Hydra) — ADR 0045's addendum,
// PR 4's card half. The hook exists ahead of them so that writing
// those cards is a Spec field and not an engine change, which is the
// same shape, and the same reason, as CatalogAdditionalLandPlays in
// land_drops.go ("No catalog card declares AdditionalLandPlays yet").
//
// Tests stub this directly, as keywords_test.go stubs
// CatalogPrintedKeywords.
//
// Keyed by CatalogAbilityKey, not CatalogKey: a block rule is a static
// ability of the permanent that prints it, so a Legolas Greenleaf that
// has lost all its abilities imposes nothing (CR 613.1f), while an
// Equipment's rule is unaffected by the equipped creature losing ITS
// abilities — the rule belongs to the Equipment.
var CatalogBlockRules func(key string) []BlockRule

// forEachBlockRuleLocked walks every block rule on the battlefield in
// battlefield order, calling fn until it returns false. Allocation-free
// beyond what the hook itself returns, because it sits on the block
// check that runs once per (attacker, blocker) pair the enumerator
// considers.
//
// Caller must hold g.mu (read or write) with fresh layers. Reads only.
func (g *Game) forEachBlockRuleLocked(fn func(rule BlockRule, source *Card) bool) {
	if CatalogBlockRules == nil || g.Battlefield == nil {
		return
	}
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		// The empty KEY is the skip, not an empty oracle ID: a token
		// has a catalog key of its own since #521, and Avatar Kuruk's
		// Spirit token ("can't block or be blocked by non-Spirit
		// creatures") is a token that prints a pair rule. ADR 0083
		// decision 3.
		key := CatalogAbilityKey(*src)
		if key == "" {
			continue
		}
		for _, r := range CatalogBlockRules(key) {
			if !fn(r, src) {
				return
			}
		}
	}
}

// forEachBlockRuleLocked's turn-scoped half. Walked after the
// battlefield so a permanent's printed rule is reported first and an
// until-end-of-turn effect cannot mask it; nothing depends on the
// order beyond which refusal a player reads.
//
// Caller must hold g.mu (read or write) with fresh layers. Reads only.
func (g *Game) forEachTurnScopedBlockRuleLocked(fn func(rule BlockRule, source *Card) bool) {
	for i := range g.TurnScopedBlockRules {
		// A turn-scoped rule's source is the effect that created it,
		// which may have left the battlefield since (CR 611.2b: the
		// effect does not end with its source). Nil is the honest
		// answer, and every predicate takes it.
		if !fn(g.TurnScopedBlockRules[i], nil) {
			return
		}
	}
}

// RegisterTurnScopedBlockRuleLocked adds an until-end-of-turn block
// rule (ADR 0045 addendum, Decision 11). It lives until the turn's
// cleanup sweep empties the registry, whoever created it and whatever
// happens to that source in the meantime.
//
// Caller must hold g.mu in write mode.
func (g *Game) RegisterTurnScopedBlockRuleLocked(rule BlockRule) {
	g.TurnScopedBlockRules = append(g.TurnScopedBlockRules, rule)
}

// ClearTurnScopedBlockRulesLocked drops every until-end-of-turn block
// rule. A fresh slice rather than a truncation, exactly as
// ClearTurnScopedReplacementsLocked does, so a clone taken before the
// sweep keeps its own backing array.
//
// Caller must hold g.mu in write mode.
func (g *Game) ClearTurnScopedBlockRulesLocked() {
	if len(g.TurnScopedBlockRules) > 0 {
		g.TurnScopedBlockRules = nil
	}
}

// blockRuleRefusalLocked is slot 4 of BlockPairRefusalLocked: the
// first pair rule on the battlefield that refuses this pair, or
// BlockOK. The refusal names the permanent the rule was read from as
// its Source — the Equipment, the lord, the attacker itself — because
// that is the card the player has to answer, and it is the one thing
// the engine knows about a rule that the client cannot derive.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) blockRuleRefusalLocked(attacker, blocker *Card) BlockRefusal {
	out := BlockOK
	refuse := func(r BlockRule, source *Card) bool {
		if r.Pair == nil || !r.Pair(g, attacker, blocker, source) {
			return true
		}
		reason := r.Reason
		if reason == "" {
			reason = BlockReasonCantBeBlockedBy
		}
		out = BlockRefusal{Reason: reason, Label: r.Label}
		if source != nil {
			out.Source = source.InstanceID
		}
		return false
	}
	g.forEachBlockRuleLocked(refuse)
	if out.Legal() {
		g.forEachTurnScopedBlockRuleLocked(refuse)
	}
	return out
}

// blockerBoundsLocked reports how many creatures may block `attacker`:
// `min` is the fewest a legal declaration may use and `max` the most,
// with 0 meaning "no bound" on either end. `src` is the permanent
// imposing the tighter of the two — the minimum's source when there is
// a minimum, the maximum's otherwise, and uuid.Nil when neither binds.
//
// Menace (CR 702.111b) is built in as a minimum of 2 rather than
// written as a catalog rule, because it is a keyword every card gets
// through Characteristic.Abilities and there is no source permanent to
// read it from but the attacker itself. Rules add their own bounds on
// top: Pathrazer of Ulamog's minimum of 3, Hungering Hydra's maximum
// of 1. This replaced BlockerCountValid's menace-only
// answer, which is gone (ADR 0045 addendum, Decision 12).
//
// A minimum greater than the maximum makes the attacker unblockable —
// a menace creature wearing Vorrac Battlehorns. That is not an error:
// blockerCountValidLocked refuses every non-empty set for it, and
// "no blockers" stays legal, as it always is.
//
// nil attacker has no bounds. Caller must hold g.mu with fresh
// layers — HasKeyword reads the effective characteristic.
func (g *Game) blockerBoundsLocked(attacker *Card) (min, max int) {
	if attacker == nil {
		return 0, 0
	}
	if HasKeyword(attacker, "menace") {
		min = 2
	}
	bound := func(r BlockRule, source *Card) bool {
		if r.Count == nil {
			return true
		}
		lo, hi := r.Count(g, attacker, source)
		if lo > min {
			min = lo
		}
		if hi > 0 && (max == 0 || hi < max) {
			max = hi
		}
		return true
	}
	g.forEachBlockRuleLocked(bound)
	g.forEachTurnScopedBlockRuleLocked(bound)
	return min, max
}

// blockerCountValidLocked reports whether `n` creatures blocking
// `attacker` is a legal COUNT (CR 509.1b). It is the whole-declaration
// half of block legality; the per-pair half is
// BlockPairRefusalLocked.
//
// An empty block is always legal: "no blockers" is a valid outcome for
// any attacker, and a minimum only ever says how many it takes to
// block at all.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) blockerCountValidLocked(attacker *Card, n int) bool {
	if attacker == nil || n <= 0 {
		return true
	}
	min, max := g.blockerBoundsLocked(attacker)
	if min > 0 && n < min {
		return false
	}
	if max > 0 && n > max {
		return false
	}
	return true
}
