package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dawn's Truce — Instant {1}{W} (EDHREC rank ~360):
//
//	"Gift a card (You may promise an opponent a gift as you cast this
//	 spell. If you do, they draw a card before its other effects.)
//	 You and permanents you control gain hexproof until end of turn.
//	 If the gift was promised, permanents you control also gain
//	 indestructible until end of turn."
//
// # Gift (CR 702.174, ADR 0089)
//
// `Gift: GiftACard()` is the whole of the keyword: the caster may
// choose an opponent at announce (CR 601.2b), and when they do that
// opponent draws a card BEFORE the hexproof lands (CR 702.174j). The
// resolution only has to read the promise back for the printed
// upgrade — indestructible for permanents you control.
//
// This card was the seam's first waiting card (#1267): it shipped as
// the hexproof half alone, never promising the gift.
//
// # The guaranteed half
//
// "You... gain hexproof" is GainPlayerKeyword (#1197's
// game.PlayerStatic machinery, KeywordHexproof, until end of turn —
// the printed duration exactly). "Permanents you control gain
// hexproof" is GrantKeywordUntilEOT with Match: YouControl() and no
// type restriction: the printed clause says "permanents", not
// "creatures", and the primitive's Match is a plain predicate over
// the whole battlefield, so no narrowing is needed to match the text.
// The indestructible grant rides the same primitive, evaluated once at
// resolution (CR 611.2c), so a permanent that enters afterwards gets
// neither — as printed.
func init() {
	Register(Spec{
		OracleID:     "37c06f89-db36-4937-9404-2b07cd22e1a6",
		Name:         "Dawn's Truce",
		Completeness: CompletenessFull,
		Gift:         GiftACard(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (GainPlayerKeyword{
				Player:   ctx.Controller(),
				Keyword:  KeywordHexproof,
				Label:    "Dawn's Truce — hexproof",
				Duration: DurationUntilEndOfTurn(ctx),
			}).Apply(ctx); err != nil {
				return err
			}
			keywords := []string{"hexproof"}
			if ctx.GiftPromised() {
				keywords = append(keywords, "indestructible")
			}
			return GrantKeywordUntilEOT{
				Match:    YouControl(),
				Keywords: keywords,
				Label:    "Dawn's Truce — permanents you control gain hexproof",
			}.Apply(ctx)
		},
	})
}
