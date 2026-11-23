package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// saddle_cards_test.go — one test per Mount (and Guidelight Matrix)
// that #2695 ships, each driving the whole path: saddle in the main
// phase, attack, resolve, and read the result off the board.

const (
	BoundingFelidarOracle  = "b00f316d-7c82-40e6-80e5-9f36434590a3"
	GloryheathLynxOracle   = "28673657-f05f-4b0a-b80c-939773abdefb"
	SeraphicSteedOracle    = "51622e89-aa5e-4dd6-a078-eb45e8ebb434"
	DistrictMascotOracle   = "5fa7cf3a-2a4d-4dd7-890b-49c497570b16"
	BulwarkOxOracle        = "bcd56fbf-fde3-41c2-83f5-9636e92d2893"
	GuardianSunmareOracle  = "3b319540-4db1-43d3-8f4d-f5a276ad78e4"
	OrnheryTumblewaggOrcl  = "f8d7bcbd-9a1d-4fc6-abdd-88a7e01b9411"
	CausticBroncoOracle    = "165bdc68-1da0-47db-a069-978fc4c04b3f"
	SaddleCardsLibraryLand = "Basic Land — Plains"
)

// saddleThenAttack saddles `mount` with `saddlers` in the first main
// phase, resolves the ability, and attacks `defender` with the Mount,
// leaving its attack trigger on the stack.
func SaddleThenAttack(t *testing.T, g *game.Game, owner, defender, mount uuid.UUID, saddlers ...uuid.UUID) {
	t.Helper()
	ToMain(t, g)
	if err := SaddleAbility(t, g, owner, mount, saddlers...); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	PassPriorityAroundTable(t, g)
	if !SaddleCard(t, g, mount).Saddled {
		t.Fatal("setup: the Mount is not saddled")
	}
	DeclareAttack(t, g, defender, mount)
}

func PushLibraryCard(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Library.PushTop(c)
	return c.InstanceID
}

func CountersOf(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return SaddleCard(t, g, id).Counters[game.CounterPlusOne]
}
