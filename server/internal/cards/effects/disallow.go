package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Disallow — Instant {1}{U}{U} (EDHREC rank 1821):
//
//	"Counter target spell, activated ability, or triggered ability.
//	 (Mana abilities can't be targeted.)"
//
// Cancel with a Stifle stapled on. The spell half is Counterspell's
// shape for a mana more.
//
// Sandbox simplification, declared: only spells can be targeted. An
// activated or triggered ability on the stack is a StackMeta item
// with no card in the stack zone, and every target clause enumerates
// cards, so an ability cannot be chosen — Siren Stormtamer's posture.
// The CounterTarget primitive can counter an ability item; what is
// missing is the target clause that would let a caster point at one.
// Weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:     "88b51e15-6630-4e14-a6b8-db0aa12e34ef",
		Name:         "Disallow",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Only spells can be countered — an activated or triggered ability on the stack can't be picked."},
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
