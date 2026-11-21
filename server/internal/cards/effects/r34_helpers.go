package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// r34PlusOnePerAuraAndEquipmentOnSource is a layer-7c Apply that gives
// +1/+1 for each Aura and Equipment attached to the source permanent
// (Gunner Conscript on itself, Heavenly Blademaster on the others).
func r34PlusOnePerAuraAndEquipmentOnSource(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
	n := 0
	for _, a := range g.BattlefieldCardsForEffect() {
		if a.IsAttachedTo(source.InstanceID) && (a.IsAura() || a.HasSubtype("Equipment")) {
			n++
		}
	}
	c.Power += n
	c.Toughness += n
}

// r34UpToOneCreatureYouControl is a ChoosePermanents candidate list for
// "you may choose a creature you control": every creature the chooser
// controls, zero or one of them (Finality, Birthing Ritual).
func r34UpToOneCreatureYouControl(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	ids := permanentsControlledByMatching(g, of, Creature())
	if len(ids) == 0 {
		return nil, 0, 0
	}
	return ids, 0, 1
}
