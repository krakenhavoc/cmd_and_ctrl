package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const (
	giantManaCakeFood    = "giant-mana-cake/leaves-food"
	giantManaCakeDestroy = "giant-mana-cake/dies-destroy"
)

// Giant Mana Cake — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Giant", "Mana", "Cake"; three art stickers.
//	{TK}{TK} — When this permanent leaves the battlefield, create two Food tokens.
//	{TK}{TK}{TK}{TK} — When this permanent dies, destroy target nonland permanent.
//	{TK}{TK} — 3/3
//	{TK}{TK}{TK} — 6/2
//
// Both stickers give a leaves-the-battlefield trigger, judged on the permanent as it last existed (CR 603.10a). "Leaves the battlefield" makes two Food tokens; "dies" destroys a target nonland permanent, chosen as the trigger goes on the stack.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "26f6170a-34dc-41c8-b3b1-20377f131e6e",
		Name:         "Giant Mana Cake",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: giantManaCakeFood,
				Triggered: []game.TriggeredAbility{
					WhenThisLeaves("Giant Mana Cake sticker — create two Food tokens", Do(CreateToken{Template: FoodToken(), N: 2})),
				},
				Text: "When this permanent leaves the battlefield, create two Food tokens.",
			},
			{
				Key: giantManaCakeDestroy,
				Triggered: []game.TriggeredAbility{
					Targeting(
						WhenThisDies("Giant Mana Cake sticker — destroy target nonland permanent", destroyFirstLegalCardTarget),
						TargetPermanent("target nonland permanent", Nonland())),
				},
				Text: "When this permanent dies, destroy target nonland permanent.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: giantManaCakeFood},
			{Grant: giantManaCakeDestroy},
		}},
	})
}
