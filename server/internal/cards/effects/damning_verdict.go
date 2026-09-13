package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Damning Verdict — Sorcery {3}{W}{W} (EDHREC rank 1173):
//
//	"Destroy all creatures with no counters on them."
//
// The counters deck's one-sided wrath. DestroyAllMatching over
// creatures whose counter map holds no positive count of ANY kind —
// a +1/+1, a -1/-1, a loyalty, a shield or a flying counter all
// spare the creature, as printed. One simultaneous event, so a
// dies-payoff sees every creature that fell together.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "355ed7ef-bfa6-4538-99fb-a2d203eb7005",
		Name:         "Damning Verdict",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: And(Creature(), b10NoCounters())}.Apply(ctx)
		},
	})
}
