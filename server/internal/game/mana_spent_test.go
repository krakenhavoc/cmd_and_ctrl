package game

import (
	"testing"

	"github.com/google/uuid"
)

// mana_spent_test.go — #761: a stack item records the mana that paid
// for it.
//
// What is pinned here is the contract ADR 0070 argues for: the tokens
// are recorded on every payment path, a waived charge says so out loud
// rather than looking like a free cast, the colour-maximising strategy
// is used only by a spell that reads colours, and the record survives
// an undo and a snapshot. The catalog half — converge, sunburst,
// adamant, "if no mana was spent" — is in
// cards/effects/mana_spent_cards_test.go.

// floatMana drops `spec` into a player's pool, one token per rune.
func floatMana(p *Player, spec string) {
	for _, r := range spec {
		p.ManaPool.AddMana(ManaToken{Color: string(r), Source: uuid.New()})
	}
}

// castForTest puts a spell of the given cost in hand and casts it
// under the given params, returning its stack item.
func castForTest(t *testing.T, g *Game, p *Player, cost string, params CastSpellParams) *StackItem {
	t.Helper()
	c := NewCard("Test Spell", p.ID)
	c.TypeLine = "Instant"
	c.ManaCost = cost
	c.Controller = p.ID
	p.Hand.PushTop(c)
	if err := g.CastSpell(p.ID, c.InstanceID, params); err != nil {
		t.Fatalf("cast: %v", err)
	}
	item := g.StackMeta[c.InstanceID]
	if item == nil {
		t.Fatalf("no stack item for the cast spell")
	}
	return item
}

// --- the record ----------------------------------------------------

// The headline: a strict cast records the tokens that left the pool,
// with their colours and their sources intact.
func TestStrictCastRecordsTheManaSpent(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	floatMana(me, "UBR")

	item := castForTest(t, g, me, "{1}{B}", CastSpellParams{Strict: true})
	if got := item.Paid.ManaSpentCount(); got != 2 {
		t.Fatalf("recorded %d mana, want 2", got)
	}
	if item.Paid.OnPaper {
		t.Error("a strict cast recorded OnPaper")
	}
	if item.Paid.SpentOfColor("B") != 1 {
		t.Errorf("the {B} pip did not come from a black token: %+v", item.Paid.Mana)
	}
	for _, tok := range item.Paid.Mana {
		if tok.Source == uuid.Nil {
			t.Error("a recorded token lost its source")
		}
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("pool has %d tokens left, want 1", len(me.ManaPool))
	}
}

// A permissive cast — the human default — records "the engine waived
// this", which is a different fact from "nothing was spent" and is
// what stops every "if no mana was spent" card firing at such a table.
func TestPermissiveCastRecordsOnPaper(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]

	item := castForTest(t, g, me, "{1}{B}", CastSpellParams{})
	if !item.Paid.OnPaper {
		t.Fatal("a permissive cast did not record OnPaper")
	}
	if item.Paid.NoManaSpent() {
		t.Error("an unrecorded payment reads as \"no mana was spent\" — the #259 direction")
	}
	if len(item.Paid.ColorsSpent()) != 0 {
		t.Error("an unrecorded payment claims colours")
	}
}

// A strict-mode ForceCast is the same fact by a different route.
func TestForceCastRecordsOnPaper(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]

	item := castForTest(t, g, me, "{5}{B}", CastSpellParams{Strict: true, ForceCast: true})
	if !item.Paid.OnPaper || item.Paid.NoManaSpent() {
		t.Errorf("force cast: OnPaper=%v NoManaSpent=%v, want true/false", item.Paid.OnPaper, item.Paid.NoManaSpent())
	}
}

// A spell that really was free — a {0} cost paid strictly — records a
// KNOWN nothing, which is what Vexing Bauble punishes.
func TestAFreeCastRecordsAKnownNothing(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	c := NewCard("Free Test Spell", me.ID)
	c.TypeLine = "Instant"
	c.ManaCost = "{0}"
	c.Controller = me.ID
	me.Hand.PushTop(c)

	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	paid := g.StackMeta[c.InstanceID].Paid
	if !paid.NoManaSpent() {
		t.Errorf("a {0} cast: NoManaSpent=%v OnPaper=%v, want true/false", paid.NoManaSpent(), paid.OnPaper)
	}
}

