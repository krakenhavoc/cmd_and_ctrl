package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const (
	deepFriedPlagueMyrScry    = "deep-fried-plague-myr/attacks-scry"
	deepFriedPlagueMyrDestroy = "deep-fried-plague-myr/leaves-destroy"
)

// Deep-Fried Plague Myr — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Deep-Fried", "Plague", "Myr"; three art stickers.
//	{TK}{TK} — Whenever this creature attacks, scry 1.
//	{TK}{TK}{TK} — Whenever this permanent leaves the battlefield, you may destroy target artifact or enchantment.
//	{TK}{TK}{TK} — 4/5
//	{TK}{TK}{TK}{TK} — 8/4
//
// The first sticker gives an attack trigger that scries 1. The second gives a leaves-the-battlefield trigger with a "you may" and a target: "target artifact or enchantment" is chosen as the trigger goes on the stack (CR 603.3d), and the may is asked as it resolves (CR 603.5).
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43f5ce56-6ad1-4a42-b1d7-26f1c7933693",
		Name:         "Deep-Fried Plague Myr",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			{
				Key: deepFriedPlagueMyrScry,
				Triggered: []game.TriggeredAbility{
					effects.WheneverThisAttacks("Deep-Fried Plague Myr sticker — scry 1", effects.Do(effects.Scry{N: 1})),
				},
				Text: "Whenever this creature attacks, scry 1.",
			},
			{
				Key: deepFriedPlagueMyrDestroy,
				Triggered: []game.TriggeredAbility{
					effects.Optional(effects.Targeting(
						effects.WhenThisLeaves("Deep-Fried Plague Myr sticker — you may destroy target artifact or enchantment", effects.DestroyFirstLegalCardTarget),
						effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment()))),
						"Destroy target artifact or enchantment?"),
				},
				Text: "Whenever this permanent leaves the battlefield, you may destroy target artifact or enchantment.",
			},
		},
		StickerSheet: &effects.StickerSheetSpec{Abilities: []effects.StickerAbilitySpec{
			{Grant: deepFriedPlagueMyrScry},
			{Grant: deepFriedPlagueMyrDestroy},
		}},
	})
}
