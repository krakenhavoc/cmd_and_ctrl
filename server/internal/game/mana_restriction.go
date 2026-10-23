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

	// SpendPurposeUnlock is paying an unlock cost (CR 709.5e, ADR
	// 0103): the special action that unlocks a Room's door. Neither a
	// cast nor an activation, so mana restricted to either does not
	// pay it; ManaRestrictUnlock is the restriction that does.
	SpendPurposeUnlock ManaSpendPurpose = "unlock"

	// SpendPurposeTurnFaceUp is paying the cost to turn a permanent face
	// up (CR 116.2g, CR 708.6: morph, megamorph, disguise, and a
	// manifested or cloaked creature's mana cost). Neither a cast nor an
	// activation, so mana restricted to either does not pay it;
	// ManaRestrictTurnFaceUp is the restriction that does (#3172,
	// Creeping Peeper).
	SpendPurposeTurnFaceUp ManaSpendPurpose = "turn_face_up"
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
	// ManaRestrictUnlock permits the token only on an unlock cost
	// (CR 709.5e, ADR 0103) — the "unlock doors" half of Smoky
	// Lounge's and Creeping Peeper's restrictions. Combine it with a
	// cast clause through ManaRestrictAnyOf.
	ManaRestrictUnlock = "purpose:unlock"
	// ManaRestrictTurnFaceUp permits the token only on the cost of
	// turning a permanent face up (CR 116.2g) — the "turn a permanent
	// face up" half of Creeping Peeper's restriction. Combine it with a
	// cast clause through ManaRestrictAnyOf.
	ManaRestrictTurnFaceUp = "purpose:turn_face_up"
	// ManaRestrictColorless permits the token only when the object
	// being paid for is colorless (CR 105.2c) — Shrine of the
	// Forsaken Gods, Eldrazi Temple.
	ManaRestrictColorless = "colorless"
	// ManaRestrictNotNonartifactSpell is a NEGATIVE clause — "this mana
	// can't be spent to cast a nonartifact spell" (Powerstone, #1727).
	// It admits an activation payment and a cast of an artifact spell,
	// and refuses a cast of anything else. Tags AND and every other tag
	// is positive, so the clause cannot be spelled as two of them.
	ManaRestrictNotNonartifactSpell = "not:nonartifact-spell"
	// ManaRestrictNotFromHand is a NEGATIVE clause — "this mana can't be
	// spent to cast spells from your hand" (Heartwood Crafter, #2811).
	// It admits any payment that is not a cast, and a cast whose spend
	// context names a zone other than the hand (a graveyard, exile, the
	// command zone, a library top). A cast whose zone is not known is
	// refused: weaker than printed, never stronger.
	ManaRestrictNotFromHand = "not:cast-from-hand"
	// ManaRestrictMonocolored permits the token only when the object
	// being paid for is exactly one colour (CR 105.2a) — Throne of
	// Eldraine's "spend this mana only to cast monocolored spells of
	// that color" (#1600), which pairs it with ManaRestrictColor. A
	// hybrid spell is every colour of its hybrid symbols (CR 202.2d),
	// so a {W/U} spell is multicolored and refused; a colourless one
	// is no colour and refused too.
	ManaRestrictMonocolored = "monocolored"
	// ManaRestrictMulticolored permits the token only when the object
	// being paid for is two or more colours (CR 105.2b) — Pillar of
	// the Paruns' and Obsidian Obelisk's "spend this mana only to cast
	// a multicolored spell" (#1600).
	ManaRestrictMulticolored = "multicolored"
	// ManaRestrictNoGeneric is a SYMBOL-level negative clause — "this
	// mana can't be spent to pay generic mana costs" (Jegantha, the
	// Wellspring, #2170). Every other tag asks what the mana pays FOR;
	// this one asks which SYMBOL of the cost it pays, so it is not
	// decided by ManaSpendContext.allows (which has no symbol) but by
	// the solvers, which refuse the token to the generic part of a cost
	// ({N}, {X}, a cost increase, commander tax: CR 107.4b) and offer
	// it to the coloured, hybrid and Phyrexian symbols (CR 107.4e-f).
	// It composes by AND with every other tag, and it must be a
	// top-level tag, never inside ManaRestrictAnyOf (which panics).
	// Weaker than printed on purpose where "as though it were mana of
	// any colour" widens a coloured symbol: the widened symbol is still
	// a coloured one, so this mana may pay it (CR 609.4b).
	ManaRestrictNoGeneric = "not:generic"
)

