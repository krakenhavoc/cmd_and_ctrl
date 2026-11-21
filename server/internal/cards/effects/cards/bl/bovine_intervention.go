package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bovine Intervention — Instant {1}{W} (Marvel Super Heroes Commander, #4455):
//
//	"Destroy target artifact or creature. Its controller creates a 2/2
//	 white Ox creature token."
//
// Beast Within's order and its stale-controller guard: the target's controller
// is read BEFORE the destroy, so a token for a permanent that changed hands
// goes to whoever held it when the spell resolved. If the target has left the
// battlefield in response, the spell has no legal target and does nothing
// (CR 608.2b). The Ox is a new row in the token table.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7402d98f-6d58-4ff2-a0db-9eec4417a75f",
		Name:         "Bovine Intervention",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		// #2679: what the target's controller is given back.
		Purpose:   effects.ForTargets(effects.RemovalReturning(0, game.TargetReturn{CreatureTokens: 1, TokenPower: 2, TokenToughness: 2})),
		OnResolve: msc4DestroyThenGiveOx,
	})
}

// msc4DestroyThenGiveOx is "Destroy target <permanent>. Its controller
// creates <token>": the controller is read before the destroy, then the
// token is created for them.
func msc4DestroyThenGiveOx(item *game.StackItem, ctx *effects.Context) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	target := item.Targets[0].ID
	controller, ok := effects.ControllerOfTarget(ctx, target)
	if !ok {
		return nil
	}
	if err := (effects.DestroyTarget{Target: target}).Apply(ctx); err != nil {
		return err
	}
	return effects.CreateToken{Controller: controller, Template: effects.TokenCard("2/2 white Ox"), N: 1}.Apply(ctx)
}
