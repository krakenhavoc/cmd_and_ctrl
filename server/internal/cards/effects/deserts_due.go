package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Desert's Due — Instant {1}{B} (Outlaws of Thunder Junction, #4468):
//
//	"Target creature gets -2/-2 until end of turn. It gets an additional
//	 -1/-1 until end of turn for each Desert you control."
//
// Defile's count with a -2/-2 floor. The Deserts are counted once, as the
// spell resolves (CR 611.2c), so a Desert played afterwards changes nothing;
// with none the creature gets only -2/-2. "Desert" is read through the
// spell's text, as Defile reads "Swamp". An illegal target counters the
// spell (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "909eacf3-05ce-4c2b-97a8-49b6ea597b64",
		Name:         "Desert's Due",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(HarmUntilEOT(0, -2, -2)), Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			n := 2 + countControlled(ctx.Game, ctx.Controller(), MatchLandSubtype(ctx.TextSubtype("Desert")))
			return BoostUntilEOT{
				Target:    t.ID,
				Power:     -n,
				Toughness: -n,
				Label:     "Desert's Due — -2/-2, and -1/-1 for each Desert you control",
			}.Apply(ctx)
		},
	})
}
