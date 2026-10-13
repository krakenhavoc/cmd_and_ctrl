package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// A creature that asks a question as it enters is still suspected once the
// question is answered: a Clone returned by It Doesn't Add Up picks what to
// copy, and the permanent that results is the one suspected (the suspect
// rides the entry's tail, not the line after the return).
func TestItDoesntAddUpSuspectsACloneAfterItsCopyPick(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	clone := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: clone, Name: "Clone", TypeLine: "Creature — Shapeshifter", OracleID: oracleClone,
		ManaCost: "{3}{U}", Power: 0, Toughness: 0, Owner: me.ID, Controller: me.ID,
	})

	castCatalogSpell(t, g, "It Doesn't Add Up", "Instant", sbItDoesntAddUp, dgCardRef(clone))
	resolveWithCopyChoice(t, g, bear)

	c, ok := battlefieldCard(g, clone)
	if !ok {
		t.Fatal("the Clone did not return")
	}
	if c.Name != "Grizzly Bears" {
		t.Errorf("the returned Clone is %q, want a copy of Grizzly Bears", c.Name)
	}
	if !sbSuspected(g, clone) {
		t.Error("the returned Clone should be suspected after answering its copy prompt")
	}
}
