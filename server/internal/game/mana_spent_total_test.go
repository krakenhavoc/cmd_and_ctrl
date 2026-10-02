package game

import (
	"testing"

	"github.com/google/uuid"
)

// mana_spent_total_test.go — #1735: "the amount of mana spent to cast"
// a permanent spell (CR 601.2h), read INSIDE the CR 614 entry window.
//
// The record itself is ADR 0068's (StackItem.Paid.Mana, pinned in
// mana_spent_test.go). What is pinned here is the figure the entry
// window hands a catalog hook — CastCounts.ManaSpent — across every way
// the total cost is settled (CR 601.2f): a plain cast, {X}, kicker, a
// reduction, an increase and the commander tax, a Phyrexian symbol paid
// with life, a free cast, a waived payment; that it survives a clone
// and a snapshot round trip; and that the copy selector reads the same
// number through EntryCastCountsForEffect, at the prompt and again when
// the answer arrives. Mockingbird, the card this was built for, is in
// cards/effects/mana_spent_total_test.go.

const (
	testSpentCountersOracle = "test-enters-with-counters-per-mana-spent"
	testSpentCopyOracle     = "test-enters-as-copy-reading-mana-spent"
)

// spentEntryCounters is "enters with a +1/+1 counter on it for each
// mana spent to cast it" (Verazol, the Split Current) — the CR 614.1c
// reader of the new field, spelled out here as xEntryCounters is.
func spentEntryCounters() EntryCountersFromCast {
	return EntryCountersFromCast{
		Kind:  CounterPlusOne,
		Count: func(cast CastCounts) int { return cast.ManaSpent },
	}
}

// seedSpentCreature puts a 1/1 creature that counts the mana spent on
// it into `p`'s hand and returns its ID. 1/1 rather than
// 0/0 so a cast that spent nothing still leaves a permanent to read.
func seedSpentCreature(t *testing.T, g *Game, p *Player, cost string) uuid.UUID {
	t.Helper()
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Test Counter Of Mana", p.ID)
	c.TypeLine = "Creature — Serpent"
	c.ManaCost = cost
	c.OracleID = testSpentCountersOracle
	c.Controller = p.ID
	c.Power, c.Toughness = 1, 1
	p.Hand.PushTop(c)
	return c.InstanceID
}

// castAndCountSpent casts the seeded creature with `params`, resolves
// it, and returns the +1/+1 counters it entered with — the figure the
// CR 614 window saw.
func castAndCountSpent(t *testing.T, g *Game, p *Player, id uuid.UUID, params CastSpellParams) int {
	t.Helper()
	if err := g.CastSpell(p.ID, id, params); err != nil {
		t.Fatalf("cast: %v", err)
	}
	resolveTop(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatalf("the creature did not reach the battlefield")
	}
	return counters(t, g, id, CounterPlusOne)
}

// treasureMana is a token minted by a Treasure: mana like any other.
func treasureMana(color string) ManaToken {
	return ManaToken{Color: color, Source: uuid.New(), SourceKinds: ManaSourceTreasure | ManaSourceArtifact}
}

// The baseline, and the "mana from a Treasure counts" clause: {2}{G}
// paid with two lands' worth and a Treasure's is three mana spent.
func TestManaSpentCountsEveryManaThatPaid(t *testing.T) {
	g := newActiveGame(t)
	stubEntryCountersFromCast(t, testSpentCountersOracle, spentEntryCounters())
	me := g.Seats[g.Turn.ActiveSeat]
	id := seedSpentCreature(t, g, me, "{2}{G}")
	floatMana(me, "GC")
	me.ManaPool.AddMana(treasureMana("R"))

	if got := castAndCountSpent(t, g, me, id, CastSpellParams{Strict: true}); got != 3 {
		t.Errorf("entered with %d counters, want 3 — one per mana spent", got)
	}
}

// {X}: X=3 on {X}{G} is four mana spent, not the printed mana value 1.
func TestManaSpentIncludesX(t *testing.T) {
	g := newActiveGame(t)
	stubEntryCountersFromCast(t, testSpentCountersOracle, spentEntryCounters())
	me := g.Seats[g.Turn.ActiveSeat]
	id := seedSpentCreature(t, g, me, "{X}{G}")
	floatMana(me, "GGGG")

	if got := castAndCountSpent(t, g, me, id, CastSpellParams{Strict: true, XValue: 3}); got != 4 {
		t.Errorf("X=3 on {X}{G}: %d counters, want 4", got)
	}
}

