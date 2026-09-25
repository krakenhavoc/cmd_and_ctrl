package game

import (
	"strconv"

	"github.com/google/uuid"
)

// cast_permission.go — S42, ADR 0066: an EFFECT lets you cast or play
// a card from a zone the card's own text does not open.
//
// cast_zones.go answers the card's half of the question: "you may
// cast this card from your graveyard" is printed on Gravecrawler, and
// flashback is the same permission with a price attached. Both are
// keyed by oracle ID, so they belong to every copy of the card.
//
// This file is the other half. Snapcaster Mage gives flashback to one
// card that never printed it; Past in Flames gives it to a SET of
// cards, locked when it resolves (CR 611.2c); Underworld Breach gives
// escape to every nonland card in a graveyard for as long as it is on
// the battlefield (CR 702.138); Bolas's Citadel lets you play the top
// card of your library for life (CR 401.5). Impulse exile, airbend,
// warp, cascade and a defeated Siege's back face are the same kind of
// thing pointed at exile.
//
// Before ADR 0066 the last group had its own type — ExilePlayPermission,
// a value on Card — and the first four had nothing. One model now
// covers all of it, because #652's checklist and Ian's standing rule
// both say the same thing: no second per-instance grant.
//
// THE SHAPE OF THE MODEL, in three decisions that carry the file:
//
//  1. TWO SCOPES, not three. "One card" and "a set locked at
//     resolution" are the same thing — CR 611.2c locks the set of
//     affected objects when the one-shot effect resolves, so the
//     resolution walks the graveyard once and writes down the
//     instances. A card that arrives later is simply not in the list.
//     That is ScopeCards. A rule over a zone ("every nonland card in
//     your graveyard", "the top card of your library") is
//     ScopeStanding, narrowed by a pure-data PermissionFilter.
//
//  2. STANDING PERMISSIONS ARE DERIVED, NEVER STORED. Every one of
//     them is granted by a permanent on the battlefield, so walking
//     the battlefield per query IS "for as long as the source
//     remains": two Underworld Breaches compose, one of them leaving
//     does not revoke the other's permission, and a Breach exiled in
//     response stops granting before the cast is validated. Exactly
//     the argument land_drops.go makes for deriving
//     CatalogAdditionalLandPlays rather than writing an allowance
//     onto the player.
//
//  3. CR 400.7 IS AN IDENTITY CHECK. A grant ends when its card
//     leaves the zone by ANY route, not only when it is cast. That
//     used to be six "zero the field on the way out" sites, one of
//     which (MoveCard) carried the other five. Card.ObjectEpoch
//     counts how many times a card has become a new object, and a
//     PermissionCardRef names {instance, epoch}. A Snapcaster target
//     exiled and returned to the graveyard has a new epoch and no
//     flashback; a cascade hit that reaches the stack has had its
//     epoch bumped by the move, so the grant is spent without anyone
//     writing a clear. One integer, and it is the rule rather than a
//     proxy for it.
//
// The type holds no closures. That is what lets CardSnapshot mirror
// it and TestEmbeddedDomainTypesStayPureData prove it, and it is why
// the filter is a struct of flags rather than a predicate: a
// permission that needs a real predicate is a ScopeStanding grant
// declared on a CardDef, which can compute one.

// PermissionScope says what a permission covers.
type PermissionScope string

const (
	// ScopeCards names specific card objects — one (Snapcaster, a
	// Siege, an impulse-exiled card) or a set locked at resolution
	// (Past in Flames, The Grim Captain's Locker). The zero value,
	// because every pre-ADR-0066 grant was this shape.
	ScopeCards PermissionScope = ""

	// ScopeStanding is a rule over the whole zone, narrowed by
	// Filter: "each nonland card in your graveyard has escape",
	// "you may play the top card of your library".
	ScopeStanding PermissionScope = "standing"
)

// GrantTiming is the timing rule a permission applies to the cast it
// opens (CR 601.2, 307.1).
type GrantTiming string

const (
	// TimingNormal leaves the card's own timing alone: a sorcery is
	// still sorcery-speed, an instant is still an instant. Every card
	// in ADR 0066 is this, and it is the zero value on purpose —
	// a grant that forgets to say must not silently open the stack.
	TimingNormal GrantTiming = ""

	// TimingFlash is "as though it had flash" (CR 702.8), the shape
	// madness (#657) and the Descendants' Path family want. Nothing
	// declares it yet; it exists so the gate in CastSpell reads a
	// permission rather than growing a second branch later.
	TimingFlash GrantTiming = "flash"

	// TimingSorcery forces sorcery speed even on an instant — the
	// "only any time you could cast a sorcery" clause several
	// graveyard permissions print, and the one Teferi, Time Raveler
	// imposes on each opponent (#1195).
	TimingSorcery GrantTiming = "sorcery"

	// TimingYourTurnOnly is "you can cast spells only during your
	// turn" (Dosan the Falling Leaf), and it is NOT TimingSorcery:
	// Dosan leaves you every instant-speed window on your own turn
	// and takes away the rest, so the two cannot share a value.
	//
	// No CastPermission declares it — it is the per-PLAYER
	// restriction's spelling (cast_timing.go), carried on this enum
	// rather than on a second one so the engine has ONE timing
	// vocabulary. Exactly the posture TimingFlash was added under.
	TimingYourTurnOnly GrantTiming = "your_turn"

	// TimingPlot is a plotted card's window (CR 702.170d, #1318): its
	// owner's main phase while the stack is empty — and NOTHING widens
	// it. That is the difference from TimingSorcery, which a per-player
	// flash grant (Vedalken Orrery, Leyline of Anticipation) overrides:
	// the plot rule is the permission's own timing, not the card's, so
	// a plotted instant or a plotted card with flash is still cast only
	// in that window. Only a CastPermission carries it; no per-player
	// statement may (CastTimingOpenLocked reads it first and stops).
	TimingPlot GrantTiming = "plot"
)

// PermissionCardRef is one card OBJECT a ScopeCards permission names:
// the instance, plus the CR 400.7 epoch it carried when the grant
// landed. A card that has changed zones since is a new object and is
// not the one the grant named.
type PermissionCardRef struct {
	ID    uuid.UUID `json:"id"`
	Epoch int       `json:"epoch"`
}

