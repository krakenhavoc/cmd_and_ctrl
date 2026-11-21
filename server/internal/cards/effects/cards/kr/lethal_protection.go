package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lethal Protection — Sorcery {3}{B} (Marvel's Spider-Man, #4460):
//
//	"Destroy target creature. Return up to one target creature card from
//	 your graveyard to your hand."
//
// Two target clauses, both chosen as the spell is cast. Each half checks
// its own target as the spell resolves (CR 608.2b): a creature that has
// gone, or gained protection, loses only the destroy, and a card that left
// the graveyard loses only the return; the spell still does the other half.
// The destroy happens first, so the creature it destroys is not a legal
// pick for the return (it was not in the graveyard when targets were
// chosen).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49a7c053-81e9-47b7-af5c-c62cd0389ed4",
		Name:         "Lethal Protection",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.TargetCardInGraveyard("up to one target creature card from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(0, 1),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				victim, legal := ctx.ClauseTarget(0)
				if !legal || victim.Kind != game.TargetCard {
					return nil
				}
				return effects.DestroyTarget{Target: victim.ID}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				card, picked := ctx.ClauseTarget(1)
				if !picked || card.Kind != game.TargetCard {
					return nil
				}
				return effects.ReturnFromGraveyard{Target: card.ID, Dest: game.ZoneHand}.Apply(ctx)
			}),
		),
	})
}
