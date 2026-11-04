package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b4_helpers.go — shared bodies for Star Trek batch 4 (#4973).

// trekB4AltCostKeyBeamMeUp is beam me up's offer key: on the wire as
// `alternative_cost`, on StackItem.AltCost and on a permanent's
// CastProvenance. An on-disk identity: never renamed, never reused.
const trekB4AltCostKeyBeamMeUp = "beam_me_up"

// trekB4BeamMeUp is "Beam me up {cost} (You may cast this card from your
// graveyard for {cost} if you also return a creature you control to its
// owner's hand. Then exile this spell.)" — Pure Energy, There's Coffee in
// that Nebula.
//
// Flashback's place and exile (FromZone graveyard, ExileOnLeavingStack,
// CR 702.34a's shape) with Daze's bounce as the card component of the
// price: the caster names one creature they control in alt_cost_ids and
// it returns to its owner's hand as the cost is paid, with the spell
// already on the stack (CR 601.2a, 601.2h). A caster who controls no
// creature cannot claim the offer. The card file still opens the zone:
//
//	CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
//	AlternativeCosts: []game.AlternativeCost{trekB4BeamMeUp("{2}{W}")},
func trekB4BeamMeUp(cost string) game.AlternativeCost {
	ac := Flashback(cost)
	ac.Key = trekB4AltCostKeyBeamMeUp
	ac.Label = "Beam me up " + cost
	ac.ReturnToHand = PermanentYouControl("a creature you control", Creature())
	ac.PayLabel = "a creature you control"
	return ac
}

// trekB4CreatureTypesAmongNonBorg is Saavik's "the number of creature
// types among non-Borg creatures you control": each creature type at
// least one such creature has counts once, read through the layered view
// (Card.HasSubtype), so a changeling that is not a Borg would have them
// all — but a changeling is a Borg, so it never counts. `borg` is the
// word as the source's text now reads it (CR 612.2).
func trekB4CreatureTypesAmongNonBorg(g *game.Game, controller uuid.UUID, borg string) int {
	var mine []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && !c.HasSubtype(borg) {
			mine = append(mine, c)
		}
	}
	n := 0
	for _, t := range game.AllCreatureTypes {
		for _, c := range mine {
			if c.HasSubtype(t) {
				n++
				break
			}
		}
	}
	return n
}

// trekB4EquipmentOrSpacecraftEnteredUnderYou is Sulu's intervening "if an
// Equipment or Spacecraft entered the battlefield under your control this
// turn", read off the turn's entry tally (the subtypes each permanent had
// as it entered).
func trekB4EquipmentOrSpacecraftEnteredUnderYou(g *game.Game, controller uuid.UUID) bool {
	return g.EnteredWithSubtypeThisTurn(controller, "Equipment") > 0 ||
		g.EnteredWithSubtypeThisTurn(controller, "Spacecraft") > 0
}