// PermissionFilter narrows which cards in the zone a ScopeStanding
// permission opens. The zero value opens every card, which is Bolas's
// Citadel.
//
// A closed set of flags rather than a predicate, because a stored
// permission must stay pure data (see the file header). It grows a
// field when a card needs one.
type PermissionFilter struct {
	// LandsOnly is Oracle of Mul Daya and Courser of Kruphix: "you
	// may play LANDS from the top of your library". Playing a land is
	// not casting (CR 305.1), so this is a play permission that opens
	// no spell.
	LandsOnly bool `json:"landsOnly,omitempty"`

	// NonLandOnly is Underworld Breach's "each NONLAND card in your
	// graveyard has escape".
	NonLandOnly bool `json:"nonLandOnly,omitempty"`

	// CreatureOnly is Realmwalker's "creature spells", and the half
	// of The Grim Captain's Locker that says "each CREATURE card".
	CreatureOnly bool `json:"creatureOnly,omitempty"`

	// InstantOrSorceryOnly is Past in Flames' "each instant and
	// sorcery card in your graveyard".
	InstantOrSorceryOnly bool `json:"instantOrSorceryOnly,omitempty"`

	// NoncreatureOnly is the half of "noncreature spells" nothing in
	// ADR 0066 needed and a timing statement does (#1195) — Borne
	// Upon a Wind's sibling clause, and the shape a "you may cast
	// noncreature spells as though they had flash" card wants.
	NoncreatureOnly bool `json:"noncreatureOnly,omitempty"`

	// SorceryOnly is Teferi, Time Raveler's +1: "you may cast SORCERY
	// spells as though they had flash" (#1195). Narrower than
	// InstantOrSorceryOnly, which would also open an instant that
	// needs no opening.
	SorceryOnly bool `json:"sorceryOnly,omitempty"`

	// FromChosenType marks a filter whose creature type is the one
	// named as the SOURCE permanent entered (CR 614.12, S26's
	// Card.NamedTribe) — Realmwalker. The catalog declares the flag;
	// the derivation reads the source's NamedTribe into CreatureType,
	// because a Realmwalker with no type named yet grants nothing.
	FromChosenType bool `json:"fromChosenType,omitempty"`

	// CreatureType is the subtype a qualifying card must have. Filled
	// by the derivation for FromChosenType; a catalog card may also
	// set it directly.
	CreatureType string `json:"creatureType,omitempty"`
}

// Matches reports whether a card in the zone qualifies under this
// filter. The zero filter matches everything.
func (f PermissionFilter) Matches(c Card) bool {
	if f.LandsOnly && !c.IsLand() {
		return false
	}
	if f.NonLandOnly && c.IsLand() {
		return false
	}
	if f.CreatureOnly && !c.IsCreature() {
		return false
	}
	if f.InstantOrSorceryOnly && !c.IsInstant() && !c.IsSorcery() {
		return false
	}
	if f.NoncreatureOnly && c.IsCreature() {
		return false
	}
	if f.SorceryOnly && !c.IsSorcery() {
		return false
	}
	if f.CreatureType != "" && !cardHasCreatureType(c, f.CreatureType) {
		return false
	}
	return true
}

