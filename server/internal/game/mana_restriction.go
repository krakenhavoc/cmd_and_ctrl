package game

import (
	"strconv"
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
	// ManaRestrictPowerUp permits the token only on the cost of a
	// power-up ability (#4571, CR 702.193, CR 106.6): Quinjet Technician's
	// "Spend this mana only to activate power-up abilities". A purpose
	// tag, narrower than ManaRestrictActivate: it needs an activation AND
	// that the ability is a power-up one (ManaSpendContext.PowerUp).
	ManaRestrictPowerUp = "purpose:power_up"
	// ManaRestrictEquip permits the token only on the cost of an equip
	// ability (#4600, CR 702.6, CR 106.6): the "activate equip abilities"
	// half of Ronin, Shadow Stalker's "Spend this mana only to cast
	// Equipment spells or activate equip abilities". A purpose tag like
	// ManaRestrictPowerUp, narrower than ManaRestrictActivate: it needs an
	// activation AND that the ability is an equip one
	// (ManaSpendContext.Equip), which the source's characteristics cannot
	// say (an Equipment has other abilities that are not equip). Combine
	// it with the cast clause through ManaRestrictAnyOf.
	ManaRestrictEquip = "purpose:equip"
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
	// ManaRestrictCastFromExile is a POSITIVE clause — "spend this mana
	// only to cast spells from exile" (Interdimensional Web Watch,
	// #4460). It admits a spell cast from the exile zone and nothing
	// else: not an activation, not a cast from the hand, a graveyard,
	// the command zone or a library, and not a cast whose zone is not
	// known (weaker than printed, never stronger). The mirror image of
	// ManaRestrictNotFromHand, read from the same CastFrom.
	ManaRestrictCastFromExile = "cast:from-exile"
	// ManaRestrictCastFromGraveyard is the same POSITIVE clause for the
	// graveyard — "spend this mana only to cast a spell from your
	// graveyard" (Lord of the Forsaken, #5368). It admits a spell cast
	// from a graveyard and nothing else: not an activation, not a cast
	// from the hand, exile, the command zone or a library, and not a cast
	// whose zone is not known (weaker than printed, never stronger).
	ManaRestrictCastFromGraveyard = "cast:from-graveyard"
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
	// ManaRestrictXCost permits the token only on a cost that contains
	// {X} (#4834, CR 107.3, CR 106.6): Elementalist's Palette's "Spend
	// this mana only on costs that contain {X}". The whole cost, not
	// only its X part: a spell whose mana cost has {X}, or an activated
	// ability whose cost does (ManaSpendContext.CostHasX). A payment
	// that does not say whether its cost has {X} refuses it — weaker
	// than printed, never stronger.
	ManaRestrictXCost = "cost:has-x"
	// ManaRestrictSpellXCost permits the token only on a SPELL that has
	// {X} in its printed mana cost (CR 107.3, CR 106.6): Troyan, Gutsy
	// Explorer's "spells with {X} in their mana costs". Unlike
	// ManaRestrictXCost it asks about the card being cast, not the cost
	// being paid, so a spell with {X} in its mana cost cast for an
	// alternative cost without one still qualifies, and an activated
	// ability never does (ManaSpendContext.ManaCostHasX). A payment
	// that does not say what it casts refuses it.
	ManaRestrictSpellXCost = "spell:has-x"
	// ManaRestrictFaceDownSpell permits the token only on a spell cast
	// face down (CR 708.4): Tin Street Gossip's "cast face-down spells".
	// It asks about the spell being cast, so an activation, a unlock or
	// a turn-face-up never satisfies it; combine it with
	// ManaRestrictTurnFaceUp through ManaRestrictAnyOf.
	ManaRestrictFaceDownSpell = "spell:face-down"
	// ManaRestrictSpellNotOwned permits the token only on a spell its
	// payer does not own (#5457, CR 106.6, CR 108.3): Thieving Varmint's
	// "Spend this mana only to cast spells you don't own". It compares
	// the spell's owner (ManaSpendContext.SpellOwner) with the player
	// paying (Payer), so an activation never satisfies it, and neither
	// does a cast whose payer or owner the payment did not name (weaker
	// than printed, never stronger).
	ManaRestrictSpellNotOwned = "spell:not-owned"
)