// CR 707.10: mana is not an object, so nothing was spent to cast a
// copy — and that zero is REAL, not unknown.
func TestASpellCopyRecordsNothing(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	floatMana(me, "BBB")
	item := castForTest(t, g, me, "{1}{B}", CastSpellParams{Strict: true})
	if item.Paid.ManaSpentCount() == 0 {
		t.Fatal("the original recorded nothing")
	}

	g.mu.Lock()
	_ = g.CopySpellForEffect(item.ID, me.ID, false)
	var copyItem *StackItem
	for id, it := range g.StackMeta {
		if id != item.ID {
			copyItem = it
		}
	}
	g.mu.Unlock()
	if copyItem == nil {
		t.Fatal("no copy on the stack")
	}
	if !copyItem.Paid.NoManaSpent() || copyItem.Paid.OnPaper {
		t.Errorf("copy: NoManaSpent=%v OnPaper=%v, want true/false", copyItem.Paid.NoManaSpent(), copyItem.Paid.OnPaper)
	}
}

// --- the event -----------------------------------------------------

// The log line stops being "mana was spent" and says what.
func TestManaSpentEventCarriesAmountAndColors(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	floatMana(me, "UBC")

	castForTest(t, g, me, "{2}{B}", CastSpellParams{Strict: true})
	var ev *Event
	for i := range g.Events {
		if g.Events[i].Kind == EventManaSpent {
			ev = &g.Events[i]
		}
	}
	if ev == nil {
		t.Fatal("no EventManaSpent")
	}
	if ev.Amount != 3 {
		t.Errorf("Amount = %d, want 3", ev.Amount)
	}
	// Colourless is not a colour (CR 105.1).
	want := map[string]bool{"U": true, "B": true}
	if len(ev.Colors) != 2 {
		t.Fatalf("Colors = %v, want two colours and no {C}", ev.Colors)
	}
	for _, c := range ev.Colors {
		if !want[c] {
			t.Errorf("Colors = %v, want only U and B", ev.Colors)
		}
	}
}

// --- the solver strategy -------------------------------------------