// CastPermission is one granted permission to cast or play cards from
// a zone. The zero value grants nothing: Player is uuid.Nil.
type CastPermission struct {
	// Player is who may cast or play — for impulse exile, usually NOT
	// the card's owner.
	Player uuid.UUID `json:"player"`

	// Zone is where the cast comes FROM: exile, a graveyard, or a
	// library. Never hand (CR 601.2 already allows it) and never the
	// command zone (CR 903.4 is the format's, not an effect's).
	Zone ZoneKind `json:"zone"`

	// ZoneOwner names the SEAT whose pile a ScopeStanding permission
	// is over, when that is not the holder's own: Xanathar, Guild
	// Kingpin's "you may play the top card of THEIR library"
	// (CR 401.5). uuid.Nil — every permission written before #1035 —
	// means the holder's own graveyard or library.
	//
	// Read by permissionReachesPileLocked, which is the ownership
	// half of a standing permission's scope, and by the CR 401.5 look
	// below. A ScopeCards permission ignores it: naming an INSTANCE
	// is naming an object wherever it sits, which is the whole
	// distinction between the two scopes (#1022).
	//
	// A DERIVED standing permission never carries one. The catalog is
	// static and cannot name a seat, so standingCastPermissionsLocked
	// zeroes the field on the way out; a card that wants "each
	// opponent's library" grants one stored permission per opponent
	// out of a resolution, the way Xanathar's upkeep trigger does.
	// The view leans on that: a foreign holder can only ever come
	// from a STORED permission, so the per-holder stamp needs no walk
	// of the battlefield to rule one out.
	ZoneOwner uuid.UUID `json:"zoneOwner,omitempty"`

	// Scope and its two payloads. Cards is read for ScopeCards,
	// Filter for ScopeStanding; the other is ignored rather than
	// asserted, because a permission is data and a half-filled one
	// should grant less, never panic.
	//
	// Filter carries no `omitzero`: it is part of GameSnapshot's graph
	// (GameSnapshot.CastPermissions[].Filter), and that option's
	// effect depends on the building Go toolchain below 1.24 (#1492)
	// — CI is pinned to 1.22.
	Scope  PermissionScope     `json:"scope,omitempty"`
	Cards  []PermissionCardRef `json:"cards,omitempty"`
	Filter PermissionFilter    `json:"filter"`

	// TopOfLibraryOnly restricts a ZoneLibrary permission to the card
	// currently on top (CR 401.5). Every library permission sets it;
	// the field exists rather than being implied by the zone so that
	// a future "play any card from your library" reads as the
	// exception it would be.
	TopOfLibraryOnly bool `json:"topOfLibraryOnly,omitempty"`

	// SeesLibraryTop is CR 401.5's other half, carried by the grant
	// rather than by a permanent: "you may LOOK at the top card of
	// their library any time" (Xanathar, Guild Kingpin), which is
	// what makes the card in the clause beside it playable at all.
	//
	// ADR 0066 decision 5 keeps the visibility rule on the POSITION
	// and derives it from the battlefield — CardDef.LibraryTopVisible,
	// read by LibraryTopVisibilityLocked. That derivation answers for
	// the permanent's controller's OWN library and cannot answer for
	// anybody else's, because the clause that opens another seat's
	// library is granted by a resolution to one chosen player and is
	// not a static ability of any permanent. So a cross-seat grant
	// says it here, and LibraryTopVisibleToLocked is the one place
	// both spellings are read — the engine's position check and the
	// view's knower stamp agree by construction (#1035).
	//
	// Meaningless without ZoneLibrary, and a permission that opens
	// another seat's library top without it opens NOTHING: a card you
	// cannot see is a card you cannot play, and the failure is a card
	// file that does nothing rather than one that cheats.
	SeesLibraryTop bool `json:"seesLibraryTop,omitempty"`

	// --- the price -------------------------------------------------

	// AltCostKey is the alternative-cost key this permission
	// synthesises — "flashback", "escape", "bolas_citadel". Shared
	// with the printed keywords on purpose (#652): CR 702.34a's
	// "if the flashback cost was paid" and CR 702.138b's "escaped"
	// read StackItem.AltCost, and they must work however the
	// permission arrived. Empty means the cast pays the printed cost
	// (impulse exile) or Cost below (airbend, cascade).
	AltCostKey string `json:"altCostKey,omitempty"`

	// Cost is a mana cost in Scryfall brace notation paid INSTEAD of
	// the card's printed mana cost — airbend's "{2}", cascade's
	// "{0}". Empty means "its mana cost", which is what flashback
	// granted by Snapcaster and escape granted by Breach both cost.
	//
	// SIMPLIFICATION, inherited verbatim from ExilePlayPermission:
	// the override is CHARGED rather than offered, so a card whose
	// printed cost is below airbend's {2} cannot be cast for the
	// cheaper printed price. Strictly weaker than printed, never
	// stronger.
	Cost string `json:"cost,omitempty"`

	// LifeEqualToManaValue is Bolas's Citadel: "pay life equal to its
	// mana value rather than pay its mana cost". A COST (CR 119.4),
	// so a player without the life cannot claim it at all —
	// which is why it becomes an AlternativeCost.Life rather than a
	// drawback on resolution.
	//
	// Not applied to a LAND: a land has no mana cost to replace, and
	// Citadel's clause says "if you cast a spell this way".
	LifeEqualToManaValue bool `json:"lifeEqualToManaValue,omitempty"`

	// ExileOtherFromGraveyard is escape's "exile N other cards from
	// your graveyard" (CR 702.138a) — Underworld Breach's three, The
	// Grim Captain's Locker's four.
	ExileOtherFromGraveyard int `json:"exileOtherFromGraveyard,omitempty"`

	// ExileOnResolution is flashback's "exile this card instead of
	// putting it anywhere else any time it would leave the stack"
	// (CR 702.34a). A REPLACEMENT, so it also catches a granted
	// flashback that fizzles or is answered by Hinder — the clause
	// that keeps Snapcaster's gift from being infinite.
	ExileOnResolution bool `json:"exileOnResolution,omitempty"`

	// AnyColor lets the holder spend mana as though it were mana of
	// any color for this card (Breeches, Brazen Plunderer).
	AnyColor bool `json:"anyColor,omitempty"`

	// AnyType is the wider clause: "mana of any TYPE can be spent to
	// cast that spell" (Hostage Taker, Gonti, Night Minister,
	// Outrageous Robbery). Colorless is a type of mana but not a color
	// (CR 106.1b), so AnyColor leaves a {C} requirement standing and
	// this does not — a stolen Thought-Knot Seer is castable off
	// Swamps. It implies AnyColor; a permission setting both is read
	// as this one. ADR 0066's 2026-09-24 amendment (#1573).
	AnyType bool `json:"anyType,omitempty"`

	// --- the window ------------------------------------------------

	// Duration is when the permission ENDS (CR 611.2), in the one
	// vocabulary ADR 0063 gives every continuous effect in the engine
	// — duration.go, swept through the same durationExpiredLocked.
	// #945 replaced the {UntilTurn, WhileInZone} pair ADR 0066
	// inherited from ExilePlayPermission with this field; the ADR
	// recorded the debt at the time and this is it paid.
	//
	// Four of the five kinds are used here:
	//
	//	UntilEndOfTurn          Snapcaster, Past in Flames, the
	//	                        Locker, impulse exile, cascade, a
	//	                        Siege's face grant — and, stamped one
	//	                        seat-turn out by
	//	                        UntilEndOfYourNextTurnDuration,
	//	                        Reckless Impulse's "until the end of
	//	                        your next turn"
	//	WhileInZone             airbend, warp, and every permission
	//	                        derived from a permanent
	//	UntilYourNextTurn       nothing prints it on a cast permission
	//	                        yet; it costs nothing to accept
	//	Indefinite              likewise
	//
	// ForAsLongAs is the one a permission must not carry: its
	// condition watches a battlefield object, and a permission's
	// "for as long as the source remains" is already free, because
	// standing permissions are derived from the battlefield on every
	// query rather than stored (see standingCastPermissionsLocked).
	//
	// THE ZERO VALUE IS "UNTIL END OF TURN", UNSTAMPED, and
	// GrantCastPermissionForEffect stamps it against the current turn
	// on the way in. That is the safe default in both directions: a
	// caller who forgets gets the shortest window rather than an
	// unbounded grant, which is the trap the old `UntilTurn int`
	// sprang (0 was both the zero value and a turn number).
	//
	// No `omitzero`, for Filter's reason above (#1492).
	Duration Duration `json:"duration"`

	// NotBeforeSeq is the earliest turn sequence the permission is
	// live on — warp's "you may cast it from exile ON A LATER TURN"
	// (CR 702.185a), and the same clause foretell prints
	// (CR 702.143a). Zero means "from now".
	//
	// IT SURVIVED THE #945 SWAP DELIBERATELY, and this is the whole
	// reason the window is a Duration PLUS one int rather than a
	// Duration alone. Duration models when an effect ENDS; CR 611.2
	// has no vocabulary for when one STARTS, because a continuous
	// effect starts as it is created. A cast permission is the one
	// thing in the engine that can be granted now and open later, so
	// the floor is a field of the permission rather than a kind, and
	// it applies whatever the Duration says: "for as long as it
	// remains exiled" does not weaken "on a later turn".
	//
	NotBeforeSeq int `json:"notBeforeSeq,omitempty"`

	// LegacyNotBeforeTurn is populated only while decoding a snapshot
	// written before ADR 0059. restoreGame converts the old round unit
	// to NotBeforeSeq and clears it before the game can be captured again.
	LegacyNotBeforeTurn int `json:"notBeforeTurn,omitempty"`

	// --- what, and when --------------------------------------------

	// Timing is the timing rule the cast obeys. See GrantTiming: the
	// zero value changes nothing, and a grant never opens the
	// sorcery-speed gate unless it says so.
	Timing GrantTiming `json:"timing,omitempty"`

	// GrantsHaste gives the permanent this cast produces haste —
	// suspend's CR 702.62a. A property of the PERMISSION rather than
	// of the card, because it is the effect that granted the cast
	// that grants the haste: the same Rift Bolt hard-cast from hand
	// has none.
	//
	// Read once, in CastSpell, where the permission is consumed. See
	// Game.grantHasteForCastLocked for the layer-6 grant it becomes
	// and for the declared duration simplification.
	GrantsHaste bool `json:"grantsHaste,omitempty"`

	// CastOnly restricts the permission to CASTING. Ragavan says "you
	// may CAST that card", and a land exiled by Ragavan is stranded
	// because playing a land is not casting (CR 305.1, 116.2a).
	// Breeches says "you may PLAY those cards" and leaves it false.
	CastOnly bool `json:"castOnly,omitempty"`

	// Faces are the printed faces this permission opens (ADR 0034) — a
	// defeated Siege's "exile it, then cast it transformed"
	// (CR 310.12b) is []int{1}, and an adventure card's "you may cast
	// the creature from exile" (CR 715.4) is []int{0}. It NARROWS
	// rather than widens: a permission that names faces opens those
	// and no other.
	//
	// EMPTY means "this permission does not speak about faces", and
	// every permission that names none leaves the card's own
	// CastableFaces to decide.
	//
	// A SLICE rather than the `Face int` this started life as (S32),
	// because zero had to carry both meanings and could not: an
	// adventure's creature half IS face 0, so "face 0 and no other"
	// was unsayable and CR 715.4 could not be expressed at all. It is
	// the same defect #945 took out of the window one field up — 0 was
	// both the zero value and a turn number — fixed the same way, by
	// giving the field a value that means "nothing to say". Pure data,
	// and the same shape Card.CastableFaces already returns, so the
	// two compose in faceForCastLocked rather than arguing.
	Faces []int `json:"faces,omitempty"`

	// --- provenance ------------------------------------------------

	// Source is the card instance whose effect granted this, and
	// SourceName / Label are what the client and the log print. A
	// derived standing permission fills all three at derivation.
	Source     uuid.UUID `json:"source,omitempty"`
	SourceName string    `json:"sourceName,omitempty"`
	Label      string    `json:"label,omitempty"`
}