// Kicker's mana is part of the total cost (CR 601.2f), so it is part
// of the amount spent.
func TestManaSpentIncludesKicker(t *testing.T) {
	g := newActiveGame(t)
	stubEntryCountersFromCast(t, testSpentCountersOracle, spentEntryCounters())
	stubOptionalCosts(t, testSpentCountersOracle, []AdditionalCost{
		{Optional: true, Key: KickerKey, ManaCost: "{2}", Label: "Kicker {2}"},
	})
	me := g.Seats[g.Turn.ActiveSeat]
	id := seedSpentCreature(t, g, me, "{1}{G}")
	floatMana(me, "GGGG")

	if got := castAndCountSpent(t, g, me, id, CastSpellParams{Strict: true, OptionalCosts: []int{0}}); got != 4 {
		t.Errorf("{1}{G} kicked for {2}: %d counters, want 4", got)
	}
}

// A reduction lowers what was spent: {2}{G} under "creature spells
// cost {1} less" spends two, and the third mana stays in the pool.
func TestManaSpentIsAfterReductions(t *testing.T) {
	const reducer = "test-spent-reducer"
	g := newActiveGame(t)
	stubEntryCountersFromCast(t, testSpentCountersOracle, spentEntryCounters())
	withCatalogCostModifiers(t, modifiersFor(reducer, CostModifier{
		Kind: CostReduction, Label: "Creature spells you cast cost {1} less to cast.", Amount: fixed(1),
	}))
	me := g.Seats[g.Turn.ActiveSeat]
	modifierSource(t, g, me, "Reducer", reducer)
	id := seedSpentCreature(t, g, me, "{2}{G}")
	floatMana(me, "GGG")

	if got := castAndCountSpent(t, g, me, id, CastSpellParams{Strict: true}); got != 2 {
		t.Errorf("{2}{G} reduced by {1}: %d counters, want 2", got)
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("pool has %d left, want 1 — the reduction's mana was not spent", len(me.ManaPool))
	}
}

// An increase and the commander tax raise it: a {1}{G} commander cast
// once before ({2} tax) under a Sphere ({1} more) spends five.
func TestManaSpentIncludesIncreasesAndCommanderTax(t *testing.T) {
	const sphere = "test-spent-sphere"
	g := newActiveGame(t)
	stubEntryCountersFromCast(t, testSpentCountersOracle, spentEntryCounters())
	withCatalogCostModifiers(t, modifiersFor(sphere, CostModifier{
		Kind: CostIncrease, Label: "Spells cost {1} more to cast.", Amount: fixed(1),
	}))
	me := g.Seats[g.Turn.ActiveSeat]
	modifierSource(t, g, me, "Sphere", sphere)
	advanceTo(t, g, StepPrecombatMain)
	cmdr := NewCard("Test Commander Of Mana", me.ID)
	cmdr.TypeLine = "Legendary Creature — Serpent"
	cmdr.ManaCost = "{1}{G}"
	cmdr.OracleID = testSpentCountersOracle
	cmdr.Controller = me.ID
	cmdr.Power, cmdr.Toughness = 1, 1
	me.Command.PushTop(cmdr)
	if me.CommanderCasts == nil {
		me.CommanderCasts = make(map[uuid.UUID]int)
	}
	me.CommanderCasts[cmdr.InstanceID] = 1
	floatMana(me, "GGGGG")

	got := castAndCountSpent(t, g, me, cmdr.InstanceID, CastSpellParams{Strict: true, FromZone: "command"})
	if got != 5 {
		t.Errorf("taxed commander under a sphere: %d counters, want 5 ({1}{G} + {2} tax + {1})", got)
	}
}

// Life is not mana (CR 107.4f): a Phyrexian symbol paid with 2 life
// adds nothing to the amount spent.
func TestManaSpentExcludesPhyrexianLife(t *testing.T) {
	g := newActiveGame(t)
	stubEntryCountersFromCast(t, testSpentCountersOracle, spentEntryCounters())
	me := g.Seats[g.Turn.ActiveSeat]
	id := seedSpentCreature(t, g, me, "{1}{G/P}")
	floatMana(me, "C")
	life := me.Life

	got := castAndCountSpent(t, g, me, id, CastSpellParams{Strict: true, PhyrexianLife: 1})
	if me.Life != life-PhyrexianLifePerSymbol {
		t.Fatalf("life = %d, want %d — the symbol was not paid with life", me.Life, life-PhyrexianLifePerSymbol)
	}
	if got != 1 {
		t.Errorf("{1}{G/P} paid {1} + 2 life: %d counters, want 1", got)
	}
}

