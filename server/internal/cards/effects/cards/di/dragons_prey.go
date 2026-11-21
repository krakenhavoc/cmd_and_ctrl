package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	dragonTax := effects.CostsMore(2, "This spell costs {2} more to cast if it targets a Dragon.", effects.ItTargets(effects.HasSubtype("Dragon")))
	dragonTax.ReadsTargets = true
	effects.Register(effects.Spec{
		OracleID:          "86cb340f-ef06-4be9-b457-703784387dbe",
		Name:              "Dragon's Prey",
		Completeness:      effects.CompletenessFull,
		Purpose:           game.Purpose{Answers: game.AnswerRemove},
		Targets:           effects.TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{dragonTax},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
