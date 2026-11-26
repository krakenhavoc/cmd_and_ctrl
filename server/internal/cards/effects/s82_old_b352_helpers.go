package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB352CreatureOrAttachedToOne is "creatures and permanents attached to
// creatures" (End Hostilities): a creature, or a permanent (an Aura, an
// Equipment, a Fortification) attached to a creature on the battlefield, read
// through the layered view.
func OldB352CreatureOrAttachedToOne(g *game.Game, _ uuid.UUID, c game.Card) bool {
	if c.IsCreature() {
		return true
	}
	if !c.IsAttached() {
		return false
	}
	host, ok := g.LookupCardForEffect(c.AttachedTo.ID)
	return ok && host.IsCreature() && OnBattlefield(g, host.InstanceID)
}
