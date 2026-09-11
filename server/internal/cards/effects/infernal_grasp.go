package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Infernal Grasp — Instant for {1}{B}:
//
//	"Destroy target creature. You lose 2 life."
//
// Unconditional two-mana instant-speed creature removal, which is
// why it is played over every "destroy target nonblack creature"
// ever printed. The two life is not a cost and not a drawback you
// can decline — it is the second sentence of the effect and it
// happens on resolution whether or not the destroy did anything.
//
// LIFE LOSS, NOT DAMAGE. ChangePlayerLifeForEffect with a negative
// delta, not DealDamageToPlayerForEffect: the card says "you lose 2
// life", so no damage-prevention shield, no damage doubler and no
// lifelink sees it, and there is no source creature to attribute it
// to. Routing it through damage would be a different card.
//
// The destroy runs first and the life loss second, in printed order.
// A target that left the battlefield between announce and resolution
// has already fizzled the whole spell (CR 608.2b runs before
// OnResolve), so the 2 life is only paid when the spell actually
// resolves.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID: "94f0a572-e91c-4b56-a5d1-6cbbeabd210d",
		Name:     "Infernal Grasp",
		Targets:  TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard {
				if err := (DestroyTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return GainLife{Player: item.Controller, Amount: -2}.Apply(ctx)
		},
	})
}
