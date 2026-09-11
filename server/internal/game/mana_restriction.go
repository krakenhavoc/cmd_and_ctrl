package game

import (
	"strings"

	"github.com/google/uuid"
)

// mana_restriction.go is the spend half of restricted mana — "spend
// this mana only to cast a creature spell" (Ancient Ziggurat), "only
// to cast colorless spells" (Shrine of the Forsaken Gods), "only to
// cast colorless Eldrazi spells or activate abilities of colorless
// Eldrazi" (Eldrazi Temple), "only to cast a legendary spell"
// (Delighted Halfling).
//
// `ManaToken.Restrictions` has carried the production half since S15
// — "a list of opaque tags the cost validator can honour" — and until
// this file nothing wrote one and nothing read one. That asymmetry is
// exactly the #259 trap: a restriction enforced on production but not
// on spend turns "add {C}, spend only on Eldrazi" into plain {C}, and
// every card in the group ships strictly better than printed. So the
// two halves land together or not at all (#352 sub-gap 2).
//
// Restrictions are TAGS, not closures, for three reasons:
//
//   - they are game state. A restricted token sits in the pool across
//     an undo boundary, and clone.go already deep-copies the slice.
//     A func field would clone as a pointer and could never be
//     compared, logged, or sent to a client.
//   - they are inspectable. `["purpose:cast", "subtype:Eldrazi"]`
//     reads in an event log and in a test failure; a closure does not.
//   - the matching rule is small and shared. One `allows` method is
//     the only thing that decides whether a token may pay, so there is
//     exactly one place a permissive bug could live.
//
// The tag vocabulary is deliberately closed. An unknown tag is a
// DENY, not an allow (see `matchesRestriction`): a card that declares
// a restriction the matcher doesn't understand must make its mana
// unspendable rather than unrestricted. Weaker than printed is
// acceptable; stronger is not.

// ManaSpendPurpose names what a mana payment is FOR. Restrictions
// that key off the kind of payment ("only to cast", "only to
// activate") match against it.
//
// The zero value is deliberately "unknown", and a spend context with
// an unknown purpose satisfies NO restriction: a caller that has not
// said what it is paying for cannot spend restricted mana. Every
// payment path in the engine names its purpose (see
// ManaSpendForCast / ManaSpendForAbility); the zero value exists for
// the handful of internal call sites that have no object at all, and
// for tests.
type ManaSpendPurpose string

const (
	// SpendPurposeUnknown is the zero value — matches nothing.
	SpendPurposeUnknown ManaSpendPurpose = ""
	// SpendPurposeCast is paying a spell's cost (CR 601.2h).
	SpendPurposeCast ManaSpendPurpose = "cast"
	// SpendPurposeActivate is paying an activated ability's cost
	// (CR 602.2b). Includes mana abilities with a mana component —
	// a Signet's {1} is an activation payment.
	SpendPurposeActivate ManaSpendPurpose = "activate"
)

// Restriction tag constants and constructors. Card files build tags
// with these rather than typing string literals, so a typo is a
// compile error instead of a silently-unspendable token.
const (
	// ManaRestrictCast permits the token only on a spell's cost.
	ManaRestrictCast = "purpose:cast"
	// ManaRestrictActivate permits the token only on an activated
	// ability's cost.
	ManaRestrictActivate = "purpose:activate"
	// ManaRestrictColorless permits the token only when the object
	// being paid for is colorless (CR 105.2c) — Shrine of the
	// Forsaken Gods, Eldrazi Temple.
	ManaRestrictColorless = "colorless"
)

// ManaRestrictType builds a "the object has this card type" tag —
// ManaRestrictType("Creature") for Ancient Ziggurat.
func ManaRestrictType(t string) string { return "type:" + t }

// ManaRestrictSubtype builds a "the object has this subtype" tag —
// ManaRestrictSubtype("Eldrazi") for Eldrazi Temple.
func ManaRestrictSubtype(t string) string { return "subtype:" + t }

// ManaRestrictSupertype builds a "the object has this supertype" tag
// — ManaRestrictSupertype("Legendary") for Delighted Halfling.
func ManaRestrictSupertype(t string) string { return "supertype:" + t }

// ManaSpendContext describes the object a mana payment is being made
// for, so restricted tokens can be admitted or refused. Built once
// per payment at the top of the spend path and threaded down into
// ManaPool.attemptSpend.
//
// The characteristics are the object's EFFECTIVE ones where the layer
// engine has computed them (a permanent on the battlefield whose
// ability is being activated) and its printed ones otherwise (a spell
// on the stack, which the layer engine does not currently recompute).
// That is the conservative direction: a type-changing continuous
// effect that would ADD a type the restriction wants goes unnoticed
// and the payment is refused, which is weaker than printed rather
// than stronger.
type ManaSpendContext struct {
	// Purpose is what is being paid for. SpendPurposeUnknown
	// matches no restriction at all.
	Purpose ManaSpendPurpose

	// Types, Subtypes and Supertypes are the object's card types
	// ("Creature", "Artifact"), subtypes ("Eldrazi", "Human") and
	// supertypes ("Legendary", "Basic").
	Types      []string
	Subtypes   []string
	Supertypes []string

	// Colors is the object's colour list — uppercase WUBRG.
	// Empty means colorless, which is what ManaRestrictColorless
	// tests for.
	Colors []string

	// AllCreatureTypes marks an object that is every creature type
	// (CR 702.73a) — a changeling spell, or a permanent under a
	// Maskwood Nexus. A "subtype:" restriction is satisfied by ANY
	// creature type when it is set.
	//
	// It rides the context rather than being folded into Subtypes for
	// the reason HasAllCreatureTypes gives: the alternative is ~345
	// strings copied per payment, to answer a question one bool
	// answers. It matters because a changeling really is a creature
	// spell of the chosen type — casting Universal Automaton with
	// Cavern of Souls mana named for Elf is legal, and refusing it
	// would be a rules error a tribal player hits immediately.
	AllCreatureTypes bool
}

