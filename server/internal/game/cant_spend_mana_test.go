package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// cant_spend_mana_test.go — #6227, "You can't spend mana to cast this
// spell" (Hogaak, Arisen Necropolis).
//
// The rules the tests are written against, from the pinned CR
// (2026-09-25):
//
//	106.1. Mana is the primary resource in the game. Players spend mana
//	to pay costs, usually when casting spells and activating abilities.
//
//	601.2h The player pays the total cost. … Partial payments are not
//	allowed. Unpayable costs can't be paid.
//
//	702.51a … "Convoke" means "For each colored mana in this spell's
//	total cost, you may tap an untapped creature of that color you
//	control rather than pay that mana. For each generic mana in this
//	spell's total cost, you may tap an untapped creature you control
//	rather than pay that mana."
//
//	702.66a … "Delve" means "For each generic mana in this spell's
//	total cost, you may exile a card from your graveyard rather than
//	pay that mana."
//
// So the mana half of the total cost has to be covered by the taps and
// the exiles alone, and anything they leave owed is unpayable.

const cantSpendManaOracle = "test-6227-cant-spend-mana"

// cantSpendManaCost is the cost under test: five coloured symbols and
// two generic, the shape of a convoke-and-delve creature. Five
// creatures of the right colours pay the coloured half and two
// graveyard cards pay the generic; six creatures (the sixth paying one
// generic) leave {1} that nothing but mana could pay.
const cantSpendManaCost = "{2}{B}{B}{G}{G}{G}"

// cantSpendManaBoard seeds the spell in hand with convoke, delve and,
// when `flagged`, the clause itself; `graveyard` cards to delve and
// `lands` untapped Swamps that make {B}.
//
// The hooks chain to whatever was installed before them, so the real
// catalog still answers for every other card.
func cantSpendManaBoard(t *testing.T, flagged bool, graveyard, lands int) (*Game, *Player, uuid.UUID, []uuid.UUID) {
	t.Helper()
	prevTap, prevDelve, prevNoMana := CatalogTapPermanentsCost, CatalogDelve, CatalogCantSpendMana
	CatalogTapPermanentsCost = func(key string) *TapPermanentsCost {
		if key == cantSpendManaOracle {
			return convokeForTest()
		}
		if prevTap != nil {
			return prevTap(key)
		}
		return nil
	}
	CatalogDelve = func(key string) bool {
		return key == cantSpendManaOracle || (prevDelve != nil && prevDelve(key))
	}
	CatalogCantSpendMana = func(key string) bool {
		return flagged && key == cantSpendManaOracle
	}
	t.Cleanup(func() {
		CatalogTapPermanentsCost, CatalogDelve, CatalogCantSpendMana = prevTap, prevDelve, prevNoMana
	})
	withCatalogHook(t, func(oracleID string) []ManaAbilityShape {
		if oracleID == "zz-6227-swamp" {
			return []ManaAbilityShape{{TapCost: true, Produced: "{B}", Label: "Add {B}"}}
		}
		return nil
	})

	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	fuel := graveyardFuel(me, graveyard)
	for i := 0; i < lands; i++ {
		pushBattlefieldForTest(g, me.ID, "Swamp", "Land", "zz-6227-swamp")
	}
	c := NewCard("Test Necropolis", me.ID)
	c.TypeLine = "Creature — Avatar"
	c.ManaCost = cantSpendManaCost
	c.OracleID = cantSpendManaOracle
	me.Hand.PushTop(c)
	return g, me, c.InstanceID, fuel
}

// cantSpendManaConvokers taps-in `black` black and `green` green
// creatures, plus `colorless` colourless ones, and returns their IDs in
// that order.
func cantSpendManaConvokers(g *Game, me *Player, black, green, colorless int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < black; i++ {
		ids = append(ids, convokeCreature(g, me, "B"))
	}
	for i := 0; i < green; i++ {
		ids = append(ids, convokeCreature(g, me, "G"))
	}
	for i := 0; i < colorless; i++ {
		ids = append(ids, convokeCreature(g, me, ""))
	}
	return ids
}

