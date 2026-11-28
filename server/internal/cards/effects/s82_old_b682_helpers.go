package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b682_helpers.go — Mercadian Masques, "Activated abilities" (S82 OLD
// batch 682, #5619).

// OldB682ExileTopFaceDownLinked is "exile the top card of your library face
// down" for the resolving trigger's source, linked to that object (Kyren
// Archive; CR 406.3, 607.2a): nobody may look at the card. An empty library
// does nothing.
func OldB682ExileTopFaceDownLinked(g *game.Game, item *game.StackItem) error {
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Library == nil || len(p.Library.Cards) == 0 {
		return nil
	}
	top := p.Library.Cards[len(p.Library.Cards)-1].InstanceID
	_, err := g.ExileFaceDownLinkedForEffect(item.SourceCardID, item.Controller, top)
	return err
}
