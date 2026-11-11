package draftpacks

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// datedSet is synthSet for a set released on the given date. The date
// goes on the set table, which is where the server reads it from when
// sets.json is loaded.
func datedSet(idx *cards.Index, set, released string, commons, uncommons, rares, mythics int, basics bool) {
	synthSet(idx, set, commons, uncommons, rares, mythics, basics)
	idx.PutSet(cards.SetInfo{Code: set, Name: "Synthetic " + set, SetType: "expansion", ReleasedAt: released})
}

func TestErasAreOrderedAndSized(t *testing.T) {
	if len(Eras) != 3 {
		t.Fatalf("%d eras, want 3", len(Eras))
	}
	if Eras[0].From != "" {
		t.Fatalf("the oldest era starts at %q, want no beginning", Eras[0].From)
	}
	for i := 1; i < len(Eras); i++ {
		if Eras[i].From <= Eras[i-1].From {
			t.Fatalf("era %q (from %q) does not come after %q (from %q)",
				Eras[i].Name, Eras[i].From, Eras[i-1].Name, Eras[i-1].From)
		}
	}
	seen := map[string]bool{}
	for _, e := range Eras {
		if e.Name == "" || seen[e.Name] {
			t.Fatalf("era name %q is empty or used twice", e.Name)
		}
		seen[e.Name] = true
	}

	pre, draft, play := Eras[0].Recipe, Eras[1].Recipe, Eras[2].Recipe

	// Before mythic rares: 11 commons, 3 uncommons, 1 rare.
	if pre.Kind != KindDraft || pre.Size() != DraftBoosterSize || pre.Size() != 15 {
		t.Fatalf("pre-mythic draft booster: kind %q, %d cards, want draft and 15", pre.Kind, pre.Size())
	}
	if pre.Commons != 11 || pre.Uncommons != 3 || pre.Lands != 0 || pre.Wildcards != 0 || pre.BonusSlots != 0 {
		t.Fatalf("pre-mythic draft booster slots: %+v", pre)
	}
	if pre.HasMythics() || pre.HasLands() || pre.UsesBonusSheet() {
		t.Fatalf("a pre-mythic draft booster has a mythic, a land slot or a bonus sheet: %+v", pre)
	}

	// With mythic rares: 10 commons, 3 uncommons, 1 rare (a mythic 1 in
	// 8), 1 basic land.
	if draft.Kind != KindDraft || draft.Size() != DraftBoosterSize {
		t.Fatalf("draft booster: kind %q, %d cards, want draft and 15", draft.Kind, draft.Size())
	}
	if draft.Commons != 10 || draft.Uncommons != 3 || draft.Lands != 1 || draft.MythicOneIn != 8 ||
		draft.Wildcards != 0 || draft.BonusSlots != 0 {
		t.Fatalf("draft booster slots: %+v", draft)
	}

	// The play booster is the 14-card recipe the package began with.
	if play != PlayBoosterRecipe() || play.Kind != KindPlay || play.Size() != 14 {
		t.Fatalf("play booster: %+v", play)
	}
	if play.Commons != 6 || play.BonusSlots != 1 || play.Uncommons != 3 || play.Lands != 1 ||
		play.Wildcards != 2 || play.MythicOneIn != 7 || play.BonusSheetOneIn != 8 {
		t.Fatalf("play booster slots: %+v", play)
	}
}

