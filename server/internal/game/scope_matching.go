package game

import (
	"fmt"

	"github.com/google/uuid"
)

// scope_matching.go — #6237: a mass "can't block this turn" over a
// FILTERED set of creatures.
//
// Goblin Locksmith: "Whenever this creature attacks, creatures with
// defender can't block this turn."
// Rumbling Ruin: "… Creatures your opponents control with power less
// than or equal to that number can't block this turn."
//
// ADR 0045 named Goblin Locksmith as the card that needed one more
// scope, and the live-scope vocabulary had grown one constant per
// printed filter — creatures without flying, creatures that aren't
// lands, odd and even mana value, power 3 or less. A filter with a
// NUMBER in it (Rumbling Ruin's computed power bound) cannot be a
// constant, so these two scopes carry the filter as DATA instead: the
// record's mods hold PermanentQuery values, which are already the
// engine's plain-data permanent filter (permanent_query.go) and already
// ride a snapshot.
//
// WHY THE SET IS STILL LIVE. CR 611.2c locks the affected set only for
// an effect that changes characteristics or control; "can't block"
// changes neither, so the set is read as the rule is asked
// (foldRuleScopedRestrictionsLocked) and a creature that arrives later
// in the turn, or that becomes a creature with defender, is covered.
//
// WHAT IS LOCKED is the NUMBER, where the card says so: Rumbling Ruin's
// bound is computed as the ability resolves and baked into the stored
// query, so a creature whose power later falls under the bound is
// covered and the bound itself never moves.

// ScopeCreaturesMatching is "creatures matching the queries its mods
// name", every player's, read live (#6237): Goblin Locksmith's
// "creatures with defender". The queries are the union of the mods'
// Queries, any of which matches.
//
// An on-disk identity: never renamed, never reused.
const ScopeCreaturesMatching AffectedScope = "creaturesMatching"

// ScopeOpponentsCreaturesMatching is the same filter narrowed to the
// record's Controller's OPPONENTS: Rumbling Ruin's "creatures your
// opponents control with power less than or equal to that number".
//
// An on-disk identity: never renamed, never reused.
const ScopeOpponentsCreaturesMatching AffectedScope = "opponentsCreaturesMatching"

// scopeMatchesQueries reports whether a scope reads its mods' queries.
func scopeMatchesQueries(scope AffectedScope) bool {
	return scope == ScopeCreaturesMatching || scope == ScopeOpponentsCreaturesMatching
}

// ScopeFiltersByQueries is scopeMatchesQueries for card code, which
// pairs a query-filtered scope with its filter and must not ship one
// without the other (effects.RestrictUntilEOT).
func ScopeFiltersByQueries(scope AffectedScope) bool { return scopeMatchesQueries(scope) }

// MatchingRestrictionMod is the restriction a query-filtered record
// carries: the bits, and the filter they apply to ("creatures with
// defender", "creatures with power 2 or less"). Any of the queries
// matching is enough.
func MatchingRestrictionMod(r Restriction, queries []PermanentQuery) Mod {
	m := AddRestrictionsMod(r)
	m.Match = append([]PermanentQuery(nil), queries...)
	return m
}

// matchingQueriesPredicate is the AppliesTo of a query-filtered record:
// every creature that matches any of the mods' queries, and — for the
// opponents' form — that the record's controller does not control.
//
// The queries read EFFECTIVE characteristics (matchesLocked), so the
// keyword in "creatures with defender" is the layered answer and a
// granted defender counts. That is a layer result, so these scopes are
// folded after the pass (scopeReadsLayerResult).
func matchingQueriesPredicate(mods []Mod, scope AffectedScope, controller uuid.UUID) func(*Card, *Game, *Card) bool {
	var queries []PermanentQuery
	for _, m := range mods {
		queries = append(queries, m.Match...)
	}
	opponentsOnly := scope == ScopeOpponentsCreaturesMatching
	return func(target *Card, g *Game, _ *Card) bool {
		if target == nil || g == nil || !target.IsCreature() {
			return false
		}
		if opponentsOnly && target.Controller == controller {
			return false
		}
		for _, q := range queries {
			if q.matchesLocked(g, target) {
				return true
			}
		}
		return false
	}
}

// matchingScopeProblem is the registration and restore check for the
// query-filtered scopes: only addRestrictions mods, each naming at
// least one query. "" when sound.
//
// A mod carrying Match under any OTHER scope is refused too: nothing
// else reads the field, so a filter there would be silently dropped.
func matchingScopeProblem(scope AffectedScope, mods []Mod) string {
	if !scopeMatchesQueries(scope) {
		for _, m := range mods {
			if len(m.Match) != 0 {
				return fmt.Sprintf("a mod with scope %q carries a filter, which only a query-filtered record reads", scope)
			}
		}
		return ""
	}
	for _, m := range mods {
		if m.Kind != ModAddRestrictions {
			return fmt.Sprintf("a %s record carries %s, which it does not read", scope, m.Kind)
		}
		if len(m.Match) == 0 {
			return fmt.Sprintf("a %s mod names no filter", scope)
		}
	}
	return ""
}

// CantBlockThisTurnMatchingForEffect is "creatures <matching> can't
// block this turn" (#6237): Goblin Locksmith's creatures with defender,
// and — with `opponentsOnly` — Rumbling Ruin's "creatures your
// opponents control with power less than or equal to that number".
//
// `controller` is the effect's "you", which the opponents' form reads.
// The filter is read live until cleanup, so a creature that arrives this
// turn and matches cannot block either (CR 611.2c, 509.1b). Reports
// false, writing nothing, for no queries.
//
// Caller must hold g.mu (write).
func (g *Game) CantBlockThisTurnMatchingForEffect(sourceID, controller uuid.UUID, queries []PermanentQuery, opponentsOnly bool, label string) bool {
	if len(queries) == 0 {
		return false
	}
	scope := ScopeCreaturesMatching
	if opponentsOnly {
		scope = ScopeOpponentsCreaturesMatching
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, scope, controller,
		[]Mod{MatchingRestrictionMod(CantBlock, queries)}, g.UntilEndOfTurnDuration(), label)
}
