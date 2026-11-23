package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// tinybones_bauble_burglar_test.go — #2179: a standing permission over
// exiled cards you don't own that carry a stash counter.

const TinybonesOracle = "fd335b2f-e6a2-45ac-949e-a67278b21cb6"

// stashed puts a card in exile owned by `owner`, carrying `n` stash
// counters, known to every seat.
func Stashed(g *game.Game, owner uuid.UUID, name, typeLine, cost string, n int) uuid.UUID {
	id := uuid.New()
	known := map[uuid.UUID]bool{}
	for _, p := range g.Seats {
		known[p.ID] = true
	}
	c := game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Owner: owner, Controller: owner, KnownBy: known}
	if n > 0 {
		c.Counters = map[string]int{"stash": n}
	}
	g.Exile.PushTop(c)
	return id
}

func ExileCard(g *game.Game, id uuid.UUID) (game.Card, bool) {
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}

func CastStash(g *game.Game, who *game.Player, id uuid.UUID) error {
	return g.CastSpell(who.ID, id, game.CastSpellParams{FromZone: "exile", Strict: true})
}

func PayAnyType(t *testing.T, g *game.Game, who *game.Player, mana string) {
	t.Helper()
	if err := g.AddManaForEffect(who.ID, uuid.Nil, mana); err != nil {
		t.Fatal(err)
	}
}

func EnumeratedFor(g *game.Game, seat, card uuid.UUID) bool {
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type == legal.TypeCastSpell && m.Source == card {
			return true
		}
	}
	return false
}

func ViewCastable(t *testing.T, g *game.Game, viewer, card uuid.UUID) bool {
	t.Helper()
	v := protocol.ViewOfGameFor(g, viewer.String())
	for _, c := range v.Exile.Cards {
		if c.InstanceID == card.String() {
			return c.CastableHere
		}
	}
	t.Fatalf("card %s is not in the exile view", card)
	return false
}
