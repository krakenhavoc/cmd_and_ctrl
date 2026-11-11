package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Expansion // Explosion — split card, oracle af6f771f-5154-4fb3-8ed4-768d71eea568:
//
//	Expansion — Instant {U/R}{U/R}: "Copy target instant or sorcery
//	  spell with mana value 4 or less. You may choose new targets for the
//	  copy."
//	Explosion — Instant {X}{U}{U}{R}{R}: "Explosion deals X damage to any
//	  target. Target player draws X cards."
//
// Both halves are ADR 0034 faces (Expansion under the bare ID, Explosion
// under "#1"). Expansion is Twincast with a mana-value bound on its target
// (CR 707.10, 707.10c). Explosion's two targets are separate clauses
// (CR 601.2c): damage to the first, then the second player draws, in
// printed order. X is paid on the cast.
//
// No simplification.
func init() {
	const oracle = "af6f771f-5154-4fb3-8ed4-768d71eea568"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Expansion",
		Completeness: CompletenessFull,
		Targets:      instantOrSorcerySpell("target instant or sorcery spell with mana value 4 or less", ManaValueLE(4)),
		OnResolve:    cmdrB21CopyTheTargetSpell,
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Explosion",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets: Clauses(
			TargetAny(),
			TargetPlayer("target player"),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			if t, ok := ctx.ClauseTarget(0); ok {
				if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: x}).Apply(ctx); err != nil {
					return err
				}
			}
			if t, ok := ctx.ClauseTarget(1); ok {
				return DrawCards{Player: t.ID, N: x}.Apply(ctx)
			}
			return nil
		},
	})
}
