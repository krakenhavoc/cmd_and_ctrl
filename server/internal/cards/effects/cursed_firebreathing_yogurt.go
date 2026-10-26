package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "535fd365-19f4-4861-8f86-d814650355e9",
		Name:         "Cursed Firebreathing Yogurt",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: cursedYogurtPing,
				Activated: []ActivatedAbility{{
					Label:   "{2}, {T}: This permanent deals 2 damage to any target.",
					Cost:    Plus(ManaCost("{2}"), TapCost()),
					Targets: TargetAny(),
					Effect: func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
							return nil
						}
						return DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: 2}.Apply(ctx)
					},
				}},
				Text: "{2}, {T}: This permanent deals 2 damage to any target.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Keywords: []string{game.KeywordProwess, game.KeywordProwess}},
			{Grant: cursedYogurtPing},
		}},
	})
}
