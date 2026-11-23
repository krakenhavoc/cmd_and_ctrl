package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_dmu_b19_helpers.go — shared bodies for set DMU, batch 19.

// dmuB19GreatestEquipmentValueNow is the greatest mana value among the
// Equipment attached to `host` right now, and whether any is (a
// battlefield read, for a trigger's "if it's equipped" as it triggers).
func DmuB19GreatestEquipmentValueNow(g *game.Game, host *game.Card) (int, bool) {
	best, any := 0, false
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if !c.IsAttachedTo(host.InstanceID) || !slices.Contains(c.Effective().Subtypes, "Equipment") {
			continue
		}
		any = true
		best = max(best, c.ManaValue())
	}
	return best, any
}

// dmuB19GreatestEquipmentValue is the same read as an ability resolves
// (CR 608.2h): the Equipment attached to the permanent as it last existed
// (PermanentInfo.Attachments), each still the object that was attached.
// One that has left the battlefield or been blinked is not counted. Read
// off the live object while `host` is still there.
func DmuB19GreatestEquipmentValue(g *game.Game, host uuid.UUID) (int, bool) {
	ref, ok := g.PermanentRefForEffect(host)
	if !ok {
		return 0, false
	}
	info, ok := g.PermanentForEffect(ref)
	if !ok {
		return 0, false
	}
	best, any := 0, false
	for _, id := range g.LiveAttachmentsForEffect(info.Attachments) {
		c, ok := g.LookupCardForEffect(id)
		if !ok || !slices.Contains(c.Effective().Subtypes, "Equipment") {
			continue
		}
		any = true
		best = max(best, c.ManaValue())
	}
	return best, any
}
