package draftpacks

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// unflaggedSet puts a made-up set into idx the way Scryfall shipped Star
// Trek, Secrets of Strixhaven and TMNT (#5689): no printing carries the
// booster flag. Its main run is the numbers 1 to 12 in the plain frame
// (6 commons, 3 uncommons, 2 rares, 1 mythic), and after it come a
// showcase and a borderless reprint of two of them, a promo, a lettered
// number, a plain printing of a new card past the run's end, and the
// basic lands, also past it. It returns the names of the cards the main
// run holds.
func unflaggedSet(idx *cards.Index, set string, universesBeyond bool) (mainRun []string) {
	put := func(name, rarity, number string, edit func(*cards.Card)) {
		c := synthCard(set, name, rarity, number)
		c.Booster = false
		if universesBeyond {
			c.PromoTypes = []string{"universesbeyond"}
		}
		if edit != nil {
			edit(&c)
		}
		idx.Put(c)
	}
	n := 1
	for _, g := range []struct {
		prefix, rarity string
		count          int
	}{{"Common", "common", 6}, {"Uncommon", "uncommon", 3}, {"Rare", "rare", 2}, {"Mythic", "mythic", 1}} {
		for i := range g.count {
			name := fmt.Sprintf("%s %s %02d", set, g.prefix, i)
			put(name, g.rarity, fmt.Sprint(n), nil)
			mainRun = append(mainRun, name)
			n++
		}
	}
	// Past the run: none of these are in it.
	put(set+" Common 00", "common", "13", func(c *cards.Card) { c.FrameEffects = []string{"showcase"} })
	put(set+" Rare 00", "rare", "14", func(c *cards.Card) { c.BorderColor = "borderless" })
	put(set+" Rare 01", "rare", "15", func(c *cards.Card) { c.FullArt = true })
	put(set+" Promo Card", "rare", "16", func(c *cards.Card) { c.Promo = true })
	put(set+" Lettered Card", "rare", "17a", nil)
	put(set+" Late Plain Card", "rare", "30", nil)
	for i, b := range []string{"Plains", "Island", "Swamp", "Mountain", "Forest"} {
		put(b, "common", fmt.Sprint(40+i), func(c *cards.Card) { c.TypeLine = "Basic Land — " + b })
	}
	return mainRun
}

func contentNames(p Pool) []string {
	var out []string
	for _, c := range p.Contents() {
		out = append(out, c.Name)
	}
	sort.Strings(out)
	return out
}

func sorted(s []string) []string {
	out := append([]string(nil), s...)
	sort.Strings(out)
	return out
}

func TestUnflaggedSetIsBuiltFromItsMainRun(t *testing.T) {
	for _, ub := range []bool{false, true} {
		t.Run(fmt.Sprintf("universes_beyond=%v", ub), func(t *testing.T) {
			idx := cards.NewIndex()
			want := sorted(unflaggedSet(idx, "unf", ub))

			resp := sets(idx, allFullExcept(), nil)
			s := findSet(t, resp, "unf")
			if !s.Draftable || len(s.Blockers) != 0 {
				t.Fatalf("an unflagged set whose cards are all Full: %+v, want draftable", s)
			}
			if s.Cards != len(want) {
				t.Fatalf("Cards = %d, want the %d cards of the main run", s.Cards, len(want))
			}

			// The one rule everywhere: PoolFor, the list's own pool and the
			// image scope read the same cards.
			if got := contentNames(PoolFor(idx, "unf")); !reflect.DeepEqual(got, want) {
				t.Fatalf("PoolFor holds %v, want %v", got, want)
			}
			l := NewListerWith(idx, allFullExcept())
			pool, _, ok := l.DraftablePool("unf")
			if !ok {
				t.Fatal("the lister has no draftable pool for the set")
			}
			if got := contentNames(pool); !reflect.DeepEqual(got, want) {
				t.Fatalf("the lister's pool holds %v, want %v", got, want)
			}
			inRun := map[string]bool{}
			for _, n := range want {
				inRun[n] = true
			}
			idx.EachPrinting(func(c cards.Card) {
				n, plain := plainNumber(c.CollectorNo)
				mainPrinting := inRun[c.Name] && plain && n <= 12
				if got := l.IsDraftPrinting(c.ID); got != mainPrinting {
					t.Errorf("IsDraftPrinting(%s #%s) = %v, want %v", c.Name, c.CollectorNo, got, mainPrinting)
				}
			})
		})
	}
}

