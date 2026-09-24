package game

import "github.com/google/uuid"

// cast_ban.go — #1316, ADR 0066's 2026-09-23 amendment and ADR 0085's:
// a CAST RESTRICTION created by a RESOLVING SPELL, with a CR 611.2
// duration.
//
// cast_gate.go's CastGateLocked already answers "may this player cast
// this spell at all" from two sources — a static on a permanent
// (CastRestriction) and the spell's own condition (CastConditionFor) —
// and its own doc comment names the gap this file closes: "BANS WITH A
// DURATION. Silence's 'this turn' and Reflector Mage's 'until your
// next turn' want the turn-scoped and permanent-duration registries...
// A third source slots into castRestrictionsLocked without changing
// this function's signature; that is the extension point."
//
// A permanent's printed ban needs no duration: the source's continued
// presence on the battlefield already IS the duration, which is
// exactly what CastRestriction already gives it. A RESOLVED SPELL's
// ban has no such source a moment later — Avatar's Wrath and Mandate
// of Peace both exile themselves — so the statement has to be STORED,
// against a player, for a stated window. That is the shape #1195 built
// for a per-player cast-TIMING statement (CastTimingRule) and #1200
// built for "your life total can't change" (LifeTotalLocked): a fourth
// payload on PlayerStatic, told apart from the other three by its own
// presence bit rather than by a discriminator field.
//
// ONE READ, THREE CALLERS FOR FREE. Unlike CastTimingRule and
// LifeTotalLocked, a cast ban has no read of its own to add at three
// call sites — CastGateLocked already IS that one read, and it already
// has exactly three callers (CastSpell, legal.castMovesForCard,
// protocol.stampLegalTargets/castStampsFor). Extending CastGateLocked's
// insides is extending all three at once, and the wire needs no new
// field either: `cant_cast` already means "an effect prevents this
// cast", stamped from CastGateLocked's own error, so a card silenced by
// Avatar's Wrath greys exactly the way one silenced by Rule of Law
// already does.

// CastBanKind is what a granted "can't cast" statement says. The zero
// value, CastBanNone, says nothing — the presence bit CastBanRule
// needs because its own zero value ("no exception, no count") would
// otherwise BE Mandate of Peace's outright ban, which is a real
// statement and not "no statement at all". The same problem
// GrantTiming's TimingNormal solves for CastTimingRule.
type CastBanKind string

const (
	// CastBanNone is "no ban" — the zero value, so a PlayerStatic
	// carrying an unrelated grant (a keyword, a timing statement, a
	// life-total lock) is invisible to castBanForbidsLocked.
	CastBanNone CastBanKind = ""

	// CastBanOutright is "can't cast spells", optionally narrowed to
	// every zone but one — Mandate of Peace's "your opponents can't
	// cast spells this turn" (no exception) and Avatar's Wrath's "your
	// opponents can't cast spells from anywhere other than their
	// hands" (ExceptFromZone: ZoneHand).
	CastBanOutright CastBanKind = "outright"

	// CastBanMaxPerTurn is "can't cast more than N spells", granted for
	// a duration rather than printed on a permanent — the shape Rule of
	// Law's clause would take if a resolving spell rather than a static
	// ability ever said it (#1316's third named case; no catalogued
	// card uses it yet). Counted off the same per-turn tally
	// EachPlayerMaxSpellsPerTurn reads (Game.CastTallyFor), so a
	// granted cap and a printed one can never disagree about what "one
	// spell this turn" means.
	CastBanMaxPerTurn CastBanKind = "max_per_turn"
)

// CastBanRule is the RULES half of a granted per-player "can't cast"
// statement (CR 101.2's "can't", CR 611.2's duration). It carries no
// player, no source and no duration — PlayerStatic already has those,
// exactly as CastTimingRule carries none of its own for the same
// reason (player_statics.go).
//
// Plain data, like every other payload on that slice: Clone copies it
// by value and the snapshot mirrors it rather than rebuilding it.
type CastBanRule struct {
	// Kind is the presence bit; see CastBanNone's comment.
	Kind CastBanKind `json:"kind,omitempty"`

	// Filter narrows which SPELLS the ban covers. The zero filter is
	// "spells", which is both proof cards on this seam; no catalogued
	// card narrows a granted ban by type yet, but the field costs
	// nothing to have (PermissionFilter's zero value already matches
	// everything, so an unset Filter is not a second thing to test).
	//
	// No `omitzero`: this is part of GameSnapshot's serialization
	// graph (PlayerStatic.CastBan.Filter), and that option's
	// behaviour depends on the building Go toolchain below 1.24
	// (#1492) — CI is pinned to 1.22.
	Filter PermissionFilter `json:"filter"`

	// ExceptFromZone is the one zone this ban does NOT reach — Avatar's
	// Wrath's "from anywhere other than their hands" is
	// ExceptFromZone: ZoneHand. Zero means the ban reaches every zone,
	// which is Mandate of Peace and every card that says only "can't
	// cast spells" with no zone clause at all.
	ExceptFromZone ZoneKind `json:"exceptFromZone,omitempty"`

	// MaxPerTurn is CastBanMaxPerTurn's printed number. Ignored for
	// CastBanOutright, where zero spells is not what either proof card
	// says — an outright ban is "always forbids", not "forbids at
	// zero", which is why the two need different Kinds rather than one
	// letting MaxPerTurn's zero value stand for "no count".
	MaxPerTurn int `json:"maxPerTurn,omitempty"`
}

