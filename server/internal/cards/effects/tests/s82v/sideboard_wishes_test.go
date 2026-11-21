package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gold is a two-colour card in a sideboard.
func gold(name, typeLine string) game.Card {
	c := effectstest.OutsideCard(name, typeLine)
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
		{"Cunning Wish", "Instant", effectstest.CunningWishOracle,
			[]game.Card{effectstest.OutsideCard("Lightning Bolt", "Instant"), effectstest.OutsideCard("Rift Bolt", "Sorcery")},
			[]string{"Lightning Bolt"}},
		{"Glittering Wish", "Sorcery", effectstest.GlitteringOracle,
			[]game.Card{effectstest.OutsideCard("Lightning Bolt", "Instant"), gold("Lyra", "Legendary Creature — Angel")},
			[]string{"Lyra"}},
		{"Golden Wish", "Sorcery", effectstest.GoldenWishOracle,
			[]game.Card{effectstest.OutsideCard("Sol Ring", "Artifact"), effectstest.OutsideCard("Pacifism", "Enchantment — Aura"), effectstest.OutsideCard("Char", "Sorcery")},
			[]string{"Sol Ring", "Pacifism"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewSideboardCatalogGame(t, 2, tc.side...)
			me := g.Seats[0]
			effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			c := effectstest.LatestChooseCardsFor(g, me.ID)
			if c == nil {
				t.Fatalf("no prompt: %+v", g.PendingChoices)
			}
			if got := effectstest.CandidateNames(g, c); !effectstest.EqualNames(got, tc.want) {
				t.Errorf("candidates %v, want %v", got, tc.want)
			}
			if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{effectstest.SideCardID(me, tc.want[0])}); err != nil {
				t.Fatalf("ResolveChooseCards: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !effectstest.OutsideHandHas(me, tc.want[0]) {
				t.Errorf("%s did not come into the hand", tc.want[0])
			}
			if !effectstest.OutsideExileHas(g, tc.name) {
				t.Errorf("%s did not exile itself", tc.name)
			}
		})
	}
}

// Death Wish's "may" covers the card only: the life is lost either way.
func TestDeathWishLosesHalfTheLifeWhetherOrNotItTakesACard(t *testing.T) {
	for name, side := range map[string][]game.Card{
		"takes a card": {effectstest.OutsideCard("Lightning Bolt", "Instant")},
		"no sideboard": nil,
	} {
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewSideboardCatalogGame(t, 2, side...)
			me := g.Seats[0]
			before := me.Life
			effectstest.CastCatalogSpell(t, g, "Death Wish", "Sorcery", effectstest.DeathWishOracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if c := effectstest.LatestChooseCardsFor(g, me.ID); c != nil {
				if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
					t.Fatal(err)
				}
				effectstest.PassPriorityAroundTable(t, g)
			}
			if want := before - (before+1)/2; me.Life != want {
				t.Errorf("life %d, want %d (half of %d, rounded up)", me.Life, want, before)
			}
			if !effectstest.OutsideExileHas(g, "Death Wish") {
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
		OracleID:   effectstest.FaeOfWishesOracle,
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
	g := effectstest.NewSideboardCatalogGame(t, 2,
		effectstest.OutsideCard("Grizzly Bears", "Creature — Bear"), effectstest.OutsideCard("Sol Ring", "Artifact"))
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
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("no prompt: %+v", g.PendingChoices)
	}
	if got := effectstest.CandidateNames(g, c); !effectstest.EqualNames(got, []string{"Sol Ring"}) {
		t.Errorf("candidates %v, want only the noncreature", got)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{effectstest.SideCardID(me, "Sol Ring")}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OutsideHandHas(me, "Sol Ring") {
		t.Error("Sol Ring did not come into the hand")
	}
	if !effectstest.OutsideExileHas(g, "Granted") && !effectstest.OutsideExileHas(g, "Fae of Wishes") {
		t.Error("the Adventure did not exile the card")
	}
}

// Every wish this file covers ships Full, the creature half included.
func TestEveryTypeClauseWishIsFull(t *testing.T) {
	for _, id := range []string{effectstest.CunningWishOracle, effectstest.GlitteringOracle, effectstest.GoldenWishOracle, effectstest.DeathWishOracle,
		effectstest.FaeOfWishesOracle, effectstest.FaeOfWishesOracle + "#1"} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Fatalf("%s is not in the catalog", id)
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s is %v, want Full", spec.Name, spec.Completeness)
		}
	}
}
