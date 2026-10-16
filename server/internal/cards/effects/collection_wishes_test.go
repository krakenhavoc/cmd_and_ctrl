package effects

import (
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// collection_wishes_test.go — ADR 0198: a deck with no sideboard takes "a
// card you own from outside the game" from the card database. Every card
// built on TakeFromOutside gains that without a change to its file; these
// pin it card by card, and that the card taken is a real catalog card.

const cultivateOracle = "8b755881-a72d-4e21-a369-d2924eb4585a"

type testCollection struct{ cards []game.Card }

func (f testCollection) Search(q string, limit int) []game.Card {
	var out []game.Card
	for _, c := range f.cards {
		if strings.Contains(strings.ToLower(c.Name), strings.ToLower(q)) && len(out) < limit {
			out = append(out, c)
		}
	}
	return out
}

func (f testCollection) Named(name string) (game.Card, bool) {
	for _, c := range f.cards {
		if strings.EqualFold(c.Name, name) {
			return c, true
		}
	}
	return game.Card{}, false
}

func (f testCollection) Suggest(fn func(game.Card) bool) {
	for _, c := range f.cards {
		if !fn(c) {
			return
		}
	}
}

// dbCard is a card-database template.
func dbCard(name, typeLine, oracle string, colors ...string) game.Card {
	c := game.NewCard(name, uuid.Nil)
	c.TypeLine = typeLine
	c.OracleID = oracle
	c.Colors = colors
	c.ScryfallID = uuid.NewSHA1(uuid.NameSpaceURL, []byte(name)).String()
	return c
}

// installTestCollection installs a database with one card of each kind a
// wish asks for.
func installTestCollection(t *testing.T) {
	t.Helper()
	prev := game.OutsideCollection
	game.OutsideCollection = testCollection{cards: []game.Card{
		dbCard("Lightning Bolt", "Instant", lightningBoltOracle, "R"),
		dbCard("Cultivate", "Sorcery", cultivateOracle, "G"),
		dbCard("Sol Ring", "Artifact", solRingOracle),
		dbCard("Grizzly Bears", "Creature — Bear", "", "G"),
		dbCard("Forest", "Basic Land — Forest", ""),
		dbCard("Pacifism", "Enchantment — Aura", "", "W"),
		dbCard("Lightning Helix", "Instant", "", "R", "W"),
		dbCard("Ulamog's Crusher", "Creature — Eldrazi", ""),
	}}
	t.Cleanup(func() { game.OutsideCollection = prev })
}

func latestCollectionPickFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceOutsideCollection && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func handCardNamed(p *game.Player, name string) (game.Card, bool) {
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			return c, true
		}
	}
	return game.Card{}, false
}

// Each wish suggests exactly its filter's matches from the database, and
// the card taken is minted into the wisher's hand, revealed, and real.
func TestEveryWishTakesFromTheDatabaseWithoutASideboard(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		want                   []string
	}{
		{"Burning Wish", "Sorcery", burningWishOracle, []string{"Cultivate"}},
		{"Cunning Wish", "Instant", cunningWishOracle, []string{"Lightning Bolt", "Lightning Helix"}},
		{"Living Wish", "Sorcery", livingWishOracle, []string{"Grizzly Bears", "Forest", "Ulamog's Crusher"}},
		{"Golden Wish", "Sorcery", goldenWishOracle, []string{"Sol Ring", "Pacifism"}},
		{"Glittering Wish", "Sorcery", glitteringOracle, []string{"Lightning Helix"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			installTestCollection(t)
			g := newSideboardCatalogGame(t, 2) // no sideboard anywhere
			me, opp := g.Seats[0], g.Seats[1]
			castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			passPriorityAroundTable(t, g)
			c := latestCollectionPickFor(g, me.ID)
			if c == nil {
				t.Fatalf("no collection pick: %+v", g.PendingChoices)
			}
			if !slices.Equal(c.CollectionSuggestions, tc.want) || c.ChooseMin != 0 || c.ChooseMax != 1 {
				t.Errorf("suggestions %v (%d..%d), want %v as a \"may\" of one", c.CollectionSuggestions, c.ChooseMin, c.ChooseMax, tc.want)
			}
			take := tc.want[0]
			if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{take}); err != nil {
				t.Fatalf("ResolveOutsideCollection: %v", err)
			}
			passPriorityAroundTable(t, g)
			card, ok := handCardNamed(me, take)
			if !ok {
				t.Fatalf("%s did not come into the hand", take)
			}
			if card.Owner != me.ID || !card.IsKnownTo(opp.ID) {
				t.Errorf("owner %v, revealed to the opponent %v", card.Owner, card.IsKnownTo(opp.ID))
			}
			if !outsideExileHas(g, tc.name) {
				t.Errorf("%s did not exile itself", tc.name)
			}
			if me.HasSideboard {
				t.Error("taking from the database gave the deck a sideboard")
			}
		})
	}
}

