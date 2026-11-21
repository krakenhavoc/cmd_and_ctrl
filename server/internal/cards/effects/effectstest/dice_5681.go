package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dice_5681_test.go — #5681: the static "roll that many dice plus one and
// ignore the lowest roll" (Pixie Guide, Barbarian Class), a roll's
// modifier beside its natural result (CR 706.2), and the cards that read
// one or the other.

const (
	PixieGuideOracle          = "b6003419-1a2a-46f3-b32c-47b825324f78"
	BarbarianClassOracle      = "59faa1b9-17aa-4f2c-a8f8-17ab50392b36"
	FaridehOracle             = "faf5c18c-2642-4df3-a21f-3b237527e6a6"
	NetheresePuzzleWardOracle = "561e8ecc-1f2b-4fd4-8f92-a0acea449816"
	CriticalHitOracle         = "992e6845-1c6c-4b74-839c-317823bde15b"
)

// d5681Key finds an RNG key under which seat 0's next n d<sides> from
// `source` are rolls `want` accepts, by peeking on a throwaway table.
func D5681Key(t *testing.T, source uuid.UUID, sides, n int, want func(rolls []int) bool) [32]byte {
	t.Helper()
	for k := 0; k < 256; k++ {
		g := NewCatalogGame(t)
		g.SetRNGKeyForTest(RandomModesKey(byte(k)))
		var rolls []int
		g.WithWriteLock(func() {
			rolls, _ = g.RollDiceForEffect(game.RandomDraw{Player: g.Seats[0].ID, Source: source}, sides, n)
		})
		if len(rolls) == n && want(rolls) {
			return RandomModesKey(byte(k))
		}
	}
	t.Fatal("no key in 0..255 gives the wanted rolls")
	return [32]byte{}
}

// d5681Roll has player roll n d<sides> plus modifier from source through
// the dice door, then lets the triggers it caused resolve.
func D5681Roll(t *testing.T, g *game.Game, player, source uuid.UUID, sides, n, modifier int) {
	t.Helper()
	var err error
	g.WithWriteLock(func() {
		err = g.RollDiceThenForEffect(game.DiceRoll{Player: player, Source: source, Sides: sides, N: n, Modifier: modifier})
	})
	if err != nil {
		t.Fatalf("roll: %v", err)
	}
	g.RunStateChecksForTest()
	PassPriorityAroundTable(t, g)
}

// d5681Rolled lists the EventRollDie events after event index from.
func D5681Rolled(g *game.Game, from int) []game.Event {
	var out []game.Event
	g.ReadSnapshot(func() {
		for _, ev := range g.Events[from:] {
			if ev.Kind == game.EventRollDie {
				out = append(out, ev)
			}
		}
	})
	return out
}

func D5681InZone(g *game.Game, id uuid.UUID, kind game.ZoneKind) bool {
	in := false
	g.WithWriteLock(func() {
		z := g.FindCardZoneForEffect(id)
		in = z != nil && z.Kind == kind
	})
	return in
}
