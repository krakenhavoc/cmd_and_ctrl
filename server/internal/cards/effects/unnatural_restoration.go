package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unnatural Restoration — Sorcery {1}{G} (Phyrexia: All Will Be One, #5479):
//
//	"Return target permanent card from your graveyard to your hand.
//	 Proliferate."
//
// Regrowth's return with a permanent-card filter, then the player's
// proliferate choice. With its only target gone the spell does not resolve
// at all (CR 608.2b), so there is no proliferate either.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "05d1e797-bda2-4411-a546-919493768650",
		Name:         "Unnatural Restoration",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target permanent card in your graveyard", YouOwn(), Permanent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			if err := returnTargetFromGraveyardToHand(item, ctx); err != nil {
				return err
			}
			return Proliferate{}.Apply(ctx)
		},
	})
}
