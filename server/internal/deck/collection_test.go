package deck

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// collection_test.go — ADR 0198: the card database as a player's
// collection, over a small in-memory index.

func collectionIndexCard(name, typeLine, layout string, legal bool) cards.Card {
	status := "not_legal"
	if legal {
		status = "legal"
	}
	return cards.Card{
		ID:         uuid.New(),
		OracleID:   uuid.NewSHA1(uuid.NameSpaceOID, []byte(name)),
		Name:       name,
		TypeLine:   typeLine,
		Layout:     layout,
		SetCode:    "tst",
		SetType:    "expansion",
		ManaCost:   "{1}",
		Legalities: map[string]string{"commander": status},
	}
}

func TestCollectionSearchesOnlyCollectibleCards(t *testing.T) {
	idx := cards.NewIndex()
	idx.Put(collectionIndexCard("Lightning Bolt", "Instant", "normal", true))
	idx.Put(collectionIndexCard("Lightning Dragon", "Creature — Dragon", "normal", true))
	idx.Put(collectionIndexCard("Lightning Token", "Token Creature — Elemental", "token", true))
	idx.Put(collectionIndexCard("Lightning Plane", "Plane — Lightning", "planar", false))
	col := NewCollection(idx)

	var names []string
	for _, c := range col.Search("lightning", 10) {
		names = append(names, c.Name)
		if c.OracleID == "" || c.ScryfallID == "" {
			t.Errorf("%s: a template without its identity (oracle %q, printing %q)", c.Name, c.OracleID, c.ScryfallID)
		}
		if c.Owner != uuid.Nil {
			t.Errorf("%s: a template with an owner", c.Name)
		}
	}
	if strings.Join(names, ",") != "Lightning Bolt,Lightning Dragon" {
		t.Errorf("search found %v", names)
	}
	if c, ok := col.Named("lightning bolt"); !ok || c.Name != "Lightning Bolt" || !c.IsInstant() {
		t.Errorf("Named: %+v %v", c.Name, ok)
	}
	for _, n := range []string{"Lightning Token", "Lightning Plane", "Nothing"} {
		if _, ok := col.Named(n); ok {
			t.Errorf("Named(%q) found a card a player cannot own", n)
		}
	}
}

func TestCollectionSuggestsAutomatedLegalStaplesFirst(t *testing.T) {
	idx := cards.NewIndex()
	automated := collectionIndexCard("Automated", "Sorcery", "normal", false)
	idx.Put(automated)
	idx.Put(collectionIndexCard("Rare Legal", "Sorcery", "normal", true))
	staple := collectionIndexCard("Staple Legal", "Sorcery", "normal", true)
	idx.Put(staple)
	// A reprint: same oracle, another printing.
	reprint := staple
	reprint.ID = uuid.New()
	idx.Put(reprint)
	idx.Put(collectionIndexCard("Banned", "Sorcery", "normal", false))

	prev := game.IsCatalogCard
	game.IsCatalogCard = func(key string) bool { return key == automated.OracleID.String() }
	t.Cleanup(func() { game.IsCatalogCard = prev })

	col := NewCollection(idx)
	var order []string
	col.Suggest(func(c game.Card) bool {
		order = append(order, c.Name)
		return true
	})
	if got := strings.Join(order, ","); got != "Automated,Staple Legal,Rare Legal,Banned" {
		t.Errorf("suggestion order %q", got)
	}
	// Stopping early stops.
	n := 0
	col.Suggest(func(game.Card) bool { n++; return n < 2 })
	if n != 2 {
		t.Errorf("Suggest called fn %d times after it returned false", n)
	}
}

// The game reads the collection through its hook, and the suggestions a
// prompt carries are the filter's matches in this order.
func TestCollectionFeedsTheGamesSuggestions(t *testing.T) {
	idx := cards.NewIndex()
	idx.Put(collectionIndexCard("Burning Wish", "Sorcery", "normal", true))
	idx.Put(collectionIndexCard("Lightning Bolt", "Instant", "normal", true))
	prev := game.OutsideCollection
	game.OutsideCollection = NewCollection(idx)
	t.Cleanup(func() { game.OutsideCollection = prev })

	got := game.CollectionSuggestionsForEffect(func(c game.Card) bool { return c.IsSorcery() }, 8)
	if strings.Join(got, ",") != "Burning Wish" {
		t.Errorf("suggestions %v", got)
	}
}

// ADR 0216 §2: the planar deck's source keeps exactly the playable
// planar rows the outside-the-game pool excludes, one per oracle ID, in
// oracle-ID order, and the pool still never sees them.
func TestPlanarCardsAreThePlanesTheCollectionExcludes(t *testing.T) {
	idx := cards.NewIndex()
	idx.Put(collectionIndexCard("Lightning Bolt", "Instant", "normal", true))
	akoum := collectionIndexCard("Akoum", "Plane — Zendikar", "planar", false)
	idx.Put(akoum)
	reprint := akoum
	reprint.ID = uuid.New()
	reprint.SetCode = "tsr"
	idx.Put(reprint)
	idx.Put(collectionIndexCard("Bant", "Plane — Alara", "planar", false))
	idx.Put(collectionIndexCard("Chaotic Aether", "Phenomenon", "planar", false))
	idx.Put(collectionIndexCard("Plane Token", "Plane — Nowhere", "token", false))
	idx.Put(collectionIndexCard("Planeswalker Friend", "Legendary Planeswalker — Friend", "normal", true))
	col := NewCollection(idx)

	planes := col.PlanarCards(game.PlanarPlanes)
	var names []string
	for i, c := range planes {
		names = append(names, c.Name)
		if c.OracleID == "" || c.ScryfallID == "" || c.Layout != "planar" {
			t.Errorf("%s: a template without its identity or layout (%q, %q, %q)", c.Name, c.OracleID, c.ScryfallID, c.Layout)
		}
		if c.Owner != uuid.Nil {
			t.Errorf("%s: a template with an owner", c.Name)
		}
		if i > 0 && planes[i-1].OracleID >= c.OracleID {
			t.Errorf("not in oracle-ID order at %s", c.Name)
		}
	}
	joined := strings.Join(names, ",")
	if len(planes) != 2 || !strings.Contains(joined, "Akoum") || !strings.Contains(joined, "Bant") {
		t.Errorf("planes = %v, want Akoum and Bant once each", names)
	}
	if got := col.PlanarCards(game.PlanarPhenomena); len(got) != 1 || got[0].Name != "Chaotic Aether" {
		t.Errorf("phenomena = %v, want Chaotic Aether", got)
	}
	if got := col.PlanarCards(game.PlanarPlanes | game.PlanarPhenomena); len(got) != 3 {
		t.Errorf("planes and phenomena = %d cards, want 3", len(got))
	}
	for _, n := range []string{"Akoum", "Bant", "Chaotic Aether"} {
		if _, ok := col.Named(n); ok {
			t.Errorf("the outside-the-game pool hands out %s", n)
		}
	}
	var nilCol *Collection
	if nilCol.PlanarCards(game.PlanarPlanes) != nil {
		t.Error("a nil collection has planar cards")
	}
}
