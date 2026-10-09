package decks

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// realdump_purpose_manual_test.go — ADR 0126 §6's manual audit of
// declared purposes, against the real Scryfall dump. Gated on
// CMDCTRL_SCRYFALL_DUMP like realdump_manual_test.go, for the same
// reason: CI has no dump.
//
//	CMDCTRL_SCRYFALL_DUMP=data/scryfall/default-cards.json \
//	  go test ./internal/decks/ -run RealDumpPurpose -v
//
// Three checks:
//
//   - TestRealDumpPurposeCuratedSpellList: curatedInstantsAndSorceries
//     (purpose_test.go) is exactly the curated decks' instants and
//     sorceries by Scryfall's type lines, so TestCuratedDeckPurposes
//     covers every one of them offline.
//   - TestRealDumpPurposeLandsUntapped: curatedLandsUntapped
//     (purpose_test.go) is exactly the curated cards that declare Lands
//     and whose search puts a card onto the battlefield without
//     "tapped" (ADR 0136 §2). It lists, without failing, the other
//     catalog cards that declare Lands, read the same way, and declare
//     no LandsUntapped.
//   - TestRealDumpPurposeAudit: every catalog card whose oracle text
//     reads as a board wipe declares a Sweep somewhere (owner decision
//     2: every catalog wipe), or is on reviewedNotAWipe with the reason.
//     It also LISTS, without failing, the catalog cards whose text reads
//     as a draw, a tutor or a land search and that declare no purpose:
//     a review aid, since ADR 0126 declares those for the curated decks
//     alone. It lists the same way the cards whose text gives a target
//     player something ("target player draws / creates / gains") or
//     deals a fixed amount of damage to a creature or any target, and
//     that declare no target entry: ADR 0126's amendment of 2026-10-08,
//     whose PR 5 works through that list.

// The oracle-text readings. Deliberately loose: a false positive is a
// line on reviewedNotAWipe, a false negative is a wipe nobody declared.
var (
	sweepText = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(destroy|exile) (all|each) `),
		regexp.MustCompile(`(?i)damage to each (other )?(\w+ )?creature`),
		regexp.MustCompile(`(?i)\b(all|each) (other )?(non-?\w+ )?creatures?\b[^.]*\bgets? -`),
		regexp.MustCompile(`(?i)creatures (your opponents control|you don't control|target (player|opponent) controls) get -`),
		regexp.MustCompile(`(?i)\breturn (all|each) [^.]*\bto (its|their) owners?'s? hands?`),
		regexp.MustCompile(`(?i)\bsacrifices? all\b`),
		regexp.MustCompile(`(?i)change "target" in its text to "each"`),
	}
	drawText     = regexp.MustCompile(`(?i)\b(you )?draws? (a|an|one|two|three|four|five|six|seven) cards?\b`)
	tutorText    = regexp.MustCompile(`(?i)search your library for`)
	landToBfText = regexp.MustCompile(`(?i)search your library for [^.]*\bland[^.]*onto the battlefield`)
	// ADR 0136 §2: a library search that puts a card onto the
	// battlefield, and whether "tapped" follows.
	searchToBfText = regexp.MustCompile(`(?i)search your library for [^.]*?onto the battlefield( tapped)?`)
	instantSorcer  = regexp.MustCompile(`\b(Instant|Sorcery)\b`)
	// ADR 0126's amendment of 2026-10-08: a target that is given
	// something, and burn at a creature or any target.
	giftText = regexp.MustCompile(`(?i)\btarget (player|opponent) (draws|creates|gains)\b`)
	burnText = regexp.MustCompile(`(?i)deals \d+ damage to (any (other )?target|(up to \w+ )?(another )?target ([\w,-]+ )*?(creature|planeswalker))`)
)

