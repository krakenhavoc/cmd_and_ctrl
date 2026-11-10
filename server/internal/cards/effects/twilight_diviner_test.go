package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const twilightDivinerOracle = "05d7222b-cc08-4f1f-8e35-6058599572ba"

func TestTwilightDivinerIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(twilightDivinerOracle)
	if !ok || spec.Name != "Twilight Diviner" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

func TestTwilightDivinerSurveilsTwoAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C")
	rfdCast(t, g, "Twilight Diviner", "Creature — Elf Cleric", twilightDivinerOracle, 3, 3)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("surveil looks at %d cards, want 2", len(c.ScryCards))
	}
}

// twilightDivinerGraveCreature puts a vanilla creature card in p's
// graveyard.
func twilightDivinerGraveCreature(p *game.Player, name string, power, toughness int) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Giant",
		Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID})
	return id
}

// Two creatures returned from graveyards together are one trigger; the
// controller picks which one to copy. A second return the same turn does
// not trigger again.
func TestTwilightDivinerCopiesOneOfACreatureBatchFromAGraveyardOnceATurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Twilight Diviner", "Creature — Elf Cleric", twilightDivinerOracle, false)
	twilightDivinerGraveCreature(me, "Hill Giant", 3, 3)
	twilightDivinerGraveCreature(opp, "Stone Giant", 3, 4)

	castCatalogSpell(t, g, "Rise of the Dark Realms", "Sorcery", b04RiseOfTheDarkRealmsOracle, nil)
	gtSettle(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to choose which creature to copy")
	}
	if len(pick.ChooseCards) != 2 {
		t.Fatalf("%d candidates, want the two creatures", len(pick.ChooseCards))
	}
	stone := findBattlefieldByName(g, "Stone Giant")
	answerChooseCards(t, g, me.ID, stone)
	gtSettle(t, g)
	if n := rfdCountNamed(g, me.ID, "Stone Giant"); n != 2 {
		t.Fatalf("%d Stone Giants, want the original and a token copy", n)
	}
	tokens := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Stone Giant" && c.IsToken() {
			tokens++
		}
	}
	if tokens != 1 {
		t.Errorf("%d Stone Giant tokens, want 1", tokens)
	}

	twilightDivinerGraveCreature(me, "Hill Troll", 2, 2)
	castCatalogSpell(t, g, "Rise of the Dark Realms", "Sorcery", b04RiseOfTheDarkRealmsOracle, nil)
	gtSettle(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil || rfdCountNamed(g, me.ID, "Hill Troll") != 1 {
		t.Error("the ability triggered a second time this turn")
	}
}

// A single creature from a graveyard is copied with no prompt; a creature
// cast from hand is not.
func TestTwilightDivinerIgnoresAHandCastAndCopiesALoneReturn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Twilight Diviner", "Creature — Elf Cleric", twilightDivinerOracle, false)
	castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	gtSettle(t, g)
	if n := rfdCountNamed(g, me.ID, "Grizzly Bears"); n != 1 {
		t.Fatalf("%d Grizzly Bears after a hand cast, want 1", n)
	}

	twilightDivinerGraveCreature(me, "Hill Giant", 3, 3)
	castCatalogSpell(t, g, "Rise of the Dark Realms", "Sorcery", b04RiseOfTheDarkRealmsOracle, nil)
	gtSettle(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("a lone candidate asked for a pick")
	}
	if n := rfdCountNamed(g, me.ID, "Hill Giant"); n != 2 {
		t.Errorf("%d Hill Giants, want the original and a token copy", n)
	}
}
