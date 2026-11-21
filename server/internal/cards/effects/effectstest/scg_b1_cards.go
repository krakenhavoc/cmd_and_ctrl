package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// scgB1Typecycle typecycles a card from the active seat's hand with the
// ability at `idx`, paying {2}, and fetches `land` from a library that also
// holds a land of another type.
func ScgB1Typecycle(t *testing.T, name, typeLine, oracle string, idx int, land, landType, other, otherType string) {
	t.Helper()
	g := NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	AdvanceToMain(t, g)
	SeedSearchLibrary(me,
		game.Card{Name: other, TypeLine: otherType},
		game.Card{Name: land, TypeLine: landType},
	)
	id := PushCatalogHandCard(me, name, typeLine, oracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("%s: typecycle: %v", name, err)
	}
	if !me.Graveyard.Contains(id) {
		t.Fatalf("%s was not discarded to typecycle", name)
	}
	PassPriorityAroundTable(t, g)
	c := SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("%s: no search prompt", name)
	}
	if SearchOptionNamed(g, c, other) != uuid.Nil {
		t.Errorf("%s offered %s", name, other)
	}
	AnswerSearchNamed(t, g, me.ID, land)
	if !B02bHandHasNamed(me, land) {
		t.Errorf("%s did not put %s into the hand", name, land)
	}
}
