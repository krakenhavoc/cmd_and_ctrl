package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB359YouControlEachCreatureWithTheGreatestPower is "if you control each
// creature on the battlefield with the greatest power" (Might Makes Right):
// among every creature on the battlefield, as the layers now stand (CR 613),
// each one with the best power is controlled by `you`. With no creature at all
// there is nothing to control, so it is false.
func S82OldB359YouControlEachCreatureWithTheGreatestPower(g *game.Game, you uuid.UUID) bool {
	g.RecomputeLayersIfStaleLocked()
	best, any := 0, false
	for _, c := range g.BattlefieldCardsForEffect() {
		if !c.IsCreature() {
			continue
		}
		if p := c.CurrentPower(); !any || p > best {
			best, any = p, true
		}
	}
	if !any {
		return false
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.CurrentPower() == best && c.Controller != you {
			return false
		}
	}
	return true
}

// S82OldB359DefenderControlsNoWall is "if defending player controls no Walls"
// (Siege Dragon): `defender` controls no permanent with the Wall subtype, by
// effective type. A nil defender (an attack aimed at nobody) controls none.
func S82OldB359DefenderControlsNoWall(g *game.Game, defender uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == defender && c.HasSubtypeWord("Wall") {
			return false
		}
	}
	return true
}

// S82OldB359DestroyEquipmentAttachedTo is "destroy all Equipment attached to
// that creature" (Blastfire Bolt): every Equipment on the battlefield whose
// host is `host` right now.
func S82OldB359DestroyEquipmentAttachedTo(ctx *Context, host uuid.UUID) error {
	return DestroyAllMatching{Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.HasSubtype("Equipment") && c.IsAttachedTo(host)
	}}.Apply(ctx)
}

// S82OldB359AbilityOfAnArtifact is "an ability of an artifact" for an
// activation event (Kurkesh): the activated ability's source is an artifact, by
// its effective types, or as it last existed if it has left the battlefield
// (CR 113.7a), so an artifact sacrificed as a cost still counts.
func S82OldB359AbilityOfAnArtifact(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && c.IsArtifact()
}