func TestUnflaggedSetLeavesOutWhatIsNotTheMainRun(t *testing.T) {
	idx := cards.NewIndex()
	unflaggedSet(idx, "unf", true)
	pool := PoolFor(idx, "unf")
	byName := map[string]Card{}
	for _, c := range pool.Contents() {
		byName[c.Name] = c
	}
	for _, name := range []string{"unf Promo Card", "unf Lettered Card", "unf Late Plain Card", "Plains", "Forest"} {
		if _, ok := byName[name]; ok {
			t.Errorf("%q is in the pool, but it is not in the main run", name)
		}
	}
	// A card that has a showcase reprint is stood for by its plain number.
	for _, name := range []string{"unf Common 00", "unf Rare 00", "unf Rare 01"} {
		c, ok := byName[name]
		if !ok {
			t.Fatalf("%q is not in the pool", name)
		}
		if n, _ := plainNumber(c.CollectorNo); n > 12 {
			t.Errorf("%q is stood for by its treated printing #%s", name, c.CollectorNo)
		}
	}
	if len(pool.Lands) != 0 {
		t.Errorf("the pool has lands %v; the basic lands lie past the main run", pool.Lands)
	}
	if got, want := len(pool.Commons), 6; got != want {
		t.Errorf("commons = %d, want %d", got, want)
	}
	if got, want := len(pool.Uncommons)+len(pool.Rares)+len(pool.Mythics), 6; got != want {
		t.Errorf("uncommons, rares and mythics = %d, want %d", got, want)
	}
}

func TestUnflaggedSetBlockersAreNamed(t *testing.T) {
	idx := cards.NewIndex()
	unflaggedSet(idx, "unf", false)
	s := findSet(t, sets(idx, allFullExcept("unf Rare 01", "unf Promo Card"), nil), "unf")
	if s.Draftable {
		t.Fatalf("a main-run card is not Full and the set is draftable: %+v", s)
	}
	// The promo is not in the main run, so it cannot block.
	if want := []string{"unf Rare 01"}; !reflect.DeepEqual(s.Blockers, want) {
		t.Fatalf("blockers = %v, want %v", s.Blockers, want)
	}
}