// Granted reports whether the permission names anyone at all.
func (p *CastPermission) Granted() bool {
	return p != nil && p.Player != uuid.Nil
}

// CastPermissionActiveForEffect reports whether playerID may use this
// permission right now: it names them, its CR 702.185a floor has been
// reached, and its CR 611.2 duration has not run out.
//
// THE one liveness test, and a method on *Game because ADR 0063's
// durationExpiredLocked needs the game — the seat-turn counts, the
// battlefield and the pin. Before #945 this was a value method
// reading the permission's own three fields, which is exactly the
// second duration vocabulary ADR 0066 recorded as debt.
//
// Nil-safe, so the cast path can ask without a guard.
//
// Caller must hold g.mu (read or write).
func (g *Game) CastPermissionActiveForEffect(p *CastPermission, playerID uuid.UUID) bool {
	if p == nil || p.Player == uuid.Nil || p.Player != playerID {
		return false
	}
	// The floor is checked first because it applies to an unbounded
	// window too (see NotBeforeSeq).
	if p.NotBeforeSeq > 0 && g.Turn.Seq < p.NotBeforeSeq {
		return false
	}
	// `false`: this is a query, not the cleanup sweep. An
	// UntilEndOfTurn permission is live for the whole of the turn it
	// names and is dropped by sweepCastPermissionsLocked at that
	// turn's cleanup step — the same split ScopedEffect lives under.
	return !g.durationExpiredLocked(p.Duration, false)
}

// NamesCard reports whether a ScopeCards permission names this exact
// card OBJECT — the instance AND the CR 400.7 epoch it carried when
// the permission was granted.
func (p *CastPermission) NamesCard(c Card) bool {
	if p == nil {
		return false
	}
	for _, ref := range p.Cards {
		if ref.ID == c.InstanceID && ref.Epoch == c.ObjectEpoch {
			return true
		}
	}
	return false
}

// CoversCard reports whether this permission opens a cast of `c` out
// of `zone`, ignoring the turn window (Active is the other half).
func (p *CastPermission) CoversCard(c Card, zone ZoneKind) bool {
	if p == nil || p.Zone != zone {
		return false
	}
	if p.Scope == ScopeStanding {
		return p.Filter.Matches(c)
	}
	return p.NamesCard(c)
}

// GrantsFaces returns the faces this permission opens, when it names
// any and it is playerID's permission.
//
// NO WINDOW CHECK, since #945, and that is a tightening rather than a
// loosening: every caller reaches a permission through
// CastPermissionForLocked, which has already asked
// CastPermissionActiveForEffect. Folding a second liveness test in
// here would be a second copy of the rule — and the copy would need
// the game, which is the one thing this pure-data type does not have.
// A permission that is not live never reaches this function, so a
// Siege back face left uncast still cannot narrow anything after its
// window shuts.
func (p *CastPermission) GrantsFaces(playerID uuid.UUID) ([]int, bool) {
	if p == nil || len(p.Faces) == 0 || p.Player == uuid.Nil || p.Player != playerID {
		return nil, false
	}
	return p.Faces, true
}

// GrantsFace is GrantsFaces for the callers that want THE face — the
// enumerator materialising the half it is about to offer, and the
// view pricing it. Every permission anything declares today names
// exactly one, so a permission naming several answers false rather
// than picking a half on the caller's behalf; faceForCastLocked is
// the one place that knows what to do with a choice.
func (p *CastPermission) GrantsFace(playerID uuid.UUID) (int, bool) {
	faces, ok := p.GrantsFaces(playerID)
	if !ok || len(faces) != 1 {
		return 0, false
	}
	return faces[0], true
}

// NamedFace is GrantsFace without the PLAYER — the read the VIEW
// wants, which asks what a grant opens before it knows whose it is
// (CastPermissionOnCardForEffect answers for any seat, because the
// grant is public information). A warp grant whose turn has not come
// yet still has to label its greyed-out button with the half it will
// open (CR 702.185a), and a client that showed the front face of a
// defeated Siege there would name the wrong card.
func (p *CastPermission) NamedFace() (int, bool) {
	if p == nil || len(p.Faces) != 1 {
		return 0, false
	}
	return p.Faces[0], true
}

// AlternativeCostFor synthesises the CR 118.9 offer this permission
// prices the cast under, or nil when the permission charges the
// printed cost (impulse exile) or its own flat Cost (airbend,
// cascade — those ride CastCostFor instead, because they are not
// claimable offers the caster names).
//
// Built on demand and never stored, which is what keeps the stored
// type free of the TargetSpec closures an AlternativeCost carries.
//
// A LAND gets nil whatever the permission says: a land has no mana
// cost to replace, and Bolas's Citadel's life clause is written about
// casting a spell.
func (p *CastPermission) AlternativeCostFor(card Card) *AlternativeCost {
	if p == nil || p.AltCostKey == "" || card.IsLand() {
		return nil
	}
	out := AlternativeCost{
		Key:      p.AltCostKey,
		Label:    p.Label,
		ManaCost: card.ManaCost,
		FromZone: p.Zone,
	}
	if p.Cost != "" {
		out.ManaCost = p.Cost
	}
	if p.LifeEqualToManaValue {
		// CR 202.3: the life is the card's mana value, read off the
		// face being cast. The mana cost is REPLACED, not added to —
		// "rather than pay its mana cost".
		out.ManaCost = ""
		out.Life = card.ManaValue()
	}
	if p.ExileOtherFromGraveyard > 0 {
		out.ExileFromGraveyard = escapeExileSpec(p.ExileOtherFromGraveyard)
	}
	out.ExileOnLeavingStack = p.ExileOnResolution
	if out.Label == "" {
		out.Label = p.AltCostKey
	}
	return &out
}

// CatalogCastPermissions returns the STANDING permissions a
// battlefield permanent with the given catalog key grants its
// controller. Populated at init from CardDef.CastPermissions.
//
// A derived hook rather than a CR 613 layer, for the reason
// CatalogAdditionalLandPlays gives: the layer engine models
// characteristics of objects, and "each nonland card in your
// graveyard has escape" is not one — it is a permission about a
// player and a zone.
var CatalogCastPermissions func(oracleID string) []CastPermission

// GrantCastPermissionForEffect stores a permission for its player. The
// one write path, so "who may cast what" has one home and one shape.
//
// A permission naming no player, no zone, or (for ScopeCards) no
// cards is dropped rather than stored: it would grant nothing and
// would still have to be swept.
//
// Caller must hold g.mu (write).
func (g *Game) GrantCastPermissionForEffect(perm CastPermission) {
	if perm.Player == uuid.Nil || perm.Zone == "" {
		return
	}
	if perm.Scope != ScopeStanding && len(perm.Cards) == 0 {
		return
	}
	if perm.Duration == (Duration{}) {
		// The zero Duration is "until end of turn", unstamped. Stamp
		// it against the turn the grant is being made in, so the
		// window is a fact about this turn rather than about whichever
		// cleanup step happens to run next (ADR 0063 Decision 2).
		perm.Duration = g.UntilEndOfTurnDuration()
	}
	p := g.playerByIDLocked(perm.Player)
	if p == nil {
		return
	}
	p.CastPermissions = append(p.CastPermissions, perm)
	g.stampPermittedViewersLocked(perm)
}

