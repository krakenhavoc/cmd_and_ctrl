package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const (
	carnivalElephantSacrifice   = "carnival-elephant-meteor/sacrifice-draw"
	carnivalElephantProliferate = "carnival-elephant-meteor/attacks-proliferate"
)

// Carnival Elephant Meteor — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Carnival", "Elephant", "Meteor"; three art stickers.
//	{TK}{TK} — Sacrifice this permanent: Draw two cards.
//	{TK}{TK}{TK} — Whenever this creature attacks, proliferate.
//	{TK}{TK} — 4/1
//	{TK}{TK}{TK}{TK}{TK} — 8/7
//
// Both stickers give an ability as a bundle: a sacrifice cost that draws two cards, and an attack trigger that proliferates (CR 701.34).
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e06b7b4-22c1-4e5b-81ee-54ab5ac756eb",
		Name:         "Carnival Elephant Meteor",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			{
				Key: carnivalElephantSacrifice,
				Activated: []effects.ActivatedAbility{{
					Label:   "Sacrifice this permanent: Draw two cards.",
					Cost:    effects.SacrificeThis(),
					Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
					Effect:  effects.Do(effects.DrawCards{N: 2}),
				}},
				Text: "Sacrifice this permanent: Draw two cards.",
			},
			{
				Key: carnivalElephantProliferate,
				Triggered: []game.TriggeredAbility{
					effects.WheneverThisAttacks("Carnival Elephant Meteor sticker — proliferate", effects.Do(effects.Proliferate{})),
				},
				Text: "Whenever this creature attacks, proliferate.",
			},
		},
		StickerSheet: &effects.StickerSheetSpec{Abilities: []effects.StickerAbilitySpec{
			{Grant: carnivalElephantSacrifice},
			{Grant: carnivalElephantProliferate},
		}},
	})
}
