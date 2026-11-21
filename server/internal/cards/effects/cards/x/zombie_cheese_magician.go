package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const zombieCheeseMagicianDraw = "zombie-cheese-magician/damage-draw"

// Zombie Cheese Magician — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Zombie", "Cheese", "Magician"; three art stickers.
//	{TK}{TK} — First strike
//	{TK}{TK}{TK}{TK} — Whenever this creature deals combat damage to a player, draw that many cards.
//	{TK}{TK} — 3/3
//	{TK}{TK}{TK} — 6/2
//
// The first sticker gives first strike. The second gives a combat-damage trigger that draws that many cards: the amount is the combat damage the triggering event dealt.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "786bdf20-b1f1-40f7-acd0-8dcb46da2bb0",
		Name:         "Zombie Cheese Magician",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			{
				Key: zombieCheeseMagicianDraw,
				Triggered: []game.TriggeredAbility{
					effects.WheneverThisDealsCombatDamageToAPlayer("Zombie Cheese Magician sticker — draw that many cards",
						func(g *game.Game, item *game.StackItem) error {
							n := item.Trigger.Event.Amount
							if n <= 0 {
								return nil
							}
							return effects.DrawCards{N: n}.Apply(effects.NewContext(g, item))
						}),
				},
				Text: "Whenever this creature deals combat damage to a player, draw that many cards.",
			},
		},
		StickerSheet: &effects.StickerSheetSpec{Abilities: []effects.StickerAbilitySpec{
			{Keywords: []string{"first strike"}},
			{Grant: zombieCheeseMagicianDraw},
		}},
	})
}
