package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gapB51CountersPlacedOnThis is "whenever one or more +1/+1 counters are put
// on this creature": one trigger per placement event that added a counter.
func GapB51CountersPlacedOnThis(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Target == source.InstanceID && B33CountersPlacedDelta(ev, game.CounterPlusOne, g) > 0
}

// gapB51EldraziSpawn is "create a 0/1 colorless Eldrazi Spawn creature token
// with 'Sacrifice this token: Add {C}.'" for the ability's controller.
func GapB51EldraziSpawn(g *game.Game, item *game.StackItem) error {
	return CreateToken{Controller: item.Controller, Template: EldraziSpawnToken(), N: 1}.Apply(NewContext(g, item))
}

// gapB51CastTriggerWithoutColorless is "When you cast this spell, if {C} wasn't
// spent to cast it": the spell's own payment is read at cast and again as the
// ability resolves (CR 603.4); the record does not change in between.
func GapB51NoColorlessSpent(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.CardID == source.InstanceID && ev.Actor == source.Controller &&
		g.StackItemPaidForEffect(ev.CardID).Spent().Count("C") == 0
}