// TestAFlaggedSetIsReadByItsFlagAlone is the guard for the fallback: a set
// with even one flagged printing is built exactly as before, so its
// treated booster printings count and its unflagged plain ones do not.
func TestAFlaggedSetIsReadByItsFlagAlone(t *testing.T) {
	idx := cards.NewIndex()
	n := synthSet(idx, "syn", 12, 5, 3, 1, true)
	// Flagged and treated, and the only printing of its card: in the pool.
	show := synthCard("syn", "Flagged Showcase", "rare", fmt.Sprint(n))
	show.FrameEffects = []string{"showcase"}
	idx.Put(show)
	// Plain and in the main run's numbers, but not flagged: out.
	plain := synthCard("syn", "Unflagged Plain", "rare", fmt.Sprint(n+1))
	plain.Booster = false
	idx.Put(plain)
	// Flagged but past a gap in the numbers: in.
	late := synthCard("syn", "Flagged After A Gap", "rare", fmt.Sprint(n+50))
	idx.Put(late)

	got := contentNames(PoolFor(idx, "syn"))
	for _, name := range []string{"Flagged Showcase", "Flagged After A Gap"} {
		if !contains(got, name) {
			t.Errorf("%q has the flag and is not in the pool", name)
		}
	}
	if contains(got, "Unflagged Plain") {
		t.Error("a printing with no flag is in a set that has flags")
	}
	if want := 12 + 5 + 3 + 1 + 5 + 2; len(got) != want {
		t.Errorf("the pool holds %d cards, want %d", len(got), want)
	}
	s := findSet(t, sets(idx, allFullExcept(), nil), "syn")
	if s.Cards != len(got) {
		t.Errorf("the list says %d cards and PoolFor holds %d", s.Cards, len(got))
	}
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

func TestAnUnflaggedSetWithNoFillableMainRunStaysUnlisted(t *testing.T) {
	// No plain printing is numbered 1: the main run is empty.
	t.Run("no number 1", func(t *testing.T) {
		idx := cards.NewIndex()
		for i := 2; i < 20; i++ {
			rarity := []string{"common", "common", "uncommon", "rare"}[i%4]
			c := synthCard("nor", fmt.Sprintf("nor Card %02d", i), rarity, fmt.Sprint(i))
			c.Booster = false
			idx.Put(c)
		}
		assertUnlisted(t, idx, "nor")
	})
	// Only treated printings.
	t.Run("only treatments", func(t *testing.T) {
		idx := cards.NewIndex()
		for i := 1; i < 20; i++ {
			rarity := []string{"common", "common", "uncommon", "rare"}[i%4]
			c := synthCard("trt", fmt.Sprintf("trt Card %02d", i), rarity, fmt.Sprint(i))
			c.Booster = false
			c.FrameEffects = []string{"extendedart"}
			idx.Put(c)
		}
		assertUnlisted(t, idx, "trt")
	})
	// A main run with no uncommon or no rare cannot fill a pack.
	t.Run("run cannot fill a pack", func(t *testing.T) {
		idx := cards.NewIndex()
		for i := 1; i <= 10; i++ {
			c := synthCard("thr", fmt.Sprintf("thr Card %02d", i), "common", fmt.Sprint(i))
			c.Booster = false
			idx.Put(c)
		}
		assertUnlisted(t, idx, "thr")
	})
	// A gap ends the run: what lies after it is not read, so the part
	// before it is all there is, and it has no rare.
	t.Run("a gap before the rares", func(t *testing.T) {
		idx := cards.NewIndex()
		put := func(name, rarity string, n int) {
			c := synthCard("gap", name, rarity, fmt.Sprint(n))
			c.Booster = false
			idx.Put(c)
		}
		for i := 1; i <= 6; i++ {
			put(fmt.Sprintf("gap Common %d", i), "common", i)
		}
		put("gap Uncommon", "uncommon", 7)
		put("gap Rare", "rare", 9)
		assertUnlisted(t, idx, "gap")
	})
}

func assertUnlisted(t *testing.T, idx *cards.Index, code string) {
	t.Helper()
	for _, s := range sets(idx, allFullExcept(), nil).Sets {
		if s.Code == code {
			t.Fatalf("set %q is listed: %+v", code, s)
		}
	}
	if PoolFor(idx, code).CanFill() {
		t.Fatalf("PoolFor(%q) can fill a pack", code)
	}
}

func TestMainRunLength(t *testing.T) {
	for _, tc := range []struct {
		nums []int
		want int
	}{
		{nil, 0},
		{[]int{2, 3, 4}, 0},
		{[]int{1}, 1},
		{[]int{1, 2, 3, 5, 6}, 3},
		{[]int{3, 1, 2, 307, 308}, 3},
	} {
		m := map[int]bool{}
		for _, n := range tc.nums {
			m[n] = true
		}
		if got := mainRunLength(m); got != tc.want {
			t.Errorf("mainRunLength(%v) = %d, want %d", tc.nums, got, tc.want)
		}
	}
}

func TestPlainNumber(t *testing.T) {
	for s, want := range map[string]int{"1": 1, "214": 214, "007": 7} {
		if n, ok := plainNumber(s); !ok || n != want {
			t.Errorf("plainNumber(%q) = %d, %v, want %d", s, n, ok, want)
		}
	}
	for _, s := range []string{"", "214a", "214★", "A-12", "1701D", "123456789", "-3"} {
		if n, ok := plainNumber(s); ok {
			t.Errorf("plainNumber(%q) = %d, true, want false", s, n)
		}
	}
}
