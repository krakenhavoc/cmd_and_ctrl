package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #6121: "Put three cards from your graveyard on the bottom of your
// library" as an activated ability's cost (Battlefield Scrounger). The
// graveyard pick of the exile-cards component with the library bottom as
// its destination (CR 401.4: the owner arranges the cards; CR 118.3: a
// graveyard of fewer than three can't pay).

const gyBottomOracle = "test-6121-graveyard-to-bottom"

func registerGraveyardToBottom(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: gyBottomOracle,
		Name:     gyBottomOracle,
		Activated: []effects.ActivatedAbility{{
			Label: "Put three cards from your graveyard on the bottom of your library: gain 1 life",
			Cost:  effects.PutFromGraveyardOnLibraryBottom(3, "three cards", nil),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GainLife{Player: item.Controller, Amount: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

func gyCard(z *game.Zone, owner uuid.UUID, name string) uuid.UUID {
	id := uuid.New()
	z.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", Owner: owner, Controller: owner})
	return id
}

func TestGraveyardToLibraryBottomCostMovesTheNamedCards(t *testing.T) {
	registerGraveyardToBottom(t)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	src := effectstest.PushCatalogPermanent(g, me.ID, "Scrounger", "Creature — Human Barbarian", gyBottomOracle, false)
	a := gyCard(me.Graveyard, me.ID, "A")
	b := gyCard(me.Graveyard, me.ID, "B")
	c := gyCard(me.Graveyard, me.ID, "C")
	d := gyCard(me.Graveyard, me.ID, "D")
	top := gyCard(me.Library, me.ID, "Library card")
	life := me.Life

	// CR 118.3: two cards can't pay "three cards".
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{a, b}}); err == nil {
		t.Fatal("two cards paid a three-card cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{a, b, c}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Graveyard.Size() != 1 || !me.Graveyard.Contains(d) {
		t.Errorf("graveyard holds %d cards, want only the unnamed one", me.Graveyard.Size())
	}
	if g.Exile.Size() != 0 {
		t.Errorf("exile holds %d cards — the cost is not an exile", g.Exile.Size())
	}
	// Each card goes beneath the one named before it, and all of them
	// beneath the library's own cards.
	want := []uuid.UUID{c, b, a}
	if me.Library.Size() < len(want)+1 {
		t.Fatalf("library holds %d cards, want at least %d", me.Library.Size(), len(want)+1)
	}
	for i, id := range want {
		if me.Library.Cards[i].InstanceID != id {
			t.Errorf("library position %d from the bottom is %s, want %s", i, me.Library.Cards[i].Name, id)
		}
	}
	if got, _ := me.Library.Top(); got.InstanceID != top {
		t.Errorf("the library's top card is %s, want the one that was there", got.Name)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life = %d, want %d", me.Life, life+1)
	}
}

// Register holds the destination to a CR 602 ability's printed count from
// the graveyard.
func TestRegisterRefusesALibraryBottomCostOffTheGraveyard(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a hand clause sent to the library bottom registered")
		}
		delete(effects.Registry, "test-6121-bad-bottom")
	}()
	cost := effects.PutFromGraveyardOnLibraryBottom(1, "a card", nil)
	cost.ExileCards.From = game.ZoneHand
	effects.Register(effects.Spec{OracleID: "test-6121-bad-bottom", Name: "test-6121-bad-bottom", Activated: []effects.ActivatedAbility{{
		Label: "x", Cost: cost, Effect: func(*game.Game, *game.StackItem) error { return nil },
	}}})
}
