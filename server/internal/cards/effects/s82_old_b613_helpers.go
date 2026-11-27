package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// SeverSoulResolve is "Destroy target creature. It can't be regenerated. You
// gain life equal to its toughness" (Sever Soul). The toughness is read
// before the destruction (CR 608.2h) and the life is gained even when the
// creature survives; a target that is illegal on resolution stops the spell
// (CR 608.2b).
func SeverSoulResolve(_ *game.StackItem, ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := ctx.Game.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		toughness := c.CurrentToughness()
		if err := (DestroyTarget{Target: t.ID, CantBeRegenerated: true}).Apply(ctx); err != nil {
			return err
		}
		if toughness <= 0 {
			return nil
		}
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), toughness)
	}
	return nil
}
