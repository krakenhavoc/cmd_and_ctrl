package game

import (
	"github.com/google/uuid"
)

// cast_timing.go — S42, #1195, ADR 0066's 2026-09-22 amendment: the
// per-PLAYER half of CR 307.1.
//
// ADR 0066 Decision 6 put a timing rule on a PERMISSION —
// `CastPermission.Timing`, read next to `HasKeyword(&card, "flash")`
// in `CastSpell` — and that is the right shape for a grant that opens
// one cast of one object. Madness (#657) and suspend's free cast
// (#659) are both exactly that: a named card, in a named zone, for a
// stated window.
//
// It is the wrong shape for the sentence Vedalken Orrery prints. "You
// may cast spells as though they had flash" names no card, names no
// zone and outlives no particular object; it is a statement about a
// PLAYER. So is its inverse, which the same seam blocked from the
// other side: "each opponent can cast spells only any time they could
// cast a sorcery" (Teferi, Time Raveler).
//
// THREE VOCABULARIES, NONE OF THEM NEW, and that is the point:
//
//   - `GrantTiming` is ADR 0066's own enum, gaining one value.
//     `TimingYourTurnOnly` is Dosan the Falling Leaf's "players can
//     cast spells only during their own turns", which is NOT
//     `TimingSorcery` — Dosan leaves you every instant-speed window
//     on your own turn and takes away the rest.
//   - `PermissionFilter` is the zone filter a standing permission
//     already narrows with, gaining `NoncreatureOnly` and
//     `SorceryOnly`. "You may cast CREATURE spells as though they had
//     flash" (Yeva) is the same predicate a standing permission
//     writes as `CreatureOnly`, and a second flag struct would be two
//     spellings of one rule.
//   - `Duration` is ADR 0063's, untouched. Emergence Zone is
//     `UntilEndOfTurn`, Teferi's +1 is `UntilYourNextTurn`, a derived
//     statement is `WhileInZone` — swept by the one
//     `durationExpiredLocked` every other continuous effect uses.
//
// TWO HOMES, the same two ADR 0066 gives a cast permission:
//
//  1. DERIVED, NEVER STORED, for a statement whose duration is a
//     permanent's presence on the battlefield (Vedalken Orrery,
//     Leyline of Anticipation, Yeva, Teferi's static). Declared on
//     `Spec.CastTimings`, read per query through `CatalogAbilityKey`
//     so a source under a CR 613.1f ability-removing effect stops
//     saying it, two Orreries compose, and one leaving cannot revoke
//     the other's grant.
//  2. STORED ON THE PLAYER, for a statement that outlives its source:
//     Emergence Zone sacrifices itself and the permission lasts the
//     turn; Teferi's +1 resolves and the planeswalker may die before
//     your next turn. Swept at the two moments
//     `sweepCastPermissionsLocked` runs, by the same expiry function.
//
// ONE READ: `CastTimingOpenLocked`, called by `CastSpell`, by the bot
// enumerator and by the view's `castable_here` derivation, so the
// three cannot disagree — the pattern `special_action.go`'s timing
// table already uses for foretell and suspend.

// CastTimingAffects says WHOM a declared timing statement is about.
//
// It exists because a catalog entry is static and cannot name a seat
// — the same reason `standingCastPermissionsLocked` zeroes
// `ZoneOwner` on the way out. `castTimingAffects` expands it against
// the source's controller and the player asking, which is the whole
// of what it means and the one place that decides; a STORED
// statement names its player outright and carries the zero value.
type CastTimingAffects string

const (
	// TimingAffectsYou is the source's controller — "YOU may cast
	// spells as though they had flash" (Vedalken Orrery, Leyline of
	// Anticipation, Yeva, Nature's Herald). The zero value, because
	// it is what a stored grant always means.
	TimingAffectsYou CastTimingAffects = ""

	// TimingAffectsEachOpponent is "EACH OPPONENT can cast spells
	// only any time they could cast a sorcery" (Teferi, Time
	// Raveler; Teferi, Mage of Zhalfir).
	TimingAffectsEachOpponent CastTimingAffects = "each_opponent"

	// TimingAffectsEachPlayer is "PLAYERS can cast spells only
	// during their own turns" (Dosan the Falling Leaf) — the source's
	// controller included.
	TimingAffectsEachPlayer CastTimingAffects = "each_player"
)

