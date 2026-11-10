package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// celestial_reunion_test.go — #5406: "As an additional cost to cast this
// spell, you may choose a creature type and behold two creatures of that
// type" (CR 601.2b, 701.4a), and the search that reads the chosen type
// back at resolution.

const (
	celestialReunionOracle = "ba0a815c-1dfe-47c1-890f-496c670124f3"
	celestialReunionCost   = "{X}{G}"
	elfCreatureLine        = "Creature — Elf Druid"
)

// reunionBoard is a table with an Elf on my battlefield, an Elf card in
// my hand, and an Elf and a Goblin creature card (each mana value 2) and a
// mana value 5 Elf in my library.
type reunionBoard struct {
	g                  *game.Game
	me, opp            *game.Player
	elfOnTable, elfInH uuid.UUID
	libElf, libGoblin  uuid.UUID
	libBigElf          uuid.UUID
}

func newReunionBoard(t *testing.T) *reunionBoard {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b := &reunionBoard{g: g, me: me, opp: opp}
	b.elfOnTable = apaPush(g, me.ID, me.ID, game.Card{Name: "Elf Scout", TypeLine: "Creature — Elf Scout", Power: 1, Toughness: 1})
	b.elfInH = handCard(me, "Elf Friend", elfCreatureLine)
	b.libElf = reunionLibraryCard(me, "Elf Lord", elfCreatureLine, "{1}{G}")
	b.libGoblin = reunionLibraryCard(me, "Goblin Chief", "Creature — Goblin Warrior", "{1}{R}")
	b.libBigElf = reunionLibraryCard(me, "Elf Titan", elfCreatureLine, "{3}{G}{G}")
	return b
}

func reunionLibraryCard(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	return id
}

func castReunion(t *testing.T, g *game.Game, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	return castWithTapParams(t, g, "Celestial Reunion", "Sorcery", celestialReunionCost, celestialReunionOracle, params)
}

func beholdElves(b *reunionBoard, x int) game.CastSpellParams {
	return game.CastSpellParams{
		XValue:        x,
		OptionalCosts: []int{beholdOfferIdx},
		BeholdType:    "Elf",
		RevealIDs:     []uuid.UUID{b.elfOnTable, b.elfInH},
	}
}

// Beheld as Elves: the type is on the stack item, the Elf card in hand is
// revealed (the permanent is chosen, not revealed), and an Elf found by the
// search goes onto the battlefield instead of into the hand.
func TestCelestialReunionBeheldPutsTheChosenTypeOntoTheBattlefield(t *testing.T) {
	b := newReunionBoard(t)
	id, err := castReunion(t, b.g, beholdElves(b, 2))
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	item := b.g.StackMeta[id]
	if item == nil || item.Paid.BeholdType != "Elf" || item.Paid.OptionalCostTimes(beholdOfferIdx) != 1 {
		t.Fatalf("Paid = %+v, want the behold offer and the type Elf recorded", item)
	}
	if got := revealedEvents(b.g, b.elfInH); got != 1 {
		t.Errorf("the Elf card in hand was revealed %d times, want 1", got)
	}
	if got := revealedEvents(b.g, b.elfOnTable); got != 0 {
		t.Errorf("the Elf permanent was revealed %d times, want 0", got)
	}
	if !b.me.Hand.Contains(b.elfInH) {
		t.Fatal("beholding moved the hand card")
	}
	passPriorityAroundTable(t, b.g)
	answerSearchByID(t, b.g, b.me.ID, b.libElf)
	if !onBattlefield(b.g, b.libElf) {
		t.Fatal("the Elf found by a beheld Elf cast is not on the battlefield")
	}
	if b.me.Hand.Contains(b.libElf) {
		t.Error("the Elf passed through the hand")
	}
	if got := revealedEvents(b.g, b.libElf); got != 1 {
		t.Errorf("the found card was revealed %d times, want 1", got)
	}
}

// Beheld as Elves but the search finds a Goblin: not the chosen type, so
// it goes to the hand.
func TestCelestialReunionBeheldOtherTypeGoesToHand(t *testing.T) {
	b := newReunionBoard(t)
	if _, err := castReunion(t, b.g, beholdElves(b, 2)); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, b.g)
	answerSearchByID(t, b.g, b.me.ID, b.libGoblin)
	if !b.me.Hand.Contains(b.libGoblin) || onBattlefield(b.g, b.libGoblin) {
		t.Fatal("a Goblin found by an Elf behold did not go to the hand")
	}
}

// Without the behold, an Elf goes to the hand like any other card.
func TestCelestialReunionUnbeheldGoesToHand(t *testing.T) {
	b := newReunionBoard(t)
	id, err := castReunion(t, b.g, game.CastSpellParams{XValue: 2})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if item := b.g.StackMeta[id]; item == nil || item.Paid.BeholdType != "" {
		t.Fatalf("Paid = %+v, want no behold type", item)
	}
	passPriorityAroundTable(t, b.g)
	answerSearchByID(t, b.g, b.me.ID, b.libElf)
	if !b.me.Hand.Contains(b.libElf) || onBattlefield(b.g, b.libElf) {
		t.Fatal("an unbeheld cast put the Elf anywhere but the hand")
	}
}

