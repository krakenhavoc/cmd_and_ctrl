package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sideboard_wishes_test.go — ADR 0191: the wishes that take one more
// clause than Burning Wish: Cunning, Glittering, Golden and Death Wish,
// and Fae of Wishes // Granted, whose Adventure half is a wish.

const (
	cunningWishOracle = "bdf80a55-b524-4d17-9be8-a0005403a0e7"
	glitteringOracle  = "45c29ac0-f9ca-4ab9-b93e-8b678509bcd7"
	goldenWishOracle  = "e5f25cbd-39e4-4806-8dba-9a552356696d"
	deathWishOracle   = "ff0ca1a4-da67-407c-a5c0-998a35eb7d8d"
	faeOfWishesOracle = "c0abbed2-d213-47ef-8d6c-4a21efb9a55f"
)

// gold is a two-colour card in a sideboard.
func gold(name, typeLine string) game.Card {
	c := outsideCard(name, typeLine)
	c.Colors = []string{"W", "U"}
	return c
}

// Each type-clause wish is Burning Wish with its own predicate: the
// prompt offers only the matches, and the spell exiles itself.
func TestTypeClauseWishesOfferOnlyTheirMatches(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		side                   []game.Card
		want                   []string
	}{
		{"Cunning Wish", "Instant", cunningWishOracle,
			[]game.Card{outsideCard("Lightning Bolt", "Instant"), outsideCard("Rift Bolt", "Sorcery")},
			[]string{"Lightning Bolt"}},
		{"Glittering Wish", "Sorcery", glitteringOracle,
			[]game.Card{outsideCard("Lightning Bolt", "Instant"), gold("Lyra", "Legendary Creature — Angel")},
			[]string{"Lyra"}},
		{"Golden Wish", "Sorcery", goldenWishOracle,
			[]game.Card{outsideCard("Sol Ring", "Artifact"), outsideCard("Pacifism", "Enchantment — Aura"), outsideCard("Char", "Sorcery")},
			[]string{"Sol Ring", "Pacifism"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newSideboardCatalogGame(t, 2, tc.side...)
			me := g.Seats[0]
			castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			passPriorityAroundTable(t, g)
			c := latestChooseCardsFor(g, me.ID)
			if c == nil {
				t.Fatalf("no prompt: %+v", g.PendingChoices)
			}
			if got := candidateNames(g, c); !equalNames(got, tc.want) {
				t.Errorf("candidates %v, want %v", got, tc.want)
			}
			if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{sideCardID(me, tc.want[0])}); err != nil {
				t.Fatalf("ResolveChooseCards: %v", err)
			}
			passPriorityAroundTable(t, g)
			if !outsideHandHas(me, tc.want[0]) {
				t.Errorf("%s did not come into the hand", tc.want[0])
			}
			if !outsideExileHas(g, tc.name) {
				t.Errorf("%s did not exile itself", tc.name)
			}
		})
	}
}

// Death Wish's "may" covers the card only: the life is lost either way.
func TestDeathWishLosesHalfTheLifeWhetherOrNotItTakesACard(t *testing.T) {
	for name, side := range map[string][]game.Card{
		"takes a card": {outsideCard("Lightning Bolt", "Instant")},
		"no sideboard": nil,
	} {
		t.Run(name, func(t *testing.T) {
			g := newSideboardCatalogGame(t, 2, side...)
			me := g.Seats[0]
			before := me.Life
			castCatalogSpell(t, g, "Death Wish", "Sorcery", deathWishOracle, nil)
			passPriorityAroundTable(t, g)
			if c := latestChooseCardsFor(g, me.ID); c != nil {
				if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
					t.Fatal(err)
				}
				passPriorityAroundTable(t, g)
			}
			if want := before - (before+1)/2; me.Life != want {
				t.Errorf("life %d, want %d (half of %d, rounded up)", me.Life, want, before)
			}
			if !outsideExileHas(g, "Death Wish") {
				t.Error("Death Wish did not exile itself")
			}
		})
	}
}

// faeOfWishes builds the card the way foulmireKnight does: both faces,
// then SetFace(0).
func faeOfWishes(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   faeOfWishesOracle,
		Layout:     game.LayoutAdventure,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Fae of Wishes", TypeLine: "Creature — Faerie Wizard", ManaCost: "{1}{U}", Colors: []string{"U"}, Power: 1, Toughness: 1},
			{Name: "Granted", TypeLine: "Sorcery — Adventure", ManaCost: "{3}{U}", Colors: []string{"U"}},
		},
	}
	c.SetFace(0)
	return c
}

// Granted's clause is a noncreature card, revealed; the card then goes
// to exile (CR 715.3d) rather than to the graveyard.
func TestGrantedFetchesANoncreatureCard(t *testing.T) {
	g := newSideboardCatalogGame(t, 2,
		outsideCard("Grizzly Bears", "Creature — Bear"), outsideCard("Sol Ring", "Artifact"))
	me := g.Seats[0]
	card := faeOfWishes(me.ID)
	me.Hand.PushTop(card)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Granted: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("no prompt: %+v", g.PendingChoices)
	}
	if got := candidateNames(g, c); !equalNames(got, []string{"Sol Ring"}) {
		t.Errorf("candidates %v, want only the noncreature", got)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{sideCardID(me, "Sol Ring")}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !outsideHandHas(me, "Sol Ring") {
		t.Error("Sol Ring did not come into the hand")
	}
	if !outsideExileHas(g, "Granted") && !outsideExileHas(g, "Fae of Wishes") {
		t.Error("the Adventure did not exile the card")
	}
}

// Every wish this file covers ships Full, the creature half included.
func TestEveryTypeClauseWishIsFull(t *testing.T) {
	for _, id := range []string{cunningWishOracle, glitteringOracle, goldenWishOracle, deathWishOracle,
		faeOfWishesOracle, faeOfWishesOracle + "#1"} {
		spec, ok := Lookup(id)
		if !ok {
			t.Fatalf("%s is not in the catalog", id)
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s is %v, want Full", spec.Name, spec.Completeness)
		}
	}
}
