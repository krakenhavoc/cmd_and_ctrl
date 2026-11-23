package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// finS3Saga puts a Saga creature onto the battlefield under `owner` with
// `lore` lore counters already on it.
func FinS3Saga(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness, lore int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
		Counters: map[string]int{game.CounterLore: lore},
	})
}

// finS3CastSaga casts a Saga creature from the active seat's hand.
func FinS3CastSaga(t *testing.T, g *game.Game, name, typeLine, oracle, cost string, power, toughness int) uuid.UUID {
	t.Helper()
	return TdCastFromHand(t, g, game.Card{Name: name, TypeLine: typeLine, OracleID: oracle,
		ManaCost: cost, Power: power, Toughness: toughness}, game.CastSpellParams{})
}

// finS3Tap taps a battlefield permanent.
func FinS3Tap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
}
