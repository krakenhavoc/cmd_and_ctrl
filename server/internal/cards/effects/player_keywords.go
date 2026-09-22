package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// player_keywords.go — the card-facing half of CR 702.16i / CR 702.11d
// (#1197, ADR 0072's 2026-09-22 amendment): giving a PLAYER an
// ability.
//
// Two shapes, and a card uses exactly one of them:
//
//   - a permanent's printed static, "You have hexproof" — declare
//     Spec.PlayerKeywords and write no code at all. The engine reads
//     the battlefield on every query, so two of them compose and one
//     leaving cannot revoke the other's grant.
//   - a resolved spell or trigger, "you gain protection from
//     everything until your next turn" — GainPlayerKeyword below,
//     which stores the grant with its CR 611.2 duration.
//
// The token is the ENGINE's, not the card's: "hexproof", or a
// protection token protection.go's closed grammar parses. A card file
// never invents one, for the reason that file gives — a token nothing
// can parse grants nothing at all, so a typo ships a card that looks
// finished and does nothing. ProtectionFromEverything is the constant
// for this seam's one quality; Register refuses anything unparseable
// at boot either way.

// ProtectionFromEverything is CR 702.16j's quality, as an engine
// token. The one protection a catalogued card grants a PLAYER —
// Teferi's Protection and The One Ring both print exactly it.
//
// A constant rather than a literal because it is spelled in three
// card files and a typo in any of them would silently grant nothing.
const ProtectionFromEverything = "protection from everything"

// KeywordHexproof is CR 702.11d's token, re-exported from the engine
// so a card file spells it once and the compiler checks it. The
// constant itself lives in game/player_statics.go, beside the reader
// that compares against it.
const KeywordHexproof = game.KeywordHexproof

// GainPlayerKeyword is "you gain <ability> until <duration>" — the
// GRANTED half of a player-level ability.
//
//	GainPlayerKeyword{
//	    Player:   ctx.Controller(),
//	    Keyword:  ProtectionFromEverything,
//	    Label:    "Teferi's Protection — protection from everything",
//	    Duration: DurationUntilYourNextTurn(ctx, ctx.Controller()),
//	}.Apply(ctx)
//
// The Duration's zero value is "until end of turn", so a card file
// that forgets one gets the shortest window rather than a permanent
// grant — the same direction every other default in this catalog
// errs in.
//
// For a PERMANENT's printed "You have hexproof", use
// Spec.PlayerKeywords instead: that half is derived from the
// battlefield and needs no code.
type GainPlayerKeyword struct {
	// Player is the seat that gains it. Zero is a no-op.
	Player uuid.UUID
	// Keyword is the engine token — ProtectionFromEverything,
	// KeywordHexproof, or any "protection from <quality>" the closed
	// grammar parses.
	Keyword string
	// Label is the log's attribution, "<card> — <clause>".
	Label string
	// Duration is the CR 611.2 window. Build it with the constructors
	// in durations.go.
	Duration game.Duration
}

func (g GainPlayerKeyword) Apply(ctx *Context) error {
	if ctx == nil || g.Player == uuid.Nil || g.Keyword == "" {
		return nil
	}
	ctx.Game.GrantPlayerStaticForEffect(g.Player, g.Keyword, g.Label, ctx.Source(), g.Duration)
	return nil
}
