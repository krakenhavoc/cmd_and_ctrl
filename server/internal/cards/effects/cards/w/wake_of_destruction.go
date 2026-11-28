package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wake of Destruction — Sorcery {3}{R}{R}{R} (Urza's Destiny, S82 #5621):
//
//	"Destroy target land and all other lands with the same name as that land."
//
// Bile Blight's group: the target is a real target (CR 115.1), the others are not,
// so hexproof on one copy changes nothing. The set is the target land plus every
// other land sharing its post-layer name as the spell resolves; a land that
// enters later is spared. With the target illegal the spell does nothing
// (CR 608.2b). Regeneration and indestructible apply to each land as usual.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "562f7f99-ee67-4a14-b5e2-1a4ae710726d",
		Name:         "Wake of Destruction",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			ctx = ctx.AsGroupMember()
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			for _, id := range effects.S82RprB17SameNameAsTarget(ctx.Game, t.ID) {
				c, ok := ctx.Game.LookupCardForEffect(id)
				if !ok || !c.IsLand() {
					continue
				}
				if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