// stampPermittedViewersLocked is FaceDownPermitted's look (#1573):
// the holder of a permission over a card exiled face down under that
// kind may look at it. "You may look at and play those cards" is one
// clause, and a card you may play but cannot see is unplayable in
// practice, so the grant is what stamps the knower — whichever
// primitive made it. A card exiled face down under any other kind is
// left alone: Necropotence's `exiled` card stays unreadable to a
// permission that happens to name it.
//
// Only ScopeCards permissions over exile can name such a card, and the
// epoch has to match: a card that left exile and came back is a new
// object, and was not what the grant named (CR 400.7).
//
// Caller must hold g.mu (write).
func (g *Game) stampPermittedViewersLocked(perm CastPermission) {
	if perm.Scope == ScopeStanding || perm.Zone != ZoneExile || g.Exile == nil {
		return
	}
	for i := range g.Exile.Cards {
		c := &g.Exile.Cards[i]
		if c.FaceDownKind == FaceDownPermitted && perm.NamesCard(*c) {
			c.AddKnower(perm.Player)
		}
	}
}

// castPermissionHoldersLocked is every seat holding a stored
// permission that names this card object — FaceDownPermitted's viewer
// set, derived. Liveness is not asked: CR 406.3 lets a player who may
// look keep looking until the card leaves exile, and every permission
// that stamps this kind lasts that long anyway.
//
// Caller must hold g.mu.
func (g *Game) castPermissionHoldersLocked(c Card) []uuid.UUID {
	var out []uuid.UUID
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		for i := range p.CastPermissions {
			if p.CastPermissions[i].NamesCard(c) {
				out = append(out, p.ID)
				break
			}
		}
	}
	return out
}

// GrantCastPermissionOverCardForEffect grants `perm` over one card
// wherever it currently sits, filling in the zone from where the card
// actually is. Reports whether the card was found.
//
// The single-card door onto the model, and the one card code should
// reach for: Snapcaster's "target instant or sorcery card in your
// graveyard gains flashback" is this call and nothing else. A
// permission whose Zone is already set is left alone, so a caller who
// means "grant this over the graveyard" can say so and get nothing
// when the card has moved.
//
// Caller must hold g.mu (write).
func (g *Game) GrantCastPermissionOverCardForEffect(cardID uuid.UUID, perm CastPermission) bool {
	z := g.findCardZoneLocked(cardID)
	if z == nil {
		return false
	}
	if perm.Zone == "" {
		perm.Zone = z.Kind
	}
	if perm.Zone != z.Kind {
		return false
	}
	for _, c := range z.Cards {
		if c.InstanceID != cardID {
			continue
		}
		if perm.Player == uuid.Nil {
			perm.Player = c.Owner
		}
		g.GrantCastPermissionToCardsForEffect(perm, []Card{c})
		return true
	}
	return false
}

// CastPermissionOnCardByIDForEffect is CastPermissionOnCardForEffect
// for a caller that has an ID rather than a card and does not know
// which zone it is in — the zone browser, the log line, and every
// test that wants to ask "is this card playable by anyone".
//
// Caller must hold g.mu (read or write).
func (g *Game) CastPermissionOnCardByIDForEffect(cardID uuid.UUID) *CastPermission {
	z := g.findCardZoneLocked(cardID)
	if z == nil {
		return nil
	}
	for _, c := range z.Cards {
		if c.InstanceID == cardID {
			return g.CastPermissionOnCardForEffect(c, z.Kind)
		}
	}
	return nil
}

// GrantCastPermissionToCardsForEffect is the CR 611.2c constructor:
// it locks the set of card objects NOW, stamping each one's current
// epoch, and stores one permission naming all of them.
//
// Past in Flames and The Grim Captain's Locker call it with a whole
// graveyard's worth of cards; Snapcaster calls it with one. That they
// are the same call is the point of ScopeCards.
//
// Caller must hold g.mu (write).
func (g *Game) GrantCastPermissionToCardsForEffect(perm CastPermission, cards []Card) {
	if len(cards) == 0 {
		return
	}
	perm.Scope = ScopeCards
	perm.Cards = make([]PermissionCardRef, 0, len(cards))
	for _, c := range cards {
		perm.Cards = append(perm.Cards, PermissionCardRef{ID: c.InstanceID, Epoch: c.ObjectEpoch})
	}
	g.GrantCastPermissionForEffect(perm)
}

// CastPermissionForLocked is THE query: does an effect let playerID
// cast or play this card out of this zone right now?
//
// It returns nil for hand and the command zone (CR 601.2 and CR 903.4
// need no effect); otherwise the stored or derived permission that
// opens the cast. The cast path, the view and the bot enumerator all
// read this one function, so none of them can disagree about what is
// legal.
//
// It does NOT answer nil for a card whose own text already opens the
// zone, and deliberately: Gravecrawler under an Underworld Breach is
// castable both ways and the caster picks. Which PRICE wins is a
// separate question, answered in one other place —
// resolveAlternativeCostLocked consults the catalog first, so a
// printed flashback cost is never repriced by a permission that
// happens to name the same key.
//
// #760 hook: a permission that carries a "cast only if" condition
// (legendary sorcery, Rakdos) is checked by the one announce-time
// cast gate that issue designs, not here. Nothing in ADR 0066
// carries one.
//
// Caller must hold g.mu (read or write).
func (g *Game) CastPermissionForLocked(playerID uuid.UUID, card Card, zone ZoneKind) *CastPermission {
	if zone == ZoneHand || zone == ZoneCommand {
		return nil
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return nil
	}
	// ADR 0090: a CR 722.3c prepare copy answers for itself, and ONLY
	// for itself. Its permission is derived from the prepared
	// permanent (prepare.go), and nothing else may open it — an
	// impulse grant or a standing rule over exile reaching a copy
	// would cast an object CR 722.3c keeps castable only by the
	// prepared permanent's controller, and only while it is prepared.
	if card.PrepareCopy {
		if perm := g.prepareCopyPermissionLocked(card, zone); perm != nil && perm.Player == playerID {
			return perm
		}
		return nil
	}
	// Stored permissions first: a named object beats a standing rule,
	// because the named one is the narrower statement and is the only
	// one that can carry a face or a per-instance price.
	for i := range p.CastPermissions {
		perm := &p.CastPermissions[i]
		if !g.CastPermissionActiveForEffect(perm, playerID) || !perm.CoversCard(card, zone) {
			continue
		}
		// #1035: the pile rule is asked of a STORED standing
		// permission too, not only a derived one. Both are "a rule
		// over a zone" and the zone is the holder's own unless the
		// grant names a seat; Xanathar's until-end-of-turn grant over
		// one chosen opponent's library is a stored ScopeStanding
		// permission, and it is the only kind that can name one.
		if !permissionReachesPileLocked(playerID, perm, card, zone) {
			continue
		}
		if !g.permissionPositionOKLocked(playerID, perm, card) {
			continue
		}
		out := *perm
		return &out
	}
	for _, perm := range g.standingCastPermissionsLocked(p) {
		if !perm.CoversCard(card, zone) {
			continue
		}
		if !permissionReachesPileLocked(p.ID, &perm, card, zone) {
			continue
		}
		if !g.permissionPositionOKLocked(p.ID, &perm, card) {
			continue
		}
		out := perm
		return &out
	}
	return nil
}

