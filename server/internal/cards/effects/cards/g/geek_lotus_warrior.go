package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const (
	geekLotusWarriorPump = "geek-lotus-warrior/pump"
	geekLotusWarriorPing = "geek-lotus-warrior/creature-enters-damage"
)

// Geek Lotus Warrior — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Geek", "Lotus", "Warrior"; three art stickers.
//	{TK}{TK} — {2}: This creature gets +2/+0 until end of turn.
//	{TK}{TK}{TK}{TK} — Whenever a creature enters under your control, this permanent deals 2 damage to target player.
//	{TK}{TK} — 4/1
//	{TK}{TK}{TK} — 3/6
//
// The first sticker gives a pump ability on the permanent itself. The second gives a trigger on each creature entering under your control (the creature itself included) that deals 2 damage to a target player from the permanent it is on.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c0a448ee-e5f9-4e57-85b0-f6d401018170",
		Name:         "Geek Lotus Warrior",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			{
				Key: geekLotusWarriorPump,
				Activated: []effects.ActivatedAbility{{
					Label:   "{2}: This creature gets +2/+0 until end of turn.",
					Cost:    effects.ManaCost("{2}"),
					Purpose: game.Purpose{Answers: game.AnswerPump},
					Effect:  effects.ThisGetsUntilEndOfTurn(2, 0, "Geek Lotus Warrior sticker — +2/+0"),
				}},
				Text: "{2}: This creature gets +2/+0 until end of turn.",
			},
			{
				Key: geekLotusWarriorPing,
				Triggered: []game.TriggeredAbility{
					effects.Targeting(
						effects.On(game.EventETB, effects.CreatureEnteredUnderYourControl,
							"Geek Lotus Warrior sticker — this permanent deals 2 damage to target player",
							effects.ThisDealsTwoToItsTarget),
						effects.TargetPlayer("target player")),
				},
				Text: "Whenever a creature enters under your control, this permanent deals 2 damage to target player.",
			},
		},
		StickerSheet: &effects.StickerSheetSpec{Abilities: []effects.StickerAbilitySpec{
			{Grant: geekLotusWarriorPump},
			{Grant: geekLotusWarriorPing},
		}},
	})
}
