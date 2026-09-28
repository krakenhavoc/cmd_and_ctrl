package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const echoOfEonsOracle = "23d3e5fe-3f82-44cf-91a1-6646a12a0255"

// Every seated player's hand and graveyard end up shuffled into their
// own library, and everyone draws exactly seven — including the
// caster's own leftover hand and graveyard, and an opponent's.
func TestEchoOfEonsShufflesHandAndGraveyardThenDrawsSeven(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]

	myGraveyardCard := seedGraveyardCard(t, g, "My Fodder", "Instant", "")
	oppHandCard := handCard(opp, "Their Fodder", "Instant")
	oppGraveyardCard := uuid.New()
	opp.Graveyard.PushTop(game.Card{
		InstanceID: oppGraveyardCard, Name: "Their Graveyard Fodder", TypeLine: "Instant",
		Owner: opp.ID, Controller: opp.ID,
	})

	castCatalogSpell(t, g, "Echo of Eons", "Sorcery", echoOfEonsOracle, nil)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != 7 {
		t.Errorf("my hand size = %d, want 7", got)
	}
	if got := opp.Hand.Size(); got != 7 {
		t.Errorf("opponent hand size = %d, want 7", got)
	}
	if me.Graveyard.Contains(myGraveyardCard) {
		t.Errorf("my graveyard card should have gone to my library")
	}
	if opp.Hand.Contains(oppHandCard) {
		t.Errorf("the opponent's hand card should have gone to their library")
	}
	if opp.Graveyard.Contains(oppGraveyardCard) {
		t.Errorf("the opponent's graveyard card should have gone to their library")
	}
	// Echo of Eons itself lands in my graveyard AFTER this resolution
	// finishes — it is still on the stack, not in the graveyard, while
	// the shuffle runs — so exactly one card (itself) is left behind.
	if me.Graveyard.Size() != 1 {
		t.Errorf("my graveyard should hold only the just-resolved sorcery, got %d", me.Graveyard.Size())
	}
	if opp.Graveyard.Size() != 0 {
		t.Errorf("the opponent's graveyard should be empty, got %d", opp.Graveyard.Size())
	}
}

func TestEchoOfEonsFlashback(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := seedGraveyardCard(t, g, "Echo of Eons", "Sorcery", echoOfEonsOracle)

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback",
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != 7 {
		t.Errorf("my hand size = %d, want 7", got)
	}
	if !g.Exile.Contains(id) {
		t.Errorf("the flashed-back sorcery should be exiled")
	}
}
