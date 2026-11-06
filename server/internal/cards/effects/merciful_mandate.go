package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merciful Mandate — Instant {1}{W} (Star Trek Commander, #4974):
//
//	"Untap all creatures you control. If it's a combat phase, you may end
//	 the combat phase. (Remove all attackers and blockers from combat.
//	 Exile all spells and abilities from the stack, including this
//	 spell.)"
//
// The untap is Aurelia's UntapAllCreaturesYouControl. "If it's a combat
// phase" is read as the spell resolves, after the untap; then the caster
// is asked, and a yes is Mandate of Peace's game.EndCombatPhaseForEffect
// (CR 724.2). The spell waits on the stack for the answer (the resolution
// tail), so the end of combat exiles it along with the rest of the stack
// (CR 724.2b), and a no lets it go to the graveyard as usual. Outside a
// combat phase nothing is asked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "201269a1-8f11-485a-a9be-6447b189b4df",
		Name:         "Merciful Mandate",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (UntapAllCreaturesYouControl{}).Apply(ctx); err != nil {
				return err
			}
			if game.PhaseOf(ctx.Game.Turn.Step) != game.PhaseCombat {
				return nil
			}
			return MayChoice{
				Question: "Merciful Mandate — end the combat phase?",
				YesLabel: "End the combat phase",
				NoLabel:  "Don't end it",
				Then:     trekB33EndTheCombatPhase,
			}.Apply(ctx)
		},
	})
}