// forbids reports whether this rule bans `playerID` from casting
// `card` out of `zone` right now. The rule's own predicate, called
// once per live entry by castBanForbidsLocked — CR 611.2's duration is
// tested by the caller, not here, exactly as CastTimingRule.CoversCast
// leaves its own duration to castTimingVerdictLocked.
func (r CastBanRule) forbids(g *Game, playerID uuid.UUID, card Card, zone ZoneKind) bool {
	switch r.Kind {
	case CastBanOutright:
		if !r.Filter.Matches(card) {
			return false
		}
		return r.ExceptFromZone == "" || zone != r.ExceptFromZone
	case CastBanMaxPerTurn:
		if !r.Filter.Matches(card) {
			return false
		}
		return g.CastTallyFor(playerID).Total >= r.MaxPerTurn
	default:
		return false
	}
}

// GrantCastBanForEffect stores a per-player cast ban for a duration —
// "until your next turn, your opponents can't cast spells from
// anywhere other than their hands" (Avatar's Wrath), "your opponents
// can't cast spells this turn" (Mandate of Peace). The *ForEffect
// surface: caller must hold g.mu (write), which a resolution frame
// already does.
//
// It writes a PlayerStatic carrying no Keyword, no Timing statement
// and no LifeTotalLocked bit — the fourth kind of entry on that slice,
// told apart from the other three by CastBanRule.Kind, exactly as
// Timing is told apart by TimingNormal and LifeTotalLocked by its own
// plain bool.
//
// `rule.Kind == CastBanNone` is a no-op, so a card file that computed
// its rule and got nothing bans nothing rather than banning "".
//
// Build the duration with the constructors in duration.go —
// DurationUntilYourNextTurn for Avatar's Wrath,
// g.UntilEndOfTurnDuration() for Mandate of Peace (also the zero-value
// default, so a card file that forgets the window gets the narrowest
// real one rather than a permanent ban — the same direction every
// other *ForEffect default in this file errs in).
func (g *Game) GrantCastBanForEffect(player uuid.UUID, rule CastBanRule, label string, source uuid.UUID, d Duration) {
	if player == uuid.Nil || rule.Kind == CastBanNone {
		return
	}
	p := g.playerByIDLocked(player)
	if p == nil {
		return
	}
	if d == (Duration{}) {
		d = g.UntilEndOfTurnDuration()
	}
	p.Statics = append(p.Statics, PlayerStatic{
		CastBan:  rule,
		Source:   source,
		Label:    label,
		Duration: d,
	})
}

// castBanForbidsLocked is THE reader: does any live granted ban stop
// `playerID` from casting `card` out of `zone` right now? Returns the
// PlayerStatic's own Label and Source for CantCastError's attribution
// — the same two fields a battlefield CastRestriction's Label and
// Source populate it with, so a refused cast names the card that said
// no whichever of the two sources answered.
//
// THE DURATION IS TESTED HERE, not folded into a separate liveness
// check, for the reason playerAbilityTokensLocked and
// castTimingVerdictLocked both give: the sweep is hygiene run at known
// moments, and this read has to be right between them — a ban that
// ends as your next turn begins must not still answer during the
// priority round that ends the previous turn.
//
// The first entry that forbids wins; CantCastError carries one reason,
// and CR 101.2 does not ask which "can't" arrived first among several.
//
// Caller must hold g.mu (read or write).
func (g *Game) castBanForbidsLocked(playerID uuid.UUID, card Card, zone ZoneKind) (label string, source uuid.UUID, forbidden bool) {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return "", uuid.Nil, false
	}
	for _, s := range p.Statics {
		if s.CastBan.Kind == CastBanNone {
			continue
		}
		if g.durationExpiredLocked(s.Duration, false) {
			continue
		}
		if s.CastBan.forbids(g, playerID, card, zone) {
			return s.Label, s.Source, true
		}
	}
	return "", uuid.Nil, false
}

// anyLiveCastBanForEffect reports whether ANY seat is carrying a live
// granted cast ban right now — the fast negative
// AnyCastRestrictionsForEffect takes before it has to know WHICH
// player or WHICH card, mirrored from that function's own battlefield
// walk one registry over.
//
// Caller must hold g.mu (read or write).
func (g *Game) anyLiveCastBanForEffect() bool {
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		for _, s := range p.Statics {
			if s.CastBan.Kind == CastBanNone {
				continue
			}
			if !g.durationExpiredLocked(s.Duration, false) {
				return true
			}
		}
	}
	return false
}