// The minted card is the card a deck import would have seated: its
// oracle ID keys its catalog entry, so it is played by the engine.
func TestACardTakenFromTheDatabaseIsACatalogCard(t *testing.T) {
	installTestCollection(t)
	g := newSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Cunning Wish", "Instant", cunningWishOracle, nil)
	passPriorityAroundTable(t, g)
	c := latestCollectionPickFor(g, me.ID)
	if c == nil {
		t.Fatal("no collection pick")
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Lightning Bolt"}); err != nil {
		t.Fatal(err)
	}
	bolt, ok := handCardNamed(me, "Lightning Bolt")
	if !ok {
		t.Fatal("no Lightning Bolt in hand")
	}
	if spec, ok := Lookup(bolt.OracleID); !ok || spec.Name != "Lightning Bolt" {
		t.Errorf("the minted card's oracle %q has no catalog entry", bolt.OracleID)
	}
	if !game.IsCatalogCard(bolt.OracleID) {
		t.Error("the engine does not see the minted card as automated")
	}
}

// Death Wish's life loss does not depend on the pool.
func TestDeathWishFromTheDatabaseStillCostsHalfTheLife(t *testing.T) {
	installTestCollection(t)
	g := newSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	before := me.Life
	castCatalogSpell(t, g, "Death Wish", "Sorcery", deathWishOracle, nil)
	passPriorityAroundTable(t, g)
	c := latestCollectionPickFor(g, me.ID)
	// "A card": every card in the test database, up to the cap of eight.
	if c == nil || len(c.CollectionSuggestions) != 8 {
		t.Fatalf("no full collection pick for \"a card\": %+v", c)
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Sol Ring"}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if want := before - (before+1)/2; me.Life != want {
		t.Errorf("life %d, want %d", me.Life, want)
	}
	if _, ok := handCardNamed(me, "Sol Ring"); !ok {
		t.Error("Sol Ring did not come into the hand")
	}
}

// Mastermind's Acquisition's second mode is mandatory and unrevealed.
func TestMastermindsAcquisitionTakesAnUnrevealedCardFromTheDatabase(t *testing.T) {
	installTestCollection(t)
	g := newSideboardCatalogGame(t, 2)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpellWithModes(t, g, "Mastermind's Acquisition", "Sorcery", mastermindsOracle, []int{1})
	passPriorityAroundTable(t, g)
	c := latestCollectionPickFor(g, me.ID)
	if c == nil {
		t.Fatalf("no collection pick: %+v", g.PendingChoices)
	}
	if c.ChooseMin != 1 {
		t.Errorf("floor %d, want a mandatory pick", c.ChooseMin)
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, nil); err == nil {
		t.Error("the mandatory pick was declined")
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Pacifism"}); err != nil {
		t.Fatal(err)
	}
	card, ok := handCardNamed(me, "Pacifism")
	if !ok {
		t.Fatal("Pacifism did not come into the hand")
	}
	if card.IsKnownTo(opp.ID) {
		t.Error("an unrevealed card became known to the opponent")
	}
}

// Extrapolate the Impossible over the database: the caster names two
// cards with different names, both are revealed, the opponent picks one,
// and the other stays outside the game.
func TestExtrapolateTheImpossibleTakesAPairFromTheDatabase(t *testing.T) {
	installTestCollection(t)
	g := newSideboardCatalogGame(t, 2)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "Extrapolate the Impossible", "Sorcery", extrapolateOracle, nil)
	passPriorityAroundTable(t, g)
	c := latestCollectionPickFor(g, me.ID)
	if c == nil {
		t.Fatalf("no collection pick: %+v", g.PendingChoices)
	}
	if c.ChooseMin != 0 || c.ChooseMax != 2 {
		t.Errorf("bounds %d..%d", c.ChooseMin, c.ChooseMax)
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Sol Ring"}); err == nil {
		t.Error("a single card was accepted")
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Sol Ring", "Sol Ring"}); err == nil {
		t.Error("a same-name pair was accepted")
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Sol Ring", "Cultivate"}); err != nil {
		t.Fatal(err)
	}
	if pendingOfKind(g, game.PendingChoiceOptionPick) != nil {
		answerChosenPlayer(t, g, me.ID, opp.ID)
	}
	rp := latestRevealPickFor(g, opp.ID)
	if rp == nil {
		t.Fatalf("the opponent was not asked: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(rp.ID, opp.ID, []uuid.UUID{sideCardID(me, "Cultivate")}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := handCardNamed(me, "Cultivate"); !ok {
		t.Error("the chosen card did not come into the hand")
	}
	if got := sideNamed(me); !equalNames(got, []string{"Sol Ring"}) {
		t.Errorf("outside the game: %v, want the other revealed card", got)
	}
	if !cardKnownToAll(g, me, sideCardID(me, "Sol Ring")) {
		t.Error("the revealed card that stayed outside is not known to the table")
	}
}
