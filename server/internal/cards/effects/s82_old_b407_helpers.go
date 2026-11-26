package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB407DyingWishDrain is "target player loses X life and you gain X life,
// where X is its power" for the creature that died (Dying Wish): the target is
// the trigger's first target if it is still legal (CR 608.2b), X is the dead
// creature's last-known power (CR 603.10a), and the Aura's controller gains.
func S82OldB407DyingWishDrain(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if item.Trigger == nil {
		return nil
	}
	target := TargetedPlayer(ctx)
	if target == uuid.Nil {
		return nil
	}
	info, ok := g.LastKnownPermanentForEffect(item.Trigger.Event.CardID)
	if !ok || info.Power <= 0 {
		return nil
	}
	if err := g.ChangePlayerLifeForEffect(item.SourceCardID, target, -info.Power); err != nil {
		return err
	}
	return GainLife{Player: ctx.Controller(), Amount: info.Power}.Apply(ctx)
}
