package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Night Brushwagg Ringmaster — Stickers (an Unfinity sticker sheet, CR
// 123.2; ADR 0208):
//
//	Name stickers "Night", "Brushwagg" and "Ringmaster"; three art stickers.
//	{TK}{TK} — Menace
//	{TK}{TK}{TK} — Persist
//	{TK}{TK} — 2/3
//	{TK}{TK}{TK}{TK}{TK}{TK} — 10/10
//
// A sheet is not a card. A deck lists it beside its cards and the game
// keeps it as data (game.Player.StickerSheetPool); three of a deck's
// sheets are chosen at random as the game begins (CR 103.2d). This entry
// says what its two ability stickers give the object they are on: menace,
// and persist, which the engine runs off the dying permanent's last-known
// abilities (CR 702.79a, undying_persist.go), so a persist sticker returns
// the creature and comes back with it (CR 123.5). The name, power and
// toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1436370f-baf0-4659-804e-90aad871b05b",
		Name:         "Night Brushwagg Ringmaster",
		Completeness: CompletenessFull,
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Keywords: []string{game.CounterMenace}},
			{Keywords: []string{game.KeywordPersist}},
		}},
	})
}
