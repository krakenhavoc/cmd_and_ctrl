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
// # Declared simplification: Gift (CR 702.170) has no primitive
//
// Checked the whole catalog and the engine: no card, no cast-time
// hook, and no resolve-time flag answers to Gift. It needs a shape
// nothing here builds today — an OPTIONAL choice made at announce
// (CR 601.2b) that names an OPPONENT rather than paying a cost
// (Kicker / Multikicker / Buyback are all `game.AdditionalCost`
// payments, never a player pick); an immediate effect on that
// opponent BEFORE the spell's own resolution body runs ("they draw a
// card before its other effects"); and a resolve-time flag ("if the
// gift was promised") in the shape `ctx.WasKicked()` already gives a
// paid optional cost, but for a naming rather than a payment. Filed
// as its own seam rather than invented here — a whole keyword
// mechanic, not a one-card shape, so it earns a tracked gap the way
// amass and storm did on this same batch rather than a card-comment
// mention.
//
// DECLARED SIMPLIFICATION (weaker than printed — the exact shape
// AGENTS.md's own worked example uses, "Flashback isn't implemented —
// the spell can only be cast from hand"): this spell never promises
// the gift. It always resolves as its un-gifted branch, so no
// opponent ever draws a card and permanents you control never gain
// indestructible. The guaranteed half — you and your permanents gain
// hexproof until end of turn — is complete and untouched, and is the
// whole reason Dawn's Truce gets played as a combat-trick fog; only
// the optional upgrade path is missing.
//
// # The guaranteed half, in full
//
// "You... gain hexproof" is GainPlayerKeyword (#1197's
// game.PlayerStatic machinery, KeywordHexproof, until end of turn —
// the printed duration exactly). "Permanents you control gain
// hexproof" is GrantKeywordUntilEOT with Match: YouControl() and no
// type restriction: the printed clause says "permanents", not
// "creatures", and the primitive's Match is a plain predicate over
// the whole battlefield, so no narrowing is needed to match the text.
func init() {
	Register(Spec{
		OracleID:     "37c06f89-db36-4937-9404-2b07cd22e1a6",
		Name:         "Dawn's Truce",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Gift isn't implemented — this spell never promises the gift. You and your permanents still gain hexproof until end of turn, but no opponent draws a card and your permanents never gain indestructible.",
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (GainPlayerKeyword{
				Player:   ctx.Controller(),
				Keyword:  KeywordHexproof,
				Label:    "Dawn's Truce — hexproof",
				Duration: DurationUntilEndOfTurn(ctx),
			}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Match:    YouControl(),
				Keywords: []string{"hexproof"},
				Label:    "Dawn's Truce — permanents you control gain hexproof",
			}.Apply(ctx)
		},
	})
}
