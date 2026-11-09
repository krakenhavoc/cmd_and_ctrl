package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b52_helpers.go — shared bodies for Standard set batch 52 (#4452).

// std52YouTappedUntappedOpposingCreature is "whenever you tap an untapped
// creature an opponent controls" (Solitary Sanctuary): the event is an
// EventTapCard (emitted only for a permanent that was untapped, CR 701.26a),
// the player who tapped it is the source's controller (Event.TappedBy
// TapTargetForEffectBy stamps), and the creature belongs to an
// opponent. A tap by an opponent's ability, an attack, or a mana ability
// leaves the TappedBy empty or someone else, so none of them count.
func std52YouTappedUntappedOpposingCreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventTapCard || ev.TappedBy != source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.Controller != source.Controller && c.Controller != uuid.Nil
}
