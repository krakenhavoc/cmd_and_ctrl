package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// vengeance_test.go — the Vengeance half of an adventure-layout card
// (Hela, Mistress of Darkness // Dark Tutor). Vengeance is not in the
// pinned Comprehensive Rules, so these tests pin the reminder text:
// "(Cast only from your graveyard. Then exile it.)".

func vengeanceAdventure(owner uuid.UUID) Card {
	c := Card{
		InstanceID: uuid.New(),
		Name:       "Hela // Tutor",
		Layout:     LayoutAdventure,
		Owner:      owner,
		Controller: owner,
		Faces: []Face{
			{Name: "Hela", TypeLine: "Legendary Creature — God", Power: 6, Toughness: 6},
			{Name: "Tutor", TypeLine: "Sorcery — Vengeance", ManaCost: "{B}",
				OracleText: "(Cast only from your graveyard. Then exile it.)\nYou gain 1 life."},
		},
	}
	c.SetFace(0)
	return c
}

func TestVengeanceHalfIsRecognisedByItsTypeLineOnly(t *testing.T) {
	c := vengeanceAdventure(uuid.New())
	if isVengeanceHalf(c) {
		t.Error("the creature half reads as Vengeance")
	}
	c.SetFace(1)
	if !isVengeanceHalf(c) {
		t.Error("the Vengeance half is not recognised")
	}
	if castAsAdventure(c) {
		t.Error("a Vengeance half reads as an Adventure: its card must not be castable from exile")
	}

	adv := vengeanceAdventure(uuid.New())
	adv.Faces[1].TypeLine = "Sorcery — Adventure"
	adv.SetFace(1)
	if isVengeanceHalf(adv) {
		t.Error("an ordinary Adventure half reads as Vengeance")
	}
	if !castAsAdventure(adv) {
		t.Error("an ordinary Adventure half stopped being an Adventure")
	}
}

func TestVengeanceHalfCastsOnlyFromAGraveyardAndIsExiled(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)

	inHand := putInHand(t, g, me, vengeanceAdventure(me.ID))
	if err := g.CastSpell(me.ID, inHand, CastSpellParams{Face: 1, Strict: true}); !errors.Is(err, ErrCastZoneNotAllowed) {
		t.Fatalf("Vengeance half from hand: err = %v, want ErrCastZoneNotAllowed", err)
	}

	inYard := putInGraveyard(t, g, me, vengeanceAdventure(me.ID))
	if err := g.CastSpell(me.ID, inYard, CastSpellParams{FromZone: "graveyard", Face: 0, Strict: true}); !errors.Is(err, ErrCastZoneNotAllowed) {
		t.Fatalf("the creature half from the graveyard: err = %v, want ErrCastZoneNotAllowed", err)
	}
	if err := g.AddManaForEffect(me.ID, uuid.New(), "{B}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.CastSpell(me.ID, inYard, CastSpellParams{FromZone: "graveyard", Face: 1, Strict: true}); err != nil {
		t.Fatalf("Vengeance half from the graveyard: %v", err)
	}
	passPriorityUntilResolvedForTest(t, g, inYard)
	if !g.Exile.Contains(inYard) {
		t.Fatal("a Vengeance half cast from a graveyard was not exiled as it left the stack")
	}
	if perm := g.CastPermissionOnCardByIDForEffect(inYard); perm != nil {
		t.Errorf("the exiled card carries a cast permission %+v: it is not an Adventure", perm)
	}
}