// CastTimingRule is the RULES half of a per-player timing statement:
// what it says, about which spells, out of which zone, about whom
// (CR 307.1, CR 702.8). It carries no player, no source and no
// duration, because both of its homes already have those:
//
//   - DECLARED on a permanent — `Spec.CastTimings`, read per query
//     through `CatalogAbilityKey`. The player is whoever `Affects`
//     names, the source is the permanent still on the battlefield,
//     and the duration is its presence there.
//   - STORED on a player — a `PlayerStatic` (player_statics.go,
//     #1197), whose `Player`, `Source`, `Label` and `Duration` are
//     this statement's. #1195 shipped its own `Player.CastTimings`
//     slice for a day; folding it onto `Statics` is what keeps the
//     engine from carrying TWO player-level registries, two sweeps
//     and two readings of one `Duration`.
//
// Pure data, no closures, to the standard `PlayerStatic` is held to
// for the same reasons: `Clone` copies it by value and the snapshot
// mirrors it rather than rebuilding it.
type CastTimingRule struct {
	// Timing is what the statement says. `TimingFlash` GRANTS
	// instant speed; `TimingSorcery` and `TimingYourTurnOnly`
	// RESTRICT. `TimingNormal` — the zero value — says nothing, and
	// is what makes a `PlayerStatic` carrying no rule (a plain
	// hexproof grant) invisible to this read.
	Timing GrantTiming `json:"timing,omitempty"`

	// Filter narrows which SPELLS the statement is about. The zero
	// filter is "spells", which is Vedalken Orrery and Teferi's
	// static; `CreatureOnly` is Yeva, `SorceryOnly` is Teferi's +1.
	//
	// `LandsOnly` and `NonLandOnly` are meaningless here and are not
	// rejected: a land PLAY is not a cast (CR 305.1) and never
	// reaches this read at all, so a filter that named one would
	// simply match nothing a cast can be.
	Filter PermissionFilter `json:"filter,omitzero"`

	// FromZone narrows the statement to casts out of one zone. Zero
	// (`ZoneKind("")`) is "from anywhere", which is every card on the
	// seam row today; the field exists because "you may cast spells
	// from your graveyard as though they had flash" is one sentence
	// away and would otherwise be a second read.
	FromZone ZoneKind `json:"fromZone,omitempty"`

	// Affects is whom a DECLARATION is about, expanded against the
	// battlefield by castTimingAffects. Declaration-only: a stored
	// statement is granted to one player by name, so its
	// `PlayerStatic.Player` is the answer and this stays zero.
	Affects CastTimingAffects `json:"affects,omitempty"`

	// Label is the printed clause, and is REQUIRED on a DECLARATION
	// (effects.Register panics at boot without it) because it is what
	// the next reader of the card file matches against the oracle
	// text. Declaration-only, like Affects: a stored statement's
	// clause is its `PlayerStatic.Label`.
	Label string `json:"label,omitempty"`
}

// Grants reports whether this statement OPENS the instant-speed
// window (CR 702.8).
func (t CastTimingRule) Grants() bool { return t.Timing == TimingFlash }

// Restricts reports whether this statement NARROWS the window
// (CR 101.2's "can't" half).
func (t CastTimingRule) Restricts() bool {
	return t.Timing == TimingSorcery || t.Timing == TimingYourTurnOnly
}

// CoversCast reports whether this statement is about a cast of `card`
// out of `zone`. Says nothing about WHOSE cast it is — that is
// `Affects` for a declaration and `PlayerStatic.Player` for a stored
// one, both checked by the walk.
func (t CastTimingRule) CoversCast(card Card, zone ZoneKind) bool {
	if t.FromZone != "" && t.FromZone != zone {
		return false
	}
	return t.Filter.Matches(card)
}

// CatalogCastTimings is the catalog hook the effects package wires at
// init, mirroring CatalogCastPermissions and CatalogCastRestrictions.
// Nil, or a nil return, means the card says nothing about anybody's
// cast timing — which is nearly every card.
var CatalogCastTimings func(oracleID string) []CastTimingRule

