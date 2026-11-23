package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const (
	stickyKavuDaredevilBounce = "sticky-kavu-daredevil/dies-bounce"
	stickyKavuDaredevilPump   = "sticky-kavu-daredevil/attacks-pump"
)

// Sticky Kavu Daredevil — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Sticky", "Kavu", "Daredevil"; three art stickers.
//	{TK}{TK} — Whenever this permanent dies, you may return target creature to its owner's hand.
//	{TK}{TK}{TK}{TK} — Whenever this creature attacks, creatures you control get +1/+1 until end of turn.
//	{TK}{TK} — 3/3
//	{TK}{TK}{TK} — 2/6
//
// The first sticker gives a dies trigger with a "you may" and a target creature that is returned to its owner's hand. The second gives an attack trigger that pumps the creatures you control; the set is fixed as it resolves (CR 611.2c).
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3def54bf-2640-47de-84e8-7a9406df007e",
		Name:         "Sticky Kavu Daredevil",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			{
				Key: stickyKavuDaredevilBounce,
				Triggered: []game.TriggeredAbility{
					effects.Optional(effects.Targeting(
						effects.WhenThisDies("Sticky Kavu Daredevil sticker — you may return target creature to its owner's hand",
							func(g *game.Game, item *game.StackItem) error {
								ctx := effects.NewContext(g, item)
								id, ok := effects.StickerFirstLegalTarget(ctx)
								if !ok {
									return nil
								}
								return effects.BounceToHand{Target: id}.Apply(ctx)
							}),
						effects.TargetCreature("target creature")),
						"Return target creature to its owner's hand?"),
				},
				Text: "Whenever this permanent dies, you may return target creature to its owner's hand.",
			},
			{
				Key: stickyKavuDaredevilPump,
				Triggered: []game.TriggeredAbility{
					effects.WheneverThisAttacks("Sticky Kavu Daredevil sticker — creatures you control get +1/+1 until end of turn",
						func(g *game.Game, item *game.StackItem) error {
							return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1, Label: "Sticky Kavu Daredevil sticker — +1/+1"}.Apply(effects.NewContext(g, item))
						}),
				},
				Text: "Whenever this creature attacks, creatures you control get +1/+1 until end of turn.",
			},
		},
		StickerSheet: &effects.StickerSheetSpec{Abilities: []effects.StickerAbilitySpec{
			{Grant: stickyKavuDaredevilBounce},
			{Grant: stickyKavuDaredevilPump},
		}},
	})
}
