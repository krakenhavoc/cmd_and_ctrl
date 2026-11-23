package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unexpected Request — Sorcery {2}{R} (Final Fantasy, #4462):
//
//	"Gain control of target creature until end of turn. Untap that
//	 creature. It gains haste until end of turn. You may attach an
//	 Equipment you control to that creature. If you do, unattach it at
//	 the beginning of the next end step."
//
// Act of Treason's control, untap and haste, then an optional attach: a
// resolution-time pick of zero or one of the Equipment you control (not a
// target, so a hexproof Equipment can be picked). Only an attach that really
// happened schedules the unattach, a delayed trigger (CR 603.7) naming the
// Equipment as the object it was, so one that has left and come back stays
// as it is (CR 400.7). It unattaches that Equipment at the beginning of the
// next end step whatever it is attached to by then. A target that is no
// longer legal as the spell resolves gets none of this (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f4497af5-8b81-46b3-bb52-eaed1999809b",
		Name:         "Unexpected Request",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				target, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (effects.GainControl{
					Target:   target,
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Unexpected Request — gain control until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.UntapTarget{Target: target}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{
					Target:   target,
					Keywords: []string{"haste"},
					Label:    "Unexpected Request — haste until end of turn",
				}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				target, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok || !effects.OnBattlefield(ctx.Game, target) {
					return nil
				}
				return effects.ChoosePermanents{
					Question: "Unexpected Request — you may attach an Equipment you control to that creature",
					Keyed:    unexpectedRequestAttach,
					Carry:    []uuid.UUID{target},
					Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
						var out []uuid.UUID
						for _, c := range g.BattlefieldCardsForEffect() {
							if c.Controller == of && c.HasSubtype("Equipment") {
								out = append(out, c.InstanceID)
							}
						}
						return out, 0, 1
					},
				}.Apply(ctx)
			}),
		),
	})
}

// unexpectedRequestAttach is the chosen Equipment attached to the creature,
// and the unattach scheduled if it really is attached to it, keyed (#4290).
var unexpectedRequestAttach = effects.RegisterPermanentsPick("prompt-run/unexpected-request-attach", effects.PermanentsDef{
	ThenRun: func(ctx *effects.Context, picked game.PromptedPicks, r game.RunDone) error {
		ids := picked.Cards()
		if len(ids) == 0 || len(r.Carry) == 0 {
			return nil
		}
		g := ctx.Game
		equipment, creature := ids[0], r.Carry[0]
		if err := g.AttachForEffect(equipment, game.TargetRef{Kind: game.TargetCard, ID: creature}); err != nil {
			return err
		}
		c, ok := g.LookupCardForEffect(equipment)
		if !ok || !c.IsAttachedTo(creature) {
			return nil
		}
		return effects.ScheduleDelayedTrigger{
			At:     game.StepEnd,
			Label:  "Unexpected Request — unattach the Equipment",
			Cards:  []uuid.UUID{equipment},
			Body:   unattachTheObjectBody,
			Params: game.EffectParams{Objects: []game.ObjectRef{{ID: equipment, Epoch: c.ObjectEpoch}}},
		}.Apply(ctx)
	},
})

// unattachTheObjectBody is "unattach it": each object listed, if it is
// still that object on the battlefield.
var unattachTheObjectBody = game.DelayedBody("fin/unattach-the-objects", func(g *game.Game, _ *game.StackItem, p game.EffectParams) error {
	for _, ref := range p.Objects {
		c, ok := g.LookupCardForEffect(ref.ID)
		if !ok || c.ObjectEpoch != ref.Epoch || !effects.OnBattlefield(g, ref.ID) {
			continue
		}
		if err := g.UnattachForEffect(ref.ID); err != nil {
			return err
		}
	}
	return nil
})