// ManaRestrictAltCost builds "spend this mana only to pay a [keyword]
// cost" for a keyword that is an alternative cost of a spell (#5680,
// CR 106.6, CR 118.9): Unblinking Observer's "to pay a disturb cost" is
// ManaRestrictAltCost("disturb"). The key is the AlternativeCost.Key the
// keyword's constructor writes (effects.Disturb's "disturb"), so the tag
// admits a cast that claims that printed offer and nothing else: not the
// same card cast for its mana cost, not a cast for another alternative
// cost (flashback and escape are graveyard casts too, which is why
// ManaRestrictCastFromGraveyard is too wide), not an activation, and not
// a cast whose payment did not name the cost it pays (weaker than
// printed, never stronger). The whole cost of such a cast is the
// alternative cost with its increases and reductions (CR 601.2f), so
// the mana pays all of it. ManaSpendContext.AltCost is what it reads.
// Combine it with a cast clause through ManaRestrictAnyOf.
func ManaRestrictAltCost(key string) string { return altCostKey + ":" + key }

// altCostKey is the tag key of ManaRestrictAltCost.
const altCostKey = "alt"

// ManaRestrictManaValueAtLeast builds "spend this mana only to cast
// spells with mana value N or greater" (#4694, CR 106.6, CR 202.3):
// Ashling, Rimebound's ManaRestrictManaValueAtLeast(4). It is a clause
// about a SPELL, judged on the spell being cast with the mana value it
// has on the stack: the face being cast (CR 709.3b, CR 712.8c, a
// transformed spell's front-face cost), both halves of a fused split
// spell (CR 202.3d), and {X} at the value announced for it (CR 202.3e).
// It refuses an activation, an unlock, a face-up turn, a payment with no
// stated purpose, and a cast whose cost the engine cannot read (weaker
// than printed, never stronger). ManaSpendContext.ManaValue is what it
// reads.
func ManaRestrictManaValueAtLeast(n int) string {
	return manaValueKey + ":" + manaValueAtLeast + strconv.Itoa(n)
}

// ManaRestrictManaValueAtMost is the other comparison, "spells with mana
// value N or less", under the same rules as
// ManaRestrictManaValueAtLeast.
func ManaRestrictManaValueAtMost(n int) string {
	return manaValueKey + ":" + manaValueAtMost + strconv.Itoa(n)
}

