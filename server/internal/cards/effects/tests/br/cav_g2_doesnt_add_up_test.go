package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// A creature that asks a question as it enters is still suspected once the
// question is answered: a Clone returned by It Doesn't Add Up picks what to
// copy, and the permanent that results is the one suspected (the suspect
// rides the entry's tail, not the line after the return).
func TestItDoesntAddUpSuspectsACloneAfterItsCopyPick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	clone := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: clone, Name: "Clone", TypeLine: "Creature — Shapeshifter", OracleID: effectstest.OracleClone,
		ManaCost: "{3}{U}", Power: 0, Toughness: 0, Owner: me.ID, Controller: me.ID,
	})

	effectstest.CastCatalogSpell(t, g, "It Doesn't Add Up", "Instant", effectstest.SbItDoesntAddUp, effectstest.DgCardRef(clone))
	effectstest.ResolveWithCopyChoice(t, g, bear)

	c, ok := effectstest.BattlefieldCard(g, clone)
	if !ok {
		t.Fatal("the Clone did not return")
	}
	if c.Name != "Grizzly Bears" {
		t.Errorf("the returned Clone is %q, want a copy of Grizzly Bears", c.Name)
	}
	if !effectstest.SbSuspected(g, clone) {
		t.Error("the returned Clone should be suspected after answering its copy prompt")
	}
}
