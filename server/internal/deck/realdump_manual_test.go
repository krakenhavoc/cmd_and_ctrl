package deck

import (
	"bufio"
	"encoding/json"
	"os"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// realdump_manual_test.go — an opt-in smoke test against the actual
// Scryfall bulk dump, gated on CMDCTRL_SCRYFALL_DUMP so CI never
// touches it (the file is ~629MB and is not in the repo).
//
//	CMDCTRL_SCRYFALL_DUMP=data/scryfall/default-cards.json \
//	  go test ./internal/deck/ -run RealDump -v
//
// Unit fixtures pin the SHAPE of the import; this pins that the shape
// is the one 117,738 real printings actually have. It is where a
// placeholder record with a null type line, or a face array the
// importer cannot parse, would show up.
func TestRealDumpMultiFaceImport(t *testing.T) {
	path := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if path == "" {
		t.Skip("set CMDCTRL_SCRYFALL_DUMP to run against the real dump")
	}
	idx := cards.NewIndex()
	if _, err := idx.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	t.Logf("loaded %d printings", idx.Count())

	cases := []struct {
		lookup    string
		name      string
		typeLine  string
		manaCost  string
		layout    string
		faces     int
		castable  int
		power     int
		toughness int
	}{
		{
			lookup: "Sea Gate Restoration", name: "Sea Gate Restoration",
			typeLine: "Sorcery", manaCost: "{4}{U}{U}{U}",
			layout: "modal_dfc", faces: 2, castable: 2,
		},
		{
			lookup: "Kazandu Mammoth", name: "Kazandu Mammoth",
			typeLine: "Creature — Elephant", manaCost: "{1}{G}{G}",
			layout: "modal_dfc", faces: 2, castable: 2,
			power: 3, toughness: 3,
		},
		{
			lookup: "Aang, Swift Savior", name: "Aang, Swift Savior",
			typeLine: "Legendary Creature — Human Avatar Ally", manaCost: "{1}{W}{U}",
			layout: "transform", faces: 2, castable: 1,
			power: 2, toughness: 3,
		},
		{
			lookup: "Jace, Vryn's Prodigy", name: "Jace, Vryn's Prodigy",
			typeLine: "Legendary Creature — Human Wizard", manaCost: "{1}{U}",
			layout: "transform", faces: 2, castable: 1,
			toughness: 2,
		},
		{
			lookup: "Fire // Ice", name: "Fire",
			typeLine: "Instant", manaCost: "{1}{R}",
			layout: "split", faces: 2, castable: 1,
		},
		{
			lookup: "Sol Ring", name: "Sol Ring",
			typeLine: "Artifact", manaCost: "{1}",
			layout: "normal", faces: 0, castable: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.lookup, func(t *testing.T) {
			c, ok := idx.FindByName(tc.lookup)
			if !ok {
				t.Fatalf("%q not in the index", tc.lookup)
			}
			got := toGameCard(c, false)
			if got.Name != tc.name {
				t.Errorf("Name = %q, want %q", got.Name, tc.name)
			}
			if got.TypeLine != tc.typeLine {
				t.Errorf("TypeLine = %q, want %q", got.TypeLine, tc.typeLine)
			}
			if got.ManaCost != tc.manaCost {
				t.Errorf("ManaCost = %q, want %q", got.ManaCost, tc.manaCost)
			}
			if got.Layout != tc.layout {
				t.Errorf("Layout = %q, want %q", got.Layout, tc.layout)
			}
			if len(got.Faces) != tc.faces {
				t.Errorf("Faces = %d, want %d", len(got.Faces), tc.faces)
			}
			if n := len(got.CastableFaces()); n != tc.castable {
				t.Errorf("CastableFaces = %d, want %d", n, tc.castable)
			}
			if got.Power != tc.power || got.Toughness != tc.toughness {
				t.Errorf("P/T = %d/%d, want %d/%d",
					got.Power, got.Toughness, tc.power, tc.toughness)
			}
			if _, err := game.ParseCost(got.ManaCost); err != nil {
				t.Errorf("ParseCost(%q): %v", got.ManaCost, err)
			}
			eff := got.Effective()
			for _, ty := range append(append([]string{}, eff.Types...), eff.Subtypes...) {
				if ty == "//" || ty == "—" {
					t.Errorf("junk type %q", ty)
				}
			}
		})
	}
}

// TestRealDumpNoFreeMultiFaceSpells sweeps every Commander-legal
// printing in the dump for the class of bug ADR 0034 existed to kill:
// a multi-face card that imports with a cost the engine reads as free
// or cannot read at all, or with the literal type "//".
//
// The population is deliberately "cards a deck could legally
// contain". The unfiltered dump also holds Role TOKENS (six
// two-faced `Token Enchantment — Aura Role` records, correctly
// costless) and Mystery Booster playtest cards (Keeper of the Crown
// costs {2}{L}, a symbol no parser knows) — neither can reach a deck,
// both would otherwise show up here as false positives, and
// isPlayablePrint / the format-legality check already exclude them
// from every path the engine takes.
//
// At the time this shipped: 0 free, 0 unparseable, 0 junk types
// across all 27,000-odd Commander-legal printings.
func TestRealDumpNoFreeMultiFaceSpells(t *testing.T) {
	path := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if path == "" {
		t.Skip("set CMDCTRL_SCRYFALL_DUMP to run against the real dump")
	}
	var free, unparseable, junkTypes, scanned int
	var examples []string
	streamDump(t, path, func(c cards.Card) {
		scanned++
		got := toGameCard(c, false)
		if len(got.Faces) < 2 {
			return
		}
		if got.IsLand() {
			return
		}
		if got.ManaCost == "" {
			free++
			if len(examples) < 10 {
				examples = append(examples, "free: "+got.Name+" ("+got.Layout+")")
			}
			return
		}
		if _, err := game.ParseCost(got.ManaCost); err != nil {
			unparseable++
			if len(examples) < 10 {
				examples = append(examples, "unparseable: "+got.Name+" "+got.ManaCost)
			}
		}
		eff := got.Effective()
		for _, ty := range append(append([]string{}, eff.Types...), eff.Subtypes...) {
			if ty == "//" || ty == "—" {
				junkTypes++
				if len(examples) < 10 {
					examples = append(examples, "junk type: "+got.Name)
				}
				break
			}
		}
	})
	t.Logf("scanned %d printings; multi-face non-land: free=%d unparseable=%d junkTypes=%d",
		scanned, free, unparseable, junkTypes)
	for _, e := range examples {
		t.Logf("  %s", e)
	}
	if free > 0 {
		t.Errorf("%d multi-face non-land cards import with NO cost — "+
			"the face model did not reach them, and they are castable "+
			"for free", free)
	}
	if unparseable > 0 {
		t.Errorf("%d multi-face cards import a cost ParseCost rejects", unparseable)
	}
	if junkTypes > 0 {
		t.Errorf("%d multi-face cards still carry a \"//\" or \"—\" type", junkTypes)
	}
}

// streamDump decodes the bulk dump one record at a time. The file is
// ~629MB; decoding it into a slice would work but is needlessly
// hostile to a laptop, and the index's own loader is not reusable
// here because it keeps only the best printing per name and this
// sweep wants every record.
func streamDump(t *testing.T, path string, visit func(cards.Card)) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open dump: %v", err)
	}
	defer f.Close()
	dec := json.NewDecoder(bufio.NewReaderSize(f, 1<<20))
	if _, err := dec.Token(); err != nil { // opening '['
		t.Fatalf("dump is not a JSON array: %v", err)
	}
	for dec.More() {
		var c cards.Card
		if err := dec.Decode(&c); err != nil {
			t.Fatalf("decode record: %v", err)
		}
		// Placeholder printings the index also refuses: art series,
		// tokens, the reversible_card novelty run, and the "Card //
		// Card" records with no real type line. ~6,456 of them.
		switch c.Layout {
		case "art_series", "token", "double_faced_token", "reversible_card":
			continue
		}
		switch c.SetType {
		case "token", "art_series", "memorabilia", "minigame", "vanguard":
			continue
		}
		if c.TypeLine == "Card // Card" {
			continue
		}
		// The engine only ever sees cards a deck could legally
		// contain, which is also what keeps Mystery Booster playtest
		// mana symbols out of the sweep.
		if c.Legalities["commander"] != "legal" {
			continue
		}
		visit(c)
	}
}