// A converge spell spreads the generic half across colours it has not
// spent; an ordinary spell keeps the colourless-first order that saves
// coloured mana for the next cast.
func TestDistinctColorStrategyIsUsedOnlyByASpellThatReadsColors(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pool := ManaPool{
		{Color: "C"}, {Color: "W"}, {Color: "U"}, {Color: "B"}, {Color: "R"}, {Color: "G"},
	}
	cost, err := ParseCost("{2}{B}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}

	def := append(ManaPool(nil), pool...)
	spentDefault, ok := def.SpendManaForWith(cost, 0, ManaSpendContext{}, SpendPreserveColors)
	if !ok {
		t.Fatal("default spend failed")
	}
	distinct := append(ManaPool(nil), pool...)
	spentDistinct, ok := distinct.SpendManaForWith(cost, 0, ManaSpendContext{}, SpendDistinctColors)
	if !ok {
		t.Fatal("distinct spend failed")
	}

	nDefault := PaidCost{Mana: spentDefault}.ColorsSpentCount()
	nDistinct := PaidCost{Mana: spentDistinct}.ColorsSpentCount()
	if nDefault != 2 {
		t.Errorf("default strategy spent %d colours (%v), want 2 — {C} first, then {B}, then one more", nDefault, spentDefault)
	}
	if nDistinct != 3 {
		t.Errorf("distinct strategy spent %d colours (%v), want 3", nDistinct, spentDistinct)
	}
	_ = me
}

// The strategy reorders; it never changes whether a cost is payable.
// Same pool, same cost, same answer, either way.
func TestSpendStrategyNeverChangesPayability(t *testing.T) {
	costs := []string{"{3}", "{2}{G}", "{G}{G}", "{6}", "{1}{W}{U}"}
	pools := []ManaPool{
		{{Color: "C"}, {Color: "G"}, {Color: "G"}},
		{{Color: "W"}, {Color: "U"}, {Color: "C"}},
		{},
		{{Color: "R"}, {Color: "R"}, {Color: "R"}, {Color: "R"}},
	}
	for _, cs := range costs {
		cost, err := ParseCost(cs)
		if err != nil {
			t.Fatalf("ParseCost(%q): %v", cs, err)
		}
		for _, pool := range pools {
			a := append(ManaPool(nil), pool...)
			b := append(ManaPool(nil), pool...)
			_, okA := a.SpendManaForWith(cost, 0, ManaSpendContext{}, SpendPreserveColors)
			_, okB := b.SpendManaForWith(cost, 0, ManaSpendContext{}, SpendDistinctColors)
			if okA != okB {
				t.Errorf("%q from %v: default=%v distinct=%v", cs, pool, okA, okB)
			}
		}
	}
}

// The distinct strategy pays one token at a time rather than emptying
// a colour's bucket: two Plains and an Island pay {2} as W + U, not
// W + W.
func TestDistinctColorStrategySpreadsWithinOneBucket(t *testing.T) {
	pool := ManaPool{{Color: "W"}, {Color: "W"}, {Color: "U"}}
	cost, _ := ParseCost("{2}")
	spent, ok := pool.SpendManaForWith(cost, 0, ManaSpendContext{}, SpendDistinctColors)
	if !ok {
		t.Fatal("spend failed")
	}
	if n := (PaidCost{Mana: spent}).ColorsSpentCount(); n != 2 {
		t.Errorf("spent %v (%d colours), want two colours", spent, n)
	}
}

// --- undo and snapshots --------------------------------------------

// The record is deep-copied, so an undo snapshot and the live game
// cannot share a backing array.
func TestPaidCostIsDeepCopiedByClone(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	floatMana(me, "BB")
	item := castForTest(t, g, me, "{1}{B}", CastSpellParams{Strict: true})

	clone := g.Clone()
	cloned := clone.StackMeta[item.ID]
	if cloned == nil {
		t.Fatal("the clone lost the stack item")
	}
	if cloned.Paid.ManaSpentCount() != item.Paid.ManaSpentCount() {
		t.Fatalf("clone recorded %d mana, want %d", cloned.Paid.ManaSpentCount(), item.Paid.ManaSpentCount())
	}
	cloned.Paid.Mana[0].Color = "ZZ"
	if item.Paid.Mana[0].Color == "ZZ" {
		t.Error("the clone shares the live record's backing array")
	}
}

// And it round-trips a snapshot, because a restore that lost it would
// resolve a converge spell for zero.
func TestPaidCostSurvivesASnapshotRoundTrip(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	floatMana(me, "URG")
	item := castForTest(t, g, me, "{2}{R}", CastSpellParams{Strict: true})
	want := item.Paid.ColorsSpentCount()

	snap := g.CaptureSnapshot()
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	back := restored.StackMeta[item.ID]
	if back == nil {
		t.Fatal("the restored game lost the stack item")
	}
	if got := back.Paid.ColorsSpentCount(); got != want {
		t.Errorf("restored %d colours, want %d", got, want)
	}
	if back.Paid.ManaSpentCount() != item.Paid.ManaSpentCount() {
		t.Errorf("restored %d mana, want %d", back.Paid.ManaSpentCount(), item.Paid.ManaSpentCount())
	}
}

// --- the readers ---------------------------------------------------

// The accessors a cast trigger uses, against a record the engine
// actually built.
func TestStackItemPaidForEffectReadsTheRecord(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	floatMana(me, "RRB")
	item := castForTest(t, g, me, "{2}{R}", CastSpellParams{Strict: true})

	g.mu.Lock()
	paid := g.StackItemPaidForEffect(item.ID)
	missing := g.StackItemPaidForEffect(uuid.New())
	g.mu.Unlock()
	if paid.ManaSpentCount() != 3 {
		t.Errorf("read %d mana, want 3", paid.ManaSpentCount())
	}
	if paid.SpentOfColor("R") != 2 {
		t.Errorf("SpentOfColor(R) = %d, want 2 — adamant reads this", paid.SpentOfColor("R"))
	}
	if !missing.NoManaSpent() {
		t.Error("an item that is not on the stack should read as a known nothing")
	}
}