// noGeneric reports whether a token may not pay generic mana — it
// carries ManaRestrictNoGeneric. The one reader every solver shares.
func (t ManaToken) noGeneric() bool {
	for _, r := range t.Restrictions {
		if r == ManaRestrictNoGeneric {
			return true
		}
	}
	return false
}

// ManaRestrictType builds a "the object has this card type" tag —
// ManaRestrictType("Creature") for Ancient Ziggurat.
func ManaRestrictType(t string) string { return "type:" + t }

// ManaRestrictNotType builds a NEGATED card-type clause — "spend this
// mana only to cast noncreature spells" (Nardole, Resourceful Cyborg,
// #2136) is ManaRestrictNotType("Creature"). It is a clause about a
// SPELL: it admits a cast whose object has none of the named types and
// refuses an activation, an unlock, and a payment with no stated
// purpose (which has no object to be "non" anything). Several types
// mean none of them ("noncreature, nonland"). Tags AND and every
// keyed tag is positive, so the clause cannot be spelled as two of
// them.
func ManaRestrictNotType(types ...string) string {
	return "nottype:" + strings.Join(types, "|")
}

// ManaRestrictSubtype builds a "the object has this subtype" tag —
// ManaRestrictSubtype("Eldrazi") for Eldrazi Temple.
func ManaRestrictSubtype(t string) string { return "subtype:" + t }

// ManaRestrictSupertype builds a "the object has this supertype" tag
// — ManaRestrictSupertype("Legendary") for Delighted Halfling.
func ManaRestrictSupertype(t string) string { return "supertype:" + t }

// ManaRestrictColor builds a "the object is this colour" tag —
// ManaRestrictColor("R") for Pyromancer's Goggles' "a red instant or
// sorcery spell" (#1547). Uppercase WUBRG, the ManaSpendContext.Colors
// alphabet.
func ManaRestrictColor(c string) string { return "color:" + c }

// ManaRestrictAnyType builds a "the object has ANY of these card types"
// tag — ManaRestrictAnyType("Instant", "Sorcery") for "an instant or
// sorcery spell" (Pyromancer's Goggles, Boseiju, #1547).
//
// One tag with a `|`-joined value rather than two tags, because tags AND
// and this clause is an OR. The alternation works for every keyed tag
// (type, subtype, supertype, color); nothing a Scryfall type line prints
// contains a `|`.
func ManaRestrictAnyType(types ...string) string {
	return "type:" + strings.Join(types, "|")
}

// ManaRestrictAnyOf builds ONE tag that admits a payment satisfying
// ANY of the alternatives, each alternative being a list of tags that
// must ALL hold — "spend this mana only to cast Room spells and unlock
// doors" (Smoky Lounge) is
//
//	ManaRestrictAnyOf(
//	    []string{ManaRestrictCast, ManaRestrictSubtype("Room")},
//	    []string{ManaRestrictUnlock},
//	)
//
// Tags on a token AND, and a restriction like that is an OR of ANDs,
// so it is one tag rather than several. ADR 0103.
func ManaRestrictAnyOf(alternatives ...[]string) string {
	parts := make([]string, 0, len(alternatives))
	for _, alt := range alternatives {
		for _, t := range alt {
			if t == ManaRestrictNoGeneric {
				panic("game: ManaRestrictNoGeneric is a symbol-level tag and cannot sit inside ManaRestrictAnyOf")
			}
		}
		parts = append(parts, strings.Join(alt, anyOfAnd))
	}
	return anyOfPrefix + strings.Join(parts, anyOfOr)
}

const (
	anyOfPrefix = "anyof:"
	anyOfOr     = "||"
	anyOfAnd    = "&&"
)

