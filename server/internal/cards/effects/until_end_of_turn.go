package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// until_end_of_turn.go — the card-facing primitives for S32's
// turn-scoped continuous effects (CR 611.2, CR 514.2). ADR 0035 has
// the design, and ADR 0063 gave the registry the other three CR 611.2
// durations.
//
// Since ADR 0041 phase 3 tier 3a (#1497) every one of them registers a
// DATA record — a `game.ScopedEffect` over the closed `game.*Mod`
// vocabulary (server/internal/game/scoped_effects.go) — so a table
// holding a Giant Growth is still a restore point. Their names and
// fields did not change, so no card calling them did either.
//
// Two primitives for the two shapes most cards want:
//
//	BoostUntilEOT         "gets +X/+Y until end of turn"     layer 7c
//	GrantKeywordUntilEOT  "gains <keyword> until end of turn" layer 6
//
// Anything else — another duration, another operation, several
// operations at one timestamp — is `ScopedEffectFor` (scoped_effect.go)
// with the mods spelled out.
//
// Each takes EITHER a pinned `Target` instance ID (Giant Growth,
// a loyalty ability's "target creature") OR a `Match` predicate
// evaluated once across the battlefield (Overrun's "creatures you
// control"). A card that does both a pump and a grant — Overrun,
// The Wandering Emperor's -2 — applies two primitives; they are two
// records with two timestamps, in different layers.
//
// WHY THE AFFECTED SET IS SNAPSHOTTED. CR 611.2c: a one-shot
// continuous effect from a resolving spell affects only the
// permanents that were on the battlefield as it resolved. Overrun
// does not pump the creature you cast after it. So the primitives
// resolve `Match` once, at registration, into a fixed set of
// instance IDs rather than re-evaluating the predicate on every
// recompute pass — which is also what makes the predicate cheap.
//
// A battlefield static (Glorious Anthem) is the opposite and stays
// that way: its AppliesTo genuinely re-runs every pass, because a
// creature that enters under an anthem does get +1/+1.

// eotAffected is the snapshot of permanents a turn-scoped effect
// locks onto: instance ID → the `EnteredBattlefieldAt` stamp that
// instance carried when the effect was created.
//
// The timestamp is half of the key, not decoration. CR 400.7: a
// permanent that leaves the battlefield and returns is a NEW
// object, and effects that were affecting the old one stop. Instance
// IDs here persist across zone changes, so an ID-only set would keep
// pumping a creature that was flickered out and back in response to
// the spell. The battlefield-entry stamp is re-minted on every
// entry, so comparing it catches exactly that case.
type eotAffected map[uuid.UUID]int64

