package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const spaceFungusBattleCry = "space-fungus-snickerdoodle/battle-cry"

// Space Fungus Snickerdoodle — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Space", "Fungus", "Snickerdoodle"; three art stickers.
//	{TK}{TK} — Skulk
//	{TK}{TK}{TK} — Battle cry
//	{TK}{TK} — 3/2
//	{TK}{TK}{TK}{TK}{TK} — 6/8
//
// The first sticker gives skulk (CR 702.118a). The second gives battle cry (CR 702.91a), the engine's keyword trigger (BattleCry): each other attacking creature gets +1/+0 until end of turn.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fd29fdbd-b4c5-4d23-9388-df0a798fbbba",
		Name:         "Space Fungus Snickerdoodle",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key:       spaceFungusBattleCry,
				Triggered: []game.TriggeredAbility{BattleCry("Space Fungus Snickerdoodle sticker")},
				Text:      "Battle cry (Whenever this creature attacks, each other attacking creature gets +1/+0 until end of turn.)",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Keywords: []string{"skulk"}},
			{Grant: spaceFungusBattleCry},
		}},
	})
}