// "Without paying its mana cost" spends nothing, and that zero is a
// KNOWN zero — the permanent remembers no mana was spent on it.
func TestManaSpentIsZeroForAFreeCast(t *testing.T) {
	g := newActiveGame(t)
	stubEntryCountersFromCast(t, testSpentCountersOracle, spentEntryCounters())
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, StepPrecombatMain)
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(Card{
			InstanceID: id, Name: "Free Counter Of Mana", TypeLine: "Creature — Serpent",
			ManaCost: "{3}{G}", OracleID: testSpentCountersOracle,
			Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		})
		g.GrantCastPermissionToCardsForEffect(CastPermission{
			Player: me.ID, Zone: ZoneExile, Cost: "{0}", CastOnly: true,
			Duration: g.UntilEndOfTurnDuration(), Label: "Test — cast it without paying its mana cost",
		}, []Card{g.Exile.Cards[len(g.Exile.Cards)-1]})
	})
	floatMana(me, "GGGG")

	if got := castAndCountSpent(t, g, me, id, CastSpellParams{Strict: true, FromZone: "exile"}); got != 0 {
		t.Errorf("a free cast entered with %d counters, want 0", got)
	}
	if len(me.ManaPool) != 4 {
		t.Errorf("pool has %d, want 4 — a free cast spends nothing", len(me.ManaPool))
	}
	perm, _ := g.LookupCardForEffect(id)
	if !perm.ManaSpentToCast().None() {
		t.Error("the permanent does not remember that no mana was spent on it")
	}
}

// A payment the engine waived (permissive mode) is unknown, and an
// unknown amount reads as zero — never as "enough" (ADR 0068 §3).
func TestManaSpentIsZeroWhenThePaymentWasWaived(t *testing.T) {
	g := newActiveGame(t)
	stubEntryCountersFromCast(t, testSpentCountersOracle, spentEntryCounters())
	me := g.Seats[g.Turn.ActiveSeat]
	id := seedSpentCreature(t, g, me, "{2}{G}")

	if got := castAndCountSpent(t, g, me, id, CastSpellParams{}); got != 0 {
		t.Errorf("a waived payment entered with %d counters, want 0", got)
	}
}

// Every entry that is not a resolving spell reads the zero CastCounts,
// as does a nil event.
func TestEntryCastCountsIsZeroOutsideACast(t *testing.T) {
	g := newActiveGame(t)
	g.mu.Lock()
	defer g.mu.Unlock()
	if got := g.EntryCastCountsForEffect(nil); got.ManaSpent != 0 || got.X != 0 {
		t.Errorf("nil event: %+v, want zero", got)
	}
	ev := &ReplacementEvent{Kind: RepEventMove, CardID: uuid.New(), OldZone: ZoneGraveyard, NewZone: ZoneBattlefield}
	if got := g.EntryCastCountsForEffect(ev); got.ManaSpent != 0 || got.X != 0 || got.Kicked != 0 {
		t.Errorf("a reanimation's entry: %+v, want zero — nothing was cast", got)
	}
}

// The figure survives undo: a clone taken with the spell on the stack
// resolves to the same total, and the clone's record is its own.
func TestManaSpentSurvivesAClone(t *testing.T) {
	g := newActiveGame(t)
	stubEntryCountersFromCast(t, testSpentCountersOracle, spentEntryCounters())
	me := g.Seats[g.Turn.ActiveSeat]
	id := seedSpentCreature(t, g, me, "{X}{G}")
	floatMana(me, "GGG")
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, XValue: 2}); err != nil {
		t.Fatalf("cast: %v", err)
	}

	clone := g.Clone()
	resolveTop(t, clone)
	if got := counters(t, clone, id, CounterPlusOne); got != 3 {
		t.Errorf("the clone's entry saw %d mana spent, want 3", got)
	}
	resolveTop(t, g)
	if got := counters(t, g, id, CounterPlusOne); got != 3 {
		t.Errorf("the live game's entry saw %d mana spent, want 3", got)
	}
}

