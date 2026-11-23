package testscm

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Each wish suggests exactly its filter's matches from the database, and
// the card taken is minted into the wisher's hand, revealed, and real.
func TestEveryWishTakesFromTheDatabaseWithoutASideboard(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		want                   []string
	}{
		{"Burning Wish", "Sorcery", effectstest.BurningWishOracle, []string{"Cultivate"}},
		{"Cunning Wish", "Instant", effectstest.CunningWishOracle, []string{"Lightning Bolt", "Lightning Helix"}},
		{"Living Wish", "Sorcery", effectstest.LivingWishOracle, []string{"Grizzly Bears", "Forest", "Ulamog's Crusher"}},
		{"Golden Wish", "Sorcery", effectstest.GoldenWishOracle, []string{"Sol Ring", "Pacifism"}},
		{"Glittering Wish", "Sorcery", effectstest.GlitteringOracle, []string{"Lightning Helix"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			effectstest.InstallTestCollection(t)
			g := effectstest.NewSideboardCatalogGame(t, 2) // no sideboard anywhere
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			c := effectstest.LatestCollectionPickFor(g, me.ID)
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
			effectstest.PassPriorityAroundTable(t, g)
			card, ok := effectstest.HandCardNamed(me, take)
			if !ok {
				t.Fatalf("%s did not come into the hand", take)
			}
			if card.Owner != me.ID || !card.IsKnownTo(opp.ID) {
				t.Errorf("owner %v, revealed to the opponent %v", card.Owner, card.IsKnownTo(opp.ID))
			}
			if !effectstest.OutsideExileHas(g, tc.name) {
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
	effectstest.InstallTestCollection(t)
	g := effectstest.NewSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Cunning Wish", "Instant", effectstest.CunningWishOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestCollectionPickFor(g, me.ID)
	if c == nil {
		t.Fatal("no collection pick")
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Lightning Bolt"}); err != nil {
		t.Fatal(err)
	}
	bolt, ok := effectstest.HandCardNamed(me, "Lightning Bolt")
	if !ok {
		t.Fatal("no Lightning Bolt in hand")
	}
	if spec, ok := effects.Lookup(bolt.OracleID); !ok || spec.Name != "Lightning Bolt" {
		t.Errorf("the minted card's oracle %q has no catalog entry", bolt.OracleID)
	}
	if !game.IsCatalogCard(bolt.OracleID) {
		t.Error("the engine does not see the minted card as automated")
	}
}

// Death Wish's life loss does not depend on the pool.
func TestDeathWishFromTheDatabaseStillCostsHalfTheLife(t *testing.T) {
	effectstest.InstallTestCollection(t)
	g := effectstest.NewSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Death Wish", "Sorcery", effectstest.DeathWishOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestCollectionPickFor(g, me.ID)
	// "A card": every card in the test database, up to the cap of eight.
	if c == nil || len(c.CollectionSuggestions) != 8 {
		t.Fatalf("no full collection pick for \"a card\": %+v", c)
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Sol Ring"}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if want := before - (before+1)/2; me.Life != want {
		t.Errorf("life %d, want %d", me.Life, want)
	}
	if _, ok := effectstest.HandCardNamed(me, "Sol Ring"); !ok {
		t.Error("Sol Ring did not come into the hand")
	}
}

// Mastermind's Acquisition's second mode is mandatory and unrevealed.
func TestMastermindsAcquisitionTakesAnUnrevealedCardFromTheDatabase(t *testing.T) {
	effectstest.InstallTestCollection(t)
	g := effectstest.NewSideboardCatalogGame(t, 2)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpellWithModes(t, g, "Mastermind's Acquisition", "Sorcery", effectstest.MastermindsOracle, []int{1})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestCollectionPickFor(g, me.ID)
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
	card, ok := effectstest.HandCardNamed(me, "Pacifism")
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
	effectstest.InstallTestCollection(t)
	g := effectstest.NewSideboardCatalogGame(t, 2)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Extrapolate the Impossible", "Sorcery", effectstest.ExtrapolateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestCollectionPickFor(g, me.ID)
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
	if effectstest.PendingOfKind(g, game.PendingChoiceOptionPick) != nil {
		effectstest.AnswerChosenPlayer(t, g, me.ID, opp.ID)
	}
	rp := effectstest.LatestRevealPickFor(g, opp.ID)
	if rp == nil {
		t.Fatalf("the opponent was not asked: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(rp.ID, opp.ID, []uuid.UUID{effectstest.SideCardID(me, "Cultivate")}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.HandCardNamed(me, "Cultivate"); !ok {
		t.Error("the chosen card did not come into the hand")
	}
	if got := effectstest.SideNamed(me); !effectstest.EqualNames(got, []string{"Sol Ring"}) {
		t.Errorf("outside the game: %v, want the other revealed card", got)
	}
	if !effectstest.CardKnownToAll(g, me, effectstest.SideCardID(me, "Sol Ring")) {
		t.Error("the revealed card that stayed outside is not known to the table")
	}
}