// The search is for a creature card with mana value X or less: the
// mana value 5 Elf is not offered at X = 2.
func TestCelestialReunionSearchIsCappedByX(t *testing.T) {
	b := newReunionBoard(t)
	if _, err := castReunion(t, b.g, beholdElves(b, 2)); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, b.g)
	c := searchChoiceFor(b.g, b.me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(b.g, c, "Elf Titan") != uuid.Nil {
		t.Error("a mana value 5 card was offered at X = 2")
	}
	if searchOptionNamed(b.g, c, "Elf Lord") == uuid.Nil {
		t.Error("the mana value 2 Elf was not offered at X = 2")
	}
}

// The chosen type is read off the stack item at resolution, not off the
// board: both beheld Elves leaving in response change nothing, and a
// restore point taken with the spell on the stack keeps the type.
func TestCelestialReunionChosenTypeIsReadAtResolution(t *testing.T) {
	b := newReunionBoard(t)
	id, err := castReunion(t, b.g, beholdElves(b, 2))
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	b.g.WithWriteLock(func() { _ = b.g.DestroyPermanentForEffect(b.elfOnTable, game.DestroyOptions{}) })
	restored := restoreRoundTrip(t, b.g, false)
	if item := restored.StackMeta[id]; item == nil || item.Paid.BeholdType != "Elf" {
		t.Fatalf("the restored item = %+v, want the type Elf", item)
	}
	passPriorityAroundTable(t, restored)
	answerSearchByID(t, restored, b.me.ID, b.libElf)
	if !onBattlefield(restored, b.libElf) {
		t.Fatal("the restored beheld cast did not put the Elf onto the battlefield")
	}
}

