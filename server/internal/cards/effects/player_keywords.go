package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// player_keywords.go — the card-facing half of CR 702.16b / CR 702.11c
// (#1197, ADR 0072's 2026-09-22 amendment): giving a PLAYER an
// ability.
//
// Two shapes, and a card uses exactly one of them:
//
//   - a permanent's printed static, "You have hexproof" — declare
//     Spec.PlayerKeywords and write no code at all. The engine reads
//     the battlefield on every query, so two of them compose and one
//     leaving cannot revoke the other's grant. A static that holds
//     only some of the time, "During your turn, you … have hexproof"
//     (#2013), is Spec.PlayerKeywordsIf, built with YouHaveDuringYourTurn
//     or YouHaveWhile below.
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

// ProtectionFromEachOfYourOpponents is "you have protection from each
// of your opponents" (Absolute Virtue, #2745), re-exported from the
// engine for the reason KeywordHexproof is. "Your" is the player the
// ability belongs to; the engine reads that, not the token.
const ProtectionFromEachOfYourOpponents = game.ProtectionFromEachOfYourOpponents

// KeywordHexproof is CR 702.11c's token, re-exported from the engine
// so a card file spells it once and the compiler checks it. The
// constant itself lives in game/player_statics.go, beside the reader
// that compares against it.
const KeywordHexproof = game.KeywordHexproof

// YouHaveWhile is a permanent's printed "<condition>, you have
// <keyword>": the controller has the ability while `while` holds, asked
// on every query with the permanent as it is now (CR 611.3a). For
// Spec.PlayerKeywordsIf. `label` is the printed sentence, for the
// tile's ability row.
//
// `while` runs under the game lock and may only read; it must not ask a
// player's abilities (that recurses into the walk that called it).
func YouHaveWhile(keyword, label string, while func(g *game.Game, source game.Card) bool) game.PlayerKeywordIf {
	return game.PlayerKeywordIf{Keyword: keyword, Label: label, While: while}
}

// YouHaveDuringYourTurn is "During your turn, you have <keyword>" —
// Gruul Spellbreaker's player half (#2013, CR 702.11c). "Your turn" is
// the granting permanent's controller's, read now, so a permanent that
// changes control grants on its new controller's turns.
func YouHaveDuringYourTurn(keyword string) game.PlayerKeywordIf {
	return YouHaveWhile(keyword, "During your turn, you have "+keyword+".",
		func(g *game.Game, source game.Card) bool { return isActivePlayer(g, source.Controller) })
}

// ThisHasDuringYourTurn is "This creature has <keyword> during your
// turn" — the object half of Gruul Spellbreaker's static, a layer-6
// grant to the source itself (firstStrikeDuringYourTurn's shape, for
// any keyword). The layer pass is recomputed as the turn changes, so
// the keyword is gone on an opponent's turn.
func ThisHasDuringYourTurn(keyword string) game.StaticAbility {
	ab := KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && isActivePlayer(g, source.Controller)
	}, keyword)
	ab.Label = "Has " + keyword + " during your turn."
	return ab
}

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
