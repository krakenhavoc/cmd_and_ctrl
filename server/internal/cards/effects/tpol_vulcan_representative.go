package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// T'Pol, Vulcan Representative — Legendary Creature — Vulcan Scientist
// {4}{U}, 4/3 (Star Trek Commander, #4974):
//
//	"When T'Pol enters, tap target creature an opponent controls and put
//	 two stun counters on it."
//
// The target, if still legal at resolution (CR 608.2b), is tapped and
// gets two stun counters; each stun counter is CR 122.1d's untap
// replacement, which the engine applies.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e8a548e8-057e-4023-ac20-bac973dbce09",
		Name:         "T'Pol, Vulcan Representative",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("T'Pol, Vulcan Representative — tap target creature an opponent controls and put two stun counters on it",
				Do(Instruction(func(ctx *Context) error {
					ts := ctx.LegalTargets()
					if len(ts) == 0 {
						return nil
					}
					if err := (TapTarget{Target: ts[0].ID}).Apply(ctx); err != nil {
						return err
					}
					return AddCounter{Target: ts[0].ID, Kind: game.CounterStun, N: 2}.Apply(ctx)
				}))), TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