// ManaSpendContext describes the object a mana payment is being made
// for, so restricted tokens can be admitted or refused. Built once
// per payment at the top of the spend path and threaded down into
// ManaPool.attemptSpend.
//
// The characteristics are the object's EFFECTIVE ones: a permanent's
// layered characteristic, and for a card being cast the types and
// colours a battlefield static gives it off the battlefield (#3862,
// zone_layers.go) — a creature card in hand under a Maskwood Nexus is
// every creature type, so Cavern of Souls mana named for Elf pays for
// it. With no such static it is the printed card.
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

	// SourceOnly is the spell's own "Spend only mana produced by basic
	// lands to cast this spell" (Imperiosaur) or "…by creatures"
	// (Myr Superion), #2556: when set, only a mana whose recorded
	// source (ManaToken.SourceKinds) has one of these kinds may pay.
	//
	// A property of the SPELL, not of any mana, so it rides this
	// context — which every cast payment, the pool solver, the
	// auto-tapper's pool top-up, internal/legal's probe and the preview
	// already thread through — rather than a tag on the token, where
	// the restriction of the cost would have to be written onto every
	// other mana source in the game. Set by ManaSpendForCast from the
	// card's catalog declaration (SpendOnlySourcesFor); zero for every
	// other payment, which means no source restriction.
	SourceOnly ManaSourceKinds

	// CastFrom is the zone a spell is being cast FROM (CR 601.2a moves
	// it to the stack first, so the zone has to be carried): what
	// ManaRestrictNotFromHand reads (#2811). Empty when the caller does
	// not know it, which that restriction treats as the hand. Set by
	// ManaSpendForCastFrom and ManaSpendForCastParams; meaningless for
	// any other purpose.
	CastFrom ZoneKind

	// AbilityOrigin is the row being paid for, for an activated or mana
	// ability's own payment (ADR 0155 Decision 7): set by the activation
	// and mana-ability payment sites, the zero origin everywhere else. A
	// spend rider's Covers reads it to reach only the rows one borrow lent
	// ("spend mana as though it were mana of any color to activate those
	// abilities").
	AbilityOrigin AbilityOrigin

	// Controller is the controller of the object whose ability is being
	// paid for (ManaSpendForAbility), and Payer the player paying, which
	// the any-colour reader fills in before it asks a Covers. Together
	// they answer Agatha's Soul Cauldron's "to activate abilities of
	// creatures YOU control" (#3201): an ability any player may activate
	// on someone else's creature is not covered. Zero for a cast.
	Controller uuid.UUID
	Payer      uuid.UUID

	// Object is the object whose ability is being paid for
	// (ManaSpendForAbility), uuid.Nil for a cast. Quicksilver Elemental's
	// "to pay the activation costs of THIS creature's abilities"
	// (AnyColorSpendStatic.OwnAbilities) compares it with the rider's
	// source.
	Object uuid.UUID

	// wish is what the spell being cast reads back from its payment
	// (#3853, ADR 0068 amendment 2026-10-15), snapshotted when the
	// context is built. Nil for every payment that is not a cast.
	//
	// It is an ORDERING input and nothing else: the auto-tapper reads it
	// to pick between sources, and the pool solver to pick between
	// tokens it was already free to take. No payability test reads it
	// (see spendWish). Plain data, so it can sit in a stored context
	// (AssistPrompt.spend) without carrying a Game or a closure with it.
	wish *spendWish
}

// ManaSpendForAbilityRow is ManaSpendForAbility for one row of the
// source's ability list: the same context, with the row's origin.
func ManaSpendForAbilityRow(c Card, o AbilityOrigin) ManaSpendContext {
	ctx := ManaSpendForAbility(c)
	ctx.AbilityOrigin = o
	return ctx
}

// ManaSpendForCastFrom is ManaSpendForCast for a cast from a known
// zone.
func ManaSpendForCastFrom(c Card, from ZoneKind) ManaSpendContext {
	ctx := ManaSpendForCast(c)
	ctx.CastFrom = from
	return ctx
}

// ManaSpendForCastParams is ManaSpendForCastFrom with the zone read off
// the cast's own parameters, as CastSpell reads it (an empty FromZone
// is the hand).
//
// #2575: and with the source restriction of the alternative cost the
// cast claims, when it prints one — Security Rhox's "You may pay {R}{G}
// rather than pay this spell's mana cost. Spend only mana produced by
// Treasures to cast it this way" restricts that payment and no other.
// Only a printed offer can carry one (AlternativeCost.SpendOnlySources),
// so the card's own offers are the ones looked in; a granted offer, or
// the printed mana cost, keeps the card's own restriction.
func ManaSpendForCastParams(c Card, params CastSpellParams) ManaSpendContext {
	from, _ := castZoneFromWire(params.FromZone)
	ctx := ManaSpendForCastFrom(c, from)
	if alt := AlternativeCostByKey(castOfferKey(c), params.AlternativeCost); alt != nil && alt.SpendOnlySources != 0 {
		ctx.SourceOnly = alt.SpendOnlySources
	}
	return ctx
}

