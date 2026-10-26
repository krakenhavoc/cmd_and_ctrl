package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const slimyBurritoBushido = "slimy-burrito-illusion/bushido-2"

// Slimy Burrito Illusion — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Slimy", "Burrito", "Illusion"; three art stickers.
//	{TK}{TK} — Bushido 2
//	{TK}{TK}{TK} — Double strike
//	{TK}{TK} — 2/4
//	{TK}{TK}{TK}{TK} — 5/6
//
// Bushido 2 is the engine's keyword trigger (Bushido): the creature gets +2/+2 until end of turn whenever it blocks or becomes blocked (CR 702.45a). The second sticker gives double strike.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2d13d399-7198-4cab-b564-dfe0e09425a3",
		Name:         "Slimy Burrito Illusion",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key:       slimyBurritoBushido,
				Triggered: []game.TriggeredAbility{Bushido(2, "Slimy Burrito Illusion sticker")},
				Text:      "Bushido 2 (Whenever this creature blocks or becomes blocked, it gets +2/+2 until end of turn.)",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: slimyBurritoBushido},
			{Keywords: []string{"double strike"}},
		}},
	})
}
