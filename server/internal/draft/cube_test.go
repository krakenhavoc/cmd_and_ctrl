package draft

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// A cube as a draft's pack source (ADR 0222 §8, #5441).

// cubeList is n different made-up cards, as a checked cube hands them
// to NewCubeSource.
func cubeList(n int) []cards.Card {
	colors := []string{"W", "U", "B", "R", "G"}
	out := make([]cards.Card, 0, n)
	for i := range n {
		name := fmt.Sprintf("Cube Card %03d", i)
		out = append(out, cards.Card{
			ID: synthID("cube print", i), OracleID: synthID("cube oracle", name), Name: name,
			SetCode: "cub", CollectorNo: fmt.Sprint(i + 1), Lang: "en", Layout: "normal",
			TypeLine: "Creature — Test", Rarity: "rare",
			Colors: []string{colors[i%5]}, ManaCost: fmt.Sprintf("{%d}{%s}", i%5, colors[i%5]),
			Power: "2", Toughness: "2",
		})
	}
	return out
}

func cubeSource(t *testing.T, n, size int) *CubeSource {
	t.Helper()
	src, err := NewCubeSource(cubeList(n), size)
	if err != nil {
		t.Fatalf("NewCubeSource: %v", err)
	}
	return src
}

// A land has no colours of its own, but a colour identity (#5704): the
// card on the wire carries it, and a card with none leaves it off.
func TestFillPrintedCarriesTheColourIdentity(t *testing.T) {
	var forest Card
	fillPrinted(&forest, cards.Card{TypeLine: "Basic Land — Forest", ColorIdentity: []string{"G"}})
	if len(forest.ColorIdentity) != 1 || forest.ColorIdentity[0] != "G" || len(forest.Colors) != 0 {
		t.Fatalf("a Forest: identity %v colours %v", forest.ColorIdentity, forest.Colors)
	}
	var rock Card
	fillPrinted(&rock, cards.Card{TypeLine: "Artifact"})
	if rock.ColorIdentity != nil {
		t.Fatalf("a colourless card has identity %v", rock.ColorIdentity)
	}
}

// The packs of a cube are a partition of the front of one seeded
// shuffle: every pack is the chosen size, no card is dealt twice, the
// same seed deals the same packs and another seed deals others.
func TestCubePacksAreAPartitionOfASeededShuffle(t *testing.T) {
	const cards, size, packs = 400, 15, 24
	src := cubeSource(t, cards, size)
	deal := func(seed uint64) [][]Card {
		got, err := src.Deal(rand.New(rand.NewPCG(seed, packStream)), packs)
		if err != nil {
			t.Fatalf("Deal: %v", err)
		}
		return got
	}
	a := deal(7)
	if len(a) != packs {
		t.Fatalf("%d packs, want %d", len(a), packs)
	}
	seen := map[uuid.UUID]bool{}
	var order []uuid.UUID
	for i, pack := range a {
		if len(pack) != size {
			t.Fatalf("pack %d has %d cards, want %d", i, len(pack), size)
		}
		for _, c := range pack {
			if seen[c.ScryfallID] {
				t.Fatalf("%s was dealt twice", c.Name)
			}
			seen[c.ScryfallID] = true
			order = append(order, c.ScryfallID)
		}
	}
	if len(seen) != packs*size {
		t.Fatalf("%d different cards dealt, want %d", len(seen), packs*size)
	}
	// The packs, end to end, are the front of the shuffle the seed
	// gives: the same permutation, cut every `size` cards.
	list := cubeList(cards)
	perm := rand.New(rand.NewPCG(7, packStream)).Perm(cards)
	for i, id := range order {
		if want := list[perm[i]].ID; id != want {
			t.Fatalf("card %d of the deal is not card %d of the seeded shuffle", i, i)
		}
	}
	b := deal(7)
	for i := range a {
		for j := range a[i] {
			if a[i][j].ScryfallID != b[i][j].ScryfallID {
				t.Fatal("one seed dealt two different sets of packs")
			}
		}
	}
	c := deal(8)
	same := true
	for i := range a {
		for j := range a[i] {
			same = same && a[i][j].ScryfallID == c[i][j].ScryfallID
		}
	}
	if same {
		t.Fatal("two seeds dealt the same packs")
	}
}