// ManaSpendForCast builds the spend context for casting `c`.
func ManaSpendForCast(c Card) ManaSpendContext {
	ch := c.Effective()
	return ManaSpendContext{
		Purpose:          SpendPurposeCast,
		Types:            ch.Types,
		Subtypes:         ch.Subtypes,
		Supertypes:       ch.Supertypes,
		Colors:           c.EffectiveColors(),
		AllCreatureTypes: HasAllCreatureTypes(&c),
	}
}

// ManaSpendForAbility builds the spend context for activating an
// ability whose source permanent is `c`. CR 106.6a: "activate
// abilities of colorless Eldrazi" restricts on the SOURCE's
// characteristics, not on what the ability does.
func ManaSpendForAbility(c Card) ManaSpendContext {
	ch := c.Effective()
	return ManaSpendContext{
		Purpose:          SpendPurposeActivate,
		Types:            ch.Types,
		Subtypes:         ch.Subtypes,
		Supertypes:       ch.Supertypes,
		Colors:           c.EffectiveColors(),
		AllCreatureTypes: HasAllCreatureTypes(&c),
	}
}

// allows reports whether a token carrying `restrictions` may be spent
// in this context. Every restriction must be satisfied (AND) — a
// token from Eldrazi Temple carries both "colorless" and
// "subtype:Eldrazi" and needs both.
//
// An empty / nil restriction list always passes: that is ordinary
// mana, which is the overwhelming majority of every pool.
func (ctx ManaSpendContext) allows(restrictions []string) bool {
	for _, r := range restrictions {
		if !ctx.matchesRestriction(r) {
			return false
		}
	}
	return true
}

// matchesRestriction decides one tag. An unrecognised tag returns
// false — see the file comment: unknown means unspendable, never
// unrestricted.
func (ctx ManaSpendContext) matchesRestriction(r string) bool {
	switch r {
	case "":
		// Defensive: an empty tag is a card-file bug, not a
		// licence.
		return false
	case ManaRestrictCast:
		return ctx.Purpose == SpendPurposeCast
	case ManaRestrictActivate:
		return ctx.Purpose == SpendPurposeActivate
	case ManaRestrictColorless:
		// An unknown purpose has no object, so it has no colour
		// either — refuse rather than read "no colours" as
		// "colorless".
		return ctx.Purpose != SpendPurposeUnknown && len(ctx.Colors) == 0
	}
	key, value, ok := strings.Cut(r, ":")
	if !ok || value == "" || ctx.Purpose == SpendPurposeUnknown {
		return false
	}
	switch key {
	case "type":
		return containsFold(ctx.Types, value)
	case "subtype":
		if containsFold(ctx.Subtypes, value) {
			return true
		}
		// CR 702.73a: a changeling object has the named subtype too,
		// as long as the name is a CREATURE type. It is not every
		// Equipment and every Aura.
		return ctx.AllCreatureTypes && IsCreatureType(value)
	case "supertype":
		return containsFold(ctx.Supertypes, value)
	}
	return false
}

// restrictionsFor resolves a mana ability's spend restrictions for
// one activation: the computed list when the ability has a
// RestrictionsFunc (Cavern of Souls' chosen type), the declared one
// otherwise.
//
// Both paths go through copyRestrictions, so the token that ends up
// in the pool never aliases either the catalog's process-lifetime
// slice or a slice the card file built and might reuse.
//
// Caller must hold g.mu (every call site is inside a mana-ability
// activation, which does).
func restrictionsFor(g *Game, ab *ManaAbilityShape, controller, source uuid.UUID) []string {
	if ab.RestrictionsFunc != nil {
		return copyRestrictions(ab.RestrictionsFunc(g, controller, source))
	}
	return copyRestrictions(ab.Restrictions)
}

// copyRestrictions returns a fresh backing array for a restriction
// list, or nil for an empty one. Every path that mints a ManaToken or
// a PendingChoice from a catalog-declared ability goes through this:
// the catalog's slice is process-lifetime and shared by every
// instance of the card, while the token is game state that clone.go
// deep-copies for undo. Aliasing the two would let a restored game
// mutate the catalog.
func copyRestrictions(rs []string) []string {
	if len(rs) == 0 {
		return nil
	}
	return append([]string(nil), rs...)
}

// containsFold is a case-insensitive membership test. Type lines come
// from Scryfall in title case and card files write them the same way,
// but a card file that wrote "creature" should not silently produce
// an unspendable token.
func containsFold(xs []string, want string) bool {
	for _, x := range xs {
		if strings.EqualFold(x, want) {
			return true
		}
	}
	return false
}