// declaresTargetEntry reports whether a spec declares a target entry
// (Purpose.Targets) in any slot.
func declaresTargetEntry(s effects.Spec) bool {
	ps := []game.Purpose{s.Purpose}
	modes := func(m *game.ModeSpec) {
		if m != nil {
			for _, o := range m.Options {
				ps = append(ps, o.Purpose)
			}
		}
	}
	modes(s.Modes)
	for _, a := range s.AlternativeCosts {
		ps = append(ps, a.Purpose)
	}
	for _, a := range s.Activated {
		ps = append(ps, a.Purpose)
		modes(a.Modes)
	}
	for _, t := range s.Triggered {
		ps = append(ps, t.Purpose)
		modes(t.Modes)
	}
	for _, p := range ps {
		if p.Targets != nil {
			return true
		}
	}
	return false
}

// reviewedNotAWipe is every catalog card the sweep reading flags that
// declares no Sweep, with why. Each was read.
var reviewedNotAWipe = map[string]string{
	"Armageddon":     "destroys lands, which no sweep class holds alone",
	"Bag of Holding": "returns the cards it exiled, not permanents",
	"Baneblade Scoundrel // Baneclaw Marauder": "-1/-1 only to the creatures blocking it",
	"Bazaar of Wonders":                        "exiles graveyards",
	"Beyeen Veil // Beyeen Coast":              "-2/-0 kills nothing",
	"Call Forth the Tempest":                   "damage equal to spells cast this turn: no fixed amount or X for a sweep to declare",
	"Crypt Incursion":                          "exiles a graveyard",
	"Eliminate the Impossible":                 "-2/-0 kills nothing",
	"Eye of Singularity":                       "destroys only permanents sharing a name, which no class says",
	"Glorious End":                             "ends the turn",
	"Jace, the Mind Sculptor":                  "exiles a library",
	"Karn Liberated":                           "restarts the game",
	"Last Laugh":                               "a ping engine that fires only when a permanent dies",
	"Living Death":                             "a symmetric mass reanimation (purpose_test.go's noPrintedAmount)",
	"Mandate of Peace":                         "ends combat and exiles the stack",
	"Obeka, Brute Chronologist":                "ends the turn",
	"Ondu Inversion // Ondu Skyruins":          "only the land face is catalogued; the sweep is played by hand",
	"Rest in Peace":                            "exiles graveyards",
	"Scavenger Grounds":                        "exiles graveyards",
	"Soul-Guide Lantern":                       "exiles graveyards",
	"Sundial of the Infinite":                  "ends the turn",
	"Time Stop":                                "ends the turn",
	"Ugin's Binding":                           "a single target, and an exile from the graveyard",
	"Ugin, the Spirit Dragon":                  "the -X sweep is not registered (its caveat says so)",
	"Watchdog":                                 "-1/-0 to attackers kills nothing",
}

// anyPurpose reports whether a spec declares a purpose in any slot.
func anyPurpose(s effects.Spec) (declared, sweep bool) {
	note := func(zero bool, hasSweep bool) {
		if !zero {
			declared = true
		}
		if hasSweep {
			sweep = true
		}
	}
	note(s.Purpose.IsZero(), !s.Purpose.Sweep.IsZero())
	if s.Modes != nil {
		for _, o := range s.Modes.Options {
			note(o.Purpose.IsZero(), !o.Purpose.Sweep.IsZero())
		}
	}
	for _, a := range s.AlternativeCosts {
		note(a.Purpose.IsZero(), !a.Purpose.Sweep.IsZero())
	}
	for _, a := range s.Activated {
		note(a.Purpose.IsZero(), !a.Purpose.Sweep.IsZero())
		if a.Modes != nil {
			for _, o := range a.Modes.Options {
				note(o.Purpose.IsZero(), !o.Purpose.Sweep.IsZero())
			}
		}
	}
	for _, t := range s.Triggered {
		note(t.Purpose.IsZero(), !t.Purpose.Sweep.IsZero())
		if t.Modes != nil {
			for _, o := range t.Modes.Options {
				note(o.Purpose.IsZero(), !o.Purpose.Sweep.IsZero())
			}
		}
	}
	return declared, sweep
}

