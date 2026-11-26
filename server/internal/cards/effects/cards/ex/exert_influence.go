package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Exert Influence — Sorcery {4}{U} (Battle for Zendikar):
//
//	"Converge — Gain control of target creature if its power is less than or
//	 equal to the number of colors of mana spent to cast this spell."
//
// The power is read as the spell resolves, from the target as it is then
// (CR 608.2b: a target that is no longer legal does nothing).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "feb9cce1-6c3c-4e04-93ee-d646a888bc29",
		Name:                "Exert Influence",
		Completeness:        effects.CompletenessFull,
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		Targets:             effects.Harmful(effects.TargetCreature("target creature")),
		Purpose:             game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if effects.B43PowerNowOrLastKnown(ctx.Game, t.ID) > ctx.ColorsSpentCount() {
					return nil
				}
				return effects.GainControl{Target: t.ID, Duration: game.IndefiniteDuration(),
					Label: "Exert Influence — gain control"}.Apply(ctx)
			}
			return nil
		},
	})
}
