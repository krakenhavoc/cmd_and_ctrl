package draft

import (
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/draftpacks"
)

// bot_fixture_test.go loads the draft bot's fixture set, c2d
// (testdata/README.md): real cards the catalog holds as Full, with
// invented rarities, beside invented vanilla creatures. It is what the
// bot is tested and measured on in CI, where there is no Scryfall dump.

const (
	botFixtureDump = "testdata/botset_dump.json"
	botFixtureSet  = "c2d"
	// botFixtureVanilla is the name every invented card of the fixture
	// starts with. They have no catalog entry and need none: a vanilla
	// creature has nothing to automate.
	botFixtureVanilla = "C2 Fixture "
)

type botFixture struct {
	idx    *cards.Index
	src    *SetSource
	rater  Rater
	byName map[string]Card
	// all is every card a pack can hold, sorted by name.
	all []Card
}

var botFixtureCache *botFixture

func loadBotFixture(t testing.TB) *botFixture {
	t.Helper()
	if botFixtureCache != nil {
		return botFixtureCache
	}
	idx := cards.NewIndex()
	if _, err := idx.Load(botFixtureDump); err != nil {
		t.Fatalf("load %s: %v", botFixtureDump, err)
	}
	vanilla := map[uuid.UUID]bool{}
	idx.EachPrinting(func(c cards.Card) {
		if strings.HasPrefix(c.Name, botFixtureVanilla) {
			vanilla[c.OracleID] = true
		}
	})
	catalogFull := draftpacks.CatalogIsFull(idx)
	isFull := func(id uuid.UUID) bool { return vanilla[id] || catalogFull(id) }
	pool := draftpacks.PoolFor(idx, botFixtureSet)
	if blockers := pool.Blockers(isFull); len(blockers) > 0 {
		t.Fatalf("the fixture set holds cards the catalog does not have as Full: %v", blockers)
	}
	src, err := NewSetSource(idx, botFixtureSet, isFull)
	if err != nil {
		t.Fatalf("NewSetSource: %v", err)
	}
	fx := &botFixture{idx: idx, src: src, rater: Rater{Facts: CatalogFacts(idx)}, byName: map[string]Card{}}
	for i, c := range pool.Contents() {
		card := src.card(c)
		card.ID = "f" + itoa(i)
		fx.byName[card.Name] = card
		fx.all = append(fx.all, card)
	}
	sort.Slice(fx.all, func(i, j int) bool { return fx.all[i].Name < fx.all[j].Name })
	botFixtureCache = fx
	return fx
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b []byte
	for ; i > 0; i /= 10 {
		b = append([]byte{byte('0' + i%10)}, b...)
	}
	return string(b)
}

// card is a fixture card by name, with a fresh ID.
func (fx *botFixture) card(t testing.TB, name, id string) Card {
	t.Helper()
	c, ok := fx.byName[name]
	if !ok {
		t.Fatalf("the fixture has no card named %q", name)
	}
	c.ID = id
	return c
}

func (fx *botFixture) rating(t testing.TB, name string) float64 {
	t.Helper()
	return fx.rater.Rating(fx.card(t, name, "x"))
}

func TestRatingTablePrint(t *testing.T) {
	fx := loadBotFixture(t)
	all := append([]Card(nil), fx.all...)
	sort.SliceStable(all, func(i, j int) bool { return fx.rater.Rating(all[i]) > fx.rater.Rating(all[j]) })
	for _, c := range all {
		t.Logf("%5.2f  %-28s %-10s %-9s", fx.rater.Rating(c), c.Name, c.ManaCost, c.Rarity)
	}
}
