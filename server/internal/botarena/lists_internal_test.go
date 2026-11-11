package botarena

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lists_internal_test.go: Config.Lists and Config.StartingLife (#5439)
// add a way to seat a deck built for one run at a life total. A run
// that sets neither is the Commander table it always was.

func cardNames(cs []game.Card) []string {
	out := make([]string, 0, len(cs))
	for _, c := range cs {
		out = append(out, c.Name)
	}
	return out
}

func TestARunWithNoListsSeatsWhatItAlwaysDid(t *testing.T) {
	got, err := seatDeck(Config{}, SeatSpec{Tier: tiers.Heuristic})
	if err != nil {
		t.Fatal(err)
	}
	want := BattleDeck(uuid.Nil)
	if !reflect.DeepEqual(cardNames(got), cardNames(want)) {
		t.Error("with no Lists the default seat is not dealt BattleDeck")
	}
	if !got[0].IsCommander {
		t.Error("BattleDeck no longer leads with its commander")
	}
	got, err = seatDeck(Config{Lists: map[string]func() []game.Card{"drafted": nil}}, SeatSpec{Deck: ExertDeckID})
	if err != nil || !reflect.DeepEqual(cardNames(got), cardNames(ExertBattleDeck(uuid.Nil))) {
		t.Errorf("a Lists entry changed another deck id (err %v)", err)
	}
}

func TestAListIsDealtByItsDeckID(t *testing.T) {
	calls := 0
	list := func() []game.Card {
		calls++
		return []game.Card{game.NewCard("Island", uuid.Nil), game.NewCard("Opt", uuid.Nil)}
	}
	cfg := Config{
		Seats: []SeatSpec{{Tier: tiers.Heuristic, Deck: "drafted"}, {Tier: tiers.Heuristic}},
		Games: 1,
		Lists: map[string]func() []game.Card{"drafted": list},
	}
	// It needs no Scryfall dump, as a synthetic deck needs none.
	if err := cfg.Validate(); err != nil {
		t.Fatalf("a run with a list is refused: %v", err)
	}
	for range 2 {
		got, err := seatDeck(cfg, cfg.Seats[0])
		if err != nil || !reflect.DeepEqual(cardNames(got), []string{"Island", "Opt"}) {
			t.Fatalf("seatDeck = %v, %v", cardNames(got), err)
		}
	}
	if calls != 2 {
		t.Errorf("the list was called %d times for two games, want 2", calls)
	}
	// A deck id that is neither a list, synthetic nor curated is still
	// refused without an index.
	cfg.Seats[1].Deck = "no-such-deck"
	if err := cfg.Validate(); err == nil {
		t.Error("an unknown deck id passed Validate")
	}
}

// Zero StartingLife leaves the table exactly as the engine makes it: a
// Commander table at Commander's life. A positive one sets the life and
// nothing else; the table still reads as Commander, because a draft's
// own format is slice 6's (#5440).
func TestStartingLifeZeroLeavesACommanderTable(t *testing.T) {
	newTable := func() *game.Game {
		g := game.NewGame()
		for _, name := range []string{"a", "b"} {
			if _, err := g.AddPlayer(name, BattleDeck(uuid.Nil)); err != nil {
				t.Fatal(err)
			}
		}
		return g
	}
	g := newTable()
	before := g.Settings
	if err := setStartingLife(g, 0); err != nil {
		t.Fatal(err)
	}
	if g.Settings != before {
		t.Errorf("a zero StartingLife changed the settings: %+v, were %+v", g.Settings, before)
	}
	if g.Settings.StartingLife != game.StartingLife || g.Settings.EffectiveFormat() != game.FormatCommander {
		t.Errorf("the default table is %d life, format %q: want Commander's %d", g.Settings.StartingLife, g.Settings.EffectiveFormat(), game.StartingLife)
	}
	if got := (Config{}).withDefaults().StartingLife; got != 0 {
		t.Errorf("withDefaults set StartingLife to %d", got)
	}

	g = newTable()
	if err := setStartingLife(g, 20); err != nil {
		t.Fatal(err)
	}
	want := before
	want.StartingLife = 20
	if g.Settings != want {
		t.Errorf("StartingLife 20 gave %+v, want only the life changed: %+v", g.Settings, want)
	}
}
