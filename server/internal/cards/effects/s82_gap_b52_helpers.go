package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gapB52EldraziSpawn is "create a 0/1 colorless Eldrazi Spawn creature token
// with 'Sacrifice this token: Add {C}.'" for the ability's controller.
func gapB52EldraziSpawn(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return CreateToken{Controller: item.Controller, Template: EldraziSpawnToken(), N: n}.Apply(NewContext(g, item))
	}
}

// gapB52CounterOnEachColorlessCreature is "put a +1/+1 counter on each
// colorless creature you control", read as the ability resolves.
func gapB52CounterOnEachColorlessCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range permanentsControlledByMatching(g, item.Controller, And(Creature(), Colorless())) {
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// gapB52OtherCreatures is "other creatures": every creature on the battlefield
// but the source, whoever controls it.
func gapB52OtherCreatures(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.InstanceID != source.InstanceID && target.IsCreature()
}

// gapB52SacrificedAnotherEldrazi is "whenever you sacrifice another Eldrazi".
func gapB52SacrificedAnotherEldrazi(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventSacrifice || ev.CardID == uuid.Nil || ev.CardID == source.InstanceID || ev.Actor != source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.HasSubtype(source.TextSubtype("Eldrazi"))
}
