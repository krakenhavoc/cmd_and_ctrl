package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Urza's Dark Cannonball — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Urza's", "Dark", "Cannonball"; three art stickers.
//	{TK}{TK} — Exalted, exalted
//	{TK}{TK}{TK} — Shadow
//	{TK}{TK} — 1/5
//	{TK}{TK}{TK}{TK} — 7/4
//
// Exalted, exalted is two instances of exalted (CR 113.2c: each functions independently; the keyword is cumulative), and each instance is its own trigger when a creature attacks alone. The second sticker gives shadow.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7abe9119-cc5d-4004-8a46-4a3153ead570",
		Name:         "Urza's Dark Cannonball",
		Completeness: CompletenessFull,
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Keywords: []string{game.KeywordExalted, game.KeywordExalted}},
			{Keywords: []string{"shadow"}},
		}},
	})
}
