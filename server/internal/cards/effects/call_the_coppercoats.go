package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Call the Coppercoats — Instant {2}{W}:
//
//	"Strive — This spell costs {1}{W} more to cast for each target
//	 beyond the first.
//	 Choose any number of target opponents. Create X 1/1 white Human
//	 Soldier creature tokens, where X is the number of creatures those
//	 opponents control."
//
// #746: strive is a coloured per-target increase. The {W} of each
// surcharge joins the cost as a coloured requirement (ADR 0048
// addendum, open question 3, option a); reductions elsewhere on the
// board still spend only generic mana.
//
// The rulings: zero targets is legal (and makes nothing), and an
// opponent who is no longer a legal target on resolution does not
// have their creatures counted.
func init() {
	Register(Spec{
		OracleID:     "73d8c33d-916a-4220-96ae-9622aad36210",
		Name:         "Call the Coppercoats",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("any number of target opponents", Opponent()).WithCount(0, 0),
		SelfCostModifiers: []game.CostModifier{
			CostsMorePerTargetBeyondFirst("{1}{W}", "Strive — This spell costs {1}{W} more to cast for each target beyond the first."),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			opponents := map[game.TargetRef]bool{}
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetPlayer {
					opponents[t] = true
				}
			}
			n := 0
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.IsCreature() && opponents[game.TargetRef{Kind: game.TargetPlayer, ID: c.Controller}] {
					n++
				}
			}
			if n == 0 {
				return nil
			}
			return CreateToken{Controller: ctx.Controller(), Template: TokenCard("1/1 white Human Soldier"), N: n}.Apply(ctx)
		},
	})
}
