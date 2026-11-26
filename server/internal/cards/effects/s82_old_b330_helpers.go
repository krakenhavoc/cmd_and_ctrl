package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b330_helpers.go — card-side helpers of set OLD, batch 330 (Dragons
// of Tarkir: Bolster, S82, #5551).

// DestroyTargetThenBolster is "Destroy target <thing>. Bolster N." (Enduring
// Victory, Pinion Feast): the first target is destroyed if it is still legal,
// then its controller bolsters (CR 701.39a), in printed order (CR 608.2c). The
// creature is chosen as the spell resolves, with no target.
func DestroyTargetThenBolster(n int) func(*game.StackItem, *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard && ctx.IsTargetLegal(item.Targets[0]) {
			if err := (DestroyTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
		}
		return Bolster{N: n}.Apply(ctx)
	}
}

// PlusOneCountersOnOtherCreaturesYouControl is the number of +1/+1 counters on
// the creatures source's controller controls other than the one the aura
// source is attached to (Gleam of Authority). It walks the battlefield
// directly, as PumpAttachedPer's counters must (a layer pass may hold only the
// read lock).
func PlusOneCountersOnOtherCreaturesYouControl(g *game.Game, source *game.Card) int {
	host := g.AttachedHostOf(source)
	if host == nil || g.Battlefield == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != source.Controller || c.InstanceID == host.InstanceID || !c.IsCreature() {
			continue
		}
		n += c.Counters[game.CounterPlusOne]
	}
	return n
}
