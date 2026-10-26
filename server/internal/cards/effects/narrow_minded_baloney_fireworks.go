package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const narrowMindedBaloneyLife = "narrow-minded-baloney-fireworks/attacks-gain-2"

// Narrow-Minded Baloney Fireworks — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Narrow-Minded", "Baloney", "Fireworks"; three art stickers.
//	{TK}{TK} — Whenever this creature attacks, you gain 2 life.
//	{TK}{TK}{TK} — Vigilance, reach
//	{TK}{TK} — 2/4
//	{TK}{TK}{TK}{TK}{TK} — 7/7
//
// The first sticker gives an attack trigger that gains 2 life. The second gives vigilance and reach.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "366e0a62-ff95-48d1-bc63-7995a393bc34",
		Name:         "Narrow-Minded Baloney Fireworks",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: narrowMindedBaloneyLife,
				Triggered: []game.TriggeredAbility{
					WheneverThisAttacks("Narrow-Minded Baloney Fireworks sticker — you gain 2 life", Do(GainLife{Amount: 2})),
				},
				Text: "Whenever this creature attacks, you gain 2 life.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: narrowMindedBaloneyLife},
			{Keywords: []string{"vigilance", "reach"}},
		}},
	})
}
