package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const (
	yawgmothMerfolkSoulDiscard = "yawgmoth-merfolk-soul/leaves-discard"
	yawgmothMerfolkSoulTokens  = "yawgmoth-merfolk-soul/leaves-clown-robots"
)

// Yawgmoth Merfolk Soul — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Yawgmoth", "Merfolk", "Soul"; three art stickers.
//	{TK}{TK} — When this permanent leaves the battlefield, target player discards a card.
//	{TK}{TK}{TK}{TK}{TK} — When this permanent leaves the battlefield, create five 1/1 white Clown Robot artifact creature tokens.
//	{TK}{TK} — 3/3
//	{TK}{TK}{TK}{TK} — 6/5
//
// Both stickers give a leaves-the-battlefield trigger, judged on the permanent as it last existed (CR 603.10a): a target player discards a card, and five 1/1 white Clown Robot artifact creature tokens are created.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4c02ee30-8681-45ed-adb5-edf06409eee4",
		Name:         "Yawgmoth Merfolk Soul",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: yawgmothMerfolkSoulDiscard,
				Triggered: []game.TriggeredAbility{
					Targeting(
						WhenThisLeaves("Yawgmoth Merfolk Soul sticker — target player discards a card",
							func(g *game.Game, item *game.StackItem) error {
								ctx := NewContext(g, item)
								if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
									return nil
								}
								return DiscardCards{Player: item.Targets[0].ID, N: 1}.Apply(ctx)
							}),
						TargetPlayer("target player")),
				},
				Text: "When this permanent leaves the battlefield, target player discards a card.",
			},
			{
				Key: yawgmothMerfolkSoulTokens,
				Triggered: []game.TriggeredAbility{
					WhenThisLeaves("Yawgmoth Merfolk Soul sticker — create five 1/1 white Clown Robot artifact creature tokens",
						Do(CreateToken{Template: TokenCard("1/1 white Clown Robot artifact"), N: 5})),
				},
				Text: "When this permanent leaves the battlefield, create five 1/1 white Clown Robot artifact creature tokens.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: yawgmothMerfolkSoulDiscard},
			{Grant: yawgmothMerfolkSoulTokens},
		}},
	})
}