// TestEraBoundaries: a set released the day before a boundary is in the
// older era, and one released on it is in the newer.
func TestEraBoundaries(t *testing.T) {
	if MythicRaresFrom != "2008-10-03" || PlayBoostersFrom != "2024-02-09" {
		t.Fatalf("era boundaries are %s and %s, want 2008-10-03 and 2024-02-09", MythicRaresFrom, PlayBoostersFrom)
	}
	pre, draft, play := PreMythicDraftBoosterRecipe(), DraftBoosterRecipe(), PlayBoosterRecipe()
	cases := []struct {
		released string
		want     Recipe
		why      string
	}{
		{"1993-08-05", pre, "the first set"},
		{"2008-10-02", pre, "the day before mythic rares"},
		{"2008-10-03", draft, "the day mythic rares arrived"},
		{"2008-10-04", draft, "the day after"},
		{"2024-02-08", draft, "the day before play boosters"},
		{"2024-02-09", play, "the day play boosters arrived"},
		{"2024-02-10", play, "the day after"},
		{"2026-11-11", play, "today"},
		// No usable date: the booster sold today.
		{"", play, "no date"},
		{"soon", play, "not a date"},
		{"2024-13-45", play, "not a real date"},
	}
	idx := cards.NewIndex()
	for i, c := range cases {
		if got := RecipeFor("anyset", c.released); got != c.want {
			t.Errorf("RecipeFor(released %q, %s) = %+v, want %+v", c.released, c.why, got, c.want)
		}
		if got := EraFor(c.released).Recipe; got != c.want {
			t.Errorf("EraFor(%q).Recipe = %+v, want %+v", c.released, got, c.want)
		}
		datedSet(idx, fmt.Sprintf("e%02d", i), c.released, 14, 5, 3, 2, true)
	}

	// The same through the pool and the list, which read the date from
	// the data.
	resp := sets(idx, allFullExcept(), nil)
	for i, c := range cases {
		code := fmt.Sprintf("e%02d", i)
		if got := PoolFor(idx, code).Recipe; got != c.want {
			t.Errorf("PoolFor of a set released %q (%s): recipe %+v, want %+v", c.released, c.why, got, c.want)
		}
		if got := findSet(t, resp, code).Booster; got != c.want.Kind {
			t.Errorf("a set released %q (%s) is listed with the %q booster, want %q", c.released, c.why, got, c.want.Kind)
		}
	}
}

// TestEraComesFromThePrintingsWithoutASetTable: a deployment with no
// sets.json still dates a set, by its printings' released_at.
func TestEraComesFromThePrintingsWithoutASetTable(t *testing.T) {
	idx := cards.NewIndex()
	n := 1
	put := func(name, rarity string) {
		c := synthCard("pri", name, rarity, fmt.Sprint(n))
		c.ReleasedAt = "2015-07-17"
		idx.Put(c)
		n++
	}
	for i := range 14 {
		put(fmt.Sprintf("pri Common %02d", i), "common")
	}
	for i := range 5 {
		put(fmt.Sprintf("pri Uncommon %02d", i), "uncommon")
	}
	put("pri Rare", "rare")
	if got := PoolFor(idx, "pri").Recipe; got != DraftBoosterRecipe() {
		t.Fatalf("recipe %+v, want the draft booster", got)
	}
	s := findSet(t, sets(idx, allFullExcept(), nil), "pri")
	if s.Booster != KindDraft || s.Released != "2015-07-17" {
		t.Fatalf("listed as %+v, want a draft booster released 2015-07-17", s)
	}
}

func TestDraftBoosterSlotsAndMythicRate(t *testing.T) {
	idx := cards.NewIndex()
	datedSet(idx, "dra", "2015-07-17", 80, 100, 60, 20, true)
	pool := PoolFor(idx, "dra")

	order := []Slot{SlotCommon, SlotCommon, SlotCommon, SlotCommon, SlotCommon,
		SlotCommon, SlotCommon, SlotCommon, SlotCommon, SlotCommon,
		SlotUncommon, SlotUncommon, SlotUncommon, SlotRare, SlotLand}
	const packs = 40_000
	mythics := 0
	for seed := range uint64(packs) {
		pack, err := pool.Pack(seeded(seed))
		if err != nil {
			t.Fatal(err)
		}
		if len(pack.Cards) != DraftBoosterSize {
			t.Fatalf("seed %d: %d cards, want %d", seed, len(pack.Cards), DraftBoosterSize)
		}
		for i, c := range pack.Cards {
			if c.Slot != order[i] {
				t.Fatalf("seed %d: card %d is in slot %s, want %s", seed, i, c.Slot, order[i])
			}
			switch c.Slot {
			case SlotCommon:
				if c.Rarity != Common || c.Basic {
					t.Fatalf("seed %d card %d: common slot holds %q (%s)", seed, i, c.Name, c.Rarity)
				}
			case SlotUncommon:
				if c.Rarity != Uncommon {
					t.Fatalf("seed %d card %d: uncommon slot holds a %s", seed, i, c.Rarity)
				}
			case SlotRare:
				if c.Rarity != Rare && c.Rarity != Mythic {
					t.Fatalf("seed %d card %d: rare slot holds a %s", seed, i, c.Rarity)
				}
				if c.Rarity == Mythic {
					mythics++
				}
			case SlotLand:
				if !c.Basic {
					t.Fatalf("seed %d card %d: land slot holds %q", seed, i, c.Name)
				}
			}
		}
	}
	within(t, "mythic in a draft booster's rare slot", mythics, packs, 1.0/DraftMythicOneIn)
}

