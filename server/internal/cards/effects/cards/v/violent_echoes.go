package cardsv

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Violent Echoes — Instant {2}{R}{R} (Reality Fracture, tracker #2795):
//
//	"Violent Echoes deals 6 damage to target creature or planeswalker. If
//	 excess damage was dealt to that permanent this way, empower Jace X,
//	 where X is that excess damage."
//
// Excess damage (CR 120.4a) is what landed beyond lethal: toughness less
// damage already marked for a creature, the loyalty for a planeswalker
// (frLethalDamage), read after the damage so prevention lowers it. With no excess the keyword action does not happen at all (the
// "if"), so no empty Jace token appears. A target that left is skipped
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e8ecec61-cc47-4051-a599-0e7470764488",
		Name:         "Violent Echoes",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 6)),
		// Empower Jace waits for a prompt the damage asks (a CR 616 order; CR 608.2c, #4155).
		// The first step files the lethal amount and the event cursor; the second reads the
		// damage dealt after the cursor from the log.
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					c, ok := ctx.Game.LookupCardForEffect(t.ID)
					if !ok {
						return nil
					}
					ctx.RecordStep(violentEchoesRecord, []uuid.UUID{t.ID}, effects.FrLethalDamage(c), int(effects.B25LastEventSeq(ctx.Game)))
					return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 6}.Apply(ctx)
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				r := ctx.StepRecord(violentEchoesRecord)
				if len(r.IDs) == 0 || len(r.Ns) < 2 {
					return nil
				}
				excess := effects.B27DamageDealtToAfter(ctx.Game, ctx.Source(), r.IDs[0], uint64(r.Ns[1])) - r.Ns[0]
				if excess <= 0 {
					return nil
				}
				return effects.EmpowerJace{N: excess}.Apply(ctx)
			}),
		),
	})
}

// violentEchoesRecord files the target, the lethal damage it needed and the
// event cursor before the damage.
const violentEchoesRecord = "violent-echoes/damage"
