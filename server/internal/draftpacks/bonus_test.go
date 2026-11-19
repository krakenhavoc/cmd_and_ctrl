package draftpacks

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// bonusIndex is a play-booster set "syn" with two made-up bonus sets
// beside it: "bga", 30 cards, and "bgb", 50.
func bonusIndex(t *testing.T) *cards.Index {
	t.Helper()
	idx := retailSized(t)
	for i := range 30 {
		idx.Put(synthCard("bga", fmt.Sprintf("Guaranteed %02d", i), "uncommon", fmt.Sprint(i+1)))
	}
	for i := range 50 {
		idx.Put(synthCard("bgb", fmt.Sprintf("Sometimes %02d", i), "mythic", fmt.Sprint(i+1)))
	}
	return idx
}

func countSlots(p Pack, slot Slot, set string) int {
	n := 0
	for _, c := range p.Cards {
		if c.Slot == slot && (set == "" || c.Set == set) {
			n++
		}
	}
	return n
}

// TestBonusSlotsOpenAtTheirRates: each bonus slot opens at its own rate,
// in the place of a common, and the pack keeps its size.
func TestBonusSlotsOpenAtTheirRates(t *testing.T) {
	slots := []BonusSlot{
		{Name: "One in five", OneIn: 5, Sources: []Source{{Set: "bga"}}},
		{Name: "One in twenty", OneIn: 20, Sources: []Source{{Set: "bgb"}}},
	}
	pool := poolFor(bonusIndex(t), "syn", Override{Bonus: &slots}.Apply(PlayBoosterRecipe()))
	const packs = 40_000
	var a, b, both int
	for seed := range uint64(packs) {
		pack, err := pool.Pack(seeded(seed))
		if err != nil {
			t.Fatal(err)
		}
		if len(pack.Cards) != PlayBoosterSize {
			t.Fatalf("seed %d: %d cards, want %d", seed, len(pack.Cards), PlayBoosterSize)
		}
		na, nb := countSlots(pack, SlotBonus, "bga"), countSlots(pack, SlotBonus, "bgb")
		a += na
		b += nb
		if na+nb != countSlots(pack, SlotBonus, "") {
			t.Fatalf("seed %d: a bonus card from neither sheet", seed)
		}
		if want := CommonSlots - na - nb; countSlots(pack, SlotCommon, "") != want {
			t.Fatalf("seed %d: %d commons with %d bonus cards, want %d", seed, countSlots(pack, SlotCommon, ""), na+nb, want)
		}
		if na == 1 && nb == 1 {
			both++
		}
	}
	within(t, "the 1-in-5 slot", a, packs, 1.0/5)
	within(t, "the 1-in-20 slot", b, packs, 1.0/20)
	within(t, "both slots", both, packs, 1.0/100)
}

// TestAGuaranteedSlotIsInEveryPack: a slot at 1 in 1 is in every pack, in
// the place of a common, as Enchanting Tales or Breaking News is.
func TestAGuaranteedSlotIsInEveryPack(t *testing.T) {
	slots := []BonusSlot{{Name: "Every pack", OneIn: 1, Sources: []Source{{Set: "bga"}}}}
	pool := poolFor(bonusIndex(t), "syn", Override{Bonus: &slots}.Apply(PlayBoosterRecipe()))
	seen := map[string]bool{}
	for seed := range uint64(3000) {
		pack, err := pool.Pack(seeded(seed))
		if err != nil {
			t.Fatal(err)
		}
		if len(pack.Cards) != PlayBoosterSize {
			t.Fatalf("seed %d: %d cards, want %d", seed, len(pack.Cards), PlayBoosterSize)
		}
		if n := countSlots(pack, SlotBonus, "bga"); n != 1 {
			t.Fatalf("seed %d: %d cards of the guaranteed slot, want 1", seed, n)
		}
		if n := countSlots(pack, SlotCommon, ""); n != CommonSlots-1 {
			t.Fatalf("seed %d: %d commons, want %d", seed, n, CommonSlots-1)
		}
		// The bonus card sits after the commons it took the place of.
		if c := pack.Cards[CommonSlots-1]; c.Slot != SlotBonus {
			t.Fatalf("seed %d: card %d is in slot %s, want the bonus card", seed, CommonSlots-1, c.Slot)
		}
		for _, c := range pack.Cards {
			if c.Slot == SlotBonus {
				seen[c.Name] = true
			}
		}
	}
	if len(seen) != 30 {
		t.Fatalf("the guaranteed slot opened %d different cards over 3000 packs, want all 30", len(seen))
	}
}