// ManaSpendForCast builds the spend context for casting `c`. It does
// not know the zone the spell is cast from; a caller that does uses
// ManaSpendForCastFrom.
func ManaSpendForCast(c Card) ManaSpendContext {
	ch := c.Effective()
	return ManaSpendContext{
		Purpose:          SpendPurposeCast,
		Types:            ch.Types,
		Subtypes:         ch.Subtypes,
		Supertypes:       ch.Supertypes,
		Colors:           c.EffectiveColors(),
		AllCreatureTypes: HasAllCreatureTypes(&c),
		SourceOnly:       SpendOnlySourcesFor(CatalogKey(c)),
		wish:             newSpendWish(c),
	}
}

// allowsToken reports whether `tok` may be spent in this context: the
// token's own restrictions (allows) AND the payment's source
// restriction. A token with no recorded source kinds (mana from a
// spell, a token minted before kinds were recorded) can never satisfy
// a source restriction — the weaker direction.
func (ctx ManaSpendContext) allowsToken(tok ManaToken) bool {
	if ctx.SourceOnly != 0 && !tok.SourceKinds.HasAny(ctx.SourceOnly) {
		return false
	}
	return ctx.allows(tok.Restrictions)
}

// ManaSpendForAbility builds the spend context for activating an
// ability whose source permanent is `c`. CR 106.6: "activate
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
		Controller:       c.Controller,
		Object:           c.InstanceID,
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
	if rest, ok := strings.CutPrefix(r, anyOfPrefix); ok {
		for _, alt := range strings.Split(rest, anyOfOr) {
			if alt != "" && ctx.allows(strings.Split(alt, anyOfAnd)) {
				return true
			}
		}
		return false
	}
	switch r {
	case "":
		// Defensive: an empty tag is a card-file bug, not a
		// licence.
		return false
	case ManaRestrictCast:
		return ctx.Purpose == SpendPurposeCast
	case ManaRestrictActivate:
		return ctx.Purpose == SpendPurposeActivate
	case ManaRestrictUnlock:
		return ctx.Purpose == SpendPurposeUnlock
	case ManaRestrictTurnFaceUp:
		return ctx.Purpose == SpendPurposeTurnFaceUp
	case ManaRestrictNoGeneric:
		// Decided per symbol by the solvers (noGeneric), not by the
		// object being paid for: nothing here to refuse.
		return true
	case ManaRestrictNotFromHand:
		switch ctx.Purpose {
		case SpendPurposeActivate, SpendPurposeUnlock, SpendPurposeTurnFaceUp:
			return true
		case SpendPurposeCast:
			return ctx.CastFrom != "" && ctx.CastFrom != ZoneHand
		}
		return false
	case ManaRestrictNotNonartifactSpell:
		switch ctx.Purpose {
		case SpendPurposeActivate:
			return true
		case SpendPurposeCast:
			return containsFold(ctx.Types, "Artifact")
		}
		return false
	case ManaRestrictColorless:
		// An unknown purpose has no object, so it has no colour
		// either — refuse rather than read "no colours" as
		// "colorless".
		return ctx.hasPaidForObject() && len(ctx.Colors) == 0
	case ManaRestrictMonocolored:
		// #1600, CR 105.2a. The same unknown-purpose refusal as
		// colorless: no object, no colour count.
		return ctx.hasPaidForObject() && distinctColorCount(ctx.Colors) == 1
	case ManaRestrictMulticolored:
		// #1600, CR 105.2b.
		return ctx.hasPaidForObject() && distinctColorCount(ctx.Colors) >= 2
	}
	key, value, ok := strings.Cut(r, ":")
	if !ok || value == "" || ctx.Purpose == SpendPurposeUnknown {
		return false
	}
	if key == "nottype" {
		// #2136: a negated type is about a spell being cast. Its `|`
		// is "none of these", not the OR the positive tags read, so it
		// is decided here rather than in the alternation loop below.
		if ctx.Purpose != SpendPurposeCast {
			return false
		}
		for _, alt := range strings.Split(value, "|") {
			if alt == "" || containsFold(ctx.Types, alt) {
				return false
			}
		}
		return true
	}
	// #1547: a `|` in the value is an alternation — "an instant OR
	// sorcery spell". Any alternative satisfying the key satisfies the
	// tag; an empty alternative is a card-file bug and matches nothing.
	for _, alt := range strings.Split(value, "|") {
		if alt != "" && ctx.matchesKeyed(key, alt) {
			return true
		}
	}
	return false
}

