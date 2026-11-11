package draftpacks

import (
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// synthNS namespaces the made-up IDs of the synthetic sets below.
var synthNS = uuid.MustParse("00000000-0000-0000-0000-0000000d4a47")

func synthID(parts ...any) uuid.UUID {
	return uuid.NewSHA1(synthNS, []byte(fmt.Sprint(parts...)))
}

// synthCard is a made-up booster printing. The oracle ID is derived from
// the name, so two printings with one name are one card.
func synthCard(set, name, rarity, number string) cards.Card {
	return cards.Card{
		ID:          synthID("print", set, number, name),
		OracleID:    synthID("oracle", name),
		Name:        name,
		SetCode:     set,
		SetName:     "Synthetic " + set,
		SetType:     "expansion",
		CollectorNo: number,
		Lang:        "en",
		Layout:      "normal",
		TypeLine:    "Creature — Test",
		BorderColor: "black",
		ReleasedAt:  "2026-03-01",
		Rarity:      rarity,
		Booster:     true,
	}
}

// synthSet puts a made-up set into idx: the given number of commons,
// uncommons, rares and mythics, and the five basic lands when basics is
// true. It returns the next free collector number.
func synthSet(idx *cards.Index, set string, commons, uncommons, rares, mythics int, basics bool) int {
	n := 1
	put := func(prefix, rarity string, count int) {
		for i := range count {
			idx.Put(synthCard(set, fmt.Sprintf("%s %s %02d", set, prefix, i), rarity, fmt.Sprint(n)))
			n++
		}
	}
	put("Common", "common", commons)
	put("Uncommon", "uncommon", uncommons)
	put("Rare", "rare", rares)
	put("Mythic", "mythic", mythics)
	if basics {
		for _, b := range []string{"Plains", "Island", "Swamp", "Mountain", "Forest"} {
			c := synthCard(set, b, "common", fmt.Sprint(n))
			c.TypeLine = "Basic Land — " + b
			idx.Put(c)
			n++
		}
	}
	return n
}

// retailSized is a made-up set about the size of a real one.
func retailSized(t *testing.T) *cards.Index {
	t.Helper()
	idx := cards.NewIndex()
	synthSet(idx, "syn", 80, 100, 60, 20, true)
	return idx
}

func seeded(seed uint64) *rand.Rand {
	return rand.New(rand.NewPCG(seed, 0x5437))
}

func slotCounts(p Pack) map[Slot]int {
	out := map[Slot]int{}
	for _, c := range p.Cards {
		out[c.Slot]++
	}
	return out
}

func TestPlayBoosterRecipeIsFourteenCards(t *testing.T) {
	if got := PlayBoosterRecipe().Size(); got != PlayBoosterSize {
		t.Fatalf("PlayBoosterRecipe().Size() = %d, want PlayBoosterSize = %d", got, PlayBoosterSize)
	}
	if PlayBoosterSize != 14 {
		t.Fatalf("PlayBoosterSize = %d, want 14 (a play booster's playable cards)", PlayBoosterSize)
	}
	// The synthetic sets of these tests are dated 2026, so the tests
	// below that take the set's own recipe run on a play booster.
	if got := RecipeFor("syn", "2026-03-01"); got != PlayBoosterRecipe() {
		t.Fatalf("a set released in 2026 gets %+v, want the play booster", got)
	}
}

func TestPackSlotCounts(t *testing.T) {
	pool := PoolFor(retailSized(t), "syn")
	for seed := range uint64(500) {
		pack, err := pool.Pack(seeded(seed))
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		if len(pack.Cards) != PlayBoosterSize {
			t.Fatalf("seed %d: %d cards, want %d", seed, len(pack.Cards), PlayBoosterSize)
		}
		got := slotCounts(pack)
		// No bonus sheet in the data: the seventh common slot is a common.
		want := map[Slot]int{SlotCommon: 7, SlotUncommon: 3, SlotRare: 1, SlotLand: 1, SlotWildcard: 2}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("seed %d: slots %v, want %v", seed, got, want)
		}
		for i, c := range pack.Cards {
			switch c.Slot {
			case SlotCommon:
				if c.Rarity != Common || c.Basic {
					t.Fatalf("seed %d card %d: common slot holds %q (%s, basic=%v)", seed, i, c.Name, c.Rarity, c.Basic)
				}
			case SlotUncommon:
				if c.Rarity != Uncommon {
					t.Fatalf("seed %d card %d: uncommon slot holds a %s", seed, i, c.Rarity)
				}
			case SlotRare:
				if c.Rarity != Rare && c.Rarity != Mythic {
					t.Fatalf("seed %d card %d: rare slot holds a %s", seed, i, c.Rarity)
				}
			case SlotLand:
				if !c.Basic {
					t.Fatalf("seed %d card %d: land slot holds %q, not a basic land", seed, i, c.Name)
				}
			case SlotWildcard:
				if c.Basic {
					t.Fatalf("seed %d card %d: wildcard is a basic land", seed, i)
				}
			}
		}
		// Slot order is the pack's order.
		order := []Slot{SlotCommon, SlotCommon, SlotCommon, SlotCommon, SlotCommon, SlotCommon, SlotCommon,
			SlotUncommon, SlotUncommon, SlotUncommon, SlotRare, SlotLand, SlotWildcard, SlotWildcard}
		for i, c := range pack.Cards {
			if c.Slot != order[i] {
				t.Fatalf("seed %d: card %d is in slot %s, want %s", seed, i, c.Slot, order[i])
			}
		}
	}
}

