package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b105_helpers.go — shared bodies for S82 GAP batch 105 (#5948).

// gapB105AurasAndEquipmentAttached counts the Auras and Equipment attached to
// the card, whoever controls them.
func gapB105AurasAndEquipmentAttached(g *game.Game, host *game.Card) int {
	n := 0
	for _, a := range g.BattlefieldCardsForEffect() {
		if a.IsAttachedTo(host.InstanceID) && (a.IsAura() || a.HasSubtype("Equipment")) {
			n++
		}
	}
	return n
}
