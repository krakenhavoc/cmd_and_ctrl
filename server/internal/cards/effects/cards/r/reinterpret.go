package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reinterpret — Instant {2}{U}{R} (Commander 2021, S82 OLD b058, #5504):
//
//	"Counter target spell. You may cast a spell with equal or lesser mana
//	 value from your hand without paying its mana cost."
//
// Press the Enemy's two steps with a counter in place of the bounce: the
// target's mana value is read before it is countered (CR 608.2h; a spell's
// counts the X it was cast with, CR 202.3e), then the controller is offered a
// free cast of any spell from their hand with mana value at most that, as part
// of the resolution (CR 608.2g; Cecily's resolution_cast offer: ONE spell, an
// {X} in its cost is 0, additional costs still owed). The cast is offered
// only if the target was still legal; with none Reinterpret is countered on
// resolution (CR 608.2b) and offers nothing. A land is never offered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e197104f-5382-4cb2-8afe-5ca0594629c2",
		Name:         "Reinterpret",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					ctx.RecordStep(reinterpretRecord, []uuid.UUID{t.ID}, effects.GapB131ReturnedManaValue(ctx, t.ID))
				}
				return effects.CounterTheTargetSpell(ctx.Item, ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				r := ctx.StepRecord(reinterpretRecord)
				if len(r.IDs) == 0 || len(r.Ns) == 0 {
					return nil
				}
				effects.OfferFreeHandCast(ctx, ctx.Controller(), reinterpretOffer, "a spell with equal or lesser mana value",
					effects.SpellWithManaValueAtMost(r.Ns[0]))
				return nil
			}),
		),
	})
}

const reinterpretRecord = "reinterpret/countered"

// reinterpretOffer is the free-cast prompt's key. An on-disk key: never
// renamed, never reused.
var reinterpretOffer = effects.RegisterFreeHandCastOffer("choose-cards/reinterpret-cast", "Reinterpret")
