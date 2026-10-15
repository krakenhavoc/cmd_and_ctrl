package game

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// text_words_test.go — the engine half of changing one word of an
// object's text (CR 612.2, ADR 0160's 2026-10-15 amendment). The cards
// and the readers they reach are tested in cards/effects.

func twChange(kind TextWordKind, from, to string) TextWordChange {
	return TextWordChange{Kind: kind, From: from, To: to}
}

// CR 613.7: a later change rewrites the text the earlier ones left. A
// word an earlier change turned INTO the replaced word changes too, and
// a word that ends up reading as itself drops out.
func TestTextWordsComposeInTimestampOrder(t *testing.T) {
	var w *TextWords
	w = w.with(twChange(TextWordColor, "B", "U"))
	if got := w.Lines(); !slices.Equal(got, []string{"black → blue"}) {
		t.Fatalf("after black → blue: %v", got)
	}
	// Now both the printed black and the printed blue read blue; both
	// become red.
	w = w.with(twChange(TextWordColor, "U", "R"))
	if got := w.Lines(); !slices.Equal(got, []string{"blue → red", "black → red"}) {
		t.Fatalf("after blue → red: %v", got)
	}
	// Red back to black: the printed black reads itself again and drops
	// out; the printed blue (and the printed red) read black.
	w = w.with(twChange(TextWordColor, "R", "B"))
	if got := w.Lines(); !slices.Equal(got, []string{"blue → black", "red → black"}) {
		t.Fatalf("after red → black: %v", got)
	}
	if w.color("B") != "B" || w.color("U") != "B" || w.color("G") != "G" {
		t.Errorf("color reads: B=%s U=%s G=%s", w.color("B"), w.color("U"), w.color("G"))
	}
}

func TestTextWordsTranslateKeywordsAndSubtypes(t *testing.T) {
	var w *TextWords
	w = w.with(twChange(TextWordColor, "B", "G"))
	w = w.with(twChange(TextWordBasicLandType, "Swamp", "Island"))
	w = w.with(twChange(TextWordCreatureType, "Goblin", "Elf"))
	for in, want := range map[string]string{
		"protection from black": "protection from green",
		"Protection from black": "Protection from green",
		"hexproof from black":   "hexproof from green",
		"protection from red":   "protection from red",
		"swampwalk":             "islandwalk",
		"forestwalk":            "forestwalk",
		"nonbasic landwalk":     "nonbasic landwalk",
		"flying":                "flying",
	} {
		if got := w.keyword(in); got != want {
			t.Errorf("keyword(%q) = %q, want %q", in, got, want)
		}
	}
	ch := Characteristic{Subtypes: []string{"Goblin", "Swamp", "Equipment"}, Abilities: []string{"swampwalk"}}
	w.applyToBaseline(&ch)
	if !slices.Equal(ch.Subtypes, []string{"Elf", "Island", "Equipment"}) {
		t.Errorf("subtypes = %v", ch.Subtypes)
	}
	if !slices.Equal(ch.Abilities, []string{"islandwalk"}) {
		t.Errorf("abilities = %v", ch.Abilities)
	}
	// Swamp → Island on an Island Swamp leaves one Island.
	dual := Characteristic{Subtypes: []string{"Island", "Swamp"}}
	w.applyToBaseline(&dual)
	if !slices.Equal(dual.Subtypes, []string{"Island"}) {
		t.Errorf("an Island Swamp under Swamp → Island has subtypes %v, want [Island]", dual.Subtypes)
	}
}

// "It can't change a word to the same word", and a word must be of its
// class (CR 612.2).
func TestTextWordChangeNormalised(t *testing.T) {
	if got, err := twChange(TextWordColor, "black", "blue").Normalised(); err != nil || got.From != "B" || got.To != "U" {
		t.Errorf("colour words normalise to letters: %+v %v", got, err)
	}
	if got, err := twChange(TextWordBasicLandType, "swamp", "island").Normalised(); err != nil || got.From != "Swamp" || got.To != "Island" {
		t.Errorf("land types normalise: %+v %v", got, err)
	}
	for _, bad := range []TextWordChange{
		twChange(TextWordColor, "B", "B"),
		twChange(TextWordColor, "B", "C"),
		twChange(TextWordBasicLandType, "Swamp", "Desert"),
		twChange(TextWordCreatureType, "Goblin", "Goblin"),
		twChange("rarity", "a", "b"),
	} {
		if _, err := bad.Normalised(); err == nil {
			t.Errorf("%+v is refused", bad)
		}
	}
	if textWordsModProblem(Mod{Kind: ModChangeText}) == "" {
		t.Error("a changeText mod with no change is refused")
	}
	if textWordsModProblem(Mod{Kind: ModAddKeywords, Words: []TextWordChange{twChange(TextWordColor, "B", "U")}}) == "" {
		t.Error("a text change on another mod kind is refused")
	}
}

