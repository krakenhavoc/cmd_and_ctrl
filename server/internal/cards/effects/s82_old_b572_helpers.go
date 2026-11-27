package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b572_helpers.go — shared shape of S82 OLD batch 572 (#5599).

// OldB572AurasAttachedTo counts the Auras attached to the permanent `id`, for
// "gets +1/+1 for each Aura attached to it" (Gatherer of Graces). It walks the
// battlefield directly, as IsEnchantedByAura does, so a layer-7 static can ask
// it for its own permanent on every pass.
func OldB572AurasAttachedTo(g *game.Game, id uuid.UUID) int {
	if g == nil || g.Battlefield == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		a := &g.Battlefield.Cards[i]
		if a.IsAttachedTo(id) && a.IsAura() {
			n++
		}
	}
	return n
}