// within fails unless got is within tol standard deviations of a
// binomial rate p over n trials. Seeds are fixed, so this never flakes:
// it fails only when the rate itself is changed.
func within(t *testing.T, what string, hits, n int, p float64) {
	t.Helper()
	const tol = 4.0
	got := float64(hits) / float64(n)
	sd := math.Sqrt(p * (1 - p) / float64(n))
	if math.Abs(got-p) > tol*sd {
		t.Errorf("%s: %d of %d = %.4f, want %.4f ± %.4f", what, hits, n, got, p, tol*sd)
	}
}

func TestPackRarityRates(t *testing.T) {
	idx := retailSized(t)
	// A bonus sheet, so the seventh-slot rate is measured too.
	for i := range 40 {
		idx.Put(synthCard("bon", fmt.Sprintf("Bonus %02d", i), "mythic", fmt.Sprint(i+1)))
	}
	sheet := "bon"
	pool := poolFor(idx, "syn", Override{BonusSheet: &sheet}.Apply(PlayBoosterRecipe()))
	if len(pool.Bonus) != 40 {
		t.Fatalf("bonus sheet has %d cards, want 40", len(pool.Bonus))
	}

	const packs = 40_000
	var mythicRare, bonus, wild, wildC, wildU, wildR, wildM int
	for seed := range uint64(packs) {
		pack, err := pool.Pack(seeded(seed))
		if err != nil {
			t.Fatal(err)
		}
		if len(pack.Cards) != PlayBoosterSize {
			t.Fatalf("seed %d: %d cards", seed, len(pack.Cards))
		}
		for _, c := range pack.Cards {
			switch c.Slot {
			case SlotRare:
				if c.Rarity == Mythic {
					mythicRare++
				}
			case SlotBonus:
				bonus++
				if c.Set != "bon" {
					t.Fatalf("seed %d: bonus slot holds a %q card", seed, c.Set)
				}
			case SlotWildcard:
				wild++
				switch c.Rarity {
				case Common:
					wildC++
				case Uncommon:
					wildU++
				case Rare:
					wildR++
				case Mythic:
					wildM++
				}
			}
		}
	}
	within(t, "mythic in the rare slot", mythicRare, packs, 1.0/MythicOneIn)
	within(t, "bonus sheet in the seventh common slot", bonus, packs, 1.0/BonusSheetOneIn)
	if wild != packs*WildcardSlots {
		t.Fatalf("%d wildcards over %d packs, want %d", wild, packs, packs*WildcardSlots)
	}
	total := float64(WildcardCommonWeight + WildcardUncommonWeight + WildcardRareOrMythicWeight)
	pRM := WildcardRareOrMythicWeight / total
	within(t, "common wildcard", wildC, wild, WildcardCommonWeight/total)
	within(t, "uncommon wildcard", wildU, wild, WildcardUncommonWeight/total)
	within(t, "rare wildcard", wildR, wild, pRM*(1-1.0/MythicOneIn))
	within(t, "mythic wildcard", wildM, wild, pRM/MythicOneIn)
}

