package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const contortionistOtterHaste = "contortionist-otter-storm/tap-haste"

// Contortionist Otter Storm — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Contortionist", "Otter", "Storm"; three art stickers.
//	{TK}{TK} — {T}: Target creature gains haste until end of turn.
//	{TK}{TK}{TK}{TK} — Deathtouch, lifelink
//	{TK}{TK} — 5/1
//	{TK}{TK}{TK} — 3/5
//
// The first sticker gives an activated ability as a bundle: "{T}: Target creature gains haste until end of turn." The second gives deathtouch and lifelink.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "67624a72-aa18-4b0c-a2cb-77e56499944a",
		Name:         "Contortionist Otter Storm",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			{
				Key: contortionistOtterHaste,
				Activated: []effects.ActivatedAbility{{
					Label:   "{T}: Target creature gains haste until end of turn.",
					Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
					Cost:    effects.TapCost(),
					Targets: effects.TargetCreature("target creature"),
					Effect:  effects.EbTargetCreatureGainsUntilEOT("haste", "Contortionist Otter Storm sticker — haste until end of turn"),
				}},
				Text: "{T}: Target creature gains haste until end of turn.",
			},
		},
		StickerSheet: &effects.StickerSheetSpec{Abilities: []effects.StickerAbilitySpec{
			{Grant: contortionistOtterHaste},
			{Keywords: []string{"deathtouch", "lifelink"}},
		}},
	})
}
