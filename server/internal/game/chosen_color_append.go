package game

import (
	"strings"

	"github.com/google/uuid"
)

// AppendChosenColorForEffect records one more colour on the permanent
// `sourceID` for an "as this enters, choose two colors" instruction (Seal of
// the Guildpact, CR 105.4): the colours are the letters of Card.ChosenColor
// in the order chosen, so a card that asks twice ends up holding two. A
// colour already held is not added twice. A permanent that has left the
// battlefield has nothing to record on. The layer version is bumped because
// the colours are an input to the permanent's cost modifiers.
//
// Caller must hold g.mu in write mode (a keyed tail runs under it).
func (g *Game) AppendChosenColorForEffect(sourceID uuid.UUID, color string) {
	color = strings.ToUpper(strings.TrimSpace(color))
	if color == "" {
		return
	}
	i := findCardOnBattlefield(g, sourceID)
	if i < 0 || strings.Contains(g.Battlefield.Cards[i].ChosenColor, color) {
		return
	}
	g.Battlefield.Cards[i].ChosenColor += color
	g.layerVersion.Add(1)
}
