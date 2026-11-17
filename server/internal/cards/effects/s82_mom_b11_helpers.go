package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b11_helpers.go — shared reads for March of the Machine, S82
// batch 11 (#5477).

// s82MomB11ArtifactOrCreatureDiedThisTurn is "an artifact or creature
// was put into a graveyard from the battlefield this turn" (Ichor Shade):
// any controller, tokens included, read off the turn's EventLTB records
// with the permanent's types as it last existed (CR 603.10a).
func s82MomB11ArtifactOrCreatureDiedThisTurn(g *game.Game) bool {
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard || ev.CardID == uuid.Nil {
			continue
		}
		c, _ := g.LookupCardForEffect(ev.CardID)
		if leftAsType(ev, c, "artifact") || leftAsType(ev, c, "creature") {
			return true
		}
	}
	return false
}

// s82MomB11ArtifactOrCreatureYouControlDied is the When for "an artifact
// or creature you control is put into a graveyard from the battlefield":
// types and controller as the permanent last existed on the battlefield.
func s82MomB11ArtifactOrCreatureYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	_, ok := b21ArtifactOrCreatureYouControlDied(ev, source, g)
	return ok
}

// s82MomB11AttackingPhyrexiansPower is the total power of the Phyrexian
// creatures `controller` has attacking right now.
func s82MomB11AttackingPhyrexiansPower(g *game.Game, controller uuid.UUID, phyrexian string) int {
	total := 0
	for _, id := range AttackingCreatures(g) {
		c, ok := g.LookupCardForEffect(id)
		if ok && c.Controller == controller && c.HasSubtype(phyrexian) {
			total += c.CurrentPower()
		}
	}
	return total
}
