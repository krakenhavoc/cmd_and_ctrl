package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cant_be_countered_view_test.go — #1553: the stack item says when the
// engine's one counter gate would refuse to counter it, whether the
// reason is the spell's own printed text or the mana that paid for it.

func pushStackSpell(g *game.Game, name, oracle string, paid game.PaidCost, kind game.StackItemKind) uuid.UUID {
	id := uuid.New()
	caster := g.Seats[1]
	g.WithWriteLock(func() {
		g.Stack.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant",
			OracleID: oracle, Owner: caster.ID, Controller: caster.ID})
		if g.StackMeta == nil {
			g.StackMeta = map[uuid.UUID]*game.StackItem{}
		}
		g.StackMeta[id] = &game.StackItem{ID: id, Kind: kind, Paid: paid,
			Controller: caster.ID, Owner: caster.ID, SourceCardID: id}
	})
	return id
}

func stackItemView(t *testing.T, g *game.Game, id uuid.UUID) StackItemView {
	t.Helper()
	for _, it := range ViewOfGame(g).StackItems {
		if it.ID == id.String() {
			return it
		}
	}
	t.Fatalf("stack item %s not in the view", id)
	return StackItemView{}
}

func TestStackItemShowsCantBeCountered(t *testing.T) {
	const printedOracle = "printed-uncounterable"
	old := game.CatalogCantBeCountered
	game.CatalogCantBeCountered = func(oracle string) bool { return oracle == printedOracle }
	t.Cleanup(func() { game.CatalogCantBeCountered = old })

	g := buildActiveGame(t)
	riderPaid := game.PaidCost{Mana: []game.ManaToken{{
		Color:  "G",
		Riders: []game.ManaSpendRider{{Kind: game.ManaRiderCantBeCountered, Applied: true}},
	}}}
	unappliedPaid := game.PaidCost{Mana: []game.ManaToken{{
		Color:  "G",
		Riders: []game.ManaSpendRider{{Kind: game.ManaRiderCantBeCountered}},
	}}}

	printed := pushStackSpell(g, "Supreme Verdict", printedOracle, game.PaidCost{}, game.StackItemSpell)
	mana := pushStackSpell(g, "Elf", "plain", riderPaid, game.StackItemSpell)
	plain := pushStackSpell(g, "Bolt", "plain", game.PaidCost{}, game.StackItemSpell)
	unapplied := pushStackSpell(g, "Elf 2", "plain", unappliedPaid, game.StackItemSpell)

	if !stackItemView(t, g, printed).CantBeCountered {
		t.Error("a printed rider must show")
	}
	if !stackItemView(t, g, mana).CantBeCountered {
		t.Error("mana that carried the applied rider must show")
	}
	if stackItemView(t, g, plain).CantBeCountered {
		t.Error("an ordinary spell must not show")
	}
	if stackItemView(t, g, unapplied).CantBeCountered {
		t.Error("a rider whose filter did not fire must not show")
	}
}

// The field is omitted when false, so pre-#1553 fixtures and clients
// are untouched.
func TestCantBeCounteredIsOmittedWhenFalse(t *testing.T) {
	g := buildActiveGame(t)
	id := pushStackSpell(g, "Bolt", "plain", game.PaidCost{}, game.StackItemSpell)
	b, err := json.Marshal(stackItemView(t, g, id))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "cant_be_countered") {
		t.Errorf("should be omitted: %s", b)
	}
}