// GrantCastTimingForEffect stores a per-player timing statement — the
// half of the model that outlives its source (Emergence Zone's "this
// turn", Teferi's "until your next turn").
//
// It writes a PlayerStatic (#1197), NOT a registry of its own. That
// is the whole of the fold: a granted timing statement and a granted
// "you have hexproof" are the same kind of thing — an ability a
// PLAYER has for a CR 611.2 duration — so they share one slice, one
// sweep (sweepPlayerStaticsLocked), one durationExpiredLocked read,
// one clone and one snapshot field. The entry carries no Keyword,
// which is what keeps it out of playerAbilityTokensLocked's answer.
//
// A statement whose duration is a permanent's presence does NOT come
// through here: it is declared on the card and derived per query, for
// the reason the file header gives.
//
// An unstamped Duration is stamped "until end of turn" on the way in,
// exactly as GrantCastPermissionForEffect does, so a card file that
// forgets the window gets the narrowest real one rather than a
// permanent grant.
//
// Caller must hold g.mu (write).
func (g *Game) GrantCastTimingForEffect(player uuid.UUID, rule CastTimingRule, label string, source uuid.UUID, d Duration) {
	if player == uuid.Nil || rule.Timing == TimingNormal {
		return
	}
	p := g.playerByIDLocked(player)
	if p == nil {
		return
	}
	if d == (Duration{}) {
		d = g.UntilEndOfTurnDuration()
	}
	// Declaration-only fields are cleared on the way in: a stored
	// statement's player is named beside it and its clause is the
	// PlayerStatic's, so leaving either here would be two places one
	// fact could disagree with itself.
	rule.Affects = TimingAffectsYou
	rule.Label = ""
	p.Statics = append(p.Statics, PlayerStatic{
		Timing:   rule,
		Source:   source,
		Label:    label,
		Duration: d,
	})
}

// castTimingVerdict is what every statement about one cast folds
// down to: the three bits the window needs, and no record of which
// card said what. Built by castTimingVerdictLocked and consumed
// immediately, so it allocates nothing.
type castTimingVerdict struct {
	// Flash is set by any live TimingFlash statement covering this
	// cast.
	Flash bool

	// SorceryOnly is set by any live TimingSorcery statement.
	SorceryOnly bool

	// YourTurnOnly is set by any live TimingYourTurnOnly statement.
	YourTurnOnly bool
}

func (v *castTimingVerdict) fold(t CastTimingRule, card Card, zone ZoneKind) {
	// A statement that says nothing is skipped before the filter runs
	// — a half-written card file costs a comparison, not a card walk.
	if !t.Grants() && !t.Restricts() {
		return
	}
	if !t.CoversCast(card, zone) {
		return
	}
	switch t.Timing {
	case TimingFlash:
		v.Flash = true
	case TimingSorcery:
		v.SorceryOnly = true
	case TimingYourTurnOnly:
		v.YourTurnOnly = true
	}
}

// castTimingVerdictLocked folds every per-player statement that
// speaks about this player's cast of this card out of this zone.
//
// Stored first, then derived, and the ORDER DOES NOT MATTER: the
// verdict is three independent bits and CR 101.2's precedence is
// applied by the reader, not by the walk. That is deliberate — a
// precedence that depended on which permanent the battlefield scan
// reached first would be a rule nobody could state.
//
// Caller must hold g.mu (read or write).
func (g *Game) castTimingVerdictLocked(playerID uuid.UUID, card Card, zone ZoneKind) castTimingVerdict {
	var v castTimingVerdict
	if p := g.playerByIDLocked(playerID); p != nil {
		// The STORED half, riding #1197's one player-level registry.
		// An entry with no rule is somebody's hexproof grant and is
		// skipped before its duration is read; the duration test is
		// the same one playerAbilityTokensLocked applies, and for the
		// same reason — the sweep is hygiene at known moments and
		// this read has to be right between them.
		for _, st := range p.Statics {
			if st.Timing.Timing == TimingNormal {
				continue
			}
			if g.durationExpiredLocked(st.Duration, false) {
				continue
			}
			v.fold(st.Timing, card, zone)
		}
	}
	if g.Battlefield == nil || CatalogCastTimings == nil {
		return v
	}
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		// CatalogAbilityKey, not CatalogKey, for the reason
		// standingCastPermissionsLocked gives: "you may cast spells
		// as though they had flash" is a static ability, and an
		// Orrery that has lost its abilities (CR 613.1f) stops
		// saying it. The EMPTY KEY is the skip — a TOKEN has one of
		// its own since #521 (ADR 0083 decision 3).
		key := CatalogAbilityKey(*src)
		if key == "" {
			continue
		}
		for _, t := range CatalogCastTimings(key) {
			if !castTimingAffects(t.Affects, src.Controller, playerID) {
				continue
			}
			v.fold(t, card, zone)
		}
	}
	return v
}