// The targeting walk lends a candidate its source's words, in both of
// its halves: the legal set and the one-target check (announce and the
// CR 608.2b re-check). A clause that reads "nonblack" through
// HasColorWord reads "nonblue" for a source whose black became blue.
func TestTargetingReadsTheSourcesChangedWords(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blue := pushBattlefieldForTest(g, opp.ID, "Blue Bear", "Creature — Bear", "")
	black := pushBattlefieldForTest(g, opp.ID, "Black Bear", "Creature — Bear", "")
	spec := &TargetSpec{
		Mode: "creature", Label: "target nonblack creature", Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsCreature() && !c.HasColorWord("B")
		},
		Min: 1, Max: 1,
	}
	var source Card
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case blue:
				g.Battlefield.Cards[i].Colors = []string{"U"}
			case black:
				g.Battlefield.Cards[i].Colors = []string{"B"}
			}
		}
		g.RecomputeLayersIfStaleLocked()
		source = Card{InstanceID: uuid.New(), Name: "Doom Blade", TypeLine: "Instant", Controller: me.ID, Owner: me.ID}
		source.TextWords = (*TextWords)(nil).with(twChange(TextWordColor, "B", "U"))
	})
	src := SourceObject(me.ID, &source)
	var legal LegalTargets
	var blueOK, blackOK bool
	g.ReadSnapshot(func() {
		legal = g.legalTargetsLocked(src, spec)
		blueOK = g.targetLegalLocked(src, spec, TargetRef{Kind: TargetCard, ID: blue})
		blackOK = g.targetLegalLocked(src, spec, TargetRef{Kind: TargetCard, ID: black})
	})
	if slices.Contains(legal.Cards, blue) || !slices.Contains(legal.Cards, black) {
		t.Errorf("the legal set of a nonblue clause = %v; want the black bear and not the blue one", legal.Cards)
	}
	if blueOK || !blackOK {
		t.Errorf("the one-target check: blue %v, black %v; want false, true", blueOK, blackOK)
	}
}

// The record is pinned: the permanent has the changed words while it
// stays, loses them when it leaves (CR 400.7), and a restore rebuilds
// them from the record.
func TestChangeTextWordIsPinnedAndSurvivesARestore(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	forest := pushBattlefieldForTest(g, me.ID, "Forest", "Basic Land — Forest", "")
	var ok bool
	g.WithWriteLock(func() {
		c, _ := g.battlefieldCardLocked(forest)
		c.EnteredBattlefieldAt = timeNowUnixNano()
		ok = g.ChangeTextWordForEffect(uuid.Nil, forest, twChange(TextWordBasicLandType, "Forest", "Island"), IndefiniteDuration(), "test")
	})
	if !ok {
		t.Fatal("the change was registered")
	}
	read := func(g *Game) (subs []string, lines []string) {
		g.ReadSnapshot(func() {
			c, _ := g.battlefieldCardLocked(forest)
			if c != nil {
				subs, lines = c.Effective().Subtypes, c.TextChanges()
			}
		})
		return subs, lines
	}
	if subs, lines := read(g); !slices.Equal(subs, []string{"Island"}) || !slices.Equal(lines, []string{"Forest → Island"}) {
		t.Fatalf("subtypes %v, text changes %v", subs, lines)
	}

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if subs, _ := read(restored); !slices.Equal(subs, []string{"Island"}) {
		t.Errorf("after a restore the Forest is still an Island: %v", subs)
	}

	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, me.Hand, forest); err != nil {
			t.Fatal(err)
		}
		g.RecomputeLayersIfStaleLocked()
	})
	for _, c := range me.Hand.Cards {
		if c.InstanceID == forest && c.TextWords != nil {
			t.Error("a card that left the battlefield is a new object with its own text (CR 400.7)")
		}
	}
}