// The issue's acceptance test, first half: five creatures of the right
// colours convoke the five coloured symbols (CR 702.51a) and two
// graveyard cards delve the two generic (CR 702.66a), which is the
// whole of the total cost — so the cast is paid in full (CR 601.2h)
// with no mana at all, and the seven untapped lands are still untapped.
func TestCantSpendManaCastsWhenTapsAndDelveCoverTheCost(t *testing.T) {
	g, me, spell, fuel := cantSpendManaBoard(t, true, 2, 7)
	taps := cantSpendManaConvokers(g, me, 2, 3, 0)

	err := g.CastSpell(me.ID, spell, CastSpellParams{
		Strict: true, AutoTap: true, TapIDs: taps, DelveIDs: fuel,
	})
	if err != nil {
		t.Fatalf("five convokers and two delved cards: %v", err)
	}
	if !g.Stack.Contains(spell) {
		t.Fatal("the paid cast is not on the stack")
	}
	for _, id := range fuel {
		if !g.Exile.Contains(id) {
			t.Errorf("delved card %s was not exiled", id)
		}
	}
	if n := cantSpendManaTappedLands(g, me.ID); n != 0 {
		t.Errorf("%d land(s) tapped for a cast that can't spend mana", n)
	}
	if n := cantSpendManaPoolSize(me); n != 0 {
		t.Errorf("pool after the cast holds %d mana (%v), want empty", n, me.ManaPool)
	}
}

// The issue's acceptance test, second half: six creatures pay the five
// coloured symbols and one generic, the graveyard is empty so delve can
// pay nothing, and the last {1} is unpayable (CR 601.2h) however many
// lands are untapped — the auto-tapper offers none of them.
func TestCantSpendManaRefusesTheCastLandsWouldHavePaid(t *testing.T) {
	g, me, spell, _ := cantSpendManaBoard(t, true, 0, 7)
	taps := cantSpendManaConvokers(g, me, 2, 3, 1)

	err := g.CastSpell(me.ID, spell, CastSpellParams{Strict: true, AutoTap: true, TapIDs: taps})
	var ime *InsufficientManaError
	if !errors.As(err, &ime) {
		t.Fatalf("cast = %v, want an InsufficientManaError", err)
	}
	if g.Stack.Contains(spell) {
		t.Error("the refused cast is on the stack")
	}
	if n := cantSpendManaTappedLands(g, me.ID); n != 0 {
		t.Errorf("the refused cast tapped %d land(s)", n)
	}
	if n := cantSpendManaPoolSize(me); n != 0 {
		t.Errorf("the refused cast left %d mana floating (%v)", n, me.ManaPool)
	}
}

// The control: the same board, the same six creatures and the same
// seven lands, without the clause. The auto-tapper taps one Swamp for
// the last {1} and the cast is paid — so it is the clause that refuses
// the cast above, and nothing else about the fixture.
func TestWithoutTheClauseALandPaysTheRemainder(t *testing.T) {
	g, me, spell, _ := cantSpendManaBoard(t, false, 0, 7)
	taps := cantSpendManaConvokers(g, me, 2, 3, 1)

	if err := g.CastSpell(me.ID, spell, CastSpellParams{Strict: true, AutoTap: true, TapIDs: taps}); err != nil {
		t.Fatalf("cast without the clause: %v", err)
	}
	if !g.Stack.Contains(spell) {
		t.Fatal("the cast is not on the stack")
	}
	if n := cantSpendManaTappedLands(g, me.ID); n != 1 {
		t.Errorf("%d land(s) tapped, want the one that paid the last {1}", n)
	}
}

