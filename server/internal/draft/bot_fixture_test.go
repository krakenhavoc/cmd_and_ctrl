package draft

import (
	"sort"
	"strconv"
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
		card.ID = "f" + strconv.Itoa(i)
		fx.byName[card.Name] = card
		fx.all = append(fx.all, card)
	}
	sort.Slice(fx.all, func(i, j int) bool { return fx.all[i].Name < fx.all[j].Name })
	botFixtureCache = fx
	return fx
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

// cards is fixture cards by name, with the IDs prefix1, prefix2, ...
func (fx *botFixture) cards(t testing.TB, prefix string, names ...string) []Card {
	t.Helper()
	out := make([]Card, 0, len(names))
	for i, n := range names {
		out = append(out, fx.card(t, n, prefix+strconv.Itoa(i+1)))
	}
	return out
}

// TestTheFixtureSetIsWhatItsReadmeSays pins the fixture: every real
// card of it is in the catalog as Full (loadBotFixture fails on one
// that is not), every invented one is a vanilla creature, and each
// colour has the same commons.
func TestTheFixtureSetIsWhatItsReadmeSays(t *testing.T) {
	fx := loadBotFixture(t)
	invented, basics := 0, 0
	for _, c := range fx.all {
		switch {
		case c.Basic:
			basics++
		case strings.HasPrefix(c.Name, botFixtureVanilla):
			invented++
			full, _ := fx.idx.Get(c.ScryfallID)
			if full.OracleText != "" || !strings.HasPrefix(c.TypeLine, "Creature") {
				t.Errorf("%s is invented and must be a vanilla creature; it is %q with text %q", c.Name, c.TypeLine, full.OracleText)
			}
		}
	}
	if invented != 30 || basics != 5 {
		t.Errorf("%d invented creatures and %d basic lands, want 30 and 5", invented, basics)
	}
	if len(fx.all) != 114 {
		t.Errorf("the fixture set holds %d cards, want 114", len(fx.all))
	}
}