// TestACollectorNumberSourcePicksOnlyThosePrintings: a range or a list
// names printings, and nothing else of the set comes along. The
// printings need no booster flag (The List and Special Guests are not
// boosters of their own).
func TestACollectorNumberSourcePicksOnlyThosePrintings(t *testing.T) {
	idx := retailSized(t)
	for i := 1; i <= 40; i++ {
		c := synthCard("gst", fmt.Sprintf("Guest %02d", i), "mythic", fmt.Sprint(i))
		c.Booster = false
		idx.Put(c)
	}
	for _, n := range []string{"ABC-12", "XYZ-3", "XYZ-4"} {
		c := synthCard("lst", "Listed "+n, "rare", n)
		c.Booster = false
		idx.Put(c)
	}
	// A treatment of a guest in the range: one printing per card, the
	// plain one.
	alt := synthCard("gst", "Guest 07", "mythic", "107")
	alt.BorderColor = "borderless"
	idx.Put(alt)

	slots := []BonusSlot{{Name: "Guests and a list", OneIn: 1, Sources: []Source{
		{Set: "gst", From: 5, To: 9},
		{Set: "lst", Numbers: []string{"ABC-12", "XYZ-4"}},
	}}}
	pool := poolFor(idx, "syn", Override{Bonus: &slots}.Apply(PlayBoosterRecipe()))
	var got []string
	for _, c := range pool.Bonus[0] {
		got = append(got, c.Set+" "+c.CollectorNo)
	}
	sort.Strings(got)
	want := []string{"gst 5", "gst 6", "gst 7", "gst 8", "gst 9", "lst ABC-12", "lst XYZ-4"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("the slot's printings are %v, want %v", got, want)
	}
	for seed := range uint64(500) {
		pack, err := pool.Pack(seeded(seed))
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range pack.Cards {
			if c.Slot == SlotBonus && !strings.HasPrefix(c.Name, "Guest 0") && !strings.HasPrefix(c.Name, "Listed ") {
				t.Fatalf("seed %d: the bonus slot holds %q", seed, c.Name)
			}
		}
	}
}

// TestABonusSlotCanReplaceTheWildcard: Edge of Eternities' Stellar
// Sights take the place of the wildcard 1 time in 8 and of the foil 1
// time in 80, not of a common.
func TestABonusSlotCanReplaceTheWildcard(t *testing.T) {
	slots := []BonusSlot{
		{Name: "Sights", OneIn: 8, Replaces: SlotWildcard, Sources: []Source{{Set: "bgb"}}},
		{Name: "Sights", OneIn: 80, Replaces: SlotFoil, Sources: []Source{{Set: "bgb"}}},
	}
	pool := poolFor(bonusIndex(t), "syn", Override{Bonus: &slots}.Apply(PlayBoosterRecipe()))
	const packs = 40_000
	var inWild, inFoil int
	for seed := range uint64(packs) {
		pack, err := pool.Pack(seeded(seed))
		if err != nil {
			t.Fatal(err)
		}
		if len(pack.Cards) != PlayBoosterSize {
			t.Fatalf("seed %d: %d cards, want %d", seed, len(pack.Cards), PlayBoosterSize)
		}
		if n := countSlots(pack, SlotCommon, ""); n != CommonSlots {
			t.Fatalf("seed %d: %d commons, want %d: the slot replaces no common", seed, n, CommonSlots)
		}
		wild, foil := countSlots(pack, SlotWildcard, ""), countSlots(pack, SlotFoil, "")
		// The pack's order says which slot each bonus card stands in.
		for i, c := range pack.Cards {
			if c.Slot != SlotBonus {
				continue
			}
			switch i {
			case PlayBoosterSize - 2:
				inWild++
			case PlayBoosterSize - 1:
				inFoil++
			default:
				t.Fatalf("seed %d: a bonus card at %d, want the wildcard's or the foil's place", seed, i)
			}
		}
		if wild+foil+countSlots(pack, SlotBonus, "") != 2 {
			t.Fatalf("seed %d: %d wildcards, %d foils and %d bonus cards, want 2 in all", seed, wild, foil, countSlots(pack, SlotBonus, ""))
		}
	}
	within(t, "a bonus card in place of the wildcard", inWild, packs, 1.0/8)
	within(t, "a bonus card in place of the foil", inFoil, packs, 1.0/80)
}

