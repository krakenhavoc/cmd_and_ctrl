package game

import "testing"

// creature_types_test.go pins the two rules S26's tribal work rests
// on: changeling is every creature type IN EVERY ZONE (CR 702.73a),
// and "shares a creature type" counts creature types only.
//
// The zone half is the one worth a test of its own. Every other
// keyword in the engine's table is a battlefield rule, so
// HasKeyword's off-battlefield fallback exists for exactly one
// consumer today (flash, gating a cast from hand). Changeling is the
// first keyword whose answer has to be the same in a graveyard, a
// library and a hand, and the code path that gets it right there is
// not the path the battlefield uses.

func TestIsCreatureTypeRejectsOtherSubtypeKinds(t *testing.T) {
	for _, want := range []string{"Elf", "Goblin", "Merfolk", "Shapeshifter", "Sliver", "Wizard"} {
		if !IsCreatureType(want) {
			t.Errorf("IsCreatureType(%q) = false, want true", want)
		}
	}
	// Land, artifact, enchantment and spell subtypes must NOT be
	// creature types: a changeling is not a Forest, and Coat of Arms
	// must not decide two Dryad Arbors share "Forest".
	for _, notType := range []string{"Forest", "Island", "Equipment", "Aura", "Saga", "Vehicle", "Arcane", ""} {
		if IsCreatureType(notType) {
			t.Errorf("IsCreatureType(%q) = true, want false", notType)
		}
	}
	// Case-insensitive: card files write title case, the deck
	// importer sometimes does not.
	if !IsCreatureType("elf") {
		t.Error(`IsCreatureType("elf") = false, want true`)
	}
}

func TestChangelingIsACanonicalKeyword(t *testing.T) {
	got, ok := CanonicalKeyword("Changeling")
	if !ok || got != KeywordChangeling {
		t.Fatalf("CanonicalKeyword(%q) = (%q, %v), want (%q, true)", "Changeling", got, ok, KeywordChangeling)
	}
}

// TestChangelingHasEveryCreatureTypeInEveryZone is CR 702.73a. The
// card is never pushed to the battlefield, so `effective` stays nil
// and HasSubtype takes its printed branch — the branch a lord
// counting cards in a graveyard, a tribal tutor and a
// creature-type-matters cost reduction all read through.
func TestChangelingHasEveryCreatureTypeInEveryZone(t *testing.T) {
	c := Card{
		Name:     "Woodland Changeling",
		TypeLine: "Creature — Shapeshifter",
		Keywords: []string{KeywordChangeling},
	}
	for _, tribe := range []string{"Elf", "Goblin", "Merfolk", "Sliver", "Shapeshifter"} {
		if !c.HasSubtype(tribe) {
			t.Errorf("changeling in hand: HasSubtype(%q) = false, want true", tribe)
		}
	}
	// Still not a land type.
	if c.HasSubtype("Forest") {
		t.Error(`changeling: HasSubtype("Forest") = true, want false — changeling grants creature types only`)
	}
	// And a plain Shapeshifter with no changeling is only a
	// Shapeshifter.
	plain := Card{Name: "Plain", TypeLine: "Creature — Shapeshifter"}
	if plain.HasSubtype("Elf") {
		t.Error("non-changeling Shapeshifter: HasSubtype(\"Elf\") = true, want false")
	}
}

// TestChangelingOnBattlefieldReadsEffectiveAbilities covers the other
// branch: once the layer engine has run, HasKeyword reads
// effective.Abilities, and printedCharacteristic is what folds
// Card.Keywords into that list.
func TestChangelingOnBattlefieldReadsEffectiveAbilities(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	id := pushTypedTestCard(g, Card{
		Name: "Universal Automaton", TypeLine: "Artifact Creature — Shapeshifter",
		Power: 1, Toughness: 1, Keywords: []string{KeywordChangeling},
		Owner: seat, Controller: seat,
	})
	c := layeredBattlefieldCard(t, g, id)
	if c.effective == nil {
		t.Fatal("expected the layer engine to have populated effective")
	}
	if !c.HasSubtype("Goblin") {
		t.Error(`battlefield changeling: HasSubtype("Goblin") = false, want true`)
	}
}

func TestSharesCreatureType(t *testing.T) {
	goblinWarrior := &Card{Name: "A", TypeLine: "Creature — Goblin Warrior"}
	goblinShaman := &Card{Name: "B", TypeLine: "Creature — Goblin Shaman"}
	elfDruid := &Card{Name: "C", TypeLine: "Creature — Elf Druid"}
	changeling := &Card{Name: "D", TypeLine: "Creature — Shapeshifter", Keywords: []string{KeywordChangeling}}
	otherChangeling := &Card{Name: "E", TypeLine: "Creature — Shapeshifter", Keywords: []string{KeywordChangeling}}
	typeless := &Card{Name: "F", TypeLine: "Creature"}
	// Dryad Arbor shares Dryad with a Dryad and must NOT share
	// Forest with a Forest — the case a raw subtype intersection
	// gets wrong.
	dryadArbor := &Card{Name: "Dryad Arbor", TypeLine: "Land Creature — Forest Dryad"}
	forest := &Card{Name: "Forest", TypeLine: "Basic Land — Forest"}

	cases := []struct {
		name string
		a, b *Card
		want bool
	}{
		{"shared goblin", goblinWarrior, goblinShaman, true},
		{"no overlap", goblinWarrior, elfDruid, false},
		{"changeling vs typed", changeling, elfDruid, true},
		{"typed vs changeling", elfDruid, changeling, true},
		{"two changelings", changeling, otherChangeling, true},
		{"changeling vs typeless", changeling, typeless, false},
		{"typeless vs typeless", typeless, typeless, false},
		{"land type does not count", dryadArbor, forest, false},
		{"nil", nil, goblinWarrior, false},
	}
	for _, tc := range cases {
		if got := SharesCreatureType(tc.a, tc.b); got != tc.want {
			t.Errorf("%s: SharesCreatureType = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestCreatureTypesOfFiltersNonCreatureSubtypes(t *testing.T) {
	arbor := &Card{Name: "Dryad Arbor", TypeLine: "Land Creature — Forest Dryad"}
	got := CreatureTypesOf(arbor)
	if len(got) != 1 || got[0] != "Dryad" {
		t.Errorf("CreatureTypesOf(Dryad Arbor) = %v, want [Dryad]", got)
	}
	changeling := &Card{Name: "D", TypeLine: "Creature — Shapeshifter", Keywords: []string{KeywordChangeling}}
	if len(CreatureTypesOf(changeling)) != len(AllCreatureTypes) {
		t.Errorf("CreatureTypesOf(changeling) = %d types, want all %d", len(CreatureTypesOf(changeling)), len(AllCreatureTypes))
	}
}

// TestNamedTribeClearsOnBattlefieldLeave — the chosen type belongs to
// the entry, not to the card (CR 614.12 re-fires on every entry).
func TestNamedTribeClearsOnBattlefieldLeave(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	c := Card{Name: "Cavern of Souls", TypeLine: "Land", Owner: seat, Controller: seat}
	id := pushTypedTestCard(g, c)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].NamedTribe = "Elf"
			}
		}
	})
	var moved Card
	g.WithWriteLock(func() {
		var err error
		moved, err = MoveCard(g.Battlefield, g.Seats[0].Graveyard, id)
		if err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
	})
	if moved.NamedTribe != "" {
		t.Errorf("NamedTribe after leaving the battlefield = %q, want empty", moved.NamedTribe)
	}
}