// A draft takes its packs from the cube in one deal, so no two seats
// in no two rounds ever open the same card, and a seed replays it.
func TestACubeDraftDealsNoCardTwice(t *testing.T) {
	src := cubeSource(t, 8*Rounds*15, 15)
	open := func(seed uint64) []uuid.UUID {
		d, err := New(Config{Name: "Cube night", SetName: "My cube", Booster: "cube", Cube: true, Mode: PickOne, Seats: 8, Seed: seed}, t0)
		if err != nil {
			t.Fatal(err)
		}
		host, err := d.AddSeat("Host", uuid.Nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := d.Start(host.ID, src, t0); err != nil {
			t.Fatalf("Start: %v", err)
		}
		var ids []uuid.UUID
		for d.State() == StatePicking {
			if err := d.Pick(host.ID, firstCards(d, host.ID), t0); err != nil {
				t.Fatal(err)
			}
		}
		for i, pool := range d.Pools() {
			if len(pool) != Rounds*15 {
				t.Fatalf("seat %d's pool has %d cards, want %d", i, len(pool), Rounds*15)
			}
			for _, c := range pool {
				ids = append(ids, c.ScryfallID)
			}
		}
		return ids
	}
	a := open(21)
	seen := map[uuid.UUID]bool{}
	for _, id := range a {
		if seen[id] {
			t.Fatal("a card of the cube reached two pools")
		}
		seen[id] = true
	}
	// A cube exactly the size of the draft is dealt whole.
	if len(seen) != 8*Rounds*15 {
		t.Fatalf("%d cards drafted, want %d", len(seen), 8*Rounds*15)
	}
	b := open(21)
	for i := range a {
		if a[i] != b[i] {
			t.Fatal("one seed gave two different cube drafts")
		}
	}
}

// A four-seat cube draft, one person and three bots, through the picks
// and the build: every seat ends with a legal 40-card deck from its own
// pool, as in a set's draft.
func TestAFourSeatCubeDraftWithBotsThroughDeckbuilding(t *testing.T) {
	for _, mode := range []Mode{PickOne, PickTwo} {
		src := cubeSource(t, 360, 15)
		d, err := New(Config{Name: "Cube night", SetName: "My cube", Booster: "cube", Cube: true, Mode: mode, Seats: 4, Seed: 99}, t0)
		if err != nil {
			t.Fatal(err)
		}
		host, err := d.AddSeat("Host", uuid.Nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := d.Start(host.ID, src, t0); err != nil {
			t.Fatalf("%s: Start: %v", mode, err)
		}
		finishPicks(t, d, []uuid.UUID{host.ID})
		if d.State() != StateDone || d.Phase() != PhaseBuilding {
			t.Fatalf("%s: after the picks: state %s, phase %q", mode, d.State(), d.Phase())
		}
		v, _ := d.ViewFor(host.ID, t0, nil)
		if !v.Cube || v.Set != "" || v.SetName != "My cube" {
			t.Fatalf("%s: the view does not say it is a cube: cube %v, set %q, set_name %q", mode, v.Cube, v.Set, v.SetName)
		}
		// The bots built when the picks ended; the person builds now.
		if err := d.SubmitDeck(host.ID, legalDeck(d, host.ID), t0); err != nil {
			t.Fatalf("%s: SubmitDeck: %v", mode, err)
		}
		if d.Phase() != PhaseRounds {
			t.Fatalf("%s: phase %q after the last deck, want rounds", mode, d.Phase())
		}
		for i := range d.Seats() {
			main, basics, _, ok := d.DeckCards(i)
			if !ok {
				t.Fatalf("%s: seat %d has no deck", mode, i)
			}
			total := len(main)
			for _, n := range basics {
				total += n
			}
			if total != 40 {
				t.Fatalf("%s: seat %d's deck has %d cards, want 40", mode, i, total)
			}
			own := map[string]bool{}
			for _, c := range d.Pools()[i] {
				own[c.ID] = true
			}
			for _, c := range main {
				if !own[c.ID] {
					t.Fatalf("%s: seat %d's deck holds a card it did not draft", mode, i)
				}
			}
		}
	}
}

// A cube too small for the draft, and a pack size outside the bounds,
// are refused, and a refused start leaves the draft in its lobby.
func TestACubeTooSmallForTheDraftIsRefused(t *testing.T) {
	if _, err := NewCubeSource(cubeList(100), MinCubePackSize-1); !errors.Is(err, ErrBadConfig) {
		t.Fatalf("a pack of %d: err = %v, want ErrBadConfig", MinCubePackSize-1, err)
	}
	if _, err := NewCubeSource(cubeList(100), MaxCubePackSize+1); !errors.Is(err, ErrBadConfig) {
		t.Fatalf("a pack of %d: err = %v, want ErrBadConfig", MaxCubePackSize+1, err)
	}
	src := cubeSource(t, 4*Rounds*15-1, 15)
	d, err := New(Config{Name: "Cube night", Cube: true, Mode: PickOne, Seats: 4, Seed: 3}, t0)
	if err != nil {
		t.Fatal(err)
	}
	host, _ := d.AddSeat("Host", uuid.Nil)
	if err := d.Start(host.ID, src, t0); !errors.Is(err, ErrCubeTooSmall) {
		t.Fatalf("Start = %v, want ErrCubeTooSmall", err)
	}
	if d.State() != StateLobby {
		t.Fatalf("state %s after a refused start, want lobby", d.State())
	}
}
