package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shatter the Oath — Sorcery {3}{B}{B}:
//
//	"Destroy target creature or enchantment. Create a Wicked Role token
//	 attached to up to one target creature you control. (If you control
//	 another Role on it, put that one into the graveyard. Enchanted
//	 creature gets +1/+0. When this token is put into a graveyard, each
//	 opponent loses 1 life.)"
//
// Two clauses by slot. The destruction comes first, as printed, so a
// creature you control that is also the destroyed one is gone before
// the Role would attach.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "f545870d-b628-4acb-88c7-37c2d72a5a85",
		Name:         "Shatter the Oath",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetPermanent("target creature or enchantment", Or(Creature(), Enchantment())),
			TargetCreature("up to one target creature you control", YouControl()).WithCount(0, 1),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			// The Role is the sentence after the destruction, so it rides
			// the destruction's keyed tail (a replaced or still-pending
			// move settles first; #4253 slice 10a).
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				return ctx.Game.DestroyPermanentsKeyedForEffect([]uuid.UUID{t.ID}, ctx.ItemTail(shatterTheOathRole))
			}
			return createRoleOnClauseTarget(ctx, 1, RoleWicked)
		},
	})
}

// shatterTheOathRole is "create a Wicked Role token attached to up to one
// target creature you control", after the destruction.
var shatterTheOathRole = CardTail("then/shatter-the-oath-role", func(ctx *Context, _ game.TailOutcome) error {
	return createRoleOnClauseTarget(ctx, 1, RoleWicked)
})
