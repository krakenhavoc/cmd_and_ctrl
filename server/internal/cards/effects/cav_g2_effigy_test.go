package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	birdsOfParadiseOracle = "d3a0b660-358c-41bd-9cd2-41fbf3491b1a"
)

func seedBirds(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Birds of Paradise", TypeLine: "Creature — Bird",
		OracleID: birdsOfParadiseOracle, ManaCost: "{G}", Power: 0, Toughness: 1,
		Keywords: []string{"flying"}, Owner: owner, Controller: owner,
	})
}

// Machine God's Effigy enters as a copy of the chosen creature except it is
// an artifact (and only an artifact: "It's not a creature") with an extra
// "{T}: Add {U}". The copied card's own ability still works.
func TestMachineGodsEffigyCopiesACreatureAsANonCreatureArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	birds := seedBirds(g, me.ID)

	id := castCatalogSpell(t, g, "Machine God's Effigy", "Artifact", machineGodsEffigyOracle, nil)
	resolveWithCopyChoice(t, g, birds)

	got := copyBattlefieldCard(t, g, id)
	if got.Name != "Birds of Paradise" {
		t.Fatalf("name = %q, want the copied creature's", got.Name)
	}
	if !got.IsArtifact() {
		t.Errorf("type line %q: want an artifact", got.TypeLine)
	}
	if got.IsCreature() {
		t.Errorf("type line %q: the Effigy is not a creature", got.TypeLine)
	}
	if got.HasSubtype("Bird") {
		t.Errorf("type line %q kept a creature subtype", got.TypeLine)
	}
	// Two mana abilities: the copied Birds' any-color one and the granted {T}: Add {U}.
	if err := g.ActivateManaAbility(me.ID, id, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the granted {T}: Add {U}: %v", err)
	}
	if blue := dcSortedPool(me); len(blue) != 1 || blue[0] != "U" {
		t.Errorf("pool = %v, want [U]", blue)
	}
}

// Declining the copy leaves the printed artifact with its own {T}: Add {U}.
func TestMachineGodsEffigyDeclinedCopyIsAPlainArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedBirds(g, me.ID)
	id := castCatalogSpell(t, g, "Machine God's Effigy", "Artifact", machineGodsEffigyOracle, nil)
	resolveWithCopyChoice(t, g, uuid.Nil)
	got := copyBattlefieldCard(t, g, id)
	if got.Name != "Machine God's Effigy" || got.IsCreature() {
		t.Fatalf("declined: got %q %q", got.Name, got.TypeLine)
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("printed {T}: Add {U}: %v", err)
	}
}

// Only creatures can be copied.
func TestMachineGodsEffigyOffersOnlyCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	birds := seedBirds(g, me.ID)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "Machine God's Effigy", "Artifact", machineGodsEffigyOracle, nil)
	for i := 0; i < 8 && copyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pc := copyPrompt(g)
	if pc == nil {
		t.Fatal("no copy prompt")
	}
	if len(pc.CopyOptions) != 1 || pc.CopyOptions[0] != birds {
		t.Errorf("CopyOptions = %v, want only the creature %s (not %s)", pc.CopyOptions, birds, rock)
	}
}
