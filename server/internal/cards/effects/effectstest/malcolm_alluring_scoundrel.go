package effectstest

import (
	"testing"

	"github.com/google/uuid"

	cardslu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/lu"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// malcolm_alluring_scoundrel_test.go — the counter, the loot and the
// threshold. The free cast is asserted all the way through to a
// resolved spell, because a permission that renders but cannot be
// claimed would look identical from the graveyard.

const malcolmOracle = "3bba3b36-f8d7-4fd6-892a-797226210b6e"

// pushMalcolm seeds Malcolm ready to act, with `chorus` counters
// already on him.
func PushMalcolm(t *testing.T, g *game.Game, controller uuid.UUID, chorus int) uuid.UUID {
	t.Helper()
	id := PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Malcolm, Alluring Scoundrel",
		TypeLine: "Legendary Creature — Siren Pirate", OracleID: malcolmOracle,
		ManaCost: "{1}{U}", Power: 2, Toughness: 1,
		Owner: controller, Controller: controller,
	})
	if chorus > 0 {
		g.WithWriteLock(func() {
			if err := g.AddCounterForEffect(id, cardslu.MalcolmChorusCounter, chorus); err != nil {
				t.Fatalf("seed chorus counters: %v", err)
			}
		})
	}
	return id
}
