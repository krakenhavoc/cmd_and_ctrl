package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b25_helpers.go — shared bodies for Foundations Jumpstart batch 25
// (#5929, S82).

// angelicCubTargetedFirstTime is "whenever this creature becomes the target
// of a spell or ability for the first time each turn": Valiant's scan with
// no restriction on who controls the spell or ability. Only the Cub's own
// events count; a zone change of the card ends the object that was
// targeted (CR 400.7). One spell naming the Cub twice emits two events
// (CR 115.3) and the second is not the first time.
func angelicCubTargetedFirstTime(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || source == nil || ev.CardID == uuid.Nil || ev.CardID != source.InstanceID {
		return false
	}
	targeted := false
	for _, prev := range g.EventsThisTurn() {
		if prev.Seq >= ev.Seq {
			break
		}
		if prev.CardID != ev.CardID {
			continue
		}
		switch prev.Kind {
		case game.EventZoneMove, game.EventETB, game.EventLTB:
			targeted = false
		case game.EventBecomesTarget:
			targeted = true
		}
	}
	return !targeted
}

// archfiendsVesselReturned is the trigger condition for Archfiend's Vessel:
// the creature entered from its controller's graveyard, or its controller
// cast it from their graveyard.
func archfiendsVesselReturned(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if !Self(ev, source, lki, g) {
		return false
	}
	if EnteredFromYourGraveyard(ev, source, lki, g) {
		return true
	}
	cast, known := source.CastByItsController()
	return known && cast && source.Provenance.FromZone == game.ZoneGraveyard
}