// castTimingAffects expands a DECLARATION's scope against one source
// controller and one caster. The whole of what CastTimingAffects
// means, in one place, so the engine and a future view stamp cannot
// read it two ways.
func castTimingAffects(a CastTimingAffects, controller, caster uuid.UUID) bool {
	switch a {
	case TimingAffectsEachOpponent:
		return controller != uuid.Nil && caster != controller
	case TimingAffectsEachPlayer:
		return true
	default:
		return caster == controller
	}
}

// CastTimingOpenLocked is CR 307.1 asked once: may `playerID` BEGIN
// to cast `card` out of `zone` right now?
//
// THREE CALLERS, ONE FUNCTION, the same arrangement ADR 0073 §7 makes
// for the cast gate and for the same reason:
//
//   - CastSpell, where it replaced an inline `requiresSorcerySpeed`
//     computation;
//   - legal.castMovesPayingOptional, so a bot is never offered a cast
//     the engine will refuse for timing, and never denied one an
//     Orrery opens;
//   - protocol.castStampsFor, where it is the third input to
//     `castable_here` (#1015 had two: a claimable price and the cast
//     gate).
//
// FOUR STEPS, AND THE ORDER IS THE RULES:
//
//  1. The CARD. An instant, or a card with flash (CR 702.8), is
//     instant-speed on its own.
//  2. The PERMISSION (ADR 0066 Decision 6, unchanged). TimingFlash
//     opens the window, TimingSorcery shuts it — the branch madness
//     and suspend's free cast reach.
//  3. The per-player GRANTS. Vedalken Orrery, Leyline of
//     Anticipation, Yeva, Emergence Zone, Teferi's +1.
//  4. The per-player RESTRICTIONS, LAST, because CR 101.2 says
//     "can't" beats "can". An opponent's Orrery does not get them
//     past your Teferi, and that falls out of the placement rather
//     than needing a rule of its own.
//
// A LAND PLAY IS NOT HERE. CR 305.1 and CR 116.2a make playing a land
// a special action rather than a cast, and every card this function
// exists for writes about casting SPELLS. CastSpell's land branch
// keeps its own sorcerySpeedOpenLocked check beside this call — the
// same split cast_gate.go documents, for the same reason: a Dosan
// that stopped a land play would be a rule nobody printed.
//
// `perm` may be nil, which is every cast out of a hand.
//
// Caller must hold g.mu (read or write).
func (g *Game) CastTimingOpenLocked(playerID uuid.UUID, card Card, zone ZoneKind, perm *CastPermission) bool {
	// 0. A plotted card (CR 702.170d, #1318) is cast in its owner's
	// main phase with the stack empty and at no other time: the window
	// belongs to the permission, so neither the card's own flash nor a
	// per-player grant opens it wider. The per-player restrictions
	// below cannot narrow it either — Dosan's "only during your turn"
	// and Teferi's "only as a sorcery" are both already true of it.
	if perm != nil && perm.Timing == TimingPlot {
		return g.sorcerySpeedOpenLocked(playerID)
	}
	// 1. The card's own timing.
	instantSpeed := card.IsInstant() || HasKeyword(&card, "flash")
	// 2. The permission's override, if it carries one.
	if perm != nil {
		switch perm.Timing {
		case TimingFlash:
			instantSpeed = true
		case TimingSorcery:
			instantSpeed = false
		}
	}
	v := g.castTimingVerdictLocked(playerID, card, zone)
	// 3. The grants.
	if v.Flash {
		instantSpeed = true
	}
	// 4. The restrictions. "Only during your own turn" refuses
	// outright rather than narrowing to the sorcery window: Dosan
	// leaves you every instant-speed window on your own turn, so it
	// is a gate on WHOSE turn it is and nothing else.
	if v.YourTurnOnly && g.activeSeatIDLocked() != playerID {
		return false
	}
	if v.SorceryOnly {
		instantSpeed = false
	}
	if instantSpeed {
		return true
	}
	return g.sorcerySpeedOpenLocked(playerID)
}
