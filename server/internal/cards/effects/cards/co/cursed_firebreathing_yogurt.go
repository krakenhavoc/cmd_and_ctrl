package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const cursedYogurtPing = "cursed-firebreathing-yogurt/tap-damage"

// Cursed Firebreathing Yogurt — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Cursed", "Firebreathing", "Yogurt"; three art stickers.
//	{TK}{TK} — Prowess, prowess
//	{TK}{TK}{TK}{TK}{TK} — {2}, {T}: This permanent deals 2 damage to any target.
//	{TK}{TK} — 4/2
//	{TK}{TK}{TK}{TK} — 4/8
//
// Prowess, prowess is two instances of prowess (CR 702.108b: each triggers separately; the keyword is cumulative). The second sticker gives an activated ability as a bundle: "This permanent" is the object the sticker is on.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "535fd365-19f4-4861-8f86-d814650355e9",
		Name:         "Cursed Firebreathing Yogurt",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			{
				Key: cursedYogurtPing,
				Activated: []effects.ActivatedAbility{{
					Label:   "{2}, {T}: This permanent deals 2 damage to any target.",
					Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
					Targets: effects.TargetAny(),
					Effect:  effects.ThisDealsTwoToItsTarget,
				}},
				Text: "{2}, {T}: This permanent deals 2 damage to any target.",
			},
		},
		StickerSheet: &effects.StickerSheetSpec{Abilities: []effects.StickerAbilitySpec{
			{Keywords: []string{game.KeywordProwess, game.KeywordProwess}},
			{Grant: cursedYogurtPing},
		}},
	})
}
