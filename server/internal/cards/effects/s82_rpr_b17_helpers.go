package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82RprB17OtherLegendaryCreatures counts the OTHER legendary creatures
// `controller` controls (Heroes' Podium). Types and supertypes are read
// post-layer; it runs inside a layer recompute, at layer 7, by which point
// layers 4 has settled every type and supertype.
func s82RprB17OtherLegendaryCreatures(g *game.Game, controller, except uuid.UUID) int {
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID == except || c.Controller != controller || !c.IsCreature() {
			continue
		}
		if isLegendary(c) {
			n++
		}
	}
	return n
}

// s82RprB17SameNameAsTarget is the set Echoing Truth returns: the target
// and every other permanent with the same name, read now, post-layer.
func s82RprB17SameNameAsTarget(g *game.Game, target uuid.UUID) []uuid.UUID {
	t, ok := g.LookupCardForEffect(target)
	if !ok {
		return nil
	}
	name := t.Effective().Name
	ids := []uuid.UUID{target}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == target || c.Effective().Name != name {
			continue
		}
		ids = append(ids, c.InstanceID)
	}
	return ids
}
