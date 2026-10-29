package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const happyDeadSquirrelMana = "happy-dead-squirrel/tap-mana"

// Happy Dead Squirrel — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Happy", "Dead", "Squirrel"; three art stickers.
//	{TK}{TK} — {T}: Add {C}{C}. Spend this mana only to cast noncreature spells.
//	{TK}{TK}{TK} — Infect
//	{TK}{TK} — 3/2
//	{TK}{TK}{TK}{TK} — 4/7
//
// The first sticker gives a mana ability with a spending restriction (CR 106.12): the two colorless mana can pay only for noncreature spells. The second gives infect.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c3e5d355-fbfb-4489-aeea-8b0ea5de6fcd",
		Name:         "Happy Dead Squirrel",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: happyDeadSquirrelMana,
				Mana: []ManaAbility{{
					Cost:         ManaAbilityCost{Tap: true},
					Produced:     "{C}{C}",
					Label:        "{T}: Add {C}{C}. Spend this mana only to cast noncreature spells.",
					Restrictions: []string{ManaRestrictCast, ManaRestrictNotType("Creature")},
				}},
				Text: "{T}: Add {C}{C}. Spend this mana only to cast noncreature spells.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: happyDeadSquirrelMana},
			{Keywords: []string{game.KeywordInfect}},
		}},
	})
}
