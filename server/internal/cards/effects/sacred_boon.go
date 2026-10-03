package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sacred Boon — Instant {1}{W}:
//
//	"Prevent the next 3 damage that would be dealt to target creature this turn. At the beginning of the next end step, put a +0/+1 counter on that creature for each 1 damage prevented this way."
//
// ADR 0108 owner decision 2 (#1906): the charged shield (CR 615.7) whose
// CR 615.5 additional effect is a delayed trigger. Every prevention the
// shield makes before the end step joins ONE delayed trigger
// (ScheduleOrJoinDelayedTriggerForEffect), which puts on the total.
// Damage that can't be prevented adds nothing and leaves the shield whole
// (CR 615.12).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "32a8d49b-cdfb-4944-bfb5-afd704cc3658",
		Name:         "Sacred Boon",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return endStepToughnessShield(ctx, 3, "Sacred Boon")
		},
	})
}

// endStepToughnessShield is "Prevent the next N damage that would be
// dealt to <the first target> this turn. At the beginning of the next end
// step, put a +0/+1 counter on that creature for each 1 damage prevented
// this way" (Sacred Boon, Scars of the Veteran). The follow-up does
// nothing for a target that isn't a creature (Scars' "If it's a
// creature"). A target gone by resolution gets no shield (CR 608.2b).
func endStepToughnessShield(ctx *Context, n int, name string) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard && t.Kind != game.TargetPlayer {
			continue
		}
		return PreventNextDamage{
			Target: t.ID,
			Amount: n,
			Then:   endStepToughnessCountersBody,
			Label:  name + " — +0/+1 counters at the next end step for damage prevented",
		}.Apply(ctx)
	}
	return nil
}
