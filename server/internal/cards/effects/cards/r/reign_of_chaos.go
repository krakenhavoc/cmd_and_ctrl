package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reign of Chaos — Sorcery {2}{R}{R} (Mirage):
//
//	"Choose one —
//	 • Destroy target Plains and target white creature.
//	 • Destroy target Island and target blue creature."
//
// Each bullet has two targets, both announced with the spell. Each is
// destroyed if it is still legal as the spell resolves; one that has become
// illegal does not stop the other (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "436b5989-9b8c-4e03-a82e-046837c891ca",
		Name:         "Reign of Chaos",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: effects.ChooseOne(
			effects.Mode("Destroy target Plains and target white creature.", effects.Clauses(
				effects.TargetPermanent("target Plains", effects.Land(), effects.Subtype("Plains")),
				effects.TargetCreature("target white creature", effects.OfColor("W")),
			)),
			effects.Mode("Destroy target Island and target blue creature.", effects.Clauses(
				effects.TargetPermanent("target Island", effects.Land(), effects.Subtype("Island")),
				effects.TargetCreature("target blue creature", effects.OfColor("U")),
			)),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for slot := 0; slot < 2; slot++ {
				t, ok := ctx.ClauseTarget(slot)
				if !ok {
					continue
				}
				if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
