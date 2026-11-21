package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// earthbend_cards_test.go — #1178, the catalog half of the AVATAR
// keyword action. The engine half (the layers, the duration, the
// delayed return, the CR 614 window on the count) is in
// game/earthbend_test.go; everything here is a real card, cast and
// resolved through the stack.
//
// Three proofs, chosen for what each one adds:
//
//	Earthbending Lesson   the keyword and nothing else — a fixed count
//	Rockalanche           a count read at RESOLUTION, plus flashback
//	The Legend of Kyoshi  the keyword beside a second clause about the
//	                      same land, and the caveat #1179 shipped

const (
	EarthbendingLessonOracle = "5e113f0f-4469-4276-82e1-0a804f3de444"
	RockalancheOracle        = "d1ecfad3-e79a-447b-a932-68ef728eaf1f"
)

// pushEarthbendLand parks an untapped land of the given subtype on the
// battlefield under `owner`, old enough that summoning sickness is
// not why anything below can or cannot attack.
func PushEarthbendLand(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   typeLine,
		Owner:      owner,
		Controller: owner,
	})
}

// earthbentBody reads the land's post-layer power/toughness and
// whether it is a Land Creature with haste.
func EarthbentBody(t *testing.T, g *game.Game, land uuid.UUID) (power, toughness int, creatureLand, hasty bool) {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID != land {
			continue
		}
		return c.CurrentPower(), c.CurrentToughness(),
			c.IsLand() && c.IsCreature(), game.HasKeyword(c, "haste")
	}
	t.Fatalf("the land is not on the battlefield")
	return 0, 0, false, false
}