// TestABonusSlotWithNoCardsNeverOpens: a sheet the dump does not hold
// leaves the slot it would replace as it is.
func TestABonusSlotWithNoCardsNeverOpens(t *testing.T) {
	slots := []BonusSlot{{Name: "Absent", OneIn: 1, Sources: []Source{{Set: "zzz"}}}}
	pool := poolFor(retailSized(t), "syn", Override{Bonus: &slots}.Apply(PlayBoosterRecipe()))
	pack, err := pool.Pack(seeded(1))
	if err != nil {
		t.Fatal(err)
	}
	if n := countSlots(pack, SlotCommon, ""); n != CommonSlots || len(pack.Cards) != PlayBoosterSize {
		t.Fatalf("%d commons in %d cards, want %d in %d", n, len(pack.Cards), CommonSlots, PlayBoosterSize)
	}
}

// collatedSets are the sets whose DRAFT product MTGJSON collates in the
// file overrides.json was read from (draft/boosters.json.gz on the
// work/s82-cards branch): each has an entry, derived from it.
var collatedSets = []string{
	"woe", "lci", "mkm", "otj", "blb", "dsk", "fdn", "dft", "tdm", "fin", "eoe",
	"spm", "tla", "mom", "one", "bro", "dmu", "snc", "neo", "vow", "mid", "afr",
}

// eraOfKind is the recipe of the newest era with the given booster kind:
// the era a collated set falls in by its date.
func eraOfKind(t *testing.T, k Kind) Recipe {
	t.Helper()
	for i := len(Eras) - 1; i >= 0; i-- {
		if Eras[i].Recipe.Kind == k {
			return Eras[i].Recipe
		}
	}
	t.Fatalf("no era sells a %s booster", k)
	return Recipe{}
}

// TestEveryCollatedSetHasTheRealPackSize: each collated set's plan,
// its entry laid over its era, holds as many cards as the retail pack:
// 14 in a play booster, 15 in a draft booster. A bonus slot takes the
// place of a slot the plan has, never adds one.
func TestEveryCollatedSetHasTheRealPackSize(t *testing.T) {
	for _, code := range collatedSets {
		o, ok := overrides[code]
		if !ok {
			t.Errorf("%s has no entry in overrides.json", code)
			continue
		}
		if o.Kind == nil {
			t.Errorf("%s: the entry names no booster kind", code)
			continue
		}
		k := Kind(*o.Kind)
		r := o.Apply(eraOfKind(t, k))
		want := map[Kind]int{KindPlay: PlayBoosterSize, KindDraft: DraftBoosterSize}[k]
		if r.Size() != want {
			t.Errorf("%s: a %s booster of %d cards, want %d", code, k, r.Size(), want)
		}
		every := map[Slot]int{}
		for _, b := range r.Bonus {
			if !r.bonusOn(b) {
				t.Errorf("%s: bonus slot %q can never open in %+v", code, b.Name, r)
			}
			if b.OneIn == 1 {
				every[b.Slot()]++
			}
		}
		for s, n := range every {
			if n > r.slotCount(s) {
				t.Errorf("%s: %d bonus slots in every pack replace the %d %s slots", code, n, r.slotCount(s), s)
			}
		}
	}
}