// Mana already FLOATING is refused too, which is the clause itself
// rather than the auto-tapper: five creatures cover the coloured half,
// the graveyard is empty, and two black mana in the pool cannot pay the
// {2} (CR 106.1). The pool is left as it was, because a refused cast
// spends nothing.
func TestCantSpendManaRefusesFloatingMana(t *testing.T) {
	g, me, spell, _ := cantSpendManaBoard(t, true, 0, 0)
	taps := cantSpendManaConvokers(g, me, 2, 3, 0)
	me.ManaPool.AddMana(ManaToken{Color: "B"}, ManaToken{Color: "B"})

	err := g.CastSpell(me.ID, spell, CastSpellParams{Strict: true, TapIDs: taps})
	var ime *InsufficientManaError
	if !errors.As(err, &ime) {
		t.Fatalf("cast = %v, want an InsufficientManaError", err)
	}
	if len(ime.Missing) == 0 {
		t.Error("the refusal names no missing symbol")
	}
	if g.Stack.Contains(spell) {
		t.Error("the refused cast is on the stack")
	}
	if n := cantSpendManaPoolSize(me); n != 2 {
		t.Errorf("pool after the refusal holds %d mana (%v), want the two black untouched", n, me.ManaPool)
	}
}

// A mana ability still works beside the clause: CR 601.2g's window is
// about the mana a payment may USE, not about adding it, so a Swamp
// tapped by hand leaves its {B} floating and the cast that cannot spend
// it still pays with taps and exiles.
func TestCantSpendManaLeavesManaAbilitiesAlone(t *testing.T) {
	g, me, spell, fuel := cantSpendManaBoard(t, true, 2, 1)
	taps := cantSpendManaConvokers(g, me, 2, 3, 0)
	var swamp uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Swamp" {
			swamp = c.InstanceID
		}
	}
	if err := g.ActivateManaAbility(me.ID, swamp, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Swamp: %v", err)
	}
	if n := cantSpendManaPoolSize(me); n != 1 {
		t.Fatalf("pool after tapping the Swamp holds %d mana (%v), want one {B}", n, me.ManaPool)
	}
	if err := g.CastSpell(me.ID, spell, CastSpellParams{Strict: true, TapIDs: taps, DelveIDs: fuel}); err != nil {
		t.Fatalf("cast beside the floating mana: %v", err)
	}
	if !g.Stack.Contains(spell) {
		t.Fatal("the paid cast is not on the stack")
	}
	if n := cantSpendManaPoolSize(me); n != 1 {
		t.Errorf("pool after the cast holds %d mana (%v), want the {B} still floating", n, me.ManaPool)
	}
}

// The clause is read off the catalog key of the face being cast, like
// delve's: no key, no clause.
func TestCantSpendManaForReadsTheCatalogKey(t *testing.T) {
	prev := CatalogCantSpendMana
	CatalogCantSpendMana = func(key string) bool { return key == cantSpendManaOracle }
	t.Cleanup(func() { CatalogCantSpendMana = prev })

	flagged := NewCard("Test Necropolis", uuid.New())
	flagged.OracleID = cantSpendManaOracle
	if !CantSpendManaFor(flagged) {
		t.Error("the flagged card does not report the clause")
	}
	if CantSpendManaFor(NewCard("Plain Bear", uuid.New())) {
		t.Error("a card with no catalog key reports the clause")
	}
	if !ManaSpendForCast(flagged).NoMana {
		t.Error("the cast's spend context does not carry NoMana")
	}
	if ManaSpendForCast(flagged).allowsToken(ManaToken{Color: "B"}) {
		t.Error("a NoMana context admitted a black mana")
	}
}

// cantSpendManaPoolSize is how much mana is floating: the pool holds
// RUNS of identical mana (#3041), so the entries are counted by amount.
func cantSpendManaPoolSize(p *Player) int {
	n := 0
	for _, tok := range p.ManaPool {
		n += tok.N()
	}
	return n
}

// cantSpendManaTappedLands counts the controller's tapped lands.
func cantSpendManaTappedLands(g *Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Tapped && c.IsLand() {
			n++
		}
	}
	return n
}
