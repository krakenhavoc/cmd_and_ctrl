package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zombify — Sorcery {3}{B}:
//
//	"Return target creature card from your graveyard to the
//	 battlefield."
//
// The plain version of the effect, and the one that shows what the
// extra words on Reanimate are buying: no life loss, but also YOUR
// graveyard only, so it can never take an opponent's creature.
//
// Because the target is restricted to the caster's own pile, the
// creature's owner and its new controller are the same player, and
// the controller argument is redundant. It is passed explicitly
// anyway — the alternative is relying on a default that is only
// correct for half this family.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bb95db4d-5017-4121-bf79-d68476602d8c",
		Name:         "Zombify",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreatureInYourGraveyard(),
		Purpose:      effects.ForTargets(effects.Reanimates(0)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			_, _ = effects.ReanimateSingleTarget(ctx, ctx.Controller())
			return nil
		},
	})
}