// matchesKeyed decides one `key:value` tag with a single value.
func (ctx ManaSpendContext) matchesKeyed(key, value string) bool {
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
	case "color":
		// #1547: "a red instant or sorcery spell". The object's
		// colours, uppercase WUBRG; a colourless object has none.
		return containsFold(ctx.Colors, value)
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
// A declared list is the ability's printed words, so a subtype or colour
// it names is read through the source's text (CR 612.2, #3984): Eldrazi
// Temple's "Eldrazi spell" under an Artificial Evolution names the new
// type. A computed list names a CHOSEN word (Cavern of Souls), which no
// text change reaches, and is used as it is.
//
// Caller must hold g.mu (every call site is inside a mana-ability
// activation, which does).
func restrictionsFor(g *Game, ab *ManaAbilityShape, controller, source uuid.UUID) []string {
	if ab.RestrictionsFunc != nil {
		return copyRestrictions(ab.RestrictionsFunc(g, controller, source))
	}
	return g.costTextLocked(source).manaTags(copyRestrictions(ab.Restrictions))
}

// spendRidersFor is a mana ability's declared spend riders for one
// production, copied, with each rider's filter read through the source's
// text as restrictionsFor reads the restrictions (CR 612.2, #3984):
// Pyromancer's Goggles' "a red instant or sorcery spell" under a Sleight
// of Mind names the new colour.
//
// Caller must hold g.mu.
func spendRidersFor(g *Game, ab *ManaAbilityShape, source uuid.UUID) []ManaSpendRider {
	out := copyManaRiders(ab.SpendRiders)
	if text := g.costTextLocked(source); text != nil {
		for i := range out {
			out[i].When = text.manaTags(out[i].When)
		}
	}
	return out
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

// distinctColorCount is how many of the five colours `colors` names,
// each counted once and in any case — the "monocolored" / "multicolored"
// question (CR 105.2a–b). Anything that is not W, U, B, R or G is not
// a colour and is not counted.
func distinctColorCount(colors []string) int {
	seen := 0
	var have [5]bool
	for _, c := range colors {
		i := strings.Index("WUBRG", strings.ToUpper(c))
		if len(c) != 1 || i < 0 || have[i] {
			continue
		}
		have[i] = true
		seen++
	}
	return seen
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

// hasPaidForObject reports whether the payment is for an object whose
// characteristics a restriction can read: a spell, an ability or a door.
// An unknown purpose has no object, and neither does turning a permanent
// face up (#3172): its face-down self has no colour a restriction should
// read as "colorless".
func (ctx ManaSpendContext) hasPaidForObject() bool {
	return ctx.Purpose != SpendPurposeUnknown && ctx.Purpose != SpendPurposeTurnFaceUp
}

// ManaAbilityRestrictionsLocked is the spend restrictions one activation
// of `a` on `source` would stamp on its mana now (restrictionsFor): the
// computed list for a RestrictionsFunc row (Cavern of Souls' chosen
// type), the declared list otherwise. The view stamps it on a battlefield
// row whose list is computed, so a reader without a game handle sees the
// restriction the mint will apply (#3976).
//
// Caller must hold g.mu (read or write).
func (g *Game) ManaAbilityRestrictionsLocked(controller, source uuid.UUID, a ManaAbilityShape) []string {
	return restrictionsFor(g, &a, controller, source)
}

// CastAdmitsRestrictedMana reports whether mana carrying `restrictions`
// may pay for casting a spell with this type line and these colours,
// from the zone `fromZone` names on the wire ("" is the hand), under the
// matcher the pool solver spends by (ManaSpendContext.allows). It also
// refuses a symbol-level tag (ManaRestrictNoGeneric), which `allows`
// cannot answer, exactly as autoTapAdmitsRestricted does.
//
// For a reader that holds no Card: the bot's turn planner asks it
// through protocol, since a policy may not import this package (ADR 0033
// §3, #3976). allCreatureTypes is the object's changeling answer
// (HasAllCreatureTypes). An estimate, not a legality answer: the cast
// itself is decided by the enumerator, under the cast's own context.
func CastAdmitsRestrictedMana(restrictions []string, typeLine string, colors []string, allCreatureTypes bool, fromZone string) bool {
	if len(restrictions) == 0 {
		return true
	}
	for _, r := range restrictions {
		if r == ManaRestrictNoGeneric {
			return false
		}
	}
	from, ok := castZoneFromWire(fromZone)
	if !ok {
		return false
	}
	supers, types, subs := parseTypeLine(typeLine)
	ctx := ManaSpendContext{
		Purpose:          SpendPurposeCast,
		Types:            types,
		Subtypes:         subs,
		Supertypes:       supers,
		Colors:           colors,
		AllCreatureTypes: allCreatureTypes,
		CastFrom:         from,
	}
	return ctx.allows(restrictions)
}
