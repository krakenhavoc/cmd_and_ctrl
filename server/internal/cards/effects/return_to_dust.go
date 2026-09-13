package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Return to Dust — Instant {2}{W}{W} (EDHREC rank 1576):
//
//	"Exile target artifact or enchantment. If you cast this spell
//	 during your main phase, you may exile up to one other target
//	 artifact or enchantment."
//
// One clause, one to two targets: the second slot exists only when
// the spell is cast in the caster's own main phase. The engine's
// target clause has no way to make a slot's legality depend on the
// step, so the clause is declared as "one or two" and the step is
// read at RESOLUTION — the same answer as at announce, because the
// stack has to be empty for a step to end, so a spell resolves in the
// step it was cast in.
//
// Sandbox simplification, declared, weaker than printed: outside the
// caster's main phase a second target can still be PICKED at
// announce, but only the first is exiled — the second is ignored,
// and the mana was spent as if it had been offered. On the caster's
// own main phase both go, as printed. Never stronger.
func init() {
	Register(Spec{
		OracleID:     "3029df1d-d02a-4fed-8ab4-000a2096f823",
		Name:         "Return to Dust",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Outside your own main phase you can still pick a second target, but only the first one is exiled."},
		Targets: TargetPermanent("target artifact or enchantment, plus up to one other during your main phase",
			Or(Artifact(), Enchantment())).WithCount(1, 2),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			limit := 1
			if b14InYourMainPhase(ctx.Game, item.Controller) {
				limit = 2
			}
			for i, t := range item.Targets {
				if i >= limit {
					break
				}
				if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
					continue
				}
				if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
