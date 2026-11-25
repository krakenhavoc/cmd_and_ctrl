package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b190_helpers.go — shared vocabulary for the S82 older-sets batch 190
// cards (#5525).

// InstantsAndSorceriesOwnedInExileAndGraveyard is "the total number of instant
// and sorcery cards you own in exile and in your graveyard" (Beacon Bolt): the
// owner's graveyard, plus the face-up exiled cards they own. A token is not a
// card.
func InstantsAndSorceriesOwnedInExileAndGraveyard(g *game.Game, owner uuid.UUID) int {
	n := FandanielInstantsAndSorceries(g, owner)
	if g == nil || g.Exile == nil {
		return n
	}
	for i := range g.Exile.Cards {
		c := &g.Exile.Cards[i]
		if c.Owner == owner && !c.FaceDown && !c.IsToken() && (c.IsInstant() || c.IsSorcery()) {
			n++
		}
	}
	return n
}