// TestEveryOverrideIsDataForItsSet: each entry says where it comes from,
// MTGJSON's collation of that set, and is one of the collated sets. When
// CMDCTRL_MTGJSON_BOOSTERS points at that file (boosters.json, or the
// .gz), every number is checked against it.
func TestEveryOverrideIsDataForItsSet(t *testing.T) {
	collated := map[string]bool{}
	for _, c := range collatedSets {
		collated[c] = true
	}
	for code, o := range overrides {
		if !collated[code] {
			t.Errorf("overrides.json has %q, which is not a collated set: say where it comes from in the test", code)
		}
		if !strings.Contains(o.Note, "MTGJSON") || !strings.Contains(o.Note, strings.ToUpper(code)+".json") {
			t.Errorf("%s: the note does not cite MTGJSON's %s.json: %q", code, strings.ToUpper(code), o.Note)
		}
	}

	path := os.Getenv("CMDCTRL_MTGJSON_BOOSTERS")
	if path == "" {
		t.Skip("CMDCTRL_MTGJSON_BOOSTERS is not set: the numbers are not checked against MTGJSON")
	}
	data := loadCollation(t, path)
	for _, code := range collatedSets {
		t.Run(code, func(t *testing.T) {
			s, ok := data.Sets[code]
			if !ok {
				t.Fatalf("%s is not in %s", code, path)
			}
			checkAgainstCollation(t, code, overrides[code], s)
		})
	}
}

type collationCard struct {
	Set      string `json:"set"`
	Number   string `json:"number"`
	Rarity   string `json:"rarity"`
	Weight   int64  `json:"weight"`
	OracleID string `json:"oracle_id"`
}

type collationSheet struct {
	TotalWeight int64           `json:"totalWeight"`
	Cards       []collationCard `json:"cards"`
}

type collationSet struct {
	Booster     string `json:"booster"`
	TotalWeight int64  `json:"boostersTotalWeight"`
	Boosters    []struct {
		Contents map[string]int `json:"contents"`
		Weight   int64          `json:"weight"`
	} `json:"boosters"`
	Sheets map[string]collationSheet `json:"sheets"`
}

type collationFile struct {
	Sets map[string]collationSet `json:"sets"`
}

func loadCollation(t *testing.T, path string) collationFile {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var r io.Reader = f
	if strings.HasSuffix(path, ".gz") {
		gz, err := gzip.NewReader(f)
		if err != nil {
			t.Fatal(err)
		}
		r = gz
	}
	var out collationFile
	if err := json.NewDecoder(r).Decode(&out); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return out
}

// nearestOneIn fails unless n is the whole number nearest 1/p.
func nearestOneIn(t *testing.T, what string, n int, p float64) {
	t.Helper()
	if p <= 0 || math.Abs(float64(n)-1/p) > 0.5 {
		t.Errorf("%s: 1 in %d, the data says 1 in %.2f", what, n, 1/p)
	}
}

func printingKey(set, number string) string { return strings.ToLower(set) + " " + number }

