package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2S14LandCardsInYourGraveyard is "for each land card in your
// graveyard" as a cost count (CR 601.2f). The spell being cast is on the
// stack (CR 601.2a) and is never counted.
func cmdr2S14LandCardsInYourGraveyard(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	p := q.Game.PlayerByIDForEffect(q.Controller)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.InstanceID != q.Card.InstanceID && c.IsLand() {
			n++
		}
	}
	return n
}

// cmdr2S14SubtypeLandEntersUnderYourControl builds "a [subtype] you
// control enters" for a land of that subtype, read through the layered
// view.
func cmdr2S14SubtypeLandEntersUnderYourControl(subtype string) When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		c, ok := enteredUnderYourControl(ev, source, g, false)
		return ok && c.IsLand() && c.HasSubtype(subtype)
	}
}

// cmdr2S14DesertCardToYourGraveyard is "a Desert card is put into your
// graveyard from anywhere" (the Gitrog Monster's reading).
func cmdr2S14DesertCardToYourGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.NewZone != game.ZoneGraveyard {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsLand() && c.HasSubtype("Desert") && c.Owner == source.Controller
}

// cmdr2S14LandCardsToYourGraveyardEvents are the event kinds that can
// put a card into a graveyard.
var cmdr2S14GraveyardEvents = []game.EventKind{game.EventMill, game.EventZoneMove, game.EventLTB}

// cmdr2S14ElementalYouControlDied is "an Elemental you control dies": a
// creature with the Elemental subtype, read from its last-known
// characteristics, that left the battlefield for a graveyard under your
// control (CR 700.4).
func cmdr2S14ElementalYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	if !ok || leftUnderControlOf(ev, dead) != source.Controller {
		return false
	}
	was, known := ev.WasSubtype("Elemental")
	if !known {
		was = dead.HasSubtype("Elemental")
	}
	return was
}
