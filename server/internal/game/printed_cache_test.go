package game

import (
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// printed_cache_test.go pins #1498's printed-characteristic cache.
//
// The cache never has to be invalidated, because every read checks
// the inputs it was built from against the card's live ones. So the
// property to prove is "the cached answer and a fresh build agree
// after every way a printed field can change" — whether or not anyone
// restamped the card. TestPrintedCacheAgreesAcrossEveryPrintedMutation
// walks those mutations one at a time; printedCacheChecks makes every
// other test in the suite do the same check on every cache hit.

// assertPrintedAgrees fails when any cached read of c disagrees with a
// fresh, uncached build of the same card.
func assertPrintedAgrees(t *testing.T, step string, c *Card) {
	t.Helper()
	want := printedFresh(c)
	if got := printedShared(c); !samePrintedCharacteristic(got, want) {
		t.Errorf("%s: printedShared = %+v\nwant (fresh build) %+v", step, got, want)
	}
	uncached := *c
	uncached.printed = nil
	if got, want := effectiveOf(c), effectiveOf(&uncached); !samePrintedCharacteristic(got, want) {
		t.Errorf("%s: Effective() = %+v\nwant (uncached) %+v", step, got, want)
	}
	gs, gt, gu := printedTypeParts(c)
	ws, wt, wu := parseTypeLine(c.TypeLine)
	if !sameStrings(gs, ws) || !sameStrings(gt, wt) || !sameStrings(gu, wu) {
		t.Errorf("%s: printedTypeParts = %v %v %v, want %v %v %v", step, gs, gt, gu, ws, wt, wu)
	}
	if got, want := c.PrintedTypeLineIs(ws, wt, wu), true; got != want {
		t.Errorf("%s: PrintedTypeLineIs(its own parse) = %v", step, got)
	}
	if got, want := c.HasSubtype("Elf"), uncached.HasSubtype("Elf"); got != want {
		t.Errorf("%s: HasSubtype(Elf) = %v, uncached %v", step, got, want)
	}
	if got, want := c.HasSupertype("Legendary"), uncached.HasSupertype("Legendary"); got != want {
		t.Errorf("%s: HasSupertype(Legendary) = %v, uncached %v", step, got, want)
	}
}

// stubPrintedKeywords installs a catalog printed-keyword hook that
// answers from the map, restoring the real one on cleanup.
func stubPrintedKeywords(t *testing.T, byKey map[string][]string) {
	t.Helper()
	prev := CatalogPrintedKeywords
	CatalogPrintedKeywords = func(key string) []string { return byKey[key] }
	t.Cleanup(func() { CatalogPrintedKeywords = prev })
}

func cacheProbeCard() Card {
	return Card{
		InstanceID: uuid.New(),
		Name:       "Probe Front",
		OracleID:   "printed-cache-probe",
		TypeLine:   "Creature — Human Werewolf",
		ManaCost:   "{1}{G}",
		Colors:     []string{"G"},
		Keywords:   []string{"trample"},
		Power:      2,
		Toughness:  2,
		Owner:      uuid.New(),
		Layout:     LayoutTransform,
		Faces: []Face{
			{Name: "Probe Front", TypeLine: "Creature — Human Werewolf", ManaCost: "{1}{G}", Colors: []string{"G"}, Power: 2, Toughness: 2, Keywords: []string{"trample"}},
			{Name: "Probe Back", TypeLine: "Legendary Creature — Elf Werewolf", Colors: []string{"G", "R"}, Power: 4, Toughness: 4, Keywords: []string{"haste", "menace"}},
		},
	}
}

// TestPrintedCacheAgreesAcrossEveryPrintedMutation walks every way a
// printed input can change on a stamped card. Each step mutates the
// card in place WITHOUT restamping — the cache must already be right —
// and then restamps and checks again, so both the stale-entry path and
// the rebuilt-entry path are covered.
func TestPrintedCacheAgreesAcrossEveryPrintedMutation(t *testing.T) {
	stubPrintedKeywords(t, map[string][]string{
		"printed-cache-probe":   {"vigilance"},
		"printed-cache-probe#1": {"flying"},
		"printed-cache-other":   {"deathtouch", "changeling"},
	})
	c := cacheProbeCard()
	c.Controller = c.Owner
	c.stampPrinted()
	if c.printed == nil {
		t.Fatal("stampPrinted left no entry on a face-up card")
	}
	assertPrintedAgrees(t, "freshly stamped", &c)

	other := Card{
		Name: "Copied Thing", OracleID: "printed-cache-other", TypeLine: "Legendary Artifact Creature — Elf Construct",
		ManaCost: "{3}", Power: 3, Toughness: 1, Keywords: []string{"reach"},
	}
	split := Card{
		Name: "Fire", OracleID: "printed-cache-split", Layout: LayoutSplit,
		Faces: []Face{
			{Name: "Fire", TypeLine: "Instant", ManaCost: "{1}{R}"},
			{Name: "Ice", TypeLine: "Instant", ManaCost: "{1}{U}"},
		},
	}
	steps := []struct {
		name   string
		mutate func(c *Card)
	}{
		{"type line rewritten", func(c *Card) { c.TypeLine = "Legendary Creature — Elf Warrior" }},
		{"type line back to an equal string", func(c *Card) { c.TypeLine = strings.Clone("Creature — Human Werewolf") }},
		{"name", func(c *Card) { c.Name = "Renamed" }},
		{"mana cost, colours derived from it", func(c *Card) { c.Colors = nil; c.ManaCost = "{2}{U}{B}" }},
		{"power", func(c *Card) { c.Power = 7 }},
		{"toughness", func(c *Card) { c.Toughness = -1 }},
		{"colours replaced", func(c *Card) { c.Colors = []string{"R"} }},
		{"colour written in place", func(c *Card) { c.Colors[0] = "W" }},
		{"keyword appended", func(c *Card) { c.Keywords = append(c.Keywords, KeywordChangeling) }},
		{"keyword written in place", func(c *Card) { c.Keywords[0] = "flying" }},
		{"keywords cleared", func(c *Card) { c.Keywords = nil }},
		{"cumulative keyword doubled", func(c *Card) { c.Keywords = []string{"prowess", "prowess"} }},
		{"transformed (SetFace 1)", func(c *Card) { c.SetFace(1) }},
		{"transformed back (SetFace 0)", func(c *Card) { c.SetFace(0) }},
		{"oracle ID (a different catalog key)", func(c *Card) { c.OracleID = "printed-cache-other" }},
		{"token key with no oracle ID", func(c *Card) { c.OracleID = ""; c.TokenKey = TokenKeyPrefix + "probe" }},
		{"oracle ID back", func(c *Card) { c.TokenKey = ""; c.OracleID = "printed-cache-probe" }},
		{"copy-granted ability (composite key)", func(c *Card) { c.GrantedAbilities = []string{"probe-grant"} }},
		{"copy grant gone", func(c *Card) { c.GrantedAbilities = nil }},
		{"turned face down (morph)", func(c *Card) { c.SetFaceDown(FaceDownMorphed) }},
		{"turned face up", func(c *Card) { c.SetFaceDown(FaceDownNone) }},
		{"foretold (face down in exile keeps its text)", func(c *Card) { c.SetFaceDown(FaceDownForetold) }},
		{"foretell ends", func(c *Card) { c.SetFaceDown(FaceDownNone) }},
		{"copy effect (ADR 0043)", func(c *Card) { c.setPrintedValues(printedValuesOf(other)) }},
		{"copy reverted", func(c *Card) {
			own := cacheProbeCard()
			c.setPrintedValues(printedValuesOf(own))
		}},
		{"becomes a split card shown whole", func(c *Card) {
			c.Layout, c.Faces, c.OracleID, c.Keywords, c.Colors = split.Layout, split.Faces, split.OracleID, nil, nil
			c.materialiseSplitWhole()
		}},
		{"split card fused", func(c *Card) { c.materialiseFused() }},
		{"split card's left half", func(c *Card) { c.Fused = false; c.materialiseHalves(true, false); c.TypeLine = c.Faces[0].TypeLine }},
		{"control changes (not a cached input)", func(c *Card) { c.Controller = uuid.New() }},
		{"base controller captured", func(c *Card) { c.BaseController = uuid.New() }},
		{"keyword counter (added on top of the baseline)", func(c *Card) { c.Counters = map[string]int{"flying": 1} }},
	}
	for _, s := range steps {
		before := c.printed
		s.mutate(&c)
		assertPrintedAgrees(t, s.name+" (unstamped)", &c)
		c.stampPrinted()
		assertPrintedAgrees(t, s.name+" (restamped)", &c)
		if c.FaceDownIsPermanent() {
			if c.printed != nil {
				t.Errorf("%s: a face-down permanent kept a cache entry", s.name)
			}
			continue
		}
		if c.printed == nil {
			t.Fatalf("%s: restamping left no entry", s.name)
		}
		// An entry is replaced only when an input changed; a stamp over
		// a matching one keeps it, which is what makes a move that
		// changes nothing cost one comparison.
		if before != nil && c.printed != before {
			in := printedInputsOf(&c)
			if before.in.sameAs(&in) {
				t.Errorf("%s: restamp replaced an entry whose inputs still matched", s.name)
			}
		}
	}

	// A catalog answer that changes under a stamped card — a test hook
	// swapped, a spec registered later — is an input too.
	c = cacheProbeCard()
	c.stampPrinted()
	stubPrintedKeywords(t, map[string][]string{"printed-cache-probe": {"lifelink", "changeling"}})
	assertPrintedAgrees(t, "catalog printed keywords swapped under a stamped card", &c)
	if !printedShared(&c).AllCreatureTypes {
		t.Error("a catalog-declared changeling did not reach the cached read")
	}
}

// TestPrintedCacheHitSharesNoWritableCapacity: a hit hands out the
// entry's own slices, clipped to length, so an append — Effective's
// keyword-counter fold, or any caller's — reallocates instead of
// writing into the entry.
func TestPrintedCacheHitSharesNoWritableCapacity(t *testing.T) {
	stubPrintedKeywords(t, nil)
	c := cacheProbeCard()
	c.Keywords = []string{"trample", "reach"}
	c.stampPrinted()
	first := printedShared(&c)
	for name, s := range map[string][]string{
		"Types": first.Types, "Subtypes": first.Subtypes, "Colors": first.Colors, "Abilities": first.Abilities,
	} {
		if cap(s) != len(s) {
			t.Errorf("cached %s has spare capacity (len %d, cap %d)", name, len(s), cap(s))
		}
	}
	_ = append(first.Abilities, "vigilance")
	_ = append(first.Subtypes, "Zombie")
	c.Counters = map[string]int{"flying": 1}
	if eff := c.Effective(); !containsKeyword(eff.Abilities, "flying") {
		t.Fatalf("keyword counter missing from Effective: %v", eff.Abilities)
	}
	c.Counters = nil
	again := printedShared(&c)
	if !sameStrings(again.Abilities, []string{"trample", "reach"}) || !sameStrings(again.Subtypes, []string{"Human", "Werewolf"}) {
		t.Errorf("an append through a cached read reached the cache: abilities %v, subtypes %v", again.Abilities, again.Subtypes)
	}
}

// TestPrintedCacheIsSafeToShareWithAClone: cloneCard copies the entry
// pointer. The entry is immutable and each card validates against its
// own fields, so changing the clone cannot change what the source
// reads, and the other way round.
func TestPrintedCacheIsSafeToShareWithAClone(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	p.Hand.PushTop(cacheProbeCard())
	id := p.Hand.Cards[len(p.Hand.Cards)-1].InstanceID
	src := &p.Hand.Cards[len(p.Hand.Cards)-1]
	entry := src.printed
	if entry == nil {
		t.Fatal("PushTop did not stamp the card")
	}

	clone := g.Clone()
	var cc *Card
	for i := range clone.Seats[0].Hand.Cards {
		if clone.Seats[0].Hand.Cards[i].InstanceID == id {
			cc = &clone.Seats[0].Hand.Cards[i]
		}
	}
	if cc == nil || cc.printed != entry {
		t.Fatal("the clone does not share the source's entry; the test is not testing sharing")
	}
	cc.SetFace(1)
	assertPrintedAgrees(t, "clone transformed", cc)
	assertPrintedAgrees(t, "source after the clone transformed", src)
	if got := src.Effective().Name; got != "Probe Front" {
		t.Errorf("source reads %q after its clone transformed; want Probe Front", got)
	}
	cc.stampPrinted()
	if src.printed != entry {
		t.Error("restamping the clone replaced the SOURCE's entry")
	}
	src.TypeLine = "Artifact"
	assertPrintedAgrees(t, "source rewritten after the clone restamped", src)
	assertPrintedAgrees(t, "clone after the source was rewritten", cc)
}

// TestZoneInsertionsAndRestoreStampThePrintedCache: the stamp sites.
func TestZoneInsertionsAndRestoreStampThePrintedCache(t *testing.T) {
	z := newZone(ZoneGraveyard, uuid.New())
	for name, insert := range map[string]func(Card){
		"PushTop":            z.PushTop,
		"PushBottom":         z.PushBottom,
		"InsertFromTop(top)": func(c Card) { z.InsertFromTop(c, 1) },
		"InsertFromTop(2)":   func(c Card) { z.InsertFromTop(c, 2) },
		"InsertFromTop(99)":  func(c Card) { z.InsertFromTop(c, 99) },
	} {
		c := cacheProbeCard()
		insert(c)
		for i := range z.Cards {
			if z.Cards[i].InstanceID != c.InstanceID {
				continue
			}
			if z.Cards[i].printed == nil {
				t.Errorf("%s did not stamp the inserted card", name)
			}
			assertPrintedAgrees(t, name, &z.Cards[i])
		}
	}

	g := newActiveGame(t)
	for _, p := range g.Seats {
		p.Graveyard.PushTop(cacheProbeCard())
	}
	snap := g.CaptureSnapshot()
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	n := 0
	for _, p := range restored.Seats {
		for _, z := range []*Zone{p.Library, p.Hand, p.Graveyard} {
			for i := range z.Cards {
				n++
				if z.Cards[i].printed == nil {
					t.Fatalf("restored %s card %q has no cache entry", z.Kind, z.Cards[i].Name)
				}
				assertPrintedAgrees(t, "restored "+string(z.Kind), &z.Cards[i])
			}
		}
	}
	if n == 0 {
		t.Fatal("the restored game has no cards to check")
	}
}

// TestPrintedCacheCheckCatchesAStaleEntry is the back-out proof for the
// test-binary check itself: an entry whose answer no longer matches its
// recorded inputs — the two ways that can happen are a builder input
// the key does not record, and a reader writing an element in place —
// panics on the next read rather than being served.
func TestPrintedCacheCheckCatchesAStaleEntry(t *testing.T) {
	prev := printedCacheChecks
	printedCacheChecks = true
	t.Cleanup(func() { printedCacheChecks = prev })
	stubPrintedKeywords(t, nil)

	c := cacheProbeCard()
	c.stampPrinted()
	c.printed.ch.Subtypes[0] = "Goblin" // a reader writing an element in place
	for name, read := range map[string]func(){
		"printedShared":    func() { _ = printedShared(&c) },
		"printedTypeParts": func() { printedTypeParts(&c) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s served a corrupted entry without panicking", name)
				}
			}()
			read()
		}()
	}
}

