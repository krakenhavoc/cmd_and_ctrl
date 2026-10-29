package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const trendyCircusPirateSquirrels = "trendy-circus-pirate/damage-squirrels"

// Trendy Circus Pirate — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Trendy", "Circus", "Pirate"; three art stickers.
//	{TK}{TK} — Deathtouch
//	{TK}{TK}{TK}{TK}{TK} — Whenever this creature deals combat damage to a player, create that many 1/1 green Squirrel creature tokens.
//	{TK}{TK} — 5/1
//	{TK}{TK}{TK} — 3/6
//
// The first sticker gives deathtouch. The second gives a combat-damage trigger that creates that many 1/1 green Squirrel creature tokens: the amount is the combat damage the triggering event dealt.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eb980d94-2860-4fe5-85e0-2757985721f1",
		Name:         "Trendy Circus Pirate",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: trendyCircusPirateSquirrels,
				Triggered: []game.TriggeredAbility{
					WheneverThisDealsCombatDamageToAPlayer("Trendy Circus Pirate sticker — create that many 1/1 green Squirrel creature tokens",
						func(g *game.Game, item *game.StackItem) error {
							n := item.Trigger.Event.Amount
							if n <= 0 {
								return nil
							}
							return CreateToken{Template: TokenCard("1/1 green Squirrel"), N: n}.Apply(NewContext(g, item))
						}),
				},
				Text: "Whenever this creature deals combat damage to a player, create that many 1/1 green Squirrel creature tokens.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Keywords: []string{"deathtouch"}},
			{Grant: trendyCircusPirateSquirrels},
		}},
	})
}
