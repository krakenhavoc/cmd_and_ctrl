package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warping Wail — Instant {1}{C} (Special Guests):
//
//	"Choose one —
//	 • Exile target creature with power or toughness 1 or less.
//	 • Counter target sorcery spell.
//	 • Create a 1/1 colorless Eldrazi Scion creature token. It has
//	   "Sacrifice this token: Add {C}.""
//
// {C} in the cost is colourless mana specifically (CR 107.4c), not generic.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8b9fa02-0a0c-4a87-9a00-3c1c7e2f4bfd",
		Name:         "Warping Wail",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Exile target creature with power or toughness 1 or less.",
				effects.TargetCreature("target creature with power or toughness 1 or less", effects.Or(effects.PowerLE(1), effects.ToughnessLE(1))),
				effects.ExileTheModesTarget), game.Purpose{Answers: game.AnswerRemove}),
			effects.ModeWithPurpose(effects.ModeDoing("Counter target sorcery spell.",
				effects.TargetSpell("target sorcery spell", effects.Sorcery()),
				effects.CounterTheModesTarget), game.Purpose{Answers: game.AnswerRemove}),
			effects.ModeWithPurpose(effects.ModeDoing("Create a 1/1 colorless Eldrazi Scion creature token. It has \"Sacrifice this token: Add {C}.\"", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.CreateToken{Template: effects.B28EldraziScionToken(), N: 1}.Apply(ctx)
				}), game.Purpose{Tokens: 1, CreatureTokens: 1}),
		),
	})
}
