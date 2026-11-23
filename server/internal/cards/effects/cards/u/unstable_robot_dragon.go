package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const (
	unstableRobotDragonSwitch = "unstable-robot-dragon/switch-pt"
	unstableRobotDragonPump   = "unstable-robot-dragon/attacks-pump"
)

// Unstable Robot Dragon — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Unstable", "Robot", "Dragon"; three art stickers.
//	{TK}{TK} — {1}: Switch this creature's power and toughness until end of turn.
//	{TK}{TK}{TK}{TK} — Whenever this creature attacks, it gets +5/+5 until end of turn.
//	{TK}{TK} — 3/2
//	{TK}{TK}{TK} — 2/7
//
// The first sticker gives an ability that switches the creature's power and toughness until end of turn (CR 613.4d, layer 7e). The second gives an attack trigger that pumps it.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "676de97a-299b-42ac-aa93-bb12dc9c4460",
		Name:         "Unstable Robot Dragon",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			{
				Key: unstableRobotDragonSwitch,
				Activated: []effects.ActivatedAbility{{
					Label:   "{1}: Switch this creature's power and toughness until end of turn.",
					Cost:    effects.ManaCost("{1}"),
					Purpose: game.Purpose{Answers: game.AnswerPump},
					Effect: func(g *game.Game, item *game.StackItem) error {
						return effects.SwitchPTUntilEOT{Target: item.SourceCardID, Label: "Unstable Robot Dragon sticker — switch power and toughness"}.Apply(effects.NewContext(g, item))
					},
				}},
				Text: "{1}: Switch this creature's power and toughness until end of turn.",
			},
			{
				Key: unstableRobotDragonPump,
				Triggered: []game.TriggeredAbility{
					effects.WheneverThisAttacks("Unstable Robot Dragon sticker — it gets +5/+5 until end of turn", effects.ThisGetsUntilEndOfTurn(5, 5, "Unstable Robot Dragon sticker — +5/+5")),
				},
				Text: "Whenever this creature attacks, it gets +5/+5 until end of turn.",
			},
		},
		StickerSheet: &effects.StickerSheetSpec{Abilities: []effects.StickerAbilitySpec{
			{Grant: unstableRobotDragonSwitch},
			{Grant: unstableRobotDragonPump},
		}},
	})
}