// TestPrintedCharacteristicSetsOnlyTheComparedFields keeps
// samePrintedCharacteristic — the check's equality — complete: the
// builder, given every input it reads, sets no field the comparison
// leaves out. A new field in the builder fails here until the
// comparison learns it.
func TestPrintedCharacteristicSetsOnlyTheComparedFields(t *testing.T) {
	in := printedInputs{
		typeLine: "Legendary Snow Artifact Creature — Elf Warrior", name: "All Inputs", manaCost: "{1}{G}{U}",
		power: 3, toughness: 4, colors: []string{"G", "U"}, keywords: []string{KeywordChangeling, "reach"},
		catalogKeywords: []string{"flying"},
	}
	ch := in.characteristic()
	ch.Controller = uuid.New()
	compared := map[string]bool{
		"Power": true, "Toughness": true, "Name": true, "AllCreatureTypes": true, "Controller": true,
		"Types": true, "Subtypes": true, "Supertypes": true, "Colors": true, "Abilities": true,
	}
	v := reflect.ValueOf(ch)
	for i := 0; i < v.NumField(); i++ {
		f := v.Type().Field(i)
		if compared[f.Name] {
			if v.Field(i).IsZero() {
				t.Errorf("compared field %s is zero for an all-inputs card; the probe no longer exercises it", f.Name)
			}
			continue
		}
		if !v.Field(i).IsZero() {
			t.Errorf("the printed builder sets Characteristic.%s, which samePrintedCharacteristic does not compare", f.Name)
		}
	}
}

// TestLandTypeManaStringsMatchTheColor pins the spelled-out strings on
// landTypeMana to the colour they were concatenated from before #1498.
func TestLandTypeManaStringsMatchTheColor(t *testing.T) {
	for _, lt := range landTypeMana {
		if lt.Produced != "{"+lt.Color+"}" || lt.Label != "Add {"+lt.Color+"}" {
			t.Errorf("%s: Produced %q, Label %q; want {%s} and Add {%s}", lt.Subtype, lt.Produced, lt.Label, lt.Color, lt.Color)
		}
	}
}
