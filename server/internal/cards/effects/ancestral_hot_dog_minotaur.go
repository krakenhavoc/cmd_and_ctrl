package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const ancestralHotDogAfflict = "ancestral-hot-dog-minotaur/afflict-2"

// Ancestral Hot Dog Minotaur — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Ancestral", "Hot Dog", "Minotaur"; three art stickers.
//	{TK}{TK} — Afflict 2
//	{TK}{TK}{TK} — Flying
//	{TK}{TK} — 1/4
//	{TK}{TK}{TK}{TK}{TK} — 8/6
//
// Afflict 2 is a triggered ability (CR 702.130a): "Whenever this creature becomes blocked, defending player loses 2 life." The first sticker gives it as a bundle (ADR 0093) with the creature it is on as "this creature"; flying is the keyword.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "481c3bfe-2293-447d-aa03-3ea6ed9321aa",
		Name:         "Ancestral Hot Dog Minotaur",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key:       ancestralHotDogAfflict,
				Triggered: []game.TriggeredAbility{Afflict(2, "Ancestral Hot Dog Minotaur sticker")},
				Text:      "Afflict 2 (Whenever this creature becomes blocked, defending player loses 2 life.)",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: ancestralHotDogAfflict},
			{Keywords: []string{"flying"}},
		}},
	})
}