// PileOwnerFor resolves WHOSE graveyard or library this permission
// speaks about: the seat it names, or the holder's own when it names
// none. The one place the uuid.Nil default is spelled out, so a
// caller cannot get the "your own pile" reading wrong.
func (p *CastPermission) PileOwnerFor(holder uuid.UUID) uuid.UUID {
	if p == nil || p.ZoneOwner == uuid.Nil {
		return holder
	}
	return p.ZoneOwner
}

// permissionReachesPileLocked is the ownership half of a STANDING
// permission's scope: Underworld Breach gives escape to the cards in
// YOUR graveyard, Bolas's Citadel opens the top of YOUR library, and
// nothing in PermissionFilter can say so.
//
// It went unwritten until #1022 because every caller asked the
// question about a pile it had already scoped — CastSpell resolved
// "graveyard" to the caster's own, the enumerator walked the caster's
// own, and the view walked each seat asking about that seat. The
// moment any of the three learned to ask about ANOTHER seat's pile (a
// ScopeCards permission can legitimately name a card there —
// Wrexial's "cast target instant or sorcery card from that player's
// graveyard"), a standing permission with no ownership clause would
// have followed it in and handed a Breach controller every graveyard
// at the table.
//
// #1035 gave the LIBRARY the same rule for the same reason. It used to
// be true by a different accident — permissionPositionOKLocked read
// CR 401.5's "the top card of your library" off the HOLDER's own pile,
// which both scoped the permission and made a cross-seat one
// impossible. Now that the position check follows the CARD, two
// Coursers on one table would otherwise have each controller playing
// lands off the other's revealed library.
//
// ZoneOwner is the way OUT of the scoping, and the only way: a grant
// that names a seat (Xanathar's chosen opponent) reaches that seat's
// pile and no other. A ScopeCards permission is deliberately not
// checked at all — naming an INSTANCE is naming an object wherever it
// sits, which is the whole distinction between the two scopes.
//
// Exile needs none of it: it is a shared zone by construction, so
// "whose exile" is not a question a permission has to answer.
func permissionReachesPileLocked(holder uuid.UUID, perm *CastPermission, card Card, zone ZoneKind) bool {
	if perm.Scope != ScopeStanding {
		return true
	}
	if zone != ZoneGraveyard && zone != ZoneLibrary {
		return true
	}
	// A card in a graveyard is in its OWNER's graveyard (CR 404.3) and
	// a library holds its owner's cards, so the card's owner IS the
	// pile's owner and no zone scan is needed on a path the view walks
	// per card per seat. An ownerless card — a token that never had
	// one — reaches nobody's standing grant.
	return card.Owner != uuid.Nil && card.Owner == perm.PileOwnerFor(holder)
}

// permissionPositionOKLocked enforces the restrictions that are about
// WHERE in the zone the card sits rather than about the card: CR
// 401.5's "the top card of your library", and the visibility that
// makes it playable at all.
//
// BOTH halves are asked about the library the CARD IS IN, not the one
// the holder owns (#1035). Reading the holder's own pile was the same
// class of accident #1022 took out of the graveyard: it happened to
// scope every permission written so far, because every printed
// library clause before Xanathar says "your library" — and it meant a
// permission over another seat's library top could never match the
// card, so it opened nothing, silently, before any surface could ask.
// The position rule is about a POSITION IN A PILE, and the pile is the
// card owner's (CR 401.1).
//
// A player may only play the top card of a library if they can SEE it
// — a permission that opens the top card without the matching "you may
// look" or "revealed" clause opens nothing. Every printed card carries
// both halves, so this is a guard against a card file that declares
// one and forgets the other. LibraryTopVisibleToLocked is the one
// place that knows the spellings.
//
// Caller must hold g.mu.
func (g *Game) permissionPositionOKLocked(holder uuid.UUID, perm *CastPermission, card Card) bool {
	if !perm.TopOfLibraryOnly {
		return true
	}
	owner := g.playerByIDLocked(card.Owner)
	if owner == nil || owner.Library == nil || len(owner.Library.Cards) == 0 {
		return false
	}
	if owner.Library.Cards[len(owner.Library.Cards)-1].InstanceID != card.InstanceID {
		return false
	}
	return g.LibraryTopVisibleToLocked(owner.ID, holder)
}

// CastPermissionGate pairs a standing permission with an ADR 0071
// designation and/or a card-specific Condition (#1314) — Fortune
// Teller's Talent's level-2 line: "as long as this Class is level 2 or
// greater" (ActiveWhen) AND "as long as you've cast a spell this turn"
// (Condition), together.
//
// CATALOG-ONLY, deliberately a SEPARATE type from CastPermission
// rather than two more fields on it: CastPermission is dual-purpose —
// a catalog declaration on one path and a value STORED on
// Player.CastPermissions, mirrored verbatim into GameSnapshot, on the
// other — and a func field on the shared type would make every
// snapshotted permission carry one, whether any instance actually set
// it or not. snapshot_drift_test.go's TestSnapshotMirrorsHaveNoFuncs
// enforces exactly this: GameSnapshot must be serialisable all the way
// down, and Condition is a closure. CastPermissionGate is reachable
// only from CardDef and CatalogGatedCastPermissions, neither of which
// the snapshot ever touches — the same posture ActivatedAbility.Condition
// and TriggeredAbility.AppliesTo already have, one struct over.
type CastPermissionGate struct {
	// Permission is the standing permission itself, exactly as an
	// entry in CatalogCastPermissions's slice would be.
	Permission CastPermission

	// ActiveWhen is the ADR 0071 designation gate, read by the same
	// activeOnly filter StaticAbilitiesForCard and CostModifiersForCard
	// already use.
	ActiveWhen Designation

	// Condition is a further "as long as …" clause that is not one of
	// ADR 0071's four designations. Nil means no further condition.
	// The mirror of ActivatedAbility.Condition (activated.go) one slot
	// over, and the same signature: *ForEffect reads only, under the
	// lock standingCastPermissionsLocked already holds. `source` is
	// the permanent contributing the permission — the "you" of "you've
	// cast a spell" is `controller`, not necessarily this card's
	// controller (a stolen Talent still asks about ITS new
	// controller).
	Condition func(g *Game, controller, source uuid.UUID) bool
}

// CatalogGatedCastPermissions returns the GATED standing permissions a
// battlefield permanent with this catalog key grants its controller —
// separate from CatalogCastPermissions so that hook's existing
// signature, and every test already stubbing it, is untouched. Nil,
// or a nil return, means every permission this card grants is
// ungated, which is every card but Fortune Teller's Talent today.
var CatalogGatedCastPermissions func(oracleID string) []CastPermissionGate

