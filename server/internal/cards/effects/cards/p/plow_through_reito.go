package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Plow Through Reito — Instant — Arcane {1}{W} (Saviors of Kamigawa, #5601):
//
//	"Sweep — Return any number of Plains you control to their owner's hand.
//	 Target creature gets +1/+1 until end of turn for each Plains returned
//	 this way."
//
// The Plains are picked as the spell resolves (OldB591AskSweep) and returned
// together; the target creature, if it is still legal, gets +N/+N for the N
// that reached a hand (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e04f151a-66bb-4bc3-a625-91cd23cb1b18",
		Name:         "Plow Through Reito",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0))},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if _, ok := effects.FirstLegalTarget(ctx); !ok {
				return nil
			}
			return effects.OldB591AskSweep(ctx, "Plains", "Plow Through Reito — return any number of Plains you control to their owner's hand", plowThroughReitoPick)
		},
	})
}

// plowThroughReitoPick returns the Plains, then pumps the target. An on-disk
// key.
var plowThroughReitoPick = effects.AskedPick("choose-cards/plow-through-reito-plains", func(ctx *effects.Context, r game.CardsChosen) error {
	n, err := effects.OldB591ReturnSwept(ctx, r.Picked)
	if err != nil || n <= 0 {
		return err
	}
	id := effects.FirstLegalBattlefieldTarget(ctx)
	if id == uuid.Nil {
		return nil
	}
	return effects.BoostUntilEOT{Target: id, Power: n, Toughness: n, Label: "Plow Through Reito — +1/+1 for each Plains returned"}.Apply(ctx)
})