func TestPackIsDeterministicForASeed(t *testing.T) {
	// Two separately built indexes: map iteration order differs between
	// them, and the pack must not.
	a, err := Generate(retailSized(t), "syn", seeded(42))
	if err != nil {
		t.Fatal(err)
	}
	for range 20 {
		b, err := Generate(retailSized(t), "SYN", seeded(42))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("the same seed gave two packs:\n%v\n%v", a, b)
		}
	}
	c, err := Generate(retailSized(t), "syn", seeded(43))
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(a, c) {
		t.Fatal("seeds 42 and 43 gave the same pack")
	}
}

func TestPackHasNoDuplicateCard(t *testing.T) {
	// Barely enough of each rarity: 9 commons cover seven common slots
	// and two common wildcards.
	idx := cards.NewIndex()
	synthSet(idx, "tiny", 9, 5, 2, 1, true)
	pool := PoolFor(idx, "tiny")
	for seed := range uint64(5000) {
		pack, err := pool.Pack(seeded(seed))
		if err != nil {
			t.Fatal(err)
		}
		seen := map[uuid.UUID]string{}
		for _, c := range pack.Cards {
			if prev, dup := seen[c.OracleID]; dup {
				t.Fatalf("seed %d: %q and %q are the same card in one pack", seed, prev, c.Name)
			}
			seen[c.OracleID] = c.Name
		}
	}
}

func TestPackRepeatsOnlyWhenTheDataForcesIt(t *testing.T) {
	// Three commons cannot fill seven common slots without a repeat. The
	// pack is still 14 cards, and the uncommons, which are enough, do
	// not repeat.
	idx := cards.NewIndex()
	synthSet(idx, "few", 3, 5, 2, 1, true)
	pool := PoolFor(idx, "few")
	for seed := range uint64(200) {
		pack, err := pool.Pack(seeded(seed))
		if err != nil {
			t.Fatal(err)
		}
		if len(pack.Cards) != PlayBoosterSize {
			t.Fatalf("seed %d: %d cards, want %d", seed, len(pack.Cards), PlayBoosterSize)
		}
		commons := map[uuid.UUID]bool{}
		uncommons := map[uuid.UUID]int{}
		for _, c := range pack.Cards {
			switch c.Rarity {
			case Common:
				if !c.Basic {
					commons[c.OracleID] = true
				}
			case Uncommon:
				uncommons[c.OracleID]++
			}
		}
		if len(commons) != 3 {
			t.Fatalf("seed %d: %d different commons, want all 3 before any repeat", seed, len(commons))
		}
		for id, n := range uncommons {
			if n > 1 {
				t.Fatalf("seed %d: uncommon %s appears %d times with spare uncommons in the set", seed, id, n)
			}
		}
	}
}

func TestLandSlotIsACommonWhenTheSetHasNoBasics(t *testing.T) {
	idx := cards.NewIndex()
	synthSet(idx, "nob", 20, 10, 5, 2, false)
	pack, err := Generate(idx, "nob", seeded(1))
	if err != nil {
		t.Fatal(err)
	}
	if len(pack.Cards) != PlayBoosterSize {
		t.Fatalf("%d cards, want %d", len(pack.Cards), PlayBoosterSize)
	}
	for _, c := range pack.Cards {
		if c.Slot == SlotLand && (c.Rarity != Common || c.Basic) {
			t.Fatalf("land slot holds %q (%s)", c.Name, c.Rarity)
		}
	}
}

func TestNoPackForAnUnknownOrUnfillableSet(t *testing.T) {
	idx := cards.NewIndex()
	synthSet(idx, "norare", 20, 10, 0, 0, true)
	for _, set := range []string{"nope", "norare"} {
		if _, err := Generate(idx, set, seeded(1)); !errors.Is(err, ErrNoPack) {
			t.Errorf("Generate(%q) error = %v, want ErrNoPack", set, err)
		}
	}
	if _, err := Generate(nil, "syn", seeded(1)); !errors.Is(err, ErrNoPack) {
		t.Errorf("Generate on a nil index: error = %v, want ErrNoPack", err)
	}
}

