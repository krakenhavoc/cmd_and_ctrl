package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dskRT2Delirium puts the first n of four card types into p's graveyard
// (four is delirium, three is not) and drops the layer cache, as a zone
// move would.
func DskRT2Delirium(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() {
		DskHL2SeedTypes(p, n)
		g.BumpLayerVersionForTest()
	})
}

// dskRT2Activate pays `mana` and activates ability idx of a permanent.
func DskRT2Activate(g *game.Game, who, id uuid.UUID, idx int, mana string, params game.ActivateAbilityParams) error {
	if mana != "" {
		if err := g.AddManaForEffect(who, uuid.Nil, mana); err != nil {
			return err
		}
	}
	params.Strict = true
	return g.ActivateCatalogAbility(who, id, idx, params)
}

// dskRT2Tapped reports whether a battlefield permanent is tapped.
func DskRT2Tapped(g *game.Game, id uuid.UUID) bool {
	c, ok := BattlefieldCard(g, id)
	return ok && c.Tapped
}