// A behold of a different type reads that type: Goblins beheld, the
// Goblin found goes onto the battlefield and an Elf would not.
func TestCelestialReunionReadsTheTypeThatWasChosen(t *testing.T) {
	b := newReunionBoard(t)
	g1 := apaPush(b.g, b.me.ID, b.me.ID, game.Card{Name: "Goblin A", TypeLine: "Creature — Goblin", Power: 1, Toughness: 1})
	g2 := handCard(b.me, "Goblin B", "Creature — Goblin Shaman")
	if _, err := castReunion(t, b.g, game.CastSpellParams{
		XValue: 2, OptionalCosts: []int{beholdOfferIdx}, BeholdType: "goblin", RevealIDs: []uuid.UUID{g1, g2},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, b.g)
	answerSearchByID(t, b.g, b.me.ID, b.libGoblin)
	if !onBattlefield(b.g, b.libGoblin) {
		t.Fatal("the Goblin found by a Goblin behold is not on the battlefield")
	}
}

// Refusals, each leaving the stack empty: one creature of the type, the
// same creature twice, a creature of another type, a Kindred card of the
// type, an opponent's creature, three creatures, no type, a word that is
// not a creature type, a type with no offer, and the offer with no cards.
func TestCelestialReunionBeholdRefusals(t *testing.T) {
	b := newReunionBoard(t)
	goblin := handCard(b.me, "Goblin Pal", "Creature — Goblin")
	kindred := handCard(b.me, "Elf Rite", "Kindred Instant — Elf")
	theirs := apaPush(b.g, b.opp.ID, b.opp.ID, game.Card{Name: "Their Elf", TypeLine: elfCreatureLine, Power: 1, Toughness: 1})
	third := handCard(b.me, "Elf Third", elfCreatureLine)
	offer := []int{beholdOfferIdx}
	for name, params := range map[string]game.CastSpellParams{
		"one Elf":              {XValue: 2, OptionalCosts: offer, BeholdType: "Elf", RevealIDs: []uuid.UUID{b.elfOnTable}},
		"the same Elf twice":   {XValue: 2, OptionalCosts: offer, BeholdType: "Elf", RevealIDs: []uuid.UUID{b.elfOnTable, b.elfOnTable}},
		"an Elf and a Goblin":  {XValue: 2, OptionalCosts: offer, BeholdType: "Elf", RevealIDs: []uuid.UUID{b.elfOnTable, goblin}},
		"a Kindred Elf card":   {XValue: 2, OptionalCosts: offer, BeholdType: "Elf", RevealIDs: []uuid.UUID{b.elfOnTable, kindred}},
		"an opponent's Elf":    {XValue: 2, OptionalCosts: offer, BeholdType: "Elf", RevealIDs: []uuid.UUID{b.elfOnTable, theirs}},
		"three Elves":          {XValue: 2, OptionalCosts: offer, BeholdType: "Elf", RevealIDs: []uuid.UUID{b.elfOnTable, b.elfInH, third}},
		"no type":              {XValue: 2, OptionalCosts: offer, RevealIDs: []uuid.UUID{b.elfOnTable, b.elfInH}},
		"not a creature type":  {XValue: 2, OptionalCosts: offer, BeholdType: "Forest", RevealIDs: []uuid.UUID{b.elfOnTable, b.elfInH}},
		"a type with no offer": {XValue: 2, BeholdType: "Elf"},
		"cards with no offer":  {XValue: 2, RevealIDs: []uuid.UUID{b.elfOnTable, b.elfInH}},
		"the offer, no cards":  {XValue: 2, OptionalCosts: offer, BeholdType: "Elf"},
		"the wrong type named": {XValue: 2, OptionalCosts: offer, BeholdType: "Goblin", RevealIDs: []uuid.UUID{b.elfOnTable, b.elfInH}},
	} {
		if _, err := castReunion(t, b.g, params); err == nil {
			t.Errorf("%s: the cast was accepted", name)
		}
		if n := stackLen(b.g); n != 0 {
			t.Fatalf("%s: a refused cast left %d object(s) on the stack", name, n)
		}
	}
}

// With only one creature of a type, the type is not offered; with two it
// is, with exactly those two, permanents and hand cards alike. Kindred
// cards and opponents' creatures are never options.
func TestCelestialReunionTypeOptions(t *testing.T) {
	b := newReunionBoard(t)
	handCard(b.me, "Lone Goblin", "Creature — Goblin")
	handCard(b.me, "Elf Rite", "Kindred Instant — Elf")
	apaPush(b.g, b.opp.ID, b.opp.ID, game.Card{Name: "Their Elf", TypeLine: elfCreatureLine, Power: 1, Toughness: 1})
	spell := handCard(b.me, "Celestial Reunion", "Sorcery")
	rc := OptionalBeholdChosenType(2).Reveal
	var opts []game.BeholdTypeOption
	b.g.ReadSnapshot(func() { opts = b.g.BeholdTypeOptionsForEffect(b.me.ID, spell, rc) })
	var types []string
	for _, o := range opts {
		types = append(types, o.Type)
	}
	// Elf Scout and Elf Friend are Elves; Elf Friend is also a Druid and
	// Elf Scout a Scout, but there is one of each.
	if len(opts) != 1 || opts[0].Type != "Elf" || len(opts[0].Cards) != 2 {
		t.Fatalf("types = %v (%+v), want only Elf with its two creatures", types, opts)
	}
	want := map[uuid.UUID]bool{b.elfOnTable: true, b.elfInH: true}
	for _, id := range opts[0].Cards {
		if !want[id] {
			t.Errorf("option %s is not one of my two Elves", id)
		}
	}
}

// --- Register's guards --------------------------------------------------

func TestRegisterRefusesMalformedChosenTypeBehold(t *testing.T) {
	good := OptionalBeholdChosenType(2)
	for name, mutate := range map[string]func(*game.AdditionalCost){
		"a printed type too": func(c *game.AdditionalCost) { r := *c.Reveal; r.Subtype = "Elf"; c.Reveal = &r },
		"a count of one":     func(c *game.AdditionalCost) { r := *c.Reveal; r.Count = 1; c.Reveal = &r },
		"any object":         func(c *game.AdditionalCost) { r := *c.Reveal; r.Creature = false; c.Reveal = &r },
		"a flash opening":    func(c *game.AdditionalCost) { c.AsThoughFlash = true },
		"exiles":             func(c *game.AdditionalCost) { r := *c.Reveal; r.Exile = true; c.Reveal = &r },
		"mana beside":        func(c *game.AdditionalCost) { c.ManaCost = "{1}" },
	} {
		bad := good
		mutate(&bad)
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: no panic", name)
				}
			}()
			checkOptionalBehold(Spec{Name: "Bad", OptionalCosts: []game.AdditionalCost{bad}})
		}()
	}
	// A fixed-type behold may not carry the chosen-type shape's fields.
	fixed := OptionalBehold("a", "Elf")
	fixedTwo := fixed
	r := *fixed.Reveal
	r.Count = 2
	fixedTwo.Reveal = &r
	mustPanic(t, "malformed behold", func() {
		checkOptionalBehold(Spec{Name: "Fixed Two", OptionalCosts: []game.AdditionalCost{fixedTwo}})
	})
	checkOptionalBehold(Spec{Name: "Celestial Reunion", OptionalCosts: []game.AdditionalCost{good}})
	checkEitherCost(Spec{Name: "Celestial Reunion", OptionalCosts: []game.AdditionalCost{good}})
}

// The "onto the battlefield instead" rides the open search prompt: a
// restore point taken while the searcher is choosing keeps it.
func TestCelestialReunionSearchPromptSurvivesARestore(t *testing.T) {
	b := newReunionBoard(t)
	if _, err := castReunion(t, b.g, beholdElves(b, 2)); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, b.g)
	if searchChoiceFor(b.g, b.me.ID) == nil {
		t.Fatal("no search prompt")
	}
	restored := restoreRoundTrip(t, b.g, false)
	answerSearchByID(t, restored, b.me.ID, b.libElf)
	if !onBattlefield(restored, b.libElf) {
		t.Fatal("after a restore mid-search the Elf did not go onto the battlefield")
	}
}