func TestPreMythicDraftBoosterSlots(t *testing.T) {
	idx := cards.NewIndex()
	// The data gives this old set mythics and basic lands anyway: the
	// recipe has no slot for either, so no pack may hold one.
	datedSet(idx, "anc", "1999-10-04", 80, 100, 60, 20, true)
	pool := PoolFor(idx, "anc")
	if len(pool.Mythics) != 0 || len(pool.Lands) != 0 || len(pool.Bonus) != 0 {
		t.Fatalf("the pool holds %d mythics, %d lands and %d bonus cards, want none",
			len(pool.Mythics), len(pool.Lands), len(pool.Bonus))
	}
	want := map[Slot]int{SlotCommon: 11, SlotUncommon: 3, SlotRare: 1}
	for seed := range uint64(5000) {
		pack, err := pool.Pack(seeded(seed))
		if err != nil {
			t.Fatal(err)
		}
		if len(pack.Cards) != DraftBoosterSize {
			t.Fatalf("seed %d: %d cards, want %d", seed, len(pack.Cards), DraftBoosterSize)
		}
		if got := slotCounts(pack); !reflect.DeepEqual(got, want) {
			t.Fatalf("seed %d: slots %v, want %v", seed, got, want)
		}
		for i, c := range pack.Cards {
			if c.Rarity == Mythic || c.Basic {
				t.Fatalf("seed %d card %d: a pre-mythic pack holds %q (%s, basic=%v)", seed, i, c.Name, c.Rarity, c.Basic)
			}
			if c.Slot == SlotRare && c.Rarity != Rare {
				t.Fatalf("seed %d: the rare slot holds a %s", seed, c.Rarity)
			}
		}
		if last := pack.Cards[len(pack.Cards)-1]; last.Slot != SlotRare {
			t.Fatalf("seed %d: the last card is in slot %s, want the rare", seed, last.Slot)
		}
	}
}

func TestEveryEraIsDeterministicForASeed(t *testing.T) {
	for _, released := range []string{"1999-10-04", "2015-07-17", "2026-03-01"} {
		build := func() Pack {
			idx := cards.NewIndex()
			datedSet(idx, "det", released, 80, 100, 60, 20, true)
			p, err := Generate(idx, "det", seeded(7))
			if err != nil {
				t.Fatal(err)
			}
			return p
		}
		a := build()
		for range 10 {
			if b := build(); !reflect.DeepEqual(a, b) {
				t.Fatalf("released %s: the same seed gave two packs:\n%v\n%v", released, a, b)
			}
		}
	}
}

// TestASlotTheRecipeLacksCannotBlock: the draftable check reads the
// set's own recipe.
func TestASlotTheRecipeLacksCannotBlock(t *testing.T) {
	notFull := allFullExcept("old Mythic 00", "Bonus Card", "Plains")
	build := func(released string) *cards.Index {
		idx := cards.NewIndex()
		datedSet(idx, "old", released, 14, 5, 3, 1, true)
		bonus := synthCard("bon", "Bonus Card", "mythic", "1")
		bonus.SetType = "masterpiece"
		idx.Put(bonus)
		return idx
	}
	sheet := "bon"
	withSheet := map[string]Override{"old": {BonusSheet: &sheet}}

	// Before mythic rares: no mythic slot, no land slot, no bonus slot.
	idx := build("1999-10-04")
	s := findSet(t, sets(idx, notFull, withSheet), "old")
	if !s.Draftable || len(s.Blockers) != 0 {
		t.Fatalf("a pre-mythic set is blocked by %v; its packs hold none of them", s.Blockers)
	}
	if s.Cards != 14+5+3 {
		t.Fatalf("Cards = %d, want %d (no mythic, no basic land)", s.Cards, 14+5+3)
	}
	if !PoolFor(idx, "old").Draftable(notFull) {
		t.Fatal("Pool.Draftable is false for the pre-mythic set")
	}

	// A draft booster with mythics: the mythic blocks, the bonus sheet
	// still cannot (the recipe has no bonus slot), basics never do.
	idx = build("2015-07-17")
	s = findSet(t, sets(idx, notFull, withSheet), "old")
	if s.Draftable || !reflect.DeepEqual(s.Blockers, []string{"old Mythic 00"}) {
		t.Fatalf("a draft-booster set: %+v, want blocked by the mythic only", s)
	}
	if s.Cards != 14+5+3+1+5 {
		t.Fatalf("Cards = %d, want %d", s.Cards, 14+5+3+1+5)
	}

	// A play booster has the bonus slot, so the named sheet blocks too.
	idx = build("2026-03-01")
	s = findSet(t, sets(idx, notFull, withSheet), "old")
	if want := []string{"Bonus Card", "old Mythic 00"}; s.Draftable || !reflect.DeepEqual(s.Blockers, want) {
		t.Fatalf("a play-booster set: blockers %v, want %v", s.Blockers, want)
	}

	// A draft-booster set whose override gives it a bonus slot.
	one, rate := 1, 4
	idx = build("2015-07-17")
	withSlot := map[string]Override{"old": {BonusSheet: &sheet, BonusSlots: &one, BonusSheetOneIn: &rate}}
	s = findSet(t, sets(idx, notFull, withSlot), "old")
	if want := []string{"Bonus Card", "old Mythic 00"}; !reflect.DeepEqual(s.Blockers, want) {
		t.Fatalf("with a bonus slot from the override: blockers %v, want %v", s.Blockers, want)
	}
}

