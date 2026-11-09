package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Protective Response — Instant {2}{W} (Lorwyn Eclipsed, #4458):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Destroy target attacking or blocking creature."
//
// Convoke is Appeal to Eirdu's TapCost. The target is any creature that is
// attacking or blocking (CR 506.4), judged as the spell resolves (CR
// 608.2b): a creature removed from combat in response is illegal. Tapping
// an attacker or blocker to convoke does not remove it from combat.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c840b89b-81be-4644-8bb5-35d14b1bc620",
		Name:         "Protective Response",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Targets:      TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return DestroyTarget{Target: id}.Apply(ctx)
		},
	})
}
