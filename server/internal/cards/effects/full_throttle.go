package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Full Throttle — Sorcery {4}{R}{R}:
//
//	"After this main phase, there are two additional combat phases.
//	 At the beginning of each combat this turn, untap all creatures
//	 that attacked this turn."
//
// Two combat phases straight after the main phase, with no main phase
// between them (the 2025-02-07 ruling), through the turn plan (ADR 0059
// sub-PR 2b, #753). Off a main phase no combat is added, but the second
// sentence still applies.
//
// The second sentence is a delayed trigger with a stated duration (CR
// 603.7b): it fires at the beginning of EVERY combat for the rest of
// the turn — the two added ones, the turn's own combat after them, and
// any other effect's — and ends at cleanup. The first beginning of
// combat finds nothing that attacked; the later ones untap the
// creatures that attacked in the combats before.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3c2f87eb-b0f4-4578-b143-6fba313a82ec",
		Name:         "Full Throttle",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (AddPhases{
				Anchor: game.PhaseAnchor{Kind: game.AnchorThisMainPhase},
				Kinds:  []game.PhaseKind{game.PhaseKindCombat, game.PhaseKindCombat},
			}).Apply(ctx); err != nil {
				return err
			}
			return ScheduleDelayedTrigger{
				At:           game.StepBeginCombat,
				Label:        "Full Throttle — untap all creatures that attacked this turn",
				Body:         untapCreaturesThatAttackedBody,
				EachThisTurn: true,
			}.Apply(ctx)
		},
	})
}
