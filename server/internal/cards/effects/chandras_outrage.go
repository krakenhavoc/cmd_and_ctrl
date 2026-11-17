package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chandra's Outrage — Instant {2}{R}{R} (Archenemy, #5580):
//
//	"Chandra's Outrage deals 4 damage to target creature and 2 damage to
//	 that creature's controller."
//
// Shocker, Unshakable's body as a spell. "That creature's controller" is
// read before the damage, while the creature is still on the battlefield,
// so a creature the 4 damage kills still has its controller hit. A target
// that has become illegal stops the spell, and no one is dealt damage
// (CR 608.2b, the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "47437865-0032-4f47-b0ab-034cc841bb84",
		Name:         "Chandra's Outrage",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			victim, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			controller := victim.Controller
			if err := (DealDamage{Source: ctx.Source(), Target: id, Amount: 4}).Apply(ctx); err != nil {
				return err
			}
			return DealDamage{Source: ctx.Source(), Target: controller, Amount: 2}.Apply(ctx)
		},
	})
}