// loadDumpForPurpose loads the dump or skips.
func loadDumpForPurpose(t *testing.T) *cards.Index {
	t.Helper()
	path := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if path == "" {
		t.Skip("set CMDCTRL_SCRYFALL_DUMP to run against the real dump")
	}
	idx := cards.NewIndex()
	if _, err := idx.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	return idx
}

// printed returns a card's type line and oracle text, every face
// joined.
func printed(c cards.Card) (typeLine, text string) {
	types, texts := []string{c.TypeLine}, []string{c.OracleText}
	for _, f := range c.CardFaces {
		types = append(types, f.TypeLine)
		texts = append(texts, f.OracleText)
	}
	return strings.Join(types, " // "), strings.Join(texts, "\n")
}

func TestRealDumpPurposeCuratedSpellList(t *testing.T) {
	idx := loadDumpForPurpose(t)
	want := map[string]bool{}
	for _, d := range All() {
		for _, c := range d.Cards() {
			if c.Basic {
				continue
			}
			id, err := uuid.Parse(c.OracleID)
			if err != nil {
				t.Fatalf("%s: %v", c.Name, err)
			}
			card, ok := idx.FindByOracleID(id)
			if !ok {
				t.Errorf("%s is not in the dump", c.Name)
				continue
			}
			if tl, _ := printed(card); instantSorcer.MatchString(tl) && !strings.Contains(tl, "Land") {
				want[c.Name] = true
			}
		}
	}
	got := map[string]bool{}
	for _, n := range curatedInstantsAndSorceries {
		got[n] = true
	}
	var diff []string
	for n := range want {
		if !got[n] {
			diff = append(diff, "missing from curatedInstantsAndSorceries: "+n)
		}
	}
	for n := range got {
		if !want[n] {
			diff = append(diff, "not an instant or sorcery in any curated deck: "+n)
		}
	}
	if len(diff) > 0 {
		sort.Strings(diff)
		t.Errorf("curatedInstantsAndSorceries is wrong:\n\t%s", strings.Join(diff, "\n\t"))
	}
}

func TestRealDumpPurposeAudit(t *testing.T) {
	idx := loadDumpForPurpose(t)

	// Every Spec under one base oracle ID is one card: its faces, a
	// split card's halves.
	type entry struct {
		name                      string
		declared, sweeps, targets bool
	}
	byBase := map[string]*entry{}
	for _, s := range effects.All() {
		base, _, _ := strings.Cut(s.OracleID, "#")
		e := byBase[base]
		if e == nil {
			e = &entry{name: s.Name}
			byBase[base] = e
		}
		d, sw := anyPurpose(s)
		e.declared = e.declared || d
		e.sweeps = e.sweeps || sw
		e.targets = e.targets || declaresTargetEntry(s)
	}

	var undeclaredWipes, staleReviews []string
	listed := map[string][]string{}
	for base, e := range byBase {
		id, err := uuid.Parse(base)
		if err != nil {
			continue // a token key, which has no printing
		}
		card, ok := idx.FindByOracleID(id)
		if !ok {
			continue
		}
		typeLine, text := printed(card)
		reads := func(re *regexp.Regexp) bool { return re.MatchString(text) }
		wipe := false
		for _, re := range sweepText {
			if reads(re) {
				wipe = true
				break
			}
		}
		_, reviewed := reviewedNotAWipe[card.Name]
		switch {
		case wipe && !e.sweeps && !reviewed:
			undeclaredWipes = append(undeclaredWipes, card.Name+" ("+base+")")
		case reviewed && (e.sweeps || !wipe):
			staleReviews = append(staleReviews, card.Name)
		}
		if !e.targets {
			switch {
			case reads(giftText):
				listed["target gift"] = append(listed["target gift"], card.Name)
			case reads(burnText):
				listed["burn at a creature or any target"] = append(listed["burn at a creature or any target"], card.Name)
			}
		}
		if e.declared {
			continue
		}
		switch {
		case reads(landToBfText):
			listed["land search"] = append(listed["land search"], card.Name)
		case reads(tutorText):
			listed["tutor"] = append(listed["tutor"], card.Name)
		case reads(drawText) && instantSorcer.MatchString(typeLine):
			listed["draw (instant or sorcery)"] = append(listed["draw (instant or sorcery)"], card.Name)
		}
	}
	if len(undeclaredWipes) > 0 {
		sort.Strings(undeclaredWipes)
		t.Errorf("%d catalog card(s) read as a board wipe and declare no Sweep (ADR 0126 owner decision 2): "+
			"declare one, or add the card to reviewedNotAWipe with the reason:\n\t%s",
			len(undeclaredWipes), strings.Join(undeclaredWipes, "\n\t"))
	}
	if len(staleReviews) > 0 {
		sort.Strings(staleReviews)
		t.Errorf("reviewedNotAWipe lists cards that now declare a Sweep, or no longer read as one:\n\t%s",
			strings.Join(staleReviews, "\n\t"))
	}
	for _, class := range []struct{ name, lacks string }{
		{"land search", "purpose"},
		{"tutor", "purpose"},
		{"draw (instant or sorcery)", "purpose"},
		{"target gift", "target entry"},
		{"burn at a creature or any target", "target entry"},
	} {
		names := listed[class.name]
		sort.Strings(names)
		t.Logf("review aid: %d catalog card(s) read as a %s and declare no %s:\n\t%s",
			len(names), class.name, class.lacks, strings.Join(names, "\n\t"))
	}
}

