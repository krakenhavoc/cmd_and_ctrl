package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// In Thrall to the Pit — Sorcery {3}{R} (Dominaria United):
//
//	"Kicker {2}{B}
//	 Gain control of target creature until end of turn. Untap that creature.
//	 It gains haste until end of turn. If this spell was kicked, sacrifice
//	 that creature at the beginning of the next end step."
//
// Act of Treason's three clauses, with a kicked tail: a delayed trigger that
// sacrifices that object (the same permanent, CR 400.7) at the next end step,
// which comes before the cleanup step where control reverts, so it is still
// the caster's to sacrifice. A target gone by resolution does nothing
// (CR 608.2b), and no sacrifice is scheduled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "5bf1c26b-12eb-4be7-b9d6-f8ad75a778aa",
		Name:          "In Thrall to the Pit",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Kicker("{2}{B}")},
		Targets:       TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (GainControl{
				Target:   target,
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "In Thrall to the Pit — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if err := (GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"haste"},
				Label:    "In Thrall to the Pit — haste until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if !ctx.WasKicked() {
				return nil
			}
			c, ok := battlefieldCardForEffect(ctx.Game, target)
			if !ok {
				return nil
			}
			return ScheduleDelayedTrigger{
				At:         game.StepEnd,
				Controller: ctx.Controller(),
				Label:      "In Thrall to the Pit — sacrifice that creature",
				Body:       cmdr2S12SoulstokeSacrifice,
				Params:     game.EffectParams{Object: game.ObjectRef{ID: target, Epoch: c.ObjectEpoch}},
			}.Apply(ctx)
		},
	})
}
