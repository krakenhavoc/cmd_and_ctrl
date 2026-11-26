package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Triton Tactics — Instant {U} (Theros, S82 OLD b389, #5561):
//
//	"Up to two target creatures each get +0/+3 until end of turn. Untap
//	 those creatures. At this turn's next end of combat, tap each creature
//	 that was blocked by one of those creatures this turn and it doesn't
//	 untap during its controller's next untap step."
//
// The pump and untap run on each target still legal. The delayed trigger
// (CR 603.7) is scheduled for the next end of combat step of this turn only
// (TritonTacticsBody counts the steps), carrying the targets; it reads this
// turn's block pairs, so an attacker blocked by one of them taps and is held
// through its controller's next untap step. With no legal target nothing is
// scheduled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "55b72176-e428-4d2a-a701-237e8c8de22e",
		Name:         "Triton Tactics",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 3)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			var refs []game.TargetRef
			for _, t := range legal {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Toughness: 3, Label: "Triton Tactics — +0/+3"}).Apply(ctx); err != nil {
					return err
				}
				refs = append(refs, t)
			}
			for _, t := range refs {
				if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			if len(refs) == 0 {
				return nil
			}
			ids := make([]uuid.UUID, 0, len(refs))
			for _, t := range refs {
				ids = append(ids, t.ID)
			}
			return effects.ScheduleDelayedTrigger{
				At:           game.StepEndCombat,
				Label:        "Triton Tactics — tap each creature blocked by one of those creatures",
				Cards:        ids,
				Body:         effects.TritonTacticsBody,
				Params:       game.EffectParams{Amount: effects.EndCombatsBegunThisTurn(ctx.Game)},
				EachThisTurn: true,
			}.Apply(ctx)
		},
	})
}