func TestOnePlainPrintingPerCard(t *testing.T) {
	idx := cards.NewIndex()
	n := synthSet(idx, "art", 12, 5, 3, 1, true)

	plain := synthCard("art", "Many Faces", "rare", fmt.Sprint(n))
	showcase := synthCard("art", "Many Faces", "rare", fmt.Sprint(n+100))
	showcase.FrameEffects = []string{"showcase"}
	showcase.PromoTypes = []string{"boosterfun"}
	borderless := synthCard("art", "Many Faces", "rare", fmt.Sprint(n+200))
	borderless.BorderColor = "borderless"
	// The treatments go in first and last, so insertion order cannot be
	// what picks the plain one.
	idx.Put(showcase)
	idx.Put(plain)
	idx.Put(borderless)

	// A second art of a basic land is the same card.
	forest2 := synthCard("art", "Forest", "common", fmt.Sprint(n+300))
	forest2.TypeLine = "Basic Land — Forest"
	forest2.FullArt = true
	idx.Put(forest2)

	// Things a retail pack has that this model does not.
	token := synthCard("art", "Soldier", "common", "T1")
	token.Layout = "token"
	token.TypeLine = "Token Creature — Soldier"
	idx.Put(token)
	notInPacks := synthCard("art", "Box Topper", "mythic", "900")
	notInPacks.Booster = false
	idx.Put(notInPacks)
	digital := synthCard("art", "Arena Only", "rare", "901")
	digital.Digital = true
	idx.Put(digital)
	special := synthCard("art", "Special Guest", "special", "902")
	idx.Put(special)

	pool := PoolFor(idx, "art")
	var found []Card
	for _, c := range pool.Rares {
		if c.Name == "Many Faces" {
			found = append(found, c)
		}
	}
	if len(found) != 1 || found[0].ID != plain.ID {
		t.Fatalf("Many Faces in the pool: %+v, want only the plain printing %s", found, plain.ID)
	}
	if len(pool.Lands) != 5 {
		t.Fatalf("%d basic lands in the pool, want 5 (one printing each)", len(pool.Lands))
	}
	for _, c := range pool.Lands {
		if c.Name == "Forest" && c.ID == forest2.ID {
			t.Fatal("the full-art Forest stands for Forest, want the plain one")
		}
	}
	for _, c := range pool.Contents() {
		switch c.Name {
		case "Soldier", "Box Topper", "Arena Only", "Special Guest":
			t.Errorf("%q is in the pool; a pack must not contain it", c.Name)
		}
	}
	if got, want := len(pool.Contents()), 12+5+3+1+5+1; got != want {
		t.Fatalf("pool holds %d cards, want %d", got, want)
	}
}

func TestOverridesAreDataPerSet(t *testing.T) {
	table, err := parseOverrides([]byte(`{"sets":{"SYN":{"note":"test","mythic_one_in":2,"bonus_sheet":"BON","wildcards":1,"uncommons":4}}}`))
	if err != nil {
		t.Fatal(err)
	}
	r := recipeFor(table, "syn", "2026-03-01")
	if r.MythicOneIn != 2 || r.BonusSheet != "bon" || r.Wildcards != 1 || r.Uncommons != 4 {
		t.Fatalf("override not applied: %+v", r)
	}
	if r.Commons != CommonSlots || r.BonusSheetOneIn != BonusSheetOneIn || r.Kind != KindPlay {
		t.Fatalf("an unset field lost its era's value: %+v", r)
	}
	if got := recipeFor(table, "other", "2026-03-01"); got != PlayBoosterRecipe() {
		t.Fatalf("a set with no override got %+v, want its era's recipe", got)
	}

	// The override changes the packs.
	pool := poolFor(retailSized(t), "syn", r)
	mythics := 0
	const packs = 4000
	for seed := range uint64(packs) {
		pack, err := pool.Pack(seeded(seed))
		if err != nil {
			t.Fatal(err)
		}
		if len(pack.Cards) != r.Size() {
			t.Fatalf("%d cards, want %d", len(pack.Cards), r.Size())
		}
		for _, c := range pack.Cards {
			if c.Slot == SlotRare && c.Rarity == Mythic {
				mythics++
			}
		}
	}
	within(t, "mythic in the rare slot at 1 in 2", mythics, packs, 0.5)
}

func TestOverridesFileParses(t *testing.T) {
	table, err := parseOverrides(overridesJSON)
	if err != nil {
		t.Fatal(err)
	}
	for code, o := range table {
		for _, e := range Eras {
			if r := o.Apply(e.Recipe); r.Size() <= 0 {
				t.Errorf("override for %q leaves an empty pack in the era %q", code, e.Name)
			}
		}
		if o.Kind != nil && *o.Kind != string(KindPlay) && *o.Kind != string(KindDraft) {
			t.Errorf("override for %q names the booster kind %q, want play or draft", code, *o.Kind)
		}
		if o.Note == "" {
			t.Errorf("override for %q has no note saying where the correction comes from", code)
		}
	}
}
