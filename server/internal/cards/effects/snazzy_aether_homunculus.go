package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const (
	snazzyAetherHomunculusTypes     = "snazzy-aether-homunculus/all-creature-types"
	snazzyAetherHomunculusMagecraft = "snazzy-aether-homunculus/magecraft-draw"
)

// Snazzy Aether Homunculus — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Snazzy", "Aether", "Homunculus"; three art stickers.
//	{TK}{TK} — {1}: Target creature gains all creature types until end of turn.
//	{TK}{TK}{TK} — Magecraft — Whenever you cast or copy an instant or sorcery spell, draw a card.
//	{TK}{TK} — 2/4
//	{TK}{TK}{TK}{TK}{TK} — 8/7
//
// The first sticker gives an ability that makes a target creature every creature type until end of turn (CR 702.73a, a layer 4 effect). The second gives magecraft (CR 207.2c, an ability word): a trigger whenever its controller casts or copies an instant or sorcery spell.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3e5ac76b-08bd-4232-9290-49742b0ff603",
		Name:         "Snazzy Aether Homunculus",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: snazzyAetherHomunculusTypes,
				Activated: []ActivatedAbility{{
					Label:   "{1}: Target creature gains all creature types until end of turn.",
					Cost:    ManaCost("{1}"),
					Targets: TargetCreature("target creature"),
					Effect: func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := stickerFirstLegalTarget(ctx)
						if !ok {
							return nil
						}
						return GrantAllCreatureTypesUntilEOT{Target: id, Label: "Snazzy Aether Homunculus sticker — all creature types"}.Apply(ctx)
					},
				}},
				Text: "{1}: Target creature gains all creature types until end of turn.",
			},
			{
				Key: snazzyAetherHomunculusMagecraft,
				Triggered: []game.TriggeredAbility{
					Magecraft("Snazzy Aether Homunculus sticker — magecraft: draw a card", Do(DrawCards{N: 1})),
				},
				Text: "Magecraft — Whenever you cast or copy an instant or sorcery spell, draw a card.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: snazzyAetherHomunculusTypes},
			{Grant: snazzyAetherHomunculusMagecraft},
		}},
	})
}