// readsLandsUntapped reports whether a card's text puts a searched-out
// card onto the battlefield without "tapped".
func readsLandsUntapped(text string) bool {
	for _, m := range searchToBfText.FindAllStringSubmatch(text, -1) {
		if m[1] == "" {
			return true
		}
	}
	return false
}

// declaredLands sums the Lands a spec declares in every slot.
func declaredLands(s effects.Spec) int {
	n := 0
	for _, p := range specPurposes(s) {
		n += p.Lands
	}
	return n
}

func TestRealDumpPurposeLandsUntapped(t *testing.T) {
	idx := loadDumpForPurpose(t)
	curated := map[string]bool{}
	var diff []string
	for _, d := range All() {
		for _, c := range d.Cards() {
			if c.Basic || curated[c.Name] {
				continue
			}
			curated[c.Name] = true
			spec, ok := effects.Lookup(c.OracleID)
			if !ok || declaredLands(spec) == 0 {
				continue
			}
			card, ok := idx.FindByOracleID(uuid.MustParse(c.OracleID))
			if !ok {
				t.Errorf("%s is not in the dump", c.Name)
				continue
			}
			_, text := printed(card)
			_, listed := curatedLandsUntapped[c.Name]
			switch untapped := readsLandsUntapped(text); {
			case untapped && !listed:
				diff = append(diff, c.Name+" puts its lands onto the battlefield untapped and is not on curatedLandsUntapped")
			case !untapped && listed:
				diff = append(diff, c.Name+" is on curatedLandsUntapped and its lands enter tapped")
			}
		}
	}
	if len(diff) > 0 {
		sort.Strings(diff)
		t.Errorf("curatedLandsUntapped is wrong (ADR 0136 §2):\n\t%s", strings.Join(diff, "\n\t"))
	}

	var aid []string
	for _, s := range effects.All() {
		if curated[s.Name] || declaredLands(s) == 0 || landsUntappedIn(s) > 0 {
			continue
		}
		base, _, _ := strings.Cut(s.OracleID, "#")
		id, err := uuid.Parse(base)
		if err != nil {
			continue
		}
		card, ok := idx.FindByOracleID(id)
		if !ok {
			continue
		}
		if _, text := printed(card); readsLandsUntapped(text) {
			aid = append(aid, s.Name)
		}
	}
	sort.Strings(aid)
	t.Logf("review aid: %d catalog card(s) outside the curated decks declare Lands, put them onto the battlefield untapped, and declare no LandsUntapped:\n\t%s",
		len(aid), strings.Join(aid, "\n\t"))
}
