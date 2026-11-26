package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crypsis — Instant {1}{U} (Born of the Gods, S82 OLD b377, #5559):
//
//	"Target creature you control gains protection from creatures your
//	 opponents control until end of turn. Untap it."
//
// The protection is the quality token ProtectionFromOpponentsCreatures (#6253;
// "your" is the protected creature's controller). A target that became illegal
// counters the whole spell, the untap included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6077e650-32f9-4954-8286-9c8a3f3227ae",
		Name:         "Crypsis",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerProtect, Targets: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, game.ProtectionFromOpponentsCreatures)).Targets},
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{game.ProtectionFromOpponentsCreatures}, Label: "Crypsis — protection from creatures your opponents control"}).Apply(ctx); err != nil {
				return err
			}
			return effects.UntapTarget{Target: t.ID}.Apply(ctx)
		},
	})
}
