package cardsw

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wine of Blood and Iron — Artifact {3} (Saviors of Kamigawa, #5601):
//
//	"{4}: Target creature gets +X/+0 until end of turn, where X is its
//	 power. Sacrifice this artifact at the beginning of the next end
//	 step."
//
// "+X/+0 where X is its power" is a doubling of its power (CR 701.10b),
// read as the ability resolves. The sacrifice is a delayed trigger
// (CR 603.7) scheduled in the same resolution, at the next end step of any
// player's turn, and it only takes this same artifact (CR 400.7). If the
// target is illegal the ability does nothing at all (CR 608.2b), so no
// sacrifice is scheduled. Each activation schedules its own.
//
// No simplification.
func init() {
	const label = "Wine of Blood and Iron — sacrifice this artifact at the beginning of the next end step"
	effects.Register(effects.Spec{
		OracleID:     "22d85711-485b-47a2-b161-eaed135f1ddc",
		Name:         "Wine of Blood and Iron",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}: Target creature gets +X/+0 until end of turn, where X is its power. Sacrifice this artifact at the beginning of the next end step.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{4}"),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id := effects.FirstLegalBattlefieldTarget(ctx)
				if id == uuid.Nil {
					return nil
				}
				if err := effects.DoublePowerUntilEOT(ctx, id, "Wine of Blood and Iron — +X/+0, X its power"); err != nil {
					return err
				}
				return effects.OldB507SacrificeThisAtTheNextEndStep(label)(g, item)
			},
		}},
	})
}