// TestAnOverrideWinsOverTheEra: overrides.json is laid over the era's
// recipe, whichever era that is.
func TestAnOverrideWinsOverTheEra(t *testing.T) {
	table, err := parseOverrides([]byte(`{"sets":{
		"OLD":{"note":"test","kind":"play","lands":1,"mythic_one_in":4},
		"MID":{"note":"test","bonus_sheet":"BON","bonus_slots":1,"bonus_sheet_one_in":4,"commons":9},
		"BAD":{"note":"test","kind":"collector"}}}`))
	if err != nil {
		t.Fatal(err)
	}

	old := recipeFor(table, "old", "1999-10-04")
	want := PreMythicDraftBoosterRecipe()
	want.Kind, want.Lands, want.MythicOneIn = KindPlay, 1, 4
	if old != want {
		t.Fatalf("override over the pre-mythic era: %+v, want %+v", old, want)
	}

	mid := recipeFor(table, "mid", "2015-07-17")
	want = DraftBoosterRecipe()
	want.BonusSheet, want.BonusSlots, want.BonusSheetOneIn, want.Commons = "bon", 1, 4, 9
	if mid != want {
		t.Fatalf("override over the draft era: %+v, want %+v", mid, want)
	}
	if mid.Size() != DraftBoosterSize || !mid.UsesBonusSheet() {
		t.Fatalf("the overridden draft booster: %d cards, uses the sheet %v", mid.Size(), mid.UsesBonusSheet())
	}

	// A kind that is not a draft product is ignored: set and collector
	// boosters are not modelled.
	if got := recipeFor(table, "bad", "2015-07-17"); got != DraftBoosterRecipe() {
		t.Fatalf("an unknown kind changed the recipe: %+v", got)
	}
	// A set with no entry keeps its era's recipe.
	if got := recipeFor(table, "none", "1999-10-04"); got != PreMythicDraftBoosterRecipe() {
		t.Fatalf("a set with no override got %+v", got)
	}

	// The overridden bonus slot changes the packs: 9 commons, then a
	// slot that is a bonus-sheet card 1 time in 4.
	idx := cards.NewIndex()
	datedSet(idx, "mid", "2015-07-17", 80, 100, 60, 20, true)
	for i := range 40 {
		idx.Put(synthCard("bon", fmt.Sprintf("Bonus %02d", i), "rare", fmt.Sprint(i+1)))
	}
	pool := poolFor(idx, "mid", mid)
	const packs = 8000
	bonus := 0
	for seed := range uint64(packs) {
		pack, err := pool.Pack(seeded(seed))
		if err != nil {
			t.Fatal(err)
		}
		if len(pack.Cards) != DraftBoosterSize {
			t.Fatalf("seed %d: %d cards", seed, len(pack.Cards))
		}
		if pack.Cards[9].Slot == SlotBonus {
			bonus++
		}
		for i, c := range pack.Cards {
			if (c.Slot == SlotBonus) != (c.Set == "bon") || (c.Slot == SlotBonus && i != 9) {
				t.Fatalf("seed %d card %d: slot %s holds a %q card", seed, i, c.Slot, c.Set)
			}
		}
	}
	within(t, "bonus sheet in an overridden draft booster", bonus, packs, 0.25)
}
