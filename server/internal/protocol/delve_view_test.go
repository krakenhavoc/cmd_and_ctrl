package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// delve_view_test.go — ADR 0100 §4: a card with delve ships `delve` on
// its cast surface — the caster's graveyard as options, in the
// engine's payment order, and the default announcement's budget as a
// hint. Both come from the engine (DelveOptionsForEffect and
// CastPrice.DelveBudget), so the picker can never offer a card or a
// count the validator refuses.

const delveViewOracle = "test-delve-view"

func installDelveForView(t *testing.T) {
	t.Helper()
	prev := game.CatalogDelve
	game.CatalogDelve = func(key string) bool {
		if key == delveViewOracle {
			return true
		}
		if prev != nil {
			return prev(key)
		}
		return false
	}
	t.Cleanup(func() { game.CatalogDelve = prev })
}

func TestDelveIsProjectedForTheCaster(t *testing.T) {
	installDelveForView(t)
	g := buildActiveGame(t)
	me := g.Seats[0]
	cruise := game.NewCard("Test Cruise", me.ID)
	cruise.TypeLine = "Sorcery"
	cruise.ManaCost = "{7}{U}"
	cruise.OracleID = delveViewOracle
	cruise.KnownBy = map[uuid.UUID]bool{me.ID: true}
	me.Hand.PushTop(cruise)

	land := game.NewCard("Dead Island", me.ID)
	land.TypeLine = "Basic Land — Island"
	spell := game.NewCard("Old Spell", me.ID)
	spell.TypeLine = "Instant"
	me.Graveyard.PushTop(spell)
	me.Graveyard.PushTop(land)

	v := FilterViewFor(ViewOfGame(g), me.ID.String())
	c := handCardView(t, v, 0, cruise.InstanceID)
	if c.Delve == nil || c.Delve.Options == nil {
		t.Fatalf("delve = %+v, want the caster's graveyard", c.Delve)
	}
	if c.Delve.Max != 7 {
		t.Errorf("delve.max = %d, want 7", c.Delve.Max)
	}
	want := []string{land.InstanceID.String(), spell.InstanceID.String()}
	got := c.Delve.Options.Cards
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("delve options = %v, want the land first, then the spell: %v", got, want)
	}
}

func TestDelveIsAbsentWithoutDelve(t *testing.T) {
	installDelveForView(t)
	g := buildActiveGame(t)
	me := g.Seats[0]
	plain := game.NewCard("Plain Sorcery", me.ID)
	plain.TypeLine = "Sorcery"
	plain.ManaCost = "{2}{U}"
	plain.OracleID = "test-no-delve-" + uuid.NewString()
	plain.KnownBy = map[uuid.UUID]bool{me.ID: true}
	me.Hand.PushTop(plain)
	v := FilterViewFor(ViewOfGame(g), me.ID.String())
	if c := handCardView(t, v, 0, plain.InstanceID); c.Delve != nil {
		t.Fatalf("delve = %+v on a card without delve", c.Delve)
	}
}