// standingCastPermissionsLocked derives every ScopeStanding
// permission the player currently holds, from the permanents they
// control. Nothing is stored, which IS the duration: a source that
// leaves stops granting on the next query, and two sources compose.
//
// CatalogAbilityKey rather than CatalogKey, for the reason
// EffectiveLandDropsLocked gives: "each nonland card in your
// graveyard has escape" is a static ability, and an Underworld Breach
// that has lost its abilities grants nothing.
//
// Two catalog hooks, walked in the same loop: the ordinary
// (ungated) permissions first, then the gated ones (#1314) — activeOnly
// applies the ADR 0071 gate and Condition runs right after, BEFORE
// either kind reaches the ZoneOwner/Source/Duration stamping
// stampStandingPermissionLocked shares between them. A level-1
// Fortune Teller's Talent never reaches that stamping for its level-2
// line, exactly as it never reaches the layer pass for an anthem it
// does not have yet.
//
// Caller must hold g.mu.
func (g *Game) standingCastPermissionsLocked(p *Player) []CastPermission {
	if p == nil || g.Battlefield == nil {
		return nil
	}
	if CatalogCastPermissions == nil && CatalogGatedCastPermissions == nil {
		return nil
	}
	var out []CastPermission
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != p.ID {
			continue
		}
		// The empty KEY is the skip, not an empty oracle ID — a token
		// has one of its own since #521 (ADR 0083 decision 3).
		key := CatalogAbilityKey(*c)
		if key == "" {
			continue
		}
		if CatalogCastPermissions != nil {
			for _, perm := range CatalogCastPermissions(key) {
				if stamped, ok := stampStandingPermissionLocked(perm, p, c); ok {
					out = append(out, stamped)
				}
			}
		}
		if CatalogGatedCastPermissions == nil {
			continue
		}
		gated := activeOnly(*c, CatalogGatedCastPermissions(key), func(gp CastPermissionGate) Designation {
			return gp.ActiveWhen
		})
		for _, gp := range gated {
			if gp.Condition != nil && !gp.Condition(g, p.ID, c.InstanceID) {
				continue
			}
			if stamped, ok := stampStandingPermissionLocked(gp.Permission, p, c); ok {
				out = append(out, stamped)
			}
		}
	}
	return out
}

// stampStandingPermissionLocked turns a catalog-declared permission
// into the derived, player-scoped value standingCastPermissionsLocked
// hands out — the stamping both of its callers share, so a gated and
// an ungated permission from the same card can never disagree about
// what a "standing" permission means.
//
// Caller must hold g.mu.
func stampStandingPermissionLocked(perm CastPermission, p *Player, c *Card) (CastPermission, bool) {
	perm.Player = p.ID
	perm.Scope = ScopeStanding
	// #1035: a catalog entry is static and cannot name a SEAT, so a
	// derived permission is always about its holder's own pile. Zeroed
	// rather than trusted, because the view's per-holder stamp relies
	// on it: a foreign holder can only come from a stored permission,
	// which is what lets the per-card walk rule one out without
	// deriving every seat's standing set.
	perm.ZoneOwner = uuid.Nil
	perm.Source = c.InstanceID
	if perm.SourceName == "" {
		perm.SourceName = c.Name
	}
	// A permission unbounded in turns, because its duration is the
	// source's presence and this slice is rebuilt from the battlefield
	// on every query. CR 611.2b's "for as long as", spelled the way a
	// derived permission can afford to spell it: nothing stored, so
	// nothing to expire.
	perm.Duration = WhileInZoneDuration()
	if perm.Filter.FromChosenType {
		// Realmwalker: "the chosen type". A permanent whose entry
		// choice has not been answered yet names no type and grants
		// nothing (CR 614.12).
		if c.NamedTribe == "" {
			return CastPermission{}, false
		}
		perm.Filter.CreatureType = c.NamedTribe
	}
	return perm, true
}

// sweepCastPermissionsLocked drops permissions whose CR 611.2
// duration has run out, and ScopeCards permissions whose every named
// object has moved on. `endOfTurn` is true only in the CR 514.2
// cleanup sweep and is handed straight to durationExpiredLocked,
// exactly as sweepScopedStaticsLocked hands it — one function decides
// what a duration means, for statics and permissions alike (#945).
//
// TWO MOMENTS, the same two a "until your next turn" static needs:
// the cleanup step (sweepTurnEndLocked) and the beginning of a turn
// (onTurnBeganLocked). The statics' third moment — the top of every
// layer recompute — buys a permission nothing, because the only kind
// that can go false between turns is ForAsLongAs and a permission
// never carries one (see CastPermission.Duration).
//
// This is HYGIENE, not correctness: CastPermissionActiveForEffect and
// NamesCard already refuse an expired or stale permission on every
// query, and standing permissions are never stored at all. Without it
// the slice would grow for the length of the game.
//
// Allocates a fresh slice rather than filtering in place with
// `s[:0]`: the backing array is shared with every undo snapshot Clone
// has taken, so an in-place compaction would rewrite history. That is
// the trap cloneCard exists to avoid on the card side, and the one
// sweepScopedStaticsLocked calls out on the static side.
//
// Caller must hold g.mu (write).
func (g *Game) sweepCastPermissionsLocked(endOfTurn bool) {
	for _, p := range g.Seats {
		if p == nil || len(p.CastPermissions) == 0 {
			continue
		}
		kept := make([]CastPermission, 0, len(p.CastPermissions))
		for _, perm := range p.CastPermissions {
			if g.durationExpiredLocked(perm.Duration, endOfTurn) {
				continue
			}
			if perm.Scope != ScopeStanding && !g.anyNamedObjectStillThereLocked(perm) {
				continue
			}
			kept = append(kept, perm)
		}
		if len(kept) == 0 {
			kept = nil
		}
		p.CastPermissions = kept
	}
}

// anyNamedObjectStillThereLocked reports whether at least one card
// object a ScopeCards permission names is still in the permission's
// zone wearing the epoch it was granted at.
//
// It looks for the card WHERE IT IS rather than in the pile the holder
// owns (#1035): a permission names an object, and Wrexial's names one
// in somebody else's graveyard, so resolving the holder's own pile
// here swept a live grant away at the next cleanup step. CR 400.7 is
// still the whole test — findCardZoneLocked answers with the zone the
// card is in now, and a card that has moved on either fails the epoch
// check or is in the wrong kind of zone.
//
// Caller must hold g.mu.
func (g *Game) anyNamedObjectStillThereLocked(perm CastPermission) bool {
	for _, ref := range perm.Cards {
		zone := g.findCardZoneLocked(ref.ID)
		if zone == nil || zone.Kind != perm.Zone {
			continue
		}
		for i := range zone.Cards {
			if perm.NamesCard(zone.Cards[i]) {
				return true
			}
		}
	}
	return false
}

// permissionZoneLocked resolves a permission's zone kind to ONE
// seat's pile: that player's graveyard or library, or the shared
// exile, which has no owner to ask about.
//
// It answers "which pile of this player's", never "whose pile" — see
// permissionReachesPileLocked and PileOwnerFor for the second
// question, which a permission can now answer about another seat
// (#1035).
//
// Caller must hold g.mu.
func (g *Game) permissionZoneLocked(playerID uuid.UUID, kind ZoneKind) *Zone {
	switch kind {
	case ZoneExile:
		return g.Exile
	case ZoneGraveyard, ZoneLibrary:
		p := g.playerByIDLocked(playerID)
		if p == nil {
			return nil
		}
		if kind == ZoneLibrary {
			return p.Library
		}
		return p.Graveyard
	}
	return nil
}

