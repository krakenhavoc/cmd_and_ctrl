package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const unsanctionedAncientJugglerBolster = "unsanctioned-ancient-juggler/attacks-bolster"

// Unsanctioned Ancient Juggler — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Unsanctioned", "Ancient", "Juggler"; three art stickers.
//	{TK}{TK} — Whenever this creature attacks, bolster 1.
//	{TK}{TK}{TK}{TK} — Indestructible
//	{TK}{TK} — 3/2
//	{TK}{TK}{TK} — 5/4
//
// The first sticker gives an attack trigger that bolsters 1 (CR 701.39a: a +1/+1 counter on a creature you control with the least toughness). The second gives indestructible.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4413cb03-d6e9-4e6c-b5fe-9240ca0ebd13",
		Name:         "Unsanctioned Ancient Juggler",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: unsanctionedAncientJugglerBolster,
				Triggered: []game.TriggeredAbility{
					WheneverThisAttacks("Unsanctioned Ancient Juggler sticker — bolster 1", Do(Bolster{N: 1})),
				},
				Text: "Whenever this creature attacks, bolster 1. (Choose a creature with the least toughness among creatures you control and put a +1/+1 counter on it.)",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: unsanctionedAncientJugglerBolster},
			{Keywords: []string{"indestructible"}},
		}},
	})
}
