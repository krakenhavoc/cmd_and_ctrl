package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seedship Impact — Instant {1}{G} (Edge of Eternities, #4461):
//
//	"Destroy target artifact or enchantment. If its mana value was 2 or
//	 less, create a Lander token."
//
// Tainted Treats' body with a Lander. The mana value is the target's as
// the spell resolves, read before it is destroyed (CR 608.2h). The Lander
// does not depend on the destruction: an indestructible target with a
// small mana value still makes one, since the text says "if its mana
// value was", not "if you do". A target that has become illegal makes
// the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c6d89c4a-351f-4d63-ba57-b978a695191e",
		Name:         "Seedship Impact",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			small := c.ManaValue() <= 2
			if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			if !small {
				return nil
			}
			return CreateToken{Controller: ctx.Controller(), Template: b16LanderToken(), N: 1}.Apply(ctx)
		},
	})
}