// And a snapshot: restore a game with the spell on the stack, resolve
// it there, and the entry window still knows what was spent. The
// permanent's own record (CR 400.7d) round-trips too.
func TestManaSpentSurvivesASnapshotRoundTrip(t *testing.T) {
	g := newActiveGame(t)
	stubEntryCountersFromCast(t, testSpentCountersOracle, spentEntryCounters())
	me := g.Seats[g.Turn.ActiveSeat]
	id := seedSpentCreature(t, g, me, "{X}{G}")
	floatMana(me, "GGGG")
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, XValue: 3}); err != nil {
		t.Fatalf("cast: %v", err)
	}

	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("restore with the spell on the stack: %v", err)
	}
	resolveTop(t, restored)
	if got := counters(t, restored, id, CounterPlusOne); got != 4 {
		t.Fatalf("after a restore the entry saw %d mana spent, want 4", got)
	}

	again, err := restored.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("restore with the permanent on the battlefield: %v", err)
	}
	perm, ok := again.LookupCardForEffect(id)
	if !ok {
		t.Fatal("the permanent is gone after the second restore")
	}
	if got := perm.ManaSpentToCast().Total(); got != 4 {
		t.Errorf("the permanent remembers %d mana spent on it, want 4", got)
	}
}

// spentCopySelector is the shape of Mockingbird's replacement, spelled
// out here so the game package can test the seam without the catalog:
// an "enters as a copy" self-replacement whose candidate filter reads
// the amount spent through EntryCastCountsForEffect. Every evaluation
// is recorded in `seen`.
func spentCopySelector(seen *[]int) ReplacementEffect {
	return ReplacementEffect{
		Watches:         []EventKind{EventZoneMove},
		SelfReplacement: true,
		Label:           "Test Mockingbird",
		AppliesTo: func(ev *ReplacementEvent, _ *Game, src *Card) bool {
			return ev.Kind == RepEventMove && ev.NewZone == ZoneBattlefield && src != nil && ev.CardID == src.InstanceID
		},
		Controller: func(ev *ReplacementEvent, _ *Game, _ *Card) uuid.UUID { return ev.Actor },
		CopySelector: &CopySelector{
			Candidates: func(ev *ReplacementEvent, g *Game, _ *Card) []uuid.UUID {
				spent := g.EntryCastCountsForEffect(ev).ManaSpent
				*seen = append(*seen, spent)
				var out []uuid.UUID
				for _, c := range g.Battlefield.Cards {
					if c.InstanceID == ev.CardID || !c.IsCreature() {
						continue
					}
					if mv, ok := g.ManaValueForEffect(c); ok && mv <= spent {
						out = append(out, c.InstanceID)
					}
				}
				return out
			},
		},
	}
}

// The seam itself: a copy selector's candidate filter reads the amount
// spent, and reads the same amount when the prompt is built and when
// the answer arrives.
func TestCopySelectorReadsTheManaSpentAtThePromptAndTheAnswer(t *testing.T) {
	var seen []int
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		testSpentCopyOracle: {spentCopySelector(&seen)},
	})
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, StepPrecombatMain)

	cheap := NewCard("Two Drop", me.ID)
	cheap.TypeLine, cheap.ManaCost, cheap.Controller = "Creature — Bear", "{1}{G}", me.ID
	cheap.Power, cheap.Toughness = 2, 2
	big := NewCard("Five Drop", me.ID)
	big.TypeLine, big.ManaCost, big.Controller = "Creature — Giant", "{4}{G}", me.ID
	big.Power, big.Toughness = 5, 5
	g.Battlefield.PushTop(cheap)
	g.Battlefield.PushTop(big)

	bird := NewCard("Test Mockingbird", me.ID)
	bird.TypeLine, bird.ManaCost, bird.OracleID, bird.Controller = "Creature — Bird", "{X}{U}", testSpentCopyOracle, me.ID
	bird.Power, bird.Toughness = 1, 1
	me.Hand.PushTop(bird)
	floatMana(me, "UUU")
	if err := g.CastSpell(me.ID, bird.InstanceID, CastSpellParams{Strict: true, XValue: 2}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	resolveTop(t, g)

	var prompt *PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceCopyTarget {
			prompt = c
		}
	}
	if prompt == nil {
		t.Fatalf("no copy prompt; candidate filter saw %v", seen)
	}
	if len(prompt.CopyOptions) != 1 || prompt.CopyOptions[0] != cheap.InstanceID {
		t.Fatalf("offered %v, want only the two-drop (three mana spent)", prompt.CopyOptions)
	}
	if err := g.ResolveCopyTarget(prompt.ID, prompt.Chooser, cheap.InstanceID); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}
	if len(seen) < 2 {
		t.Fatalf("the filter ran %d times, want the prompt and the answer", len(seen))
	}
	for i, n := range seen {
		if n != 3 {
			t.Errorf("evaluation %d read %d mana spent, want 3", i, n)
		}
	}
	got, ok := g.LookupCardForEffect(bird.InstanceID)
	if !ok || got.Name != "Two Drop" {
		t.Errorf("the copy did not land: %+v", got.Name)
	}
}
