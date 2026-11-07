package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragon's Prey — Instant {2}{B} (Tarkir: Dragonstorm, #4463):
//
//	"This spell costs {2} more to cast if it targets a Dragon.
//	 Destroy target creature."
//
// Murder with Monastery Siege's target-reading increase in the self slot:
// the announced target is judged as the cost is determined (CR 601.2f,
// after 601.2c), so a Dragon costs {4}{B} and anything else {2}{B}.
//
// No simplification.
func init() {
	dragonTax := CostsMore(2, "This spell costs {2} more to cast if it targets a Dragon.", ItTargets(HasSubtype("Dragon")))
	dragonTax.ReadsTargets = true
	Register(Spec{
		OracleID:          "86cb340f-ef06-4be9-b457-703784387dbe",
		Name:              "Dragon's Prey",
		Completeness:      CompletenessFull,
		Purpose:           game.Purpose{Answers: game.AnswerRemove},
		Targets:           TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{dragonTax},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return destroyFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
