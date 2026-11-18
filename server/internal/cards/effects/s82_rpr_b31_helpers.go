package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82RprB31ZombiesYouControl is the battlefield Zombie creatures ctx's
// controller controls now. "Zombie" is a word of the resolving ability's
// text (CR 612.2).
func s82RprB31ZombiesYouControl(ctx *Context, controller uuid.UUID) []uuid.UUID {
	zombie := ctx.TextSubtype("Zombie")
	var ids []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.HasSubtype(zombie) {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids
}

// s82RprB31CounterOnEachZombieYouControl is "put a +1/+1 counter on each
// Zombie creature you control", one placement per creature.
func s82RprB31CounterOnEachZombieYouControl(ctx *Context) error {
	for _, id := range s82RprB31ZombiesYouControl(ctx, ctx.Controller()) {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
			continue
		}
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
