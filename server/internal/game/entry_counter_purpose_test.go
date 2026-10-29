package game

import "testing"

func TestEntryCounterProjectionDoesNotExecuteCatalogFunctions(t *testing.T) {
	previous := CatalogLookup
	t.Cleanup(func() { CatalogLookup = previous })
	panicReplacement := func(*ReplacementEvent, *Game, *Card) bool { panic("view called a replacement predicate") }
	def := &CardDef{
		Purpose: Purpose{Draws: 2},
		Replacements: []ReplacementEffect{
			{SelfEntryCounters: &EntryCounterAmount{Kind: CounterPlusOne, Amount: 4}, AppliesTo: panicReplacement},
			{AppliesTo: panicReplacement, Replace: func(*ReplacementEvent, *Game, *Card) error { panic("view executed replacement") }},
			{SelfEntryCounters: &EntryCounterAmount{Kind: CounterPlusOne, Amount: 99}, ActiveWhen: Designation{Kind: 1}},
		},
		EntersWithCountersFromCast: []EntryCountersFromCast{
			{Kind: CounterPlusOne, CountIsX: true, Count: func(CastCounts) int { panic("view evaluated X function") }},
			{Kind: CounterPlusOne, Count: func(CastCounts) int { panic("view evaluated conditional function") }},
		},
	}
	CatalogLookup = func(string) *CardDef { return def }
	p := CardPurposeOf(Card{OracleID: "entry-counter-projection-test"})
	if p.Draws != 2 || p.EntryCounters == nil || len(*p.EntryCounters) != 2 {
		t.Fatalf("purpose %+v", p)
	}
	if (*p.EntryCounters)[0] != (EntryCounterAmount{Kind: CounterPlusOne, Amount: 4}) || (*p.EntryCounters)[1] != (EntryCounterAmount{Kind: CounterPlusOne, AmountIsX: true}) {
		t.Fatalf("entries %+v", *p.EntryCounters)
	}
	if def.Purpose.EntryCounters != nil {
		t.Fatal("projection mutated catalog purpose")
	}
}
