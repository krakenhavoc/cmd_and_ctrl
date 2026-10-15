package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cloud_deck_helpers.go — shared by the Cloud, Ex-SOLDIER deck cards
// (#3656): "equipped" readings over the attachment relation.

// isEquippedCreature is "equipped": the creature has an Equipment
// attached to it (CR 301.5). Caller must hold g.mu.
func isEquippedCreature(g *game.Game, c game.Card) bool {
	if !c.IsCreature() {
		return false
	}
	for _, a := range g.BattlefieldCardsForEffect() {
		if a.IsAttachedTo(c.InstanceID) && a.HasSubtype("Equipment") {
			return true
		}
	}
	return false
}

// equippedCreaturesYouControl counts the creatures `controller`
// controls that are equipped; attackingOnly narrows it to those that
// are attacking. Caller must hold g.mu.
func equippedCreaturesYouControl(g *game.Game, controller uuid.UUID, attackingOnly bool) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller || !c.IsCreature() {
			continue
		}
		if attackingOnly && c.AttackingTarget == uuid.Nil {
			continue
		}
		if isEquippedCreature(g, c) {
			n++
		}
	}
	return n
}

// youCastAHistoricSpell is "whenever you cast a historic spell" (CR 700.6):
// the cast event is the source controller's and the spell on the stack is
// an artifact, a legendary or a Saga. Jhoira and Elena share it.
func youCastAHistoricSpell(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && b09IsHistoric(spell)
}
