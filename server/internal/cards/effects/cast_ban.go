package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cast_ban.go — the card-facing half of #1316 (ADR 0066's 2026-09-23
// amendment, ADR 0085's): a cast restriction created by a RESOLVING
// SPELL, with a CR 611.2 duration.
//
// Two shapes, and a card uses exactly one:
//
//	RestrictCasting{
//	    Player:   opp,
//	    Rule:     game.CastBanRule{Kind: game.CastBanOutright, ExceptFromZone: game.ZoneHand},
//	    Label:    "Avatar's Wrath — can't cast spells from anywhere other than their hand",
//	    Duration: DurationUntilYourNextTurn(ctx, ctx.Controller()),
//	}.Apply(ctx)
//
//	RestrictCasting{
//	    Player:   opp,
//	    Rule:     game.CastBanRule{Kind: game.CastBanOutright},
//	    Label:    "Mandate of Peace — can't cast spells this turn",
//	    Duration: DurationUntilEndOfTurn(ctx),
//	}.Apply(ctx)
//
// For a PERMANENT's printed "players can't cast spells" (Rule of Law,
// Grafdigger's Cage), use Spec.CastRestrictions instead
// (cast_restriction.go): that half needs no duration at all, because
// the source's continued presence on the battlefield already is one.

// RestrictCasting is "<player> can't cast <rule>, until <duration>" —
// the GRANTED half of a per-player cast ban.
//
// The Duration's zero value is "until end of turn", so a card file
// that forgets one gets the shortest window rather than a permanent
// ban — the same direction every other default in this catalog errs
// in.
type RestrictCasting struct {
	// Player is the seat that is banned. Zero is a no-op.
	Player uuid.UUID
	// Rule is what the ban says. Build it with game.CastBanRule{...}
	// literals; there is no wrapper constructor because the two
	// catalogued shapes are one field apart (ExceptFromZone) and a
	// third constructor per shape would be more code than the struct
	// literal it replaces.
	Rule game.CastBanRule
	// Label is the log's and the refusal's attribution, "<card> — <clause>".
	Label string
	// Duration is the CR 611.2 window. Build it with the constructors
	// in durations.go.
	Duration game.Duration
}

func (r RestrictCasting) Apply(ctx *Context) error {
	if ctx == nil || r.Player == uuid.Nil {
		return nil
	}
	ctx.Game.GrantCastBanForEffect(r.Player, r.Rule, r.Label, ctx.Source(), r.Duration)
	return nil
}
