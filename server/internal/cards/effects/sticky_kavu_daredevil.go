package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "3def54bf-2640-47de-84e8-7a9406df007e",
		Name:         "Sticky Kavu Daredevil",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: stickyKavuDaredevilBounce,
				Triggered: []game.TriggeredAbility{
					Optional(Targeting(
						WhenThisDies("Sticky Kavu Daredevil sticker — you may return target creature to its owner's hand",
							func(g *game.Game, item *game.StackItem) error {
								ctx := NewContext(g, item)
								id, ok := stickerFirstLegalTarget(ctx)
								if !ok {
									return nil
								}
								return BounceToHand{Target: id}.Apply(ctx)
							}),
						TargetCreature("target creature")),
						"Return target creature to its owner's hand?"),
				},
				Text: "Whenever this permanent dies, you may return target creature to its owner's hand.",
			},
			{
				Key: stickyKavuDaredevilPump,
				Triggered: []game.TriggeredAbility{
					WheneverThisAttacks("Sticky Kavu Daredevil sticker — creatures you control get +1/+1 until end of turn",
						func(g *game.Game, item *game.StackItem) error {
							return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Toughness: 1, Label: "Sticky Kavu Daredevil sticker — +1/+1"}.Apply(NewContext(g, item))
						}),
				},
				Text: "Whenever this creature attacks, creatures you control get +1/+1 until end of turn.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: stickyKavuDaredevilBounce},
			{Grant: stickyKavuDaredevilPump},
		}},
	})
}
