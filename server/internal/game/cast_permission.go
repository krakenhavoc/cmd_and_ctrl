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
	// graveyard permissions print.
	TimingSorcery GrantTiming = "sorcery"
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
	// library. Never hand (CR 601.1 already allows it) and never the
	// command zone (CR 903.4 is the format's, not an effect's).
	Zone ZoneKind `json:"zone"`

	// Scope and its two payloads. Cards is read for ScopeCards,
	// Filter for ScopeStanding; the other is ignored rather than
	// asserted, because a permission is data and a half-filled one
	// should grant less, never panic.
	Scope  PermissionScope     `json:"scope,omitempty"`
	Cards  []PermissionCardRef `json:"cards,omitempty"`
	Filter PermissionFilter    `json:"filter,omitzero"`

	// TopOfLibraryOnly restricts a ZoneLibrary permission to the card
	// currently on top (CR 401.5). Every library permission sets it;
	// the field exists rather than being implied by the zone so that
	// a future "play any card from your library" reads as the
	// exception it would be.
	TopOfLibraryOnly bool `json:"topOfLibraryOnly,omitempty"`

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
	// mana value rather than pay its mana cost". A COST (CR 118.4,
	// 119.4), so a player without the life cannot claim it at all —
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
	Duration Duration `json:"duration,omitzero"`

	// NotBeforeTurn is the earliest turn NUMBER the permission is
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
	// Turn.Number is a ROUND counter, so this reads as "not before
	// round N" rather than "not on the turn it was granted". Every
	// printed warp card is cast at sorcery speed on its controller's
	// own turn, where the two agree; see #945's PR for the case where
	// they do not.
	NotBeforeTurn int `json:"notBeforeTurn,omitempty"`

	// --- what, and when --------------------------------------------

	// Timing is the timing rule the cast obeys. See GrantTiming: the
	// zero value changes nothing, and a grant never opens the
	// sorcery-speed gate unless it says so.
	Timing GrantTiming `json:"timing,omitempty"`

	// GrantsHaste gives the permanent this cast produces haste —
	// suspend's CR 702.62e. A property of the PERMISSION rather than
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
	// window too (see NotBeforeTurn).
	if p.NotBeforeTurn > 0 && g.Turn.Number < p.NotBeforeTurn {
		return false
	}
	// `false`: this is a query, not the cleanup sweep. An
	// UntilEndOfTurn permission is live for the whole of the turn it
	// names and is dropped by sweepCastPermissionsLocked at that
	// turn's cleanup step — the same split ScopedStatic lives under.
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
// It returns nil for hand and the command zone (CR 601.1 and CR 903.4
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
	// Stored permissions first: a named object beats a standing rule,
	// because the named one is the narrower statement and is the only
	// one that can carry a face or a per-instance price.
	for i := range p.CastPermissions {
		perm := &p.CastPermissions[i]
		if !g.CastPermissionActiveForEffect(perm, playerID) || !perm.CoversCard(card, zone) {
			continue
		}
		if !g.permissionPositionOKLocked(p, perm, card) {
			continue
		}
		out := *perm
		return &out
	}
	for _, perm := range g.standingCastPermissionsLocked(p) {
		if !perm.CoversCard(card, zone) {
			continue
		}
		if !g.permissionPositionOKLocked(p, &perm, card) {
			continue
		}
		out := perm
		return &out
	}
	return nil
}

// permissionPositionOKLocked enforces the restrictions that are about
// WHERE in the zone the card sits rather than about the card: CR
// 401.5's "the top card of your library", and the visibility that
// makes it playable at all.
//
// A player may only play the top card of their library if they can
// SEE it — a permission that opens the top card without the matching
// "you may look" or "revealed" clause opens nothing. Every printed
// card carries both halves, so this is a guard against a card file
// that declares one and forgets the other.
//
// Caller must hold g.mu.
func (g *Game) permissionPositionOKLocked(p *Player, perm *CastPermission, card Card) bool {
	if !perm.TopOfLibraryOnly {
		return true
	}
	if p.Library == nil || len(p.Library.Cards) == 0 {
		return false
	}
	if p.Library.Cards[len(p.Library.Cards)-1].InstanceID != card.InstanceID {
		return false
	}
	return g.LibraryTopVisibilityLocked(p.ID) != LibraryTopHidden
}

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
// Caller must hold g.mu.
func (g *Game) standingCastPermissionsLocked(p *Player) []CastPermission {
	if p == nil || g.Battlefield == nil || CatalogCastPermissions == nil {
		return nil
	}
	var out []CastPermission
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != p.ID || c.OracleID == "" {
			continue
		}
		for _, perm := range CatalogCastPermissions(CatalogAbilityKey(*c)) {
			perm.Player = p.ID
			perm.Scope = ScopeStanding
			perm.Source = c.InstanceID
			if perm.SourceName == "" {
				perm.SourceName = c.Name
			}
			// A permission unbounded in turns, because its duration is
			// the source's presence and this slice is rebuilt from the
			// battlefield on every query. CR 611.2b's "for as long
			// as", spelled the way a derived permission can afford to
			// spell it: nothing stored, so nothing to expire.
			perm.Duration = WhileInZoneDuration()
			if perm.Filter.FromChosenType {
				// Realmwalker: "the chosen type". A permanent whose
				// entry choice has not been answered yet names no
				// type and grants nothing (CR 614.12).
				if c.NamedTribe == "" {
					continue
				}
				perm.Filter.CreatureType = c.NamedTribe
			}
			out = append(out, perm)
		}
	}
	return out
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
// zone wearing the epoch it was granted at. Caller must hold g.mu.
func (g *Game) anyNamedObjectStillThereLocked(perm CastPermission) bool {
	zone := g.permissionZoneLocked(perm.Player, perm.Zone)
	if zone == nil {
		return false
	}
	for i := range zone.Cards {
		if perm.NamesCard(zone.Cards[i]) {
			return true
		}
	}
	return false
}

// permissionZoneLocked resolves a permission's zone to the pile it
// names: the granting player's own graveyard or library, or the
// shared exile. Caller must hold g.mu.
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
// in play right now — a stored one on any seat, or a standing one from
// a permanent on the battlefield.
//
// A fast negative for the view and the bot enumerator, and it is not a
// micro-optimisation: without it every frame walked every graveyard,
// every library and the whole of exile asking a question whose answer
// is "no" in the overwhelming majority of games, and the cost of that
// walk is what a bot's decision loop pays. The check is one pass over
// the seats plus, only if that finds nothing, one over the
// battlefield.
//
// Caller must hold g.mu (read or write).
func (g *Game) AnyCastPermissionsForEffect() bool {
	for _, p := range g.Seats {
		if p != nil && len(p.CastPermissions) > 0 {
			return true
		}
	}
	if g.Battlefield == nil || CatalogCastPermissions == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.OracleID == "" {
			continue
		}
		if len(CatalogCastPermissions(CatalogAbilityKey(*c))) > 0 {
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
			// reads NotBeforeTurn and greys the button.
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
