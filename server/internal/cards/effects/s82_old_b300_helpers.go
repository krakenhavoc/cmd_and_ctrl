package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b300_helpers.go — shared pieces of the Oath of the Gatewatch batch
// 300 cards (#5545). Both run inside a layer recompute, where the caller may
// hold only the read lock, so they walk g.Battlefield.Cards directly.

// OldB300SharedTypeCreaturesYouControl counts the creatures `source`'s
// controller controls, other than `host`, that share a creature type with it
// (Stoneforge Masterwork). Each candidate is read as the earlier layers have
// left it this pass.
func OldB300SharedTypeCreaturesYouControl(g *game.Game, host, source *game.Card) int {
	if g == nil || g.Battlefield == nil || host == nil || source == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		other := &g.Battlefield.Cards[i]
		if other.InstanceID == host.InstanceID || other.Controller != source.Controller || !other.IsCreature() {
			continue
		}
		if game.SharesCreatureType(host, other) {
			n++
		}
	}
	return n
}

// OldB300ControlsAnEquipment is "as long as you control an Equipment" for
// `source`'s controller (Weapons Trainer).
func OldB300ControlsAnEquipment(g *game.Game, source *game.Card) bool {
	return CountPermanentsYouControl(g, source, func(c *game.Card) bool { return c.HasSubtype("Equipment") }) > 0
}