func checkAgainstCollation(t *testing.T, code string, o Override, s collationSet) {
	kind := map[string]Kind{"play": KindPlay, "draft": KindDraft}[s.Booster]
	if o.Kind == nil || Kind(*o.Kind) != kind {
		t.Errorf("booster kind %v, the data's product is %q", o.Kind, s.Booster)
	}

	// The rare slot: the mythic share of every rare-or-mythic sheet, by
	// the share of packs that have it.
	var mythic, slots float64
	for _, b := range s.Boosters {
		share := float64(b.Weight) / float64(s.TotalWeight)
		for name, n := range b.Contents {
			if !strings.Contains(strings.ToLower(name), "raremythic") {
				continue
			}
			sh := s.Sheets[name]
			var m int64
			for _, c := range sh.Cards {
				if c.Rarity == "mythic" {
					m += c.Weight
				}
			}
			mythic += share * float64(n) * float64(m) / float64(sh.TotalWeight)
			slots += share * float64(n)
		}
	}
	if o.MythicOneIn == nil {
		t.Errorf("no mythic rate")
	} else {
		nearestOneIn(t, "mythic in the rare slot", *o.MythicOneIn, mythic/slots)
	}

	// The wildcard and foil splits: the sheet's weights of the set's own
	// cards, per 10,000.
	for _, tc := range []struct {
		sheet string
		split *RaritySplit
	}{{"wildcard", o.Wildcard}, {"foil", o.Foil}} {
		sh, ok := s.Sheets[tc.sheet]
		if kind != KindPlay {
			if tc.split != nil {
				t.Errorf("a draft booster with a %s split", tc.sheet)
			}
			continue
		}
		if !ok || tc.split == nil {
			t.Errorf("the %s split: sheet %v, override %v", tc.sheet, ok, tc.split)
			continue
		}
		by := map[string]int64{}
		var total int64
		for _, c := range sh.Cards {
			if c.Set == code {
				by[c.Rarity] += c.Weight
				total += c.Weight
			}
		}
		got := map[string]int{"common": tc.split.Common, "uncommon": tc.split.Uncommon, "rare": tc.split.Rare, "mythic": tc.split.Mythic}
		for rarity, w := range got {
			if want := math.Round(10000 * float64(by[rarity]) / float64(total)); float64(w) != want {
				t.Errorf("%s %s: %d per 10,000, the data says %.0f", tc.sheet, rarity, w, want)
			}
		}
	}

	// Each bonus slot: its printings are exactly a sheet's (or, in place
	// of the wildcard or foil, that sheet's cards from outside the set),
	// at the sheet's rate.
	var bonus []BonusSlot
	if o.Bonus != nil {
		bonus = *o.Bonus
	}
	matched := map[string]bool{}
	for _, b := range bonus {
		named := map[string]bool{}
		for _, src := range b.Sources {
			for _, num := range src.Numbers {
				named[printingKey(src.Set, num)] = true
			}
			for n := src.From; n <= src.To && n > 0; n++ {
				named[printingKey(src.Set, fmt.Sprint(n))] = true
			}
		}
		switch b.Slot() {
		case SlotWildcard, SlotFoil:
			sh := s.Sheets[string(b.Slot())]
			var w int64
			outside := map[string]bool{}
			for _, c := range sh.Cards {
				if c.Set != code {
					w += c.Weight
					outside[printingKey(c.Set, c.Number)] = true
				}
			}
			if !reflect.DeepEqual(named, outside) {
				t.Errorf("%q: printings %d, the %s sheet's from outside the set %d", b.Name, len(named), b.Slot(), len(outside))
			}
			nearestOneIn(t, b.Name+" in place of the "+string(b.Slot()), b.OneIn, float64(w)/float64(sh.TotalWeight))
		default:
			found := ""
			for name, sh := range s.Sheets {
				in := map[string]bool{}
				for _, c := range sh.Cards {
					in[printingKey(c.Set, c.Number)] = true
				}
				if reflect.DeepEqual(in, named) {
					found = name
				}
			}
			if found == "" {
				t.Errorf("%q: no sheet holds exactly its %d printings", b.Name, len(named))
				continue
			}
			matched[found] = true
			var w int64
			for _, bo := range s.Boosters {
				if bo.Contents[found] > 0 {
					w += bo.Weight
				}
			}
			nearestOneIn(t, b.Name, b.OneIn, float64(w)/float64(s.TotalWeight))
		}
	}
	// No sheet of cards from outside the set is left out, but the foil
	// and wildcard (above) and ONE's prototype praetors, which take the
	// place of the rare in 3 packs of 180 and are not modelled.
	for name, sh := range s.Sheets {
		if matched[name] || name == "wildcard" || strings.HasPrefix(name, "foil") ||
			(code == "one" && name == "prototypePraetor") {
			continue
		}
		for _, c := range sh.Cards {
			if c.Set != code && c.Set != "" {
				t.Errorf("sheet %s holds %s %s from outside the set, and no bonus slot names it", name, c.Set, c.Number)
				break
			}
		}
	}
}
