package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Windgrace's Judgment — Instant {3}{B}{G} (EDHREC rank 3173):
//
//	"For any number of opponents, destroy target nonland permanent
//	 that player controls."
//
// One removal spell per opponent at instant speed — the multiplayer
// Putrefy. The target clause is "any number of target nonland
// permanents your opponents control", and the "one per opponent" is
// enforced as the spell resolves: of the announced targets still
// legal, the first named under each controller is destroyed and any
// later one under the same controller is skipped. That is the
// declared simplification, and it runs the weaker way — the caster
// can never destroy more than one permanent of one opponent's, but
// the picker lets them name two and lose the second.
//
// The controller of each target is read as the spell resolves, so a
// permanent that changed hands in response is judged by who holds it
// then (a target that stopped being an opponent's is illegal and
// skipped by the CR 608.2b re-check).
func init() {
	Register(Spec{
		OracleID:     "de1ca6ed-b275-4f62-ba05-f31b3659b352",
		Name:         "Windgrace's Judgment",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"You can name more than one permanent of the same opponent's, but only the first one you chose is destroyed — one per opponent, as printed."},
		Targets:      TargetPermanent("any number of target nonland permanents your opponents control", Nonland(), OpponentControls()).WithCount(1, 0),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b30DestroyOnePerController(ctx)
		},
	})
}
