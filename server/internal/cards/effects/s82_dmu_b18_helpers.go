package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// youCastASpellThatTargetsThis is heroic (Rosnakht, Heir of Rohgahh):
// "whenever you cast a spell that targets <this creature>". The spell's
// targets are read off its stack item as it is cast (CR 601.2c).
func youCastASpellThatTargetsThis(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	item := g.StackItemForEffect(ev.CardID)
	if item == nil || item.Kind != game.StackItemSpell {
		return false
	}
	for _, t := range item.Targets {
		if t.Kind == game.TargetCard && t.ID == source.InstanceID {
			return true
		}
	}
	return false
}

// greatestPowerAmongYourWurms is Baru, Wurmspeaker's X: the greatest power
// among Wurms the activator controls, 0 with none (CR 107.1b).
func greatestPowerAmongYourWurms(q game.CostQuery) int {
	if q.Game == nil || q.Game.Battlefield == nil {
		return 0
	}
	best := 0
	for i := range q.Game.Battlefield.Cards {
		c := &q.Game.Battlefield.Cards[i]
		if c.Controller != q.Controller || !c.IsCreature() || !c.HasSubtypeWord(q.Source.TextSubtype("Wurm")) {
			continue
		}
		best = max(best, c.PowerForComparison())
	}
	return best
}

// nontokenCreaturesYouControlledThatDiedThisTurn counts the nontoken creatures
// `you` controlled as they died this turn (Tobias, Doomed Conqueror): each
// battlefield-to-graveyard move of the turn, read off its last-known
// information (CR 608.2h), so a creature that has since left the graveyard
// still counts and Tobias, who died to put the trigger there, counts himself.
func nontokenCreaturesYouControlledThatDiedThisTurn(g *game.Game, you uuid.UUID) int {
	n := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
			continue
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if !ok || c.IsToken() || !leftAsType(ev, c, "creature") || leftUnderControlOf(ev, c) != you {
			continue
		}
		n++
	}
	return n
}