// AnyCastPermissionsForEffect reports whether ANY permission could be
// in play right now — a stored one on any seat, or a standing one
// (gated or not, #1314) from a permanent on the battlefield.
//
// A fast negative for the view and the bot enumerator, and it is not a
// micro-optimisation: without it every frame walked every graveyard,
// every library and the whole of exile asking a question whose answer
// is "no" in the overwhelming majority of games, and the cost of that
// walk is what a bot's decision loop pays. The check is one pass over
// the seats plus, only if that finds nothing, one over the
// battlefield.
//
// Deliberately does NOT evaluate a gated permission's ActiveWhen or
// Condition — this is "could anything open one of the expensive
// zones", not "does one actually apply right now", and a card whose
// gate is not yet satisfied still has to make the enumerator walk the
// zone once to find that out. Answering "no" for an unsatisfied gate
// would be the wrong direction: it would make the fast path decide the
// question the walk exists to ask.
//
// Caller must hold g.mu (read or write).
func (g *Game) AnyCastPermissionsForEffect() bool {
	for _, p := range g.Seats {
		if p != nil && len(p.CastPermissions) > 0 {
			return true
		}
	}
	// ADR 0090: a prepared permanent grants a cast of its CR 722.3c
	// copy out of exile, derived rather than stored.
	if g.anyPreparedPermanentLocked() {
		return true
	}
	if g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		key := CatalogAbilityKey(*c)
		if key == "" {
			continue
		}
		if CatalogCastPermissions != nil && len(CatalogCastPermissions(key)) > 0 {
			return true
		}
		if CatalogGatedCastPermissions != nil && len(CatalogGatedCastPermissions(key)) > 0 {
			return true
		}
	}
	return false
}

// CastPermissionOnCardForEffect is the view and client half: the
// permission (if any) that lets ANY player cast this card from the
// zone it is sitting in. Used to project the `exile_play` wire field,
// which is public information — the trigger that created it resolved
// in the open.
//
// Caller must hold g.mu (read or write).
func (g *Game) CastPermissionOnCardForEffect(card Card, zone ZoneKind) *CastPermission {
	// ADR 0090: the CR 722.3c copy's permission is derived, and it is
	// the only one that may name the copy (see CastPermissionForLocked).
	if card.PrepareCopy {
		return g.prepareCopyPermissionLocked(card, zone)
	}
	var fallback *CastPermission
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		for i := range p.CastPermissions {
			perm := &p.CastPermissions[i]
			if !perm.CoversCard(card, zone) {
				continue
			}
			if g.CastPermissionActiveForEffect(perm, p.ID) {
				out := *perm
				return &out
			}
			// A permission whose window has not OPENED yet is still
			// worth showing: warp's "you may cast it from exile on a
			// later turn" (CR 702.185a) is stamped during the end step
			// of the turn the creature was warped in, and a client
			// that could not see it until the next turn would show a
			// blank card in exile with no explanation. The client
			// reads NotBeforeSeq and greys the button.
			if fallback == nil {
				out := *perm
				fallback = &out
			}
		}
	}
	return fallback
}

// cardHasCreatureType reports whether a card in a non-battlefield
// zone has the given creature type. Realmwalker's "creature spells of
// the chosen type" is checked against a card in a LIBRARY, where the
// layer engine has nothing to say, so this reads the subtypes through
// the same CreatureTypesOf vocabulary the battlefield uses — a
// changeling in the library is every type, which is correct.
func cardHasCreatureType(c Card, want string) bool {
	for _, t := range CreatureTypesOf(&c) {
		if t == want {
			return true
		}
	}
	return false
}

// escapeExileSpec builds the "exile N other cards from your
// graveyard" component of a GRANTED escape cost (CR 702.138a).
//
// The game package's own copy of what effects.CardsInYourGraveyard
// builds for a printed escape cost, and it has to be: the catalog
// imports game, so a cost synthesised inside the engine cannot reach
// the catalog's constructors. What the two must agree about is small
// — owner, zone, count.
//
// "Other" needs no clause here: CR 601.2a has already moved the spell
// to the stack, and validateAlternativeCostPaymentLocked refuses the
// cast ID outright.
func escapeExileSpec(n int) *TargetSpec {
	label := "Exile " + countWord(n) + " other cards from your graveyard"
	return &TargetSpec{
		Mode:  "card_in_graveyard",
		Label: label,
		Zones: []ZoneKind{ZoneGraveyard},
		CardOK: func(_ *Game, caster uuid.UUID, c Card, _ ZoneKind) bool {
			return c.Owner == caster
		},
		Min: n, Max: n,
	}
}

// countWord spells a small count the way an oracle line does —
// "three other cards", not "3 other cards".
func countWord(n int) string {
	words := []string{"zero", "one", "two", "three", "four", "five",
		"six", "seven", "eight", "nine", "ten"}
	if n >= 0 && n < len(words) {
		return words[n]
	}
	return strconv.Itoa(n)
}

// asAnyColorCost rewrites a cost so every colored requirement is
// payable by any mana — "you may spend mana as though it were mana
// of any color" (Breeches). Folding the colored slots into the
// generic demand is exactly equivalent for the pool solver: a
// requirement any token can satisfy IS a generic requirement.
//
// Colorless {C} requirements are left alone. "Mana of any color"
// does not include colorless (CR 106.1b), so a {C} slot still needs
// real colorless mana.
func asAnyColorCost(cost ParsedCost) ParsedCost {
	out := cost
	out.Required = nil
	for _, req := range cost.Required {
		if requiresColorless(req) {
			out.Required = append(out.Required, req)
			continue
		}
		out.Generic++
	}
	return out
}

// asAnyTypeCost is asAnyColorCost without the {C} exception — "mana
// of any TYPE can be spent" (Hostage Taker). Every requirement,
// colorless included, folds into the generic demand, because any mana
// in the pool can now pay any symbol. A hybrid or two-for-one hybrid
// slot folds to one generic, the cheapest way to pay it when any mana
// counts as its colour. A Phyrexian slot folds too and so loses its
// "or 2 life" half — the same trade asAnyColorCost has always made,
// and weaker than printed rather than stronger.
func asAnyTypeCost(cost ParsedCost) ParsedCost {
	out := cost
	out.Required = nil
	out.Generic += len(cost.Required)
	return out
}

// spendAsThoughAny applies a permission's "spend mana as though"
// clause to the cost a cast owes: the any-type fold when the grant
// says any TYPE, the any-color fold when it says any colour, and
// nothing otherwise (nil grant included). The one reading, so the
// payment, the auto-tapper, the preview and the view cannot disagree
// about which clause a grant carries.
func spendAsThoughAny(grant *CastPermission, cost ParsedCost) ParsedCost {
	switch {
	case grant == nil:
		return cost
	case grant.AnyType:
		return asAnyTypeCost(cost)
	case grant.AnyColor:
		return asAnyColorCost(cost)
	}
	return cost
}

// requiresColorless reports whether a requirement can only be paid
// with colorless mana — i.e. every option it admits is {C}.
func requiresColorless(req ColorRequirement) bool {
	if len(req.Options) == 0 {
		return false
	}
	for _, c := range req.Options {
		if c != "C" {
			return false
		}
	}
	return true
}
