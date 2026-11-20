package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// crtBlock puts `attacker` in combat attacking `defender` and has
// `blocker` block it.
func CrtBlock(g *game.Game, blocker, attacker, defender uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case attacker:
				g.Battlefield.Cards[i].AttackingTarget = defender
			case blocker:
				g.Battlefield.Cards[i].BlockingTarget = attacker
			}
		}
	})
}

func CrtActivateRaw(t *testing.T, g *game.Game, who, source uuid.UUID, index int, targets []game.TargetRef, mana ...string) error {
	t.Helper()
	B06AddMana(g.PlayerByIDForEffect(who), mana...)
	return g.ActivateCatalogAbility(who, source, index, game.ActivateAbilityParams{Targets: targets, Strict: true})
}
