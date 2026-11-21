package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Return Triumphant — Sorcery {1}{W}:
//
//	"Return target creature card with mana value 3 or less from your
//	 graveyard to the battlefield. Create a Young Hero Role token
//	 attached to it. (Enchanted creature has "Whenever this creature
//	 attacks, if its toughness is 3 or less, put a +1/+1 counter on it."
//	 If you put another Role on the creature later, put this one into
//	 the graveyard.)"
//
// The Role is attached to the creature only if it actually came back: a
// card that a replacement sent elsewhere is not a creature on the
// battlefield, and CreateRoleToken creates nothing for it.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a4814615-ffa2-4692-86bc-3a1b7fc2acc8",
		Name:         "Return Triumphant",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCardInGraveyard("target creature card with mana value 3 or less from your graveyard",
			effects.YouOwn(), effects.Creature(), effects.ManaValueLE(3)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			id := targets[0].ID
			if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateRoleToken{Role: effects.RoleYoungHero, Host: id}.Apply(ctx)
		},
	})
}
