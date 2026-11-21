package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gapB87YouCastFromOutsideYourHand is the Paradox condition (Doctor Who): the
// source's controller cast a spell from anywhere other than their hand, read off
// the zone the spell left.
func GapB87YouCastFromOutsideYourHand(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventCast && ev.Actor == source.Controller &&
		ev.OldZone != "" && ev.OldZone != game.ZoneHand
}

// gapB87UntapCreaturesWithACounter is "untap each creature you control with a
// counter on it": a counter of any kind, read as the ability resolves.
func GapB87UntapCreaturesWithACounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return B16UntapAllYouControlMatching(ctx, item.Controller, func(c game.Card) bool {
		if !c.IsCreature() {
			return false
		}
		for _, n := range c.Counters {
			if n > 0 {
				return true
			}
		}
		return false
	})
}

// gapB87HistoricPermanentsYouControl counts the historic permanents (artifacts,
// legendaries, Sagas) a player controls now.
func GapB87HistoricPermanentsYouControl(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && B09IsHistoric(c) {
			n++
		}
	}
	return n
}

// gapB87NoncreatureNonEquipmentArtifactsYouControl is Dan Lewis's set for the
// layer-4 type addition: the source's controller's artifacts that are neither
// creatures nor Equipment, read post-layer (the addition runs before any later
// layer, so an artifact that is already an Equipment is left alone).
func GapB87NoncreatureNonEquipmentArtifactsYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.Controller == source.Controller && target.HasCardType("artifact") && !target.IsCreature() &&
		!target.HasSubtype("Equipment")
}

// gapB87DanLewisGrantSet is the same set for the layer-6 grants. One effect that
// applies in several layers keeps applying to the set of objects it began with
// (CR 613.6), so an artifact Dan Lewis himself made an Equipment in layer 4 must
// still receive the grants: "non-Equipment" is judged on the object as it was
// before that addition, here its printed subtypes.
func GapB87DanLewisGrantSet(target *game.Card, _ *game.Game, source *game.Card) bool {
	if target.Controller != source.Controller || !target.HasCardType("artifact") || target.IsCreature() {
		return false
	}
	_, _, printed := game.ParseTypeLine(target.TypeLine)
	for _, st := range printed {
		if st == "Equipment" {
			return false
		}
	}
	return true
}
