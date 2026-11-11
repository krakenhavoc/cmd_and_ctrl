package botarena_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/botarena"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/draft"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/draftpacks"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// draft_decks_test.go is the games-won half of the draft bot's measure
// (ADR 0222 §7, #5439): a deck the bot drafted and built against a deck
// of random picks built from the same packs, played by the same
// heuristic on both sides. The rating half, and the two-colours and
// curve half, are internal/draft's bot_measure_test.go and run in CI.
//
// It is a whole-game test and gated like the others (AISEAT_GAME_TESTS):
// the nightly runs it.

const draftFixtureDump = "../draft/testdata/botset_dump.json"

// draftFixture is the draft bot's fixture set, c2d, as a pack source.
func draftFixture(t *testing.T) (*cards.Index, *draft.SetSource) {
	t.Helper()
	idx := cards.NewIndex()
	if _, err := idx.Load(draftFixtureDump); err != nil {
		t.Fatalf("load %s: %v", draftFixtureDump, err)
	}
	vanilla := map[uuid.UUID]bool{}
	idx.EachPrinting(func(c cards.Card) {
		if strings.HasPrefix(c.Name, "C2 Fixture ") {
			vanilla[c.OracleID] = true
		}
	})
	catalogFull := draftpacks.CatalogIsFull(idx)
	src, err := draft.NewSetSource(idx, "c2d", func(id uuid.UUID) bool { return vanilla[id] || catalogFull(id) })
	if err != nil {
		t.Fatalf("NewSetSource: %v", err)
	}
	return idx, src
}

// draftPools runs one eight-seat draft with no person picking and
// returns each seat's pool.
func draftPools(t *testing.T, src draft.Source, seed uint64, picker draft.Picker) [][]draft.Card {
	t.Helper()
	now := time.Date(2026, 11, 11, 19, 0, 0, 0, time.UTC)
	d, err := draft.New(draft.Config{Name: "arena", Set: "c2d", Mode: draft.PickOne, Seats: 8, Seed: seed, Picker: picker}, now)
	if err != nil {
		t.Fatalf("draft.New: %v", err)
	}
	host, err := d.AddSeat("host", uuid.Nil)
	if err != nil {
		t.Fatalf("AddSeat: %v", err)
	}
	if err := d.Start(host.ID, src, now); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for i := 0; d.State() == draft.StatePicking; i++ {
		if i > 1000 {
			t.Fatal("the draft did not finish")
		}
		d.Expire(now)
	}
	return d.Pools()
}

// draftedDeck is the deck BuildDeck makes of a pool, as a list the
// engine seats: the printings of the main deck and the basic lands.
func draftedDeck(t *testing.T, idx *cards.Index, rater draft.Rater, pool []draft.Card, seed uint64) func() []game.Card {
	t.Helper()
	main, lands := rater.BuildDeck(pool, nil, seed)
	byID := map[string]draft.Card{}
	for _, c := range pool {
		byID[c.ID] = c
	}
	list := &deck.List{}
	for _, id := range main {
		c, ok := idx.Get(byID[id].ScryfallID)
		if !ok {
			t.Fatalf("the index does not hold %q", byID[id].Name)
		}
		list.Mainboard = append(list.Mainboard, c)
	}
	for name, n := range lands {
		c, ok := idx.FindByName(name)
		if !ok {
			t.Fatalf("the index holds no %s", name)
		}
		for range n {
			list.Mainboard = append(list.Mainboard, c)
		}
	}
	if len(list.Mainboard) != draft.DeckSize {
		t.Fatalf("a deck of %d cards, want %d", len(list.Mainboard), draft.DeckSize)
	}
	return list.ToGameCards
}

// TestADraftedDeckBeatsADeckOfRandomPicks plays, for each seat of each
// seeded draft, the deck the bot drafted against the deck built from
// that seat's picks when every seat picked at random from the same
// packs. Each pair plays twice, once on the play each. "Well over half"
// (ADR 0222 §7) is read as at least 60% of the decided games.
func TestADraftedDeckBeatsADeckOfRandomPicks(t *testing.T) {
	requireGameTests(t)
	const (
		drafts = 10
		seed0  = 5439
	)
	idx, src := draftFixture(t)
	rater := draft.Rater{Facts: draft.CatalogFacts(idx)}
	wins, losses, undecided := 0, 0, 0
	started := time.Now()
	for i := range drafts {
		seed := uint64(seed0 + i)
		rated := draftPools(t, src, seed, draft.BotPicker{Seed: seed, Rater: &rater})
		random := draftPools(t, src, seed, draft.RandomPicker{Seed: seed})
		for seat := range rated {
			cfg := botarena.Config{
				Seats: []botarena.SeatSpec{
					{Tier: tiers.Heuristic, Name: "drafted", Deck: "drafted"},
					{Tier: tiers.Heuristic, Name: "random", Deck: "random"},
				},
				Lists: map[string]func() []game.Card{
					"drafted": draftedDeck(t, idx, rater, rated[seat], seed),
					"random":  draftedDeck(t, idx, rater, random[seat], seed),
				},
				Games:        1,
				Lockstep:     true,
				TurnBudget:   40,
				Wall:         2 * time.Minute,
				StartingLife: 20,
			}
			for g, order := range [][]int{{0, 1}, {1, 0}} {
				res, err := botarena.Play(context.Background(), cfg, seed*100+uint64(seat*2+g), order)
				if err != nil {
					t.Fatalf("Play: %v", err)
				}
				switch {
				case res.Winner < 0:
					undecided++
				case order[res.Winner] == 0:
					wins++
				default:
					losses++
				}
			}
		}
	}
	decided := wins + losses
	t.Logf("drafted decks won %d and lost %d of %d decided games (%d undecided) in %s",
		wins, losses, decided, undecided, time.Since(started).Round(time.Second))
	if decided == 0 {
		t.Fatal("no game was decided")
	}
	if rate := float64(wins) / float64(decided); rate < 0.60 {
		t.Errorf("drafted decks won %.0f%% of decided games, want at least 60%%", 100*rate)
	}
}
