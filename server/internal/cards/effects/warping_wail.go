package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c8b9fa02-0a0c-4a87-9a00-3c1c7e2f4bfd",
		Name:         "Warping Wail",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Exile target creature with power or toughness 1 or less.",
				TargetCreature("target creature with power or toughness 1 or less", Or(PowerLE(1), ToughnessLE(1))),
				ExileTheModesTarget), game.Purpose{Answers: game.AnswerRemove}),
			ModeWithPurpose(ModeDoing("Counter target sorcery spell.",
				TargetSpell("target sorcery spell", Sorcery()),
				CounterTheModesTarget), game.Purpose{Answers: game.AnswerRemove}),
			ModeWithPurpose(ModeDoing("Create a 1/1 colorless Eldrazi Scion creature token. It has \"Sacrifice this token: Add {C}.\"", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return CreateToken{Template: b28EldraziScionToken(), N: 1}.Apply(ctx)
				}), game.Purpose{Tokens: 1, CreatureTokens: 1}),
		),
	})
}
