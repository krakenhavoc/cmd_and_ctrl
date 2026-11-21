package game

import "github.com/google/uuid"

// noted_mana_type.go — "Note the type of mana spent to pay this activation
// cost" (Jeweled Amulet, #5948).
//
// The noted type is a single mana type of the permanent that noted it (W, U,
// B, R, G, or C for colorless), kept where a chosen color is kept
// (Card.ChosenColor): per instance, carried by the snapshot, cleared when the
// permanent leaves the battlefield, and shown on the tile as the same chip. A
// permanent with no such note has "" and a later note replaces an earlier one,
// which is "last noted".

// NoteManaTypeForEffect records color (W, U, B, R, G or C) as the mana type
// the permanent sourceID last noted. A source that is not on the battlefield
// notes nothing. Caller must hold g.mu.
func (g *Game) NoteManaTypeForEffect(sourceID uuid.UUID, color string) {
	switch color {
	case "W", "U", "B", "R", "G", "C":
	default:
		return
	}
	if i := findCardOnBattlefield(g, sourceID); i >= 0 {
		g.Battlefield.Cards[i].ChosenColor = color
		g.layerVersion.Add(1)
	}
}
