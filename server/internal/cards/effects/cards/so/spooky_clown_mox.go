package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const spookyClownMoxTap = "spooky-clown-mox/tap-tap-creature"

// Spooky Clown Mox — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Spooky", "Clown", "Mox"; three art stickers.
//	{TK}{TK} — Vigilance
//	{TK}{TK}{TK}{TK} — {1}, {T}: Tap target creature.
//	{TK}{TK} — 1/5
//	{TK}{TK}{TK} — 5/4
//
// The first sticker gives vigilance. The second gives an ability that taps the permanent and pays {1} to tap a target creature.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "785e6563-51c4-4063-9ad9-0392954e135e",
		Name:         "Spooky Clown Mox",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			{
				Key: spookyClownMoxTap,
				Activated: []effects.ActivatedAbility{{
					Label:   "{1}, {T}: Tap target creature.",
					Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
					Targets: effects.TargetCreature("target creature"),
					Effect:  effects.TapLegalCardTargets,
				}},
				Text: "{1}, {T}: Tap target creature.",
			},
		},
		StickerSheet: &effects.StickerSheetSpec{Abilities: []effects.StickerAbilitySpec{
			{Keywords: []string{"vigilance"}},
			{Grant: spookyClownMoxTap},
		}},
	})
}
