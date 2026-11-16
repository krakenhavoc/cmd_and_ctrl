package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unfathomable Truths — Instant {4}{U} (Modern Horizons 3, #5467):
//
//	"Devoid (This card has no color.)
//	 Draw three cards and create a 0/1 colorless Eldrazi Spawn creature
//	 token with "Sacrifice this token: Add {C}.""
//
// Devoid is the engine's colour-removing keyword (CR 702.114a). The
// draw is the plain primitive and the token is the catalog's own Eldrazi
// Spawn (Pawn of Ulamog's), whose sacrifice is a real mana ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "63765157-e2ba-4cec-9501-8a405259de5b",
		Name:            "Unfathomable Truths",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Purpose:         game.Purpose{Draws: 3, Tokens: 1, CreatureTokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (DrawCards{N: 3}).Apply(ctx); err != nil {
				return err
			}
			return CreateToken{Controller: ctx.Controller(), Template: EldraziSpawnToken(), N: 1}.Apply(ctx)
		},
	})
}