const (
	manaValueKey     = "mv"
	manaValueAtLeast = ">="
	manaValueAtMost  = "<="
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

// ManaRestrictNotSubtype builds a NEGATED subtype clause about a spell
// being cast — "a non-Human creature spell" (Animal Attendant) is
// ManaRestrictCast, ManaRestrictType("Creature") and
// ManaRestrictNotSubtype("Human"). Like ManaRestrictNotType it admits
// only a cast, and several subtypes mean none of them. A changeling
// spell has every creature type (CR 702.73a), so it is never "non-" one.
func ManaRestrictNotSubtype(subtypes ...string) string {
	return "notsubtype:" + strings.Join(subtypes, "|")
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

	// PowerUp is set when the ability being paid for is a power-up
	// ability (CR 702.193): what ManaRestrictPowerUp reads. Set by the
	// activation sites from ActivatedAbilityShape.PowerUp; false for a
	// cast, a mana ability and every unnamed payment, which therefore
	// can't spend power-up-only mana.
	PowerUp bool

	// Equip is set when the ability being paid for is an equip ability
	// (CR 702.6): what ManaRestrictEquip reads. Set by the same sites as
	// PowerUp, from ActivatedAbilityShape.Equip; false for a cast, a mana
	// ability and every unnamed payment, which therefore can't spend
	// equip-only mana. The {0} twin of an equip ability (#4475) is a
	// mirrored row and carries the bit.
	Equip bool

	// CostHasX is set when the cost being paid contains {X}: what
	// ManaRestrictXCost reads (#4834). ManaSpendForCast sets it from the
	// spell's mana cost (ManaSpendForCastParams from the alternative
	// cost the cast claims, when it claims one), and the activation sites
	// from the {X} symbols of the row's mana (AbilityCost.XSlots), beside PowerUp. False for
	// every other payment.
	CostHasX bool

	// ManaCostHasX is set when the spell being cast has {X} in its
	// printed mana cost, whatever cost is being paid for it: what
	// ManaRestrictSpellXCost reads (Troyan, Gutsy Explorer). Set by
	// ManaSpendForCast beside CostHasX and, unlike it, not rewritten by an
	// alternative cost. False for every payment that is not a cast.
	ManaCostHasX bool

	// AltCost is the key of the printed alternative cost the cast claims
	// (AlternativeCost.Key, CR 118.9): what ManaRestrictAltCost reads
	// (#5680). Set by ManaSpendForCastParams when the card prints the
	// offer the cast names; empty for a cast at the mana cost, for a
	// granted offer, and for every payment that is not a cast.
	AltCost string

	// FaceDownSpell is set when the spell being cast is cast face down
	// (CR 708.4): what ManaRestrictFaceDownSpell reads. False for every
	// payment that is not a cast.
	FaceDownSpell bool

	// YourCommander is set when the spell being cast is a commander its
	// caster owns (CR 903.3d): what ManaRestrictYourCommander reads.
	// Set by ManaSpendForCast; false for every payment that is not a
	// cast.
	YourCommander bool

	// SpellOwner is the owner of the spell being cast (CR 108.3): what
	// ManaRestrictSpellNotOwned compares with Payer (#5457). Set by
	// ManaSpendForCast; zero for every payment that is not a cast.
	SpellOwner uuid.UUID

	// ManaValue is the mana value of the spell being cast as it is on
	// the stack, and ManaValueKnown whether the engine could read its
	// cost: what ManaRestrictManaValueAtLeast reads (#4694, CR 202.3).
	// ManaSpendForCast sets both from the face being cast with {X} at 0
	// (manaCostForValue: a transformed spell's front face, a face-down
	// spell's 0), and ManaSpendForCastParams counts {X} at the announced
	// value (CR 202.3e). WithX recounts it for a caller that tries
	// several X values. Zero and false for every other payment.
	ManaValue      int
	ManaValueKnown bool

	// manaValueBase and manaValueXSlots are ManaValue's parts, the value
	// with {X} at 0 and the number of {X} in the mana cost, kept so WithX
	// can recount it. Unexported, so a stored context (AssistPrompt.Spend)
	// keeps only ManaValue, which is settled by then.
	manaValueBase   int
	manaValueXSlots int

	// Controller is the controller of the object whose ability is being
	// paid for (ManaSpendForAbility), and Payer the player paying, which
	// the any-colour reader fills in before it asks a Covers. Together
	// they answer Agatha's Soul Cauldron's "to activate abilities of
	// creatures YOU control" (#3201): an ability any player may activate
	// on someone else's creature is not covered. Controller is zero for
	// a cast. A cast's payment names its Payer with PaidBy (#5457), for
	// ManaRestrictSpellNotOwned.
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
	// (AssistPrompt.Spend) without carrying a Game or a closure with it.
	wish *spendWish
}

// ManaSpendForAbilityRow is ManaSpendForAbility for one row of the
// source's ability list: the same context, with the row's origin.
func ManaSpendForAbilityRow(c Card, o AbilityOrigin) ManaSpendContext {
	ctx := ManaSpendForAbility(c)
	ctx.AbilityOrigin = o
	return ctx
}

// PaidBy is the context with the player paying named (#5457). Every
// cast payment names its payer, the caster, or the helper of an assist
// (CR 702.132a), so a restriction that asks whose spell it is
// (ManaRestrictSpellNotOwned) can answer.
func (ctx ManaSpendContext) PaidBy(payer uuid.UUID) ManaSpendContext {
	ctx.Payer = payer
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
	alt := AlternativeCostByKey(castOfferKey(c), params.AlternativeCost)
	if alt == nil {
		// CR 712.11d, ADR 0107 §4: a disturb cast has already turned the
		// card to its back face, whose own entry prints no offer; the
		// claim is the front face's (#5680).
		alt = frontFaceCastOffer(c, params.AlternativeCost)
	}
	if alt != nil {
		if alt.SpendOnlySources != 0 {
			ctx.SourceOnly = alt.SpendOnlySources
		}
		// #4834: the cost paid is the alternative one.
		ctx.CostHasX = strings.Contains(alt.ManaCost, "{X}")
		// #5680: and it is this one.
		ctx.AltCost = alt.Key
	}
	// #4694: the spell's mana value on the stack counts the X announced
	// for it (CR 202.3e). Only an {X} in its MANA cost counts, so an
	// alternative cost's X changes nothing here.
	return ctx.WithX(params.XValue)
}

// WithX is the context with the spell's mana value recounted for {X}
// announced as x (CR 202.3e), for a caller that weighs several values of
// X against one context (the legal-move enumerator's affordable-X
// search). A no-op for any payment that is not a cast, and for a cast
// whose cost could not be read.
func (ctx ManaSpendContext) WithX(x int) ManaSpendContext {
	if ctx.Purpose != SpendPurposeCast || !ctx.ManaValueKnown || x < 0 {
		return ctx
	}
	ctx.ManaValue = ctx.manaValueBase + ctx.manaValueXSlots*x
	return ctx
}

// NextManaValueX is the least X at or above `from` at which this cast's
// mana value first reaches the "N or greater" threshold of a
// ManaRestrictManaValueAtLeast tag on some mana in `pool` (#4694), and
// false when there is none: the cast is not an X spell, its mana value
// is not known, or no floating mana carries such a tag.
//
// For an X search that stops at the first X it cannot pay because cost
// grows with X: mana for a spell with mana value 4 or greater makes that
// untrue, since {X}{R} at X = 0 can't spend it and at X = 3 can. The
// search jumps to the X this names and goes on from there. A hint, not a
// verdict: the X it names is still asked of the payer.
func (ctx ManaSpendContext) NextManaValueX(pool ManaPool, from int) (int, bool) {
	if ctx.Purpose != SpendPurposeCast || !ctx.ManaValueKnown || ctx.manaValueXSlots <= 0 {
		return 0, false
	}
	best, ok := 0, false
	for _, tok := range pool {
		for _, r := range tok.Restrictions {
			if !strings.Contains(r, manaValueKey+":") {
				continue
			}
			for _, piece := range strings.FieldsFunc(r, func(c rune) bool { return c == '|' || c == '&' }) {
				piece = strings.TrimPrefix(strings.TrimPrefix(piece, anyOfPrefix), manaValueKey+":")
				n, found := strings.CutPrefix(piece, manaValueAtLeast)
				if !found {
					continue
				}
				v, err := strconv.Atoi(n)
				if err != nil {
					continue
				}
				// The least X with base + slots*X >= v.
				need := v - ctx.manaValueBase
				x := 0
				if need > 0 {
					x = (need + ctx.manaValueXSlots - 1) / ctx.manaValueXSlots
				}
				if x >= from && (!ok || x < best) {
					best, ok = x, true
				}
			}
		}
	}
	return best, ok
}

// withCastManaValue stamps the mana value of `c` as a spell with {X} at
// 0 (CR 202.3e), read through manaCostForValue as every mana-value read
// is.
func (ctx ManaSpendContext) withCastManaValue(c Card) ManaSpendContext {
	return ctx.withManaValueOf(manaCostForValue(c))
}

// withManaValueOf stamps the mana value of a spell whose mana cost is
// manaCost, {X} at 0; an unreadable cost leaves it unknown.
func (ctx ManaSpendContext) withManaValueOf(manaCost string) ManaSpendContext {
	cost, err := ParseCost(manaCost)
	if err != nil {
		return ctx
	}
	ctx.ManaValueKnown = true
	ctx.manaValueBase = cost.ManaValue()
	ctx.manaValueXSlots = cost.XSlots
	ctx.ManaValue = ctx.manaValueBase
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
		CostHasX:         strings.Contains(manaCostForValue(c), "{X}"),
		ManaCostHasX:     strings.Contains(manaCostForValue(c), "{X}"),
		FaceDownSpell:    c.FaceDown,
		YourCommander:    c.IsCommander && c.Owner != uuid.Nil && c.Owner == c.Controller,
		SpellOwner:       c.Owner,
		wish:             newSpendWish(c),
	}.withCastManaValue(c)
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
	case ManaRestrictPowerUp:
		return ctx.Purpose == SpendPurposeActivate && ctx.PowerUp
	case ManaRestrictEquip:
		return ctx.Purpose == SpendPurposeActivate && ctx.Equip
	case ManaRestrictXCost:
		return (ctx.Purpose == SpendPurposeCast || ctx.Purpose == SpendPurposeActivate) && ctx.CostHasX
	case ManaRestrictYourCommander:
		return ctx.Purpose == SpendPurposeCast && ctx.YourCommander
	case ManaRestrictSpellXCost:
		return ctx.Purpose == SpendPurposeCast && ctx.ManaCostHasX
	case ManaRestrictFaceDownSpell:
		return ctx.Purpose == SpendPurposeCast && ctx.FaceDownSpell
	case ManaRestrictSpellNotOwned:
		return ctx.Purpose == SpendPurposeCast && ctx.Payer != uuid.Nil && ctx.SpellOwner != uuid.Nil && ctx.Payer != ctx.SpellOwner
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
	case ManaRestrictCastFromExile:
		return ctx.Purpose == SpendPurposeCast && ctx.CastFrom == ZoneExile
	case ManaRestrictCastFromGraveyard:
		return ctx.Purpose == SpendPurposeCast && ctx.CastFrom == ZoneGraveyard
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
	if key == "notsubtype" {
		// A negated subtype is about a spell being cast, as nottype is.
		if ctx.Purpose != SpendPurposeCast {
			return false
		}
		for _, alt := range strings.Split(value, "|") {
			if alt == "" || ctx.matchesKeyed("subtype", alt) {
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
	case manaValueKey:
		return ctx.matchesManaValue(value)
	case altCostKey:
		// #5680: a clause about the cost a spell is cast for.
		return ctx.Purpose == SpendPurposeCast && ctx.AltCost == value
	}
	return false
}

// matchesManaValue decides one ManaRestrictManaValueAtLeast /
// ManaRestrictManaValueAtMost comparison (#4694). A clause about a spell:
// only a cast whose mana value is known can satisfy it, and a value that
// does not parse matches nothing.
func (ctx ManaSpendContext) matchesManaValue(value string) bool {
	if ctx.Purpose != SpendPurposeCast || !ctx.ManaValueKnown {
		return false
	}
	if n, ok := strings.CutPrefix(value, manaValueAtLeast); ok {
		v, err := strconv.Atoi(n)
		return err == nil && ctx.ManaValue >= v
	}
	if n, ok := strings.CutPrefix(value, manaValueAtMost); ok {
		v, err := strconv.Atoi(n)
		return err == nil && ctx.ManaValue <= v
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
//
// manaCost is the cost the spell's mana value is computed from (the face
// cast, a transformed spell's front face) and x the X announced for it
// (CR 202.3e), for ManaRestrictManaValueAtLeast (#4694); a cost that
// does not parse leaves the mana value unknown, which that restriction
// refuses.
func CastAdmitsRestrictedMana(restrictions []string, typeLine string, colors []string, allCreatureTypes bool, fromZone, manaCost string, x int) bool {
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
	return ctx.withManaValueOf(manaCost).WithX(x).allows(restrictions)
}