// eotSnapshot resolves a primitive's target clause into an
// `eotAffected` set against the current battlefield. Exactly one of
// `target` / `match` is meaningful: a non-nil `match` wins and
// `target` is ignored.
//
// Returns nil when nothing matches — the caller then registers
// nothing at all rather than a dead entry the recompute would walk
// for the rest of the turn.
//
// Caller must be inside the resolution frame (holds g.mu write).
func eotSnapshot(ctx *Context, target uuid.UUID, match CardPredicate) eotAffected {
	// #1432: "this creature gets +1/+1" pinned to a source that left
	// and came back pins nothing — the new object is not "this". A
	// Match selection is a set, not "this", and is not asked.
	if match == nil && ctx.isNewSourceObject(target) {
		return nil
	}
	out := eotAffected{}
	caster := ctx.Controller()
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if match != nil {
			if !match(ctx.Game, caster, c) {
				continue
			}
		} else if c.InstanceID != target {
			continue
		}
		out[c.InstanceID] = c.EnteredBattlefieldAt
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// BoostUntilEOT is "target creature gets +X/+Y until end of turn"
// (Giant Growth) or "creatures you control get +X/+Y until end of
// turn" (Overrun). Layer 7c — it MODIFIES power and toughness, so
// it composes additively with anthems and is overwritten by nothing
// except a later layer (a 7b "base P/T becomes N/N" applies first
// and this still adds on top, which is the printed behaviour).
//
// Negative values are legal and are how a shrink effect is written
// ("target creature gets -2/-0"). The layer engine does not clamp;
// CurrentPower does, so a creature pumped below zero deals no
// damage rather than negative damage, and CurrentToughness stays
// unclamped so the 0-toughness SBA can see it.
type BoostUntilEOT struct {
	// Target pins the effect to one permanent. Ignored when Match
	// is set.
	Target uuid.UUID

	// Match selects the affected permanents from the battlefield,
	// evaluated ONCE at resolution (CR 611.2c). "Creatures you
	// control" is And(Creature(), YouControl()).
	Match CardPredicate

	Power     int
	Toughness int

	// Label is attribution for logs and tests; defaults to a
	// generic string when empty.
	Label string
}

func (b BoostUntilEOT) Apply(ctx *Context) error {
	if b.Power == 0 && b.Toughness == 0 {
		return nil
	}
	return untilEndOfTurn(ctx, b.Target, b.Match,
		eotLabel(b.Label, "pump until end of turn"),
		game.ModifyPTMod(b.Power, b.Toughness))
}

// untilEndOfTurn is the one body every until-end-of-turn primitive
// shares: `mods` over the set `target` / `match` names, locked now
// (CR 611.2c), ending in this turn's cleanup step (CR 514.2).
func untilEndOfTurn(ctx *Context, target uuid.UUID, match CardPredicate, label string, mods ...game.Mod) error {
	return ScopedEffectFor{
		Target:   target,
		Match:    match,
		Mods:     mods,
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    label,
	}.Apply(ctx)
}

// GrantKeywordUntilEOT is "target creature gains <keyword> until end
// of turn" (The Wandering Emperor's -2 lifelink half) or the mass
// form (Overrun's trample). Layer 6 — ability-adding.
//
// WHICH KEYWORDS ACTUALLY DO ANYTHING. The engine honours fifteen
// tokens — the twelve combat keywords (flying, reach, first strike,
// double strike, deathtouch, lifelink, trample, vigilance, menace,
// defender, haste, flash), the S23 targeting pair (hexproof,
// shroud), and S25's indestructible. See `canonicalKeywords` in
// game/keywords.go, which is the authoritative list. Granting any of
// those is real.
//
// This is also the answer to the S25 (#77) checklist items that
// named `HexproofUntilEOT` and `IndestructibleUntilEOT` as separate
// primitives. They are not separate primitives and deliberately were
// not written as such: a duration is orthogonal to which keyword is
// granted, so both are
//
//	GrantKeywordUntilEOT{Keywords: []string{"hexproof"}}
//	GrantKeywordUntilEOT{Keywords: []string{"indestructible"}}
//
// and a card that grants both at once (Heroic Intervention,
// Tamiyo's Safekeeping) passes both tokens to ONE registry entry
// rather than stacking two named wrappers. What S25 actually had to
// build was the other half — the consumer that reads the token,
// which for indestructible is the destruction path
// (game/indestructible.go). Hexproof's consumer, the targeting gate
// in game/targets.go, shipped in S23.
//
// Granting a token OUTSIDE the canonical set — ward, say — still
// appends a string that nothing reads, so such
// a card ships weaker than printed and MUST say so in its comment.
// This primitive deliberately does not reject unknown tokens: a
// declared-but-inert grant is how The Wandering Rescuer was written,
// so that the day the keyword lands in its consumer the card starts
// working untouched. Boros Charm's indestructible mode and Darksteel
// Citadel both took that bet and both collected in S25 (and an infect
// or wither grant collected with #748) without a
// line of card code changing.
type GrantKeywordUntilEOT struct {
	// Target pins the effect to one permanent. Ignored when Match
	// is set.
	Target uuid.UUID

	// Match selects the affected permanents, evaluated ONCE at
	// resolution (CR 611.2c).
	Match CardPredicate

	// Keywords are canonical lowercase tokens ("first strike", not
	// "First Strike"). HasKeyword compares byte-for-byte.
	Keywords []string

	Label string
}

func (k GrantKeywordUntilEOT) Apply(ctx *Context) error {
	if len(k.Keywords) == 0 {
		return nil
	}
	return untilEndOfTurn(ctx, k.Target, k.Match,
		eotLabel(k.Label, "keyword grant until end of turn"),
		game.AddKeywordsMod(k.Keywords...))
}

// eotHasAbility reports whether the keyword is already present.
// Keeps a Layer 6 grant idempotent when two copies of the same
// effect stack — "gains trample" twice is still just trample, and a
// duplicated string would survive into the wire view.
func eotHasAbility(abilities []string, kw string) bool {
	for _, a := range abilities {
		if a == kw {
			return true
		}
	}
	return false
}

// eotLabel falls back to a generic description when a card doesn't
// bother naming its effect.
func eotLabel(label, fallback string) string {
	if label != "" {
		return label
	}
	return fallback
}
