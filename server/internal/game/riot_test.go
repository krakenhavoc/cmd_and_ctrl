package game

import (
	"testing"

	"github.com/google/uuid"
)

// riot_test.go — riot (CR 702.136), unleash (CR 702.98) and the CR
// 614.12 look-ahead they are read through (ADR 0109 §10, #1556). The
// mechanic, with keywords stamped on the card the way the deck importer
// stamps them and statics stubbed through the catalog hook; the cards
// are pinned in cards/effects/riot_test.go.

const (
	testRiotGrantOracle    = "test-rhythm-of-the-wild"
	testRiotRemovalOracle  = "test-dress-down"
	testUnleashGrantOracle = "test-unleash-grant"
)

// seedKeywordCreature puts a 2/2 creature with `keywords` printed into
// the active seat's hand, in its precombat main phase.
func seedKeywordCreature(t *testing.T, g *Game, name string, keywords ...string) uuid.UUID {
	t.Helper()
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	c := NewCard(name, me.ID)
	c.TypeLine = "Creature — Goblin"
	c.ScryfallID = "test-printing"
	c.Power, c.Toughness = 2, 2
	c.Keywords = keywords
	me.Hand.PushTop(c)
	return c.InstanceID
}

// castAndResolve casts the seeded creature and resolves it, leaving any
// entry question open.
func castAndResolve(t *testing.T, g *Game, id uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	if err := g.CastSpell(me.ID, id, CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	resolveTop(t, g)
}

// openChoice is the one open prompt, which must be of `kind`.
func openChoice(t *testing.T, g *Game, kind PendingChoiceKind) *PendingChoice {
	t.Helper()
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != kind {
		var got []PendingChoiceKind
		for _, c := range g.PendingChoices {
			got = append(got, c.Kind)
		}
		t.Fatalf("want one open %s prompt, have %v", kind, got)
	}
	return g.PendingChoices[0]
}

// answerRiot answers the open riot question.
func answerRiot(t *testing.T, g *Game, counter bool) {
	t.Helper()
	c := openChoice(t, g, PendingChoiceEntryRiot)
	if err := g.ResolveEntryRiot(c.ID, c.Chooser, counter); err != nil {
		t.Fatalf("ResolveEntryRiot: %v", err)
	}
}

// withRiotStatics stubs the catalog statics: a "Rhythm" granting riot to
// nontoken creatures its controller controls, a "Dress Down" removing
// every creature's abilities, and an unleash grant.
func withRiotStatics(t *testing.T) {
	t.Helper()
	prev := CatalogStaticAbilities
	CatalogStaticAbilities = func(key string) []StaticAbility {
		switch key {
		case testRiotGrantOracle:
			return []StaticAbility{{
				Layer: Layer6Ability,
				AppliesTo: func(target *Card, _ *Game, source *Card) bool {
					return target.IsCreature() && !target.IsToken() && target.Controller == source.Controller
				},
				Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
					c.Abilities = AppendKeywordAbility(c.Abilities, KeywordRiot)
				},
			}}
		case testUnleashGrantOracle:
			return []StaticAbility{{
				Layer: Layer6Ability,
				AppliesTo: func(target *Card, _ *Game, source *Card) bool {
					return target.IsCreature() && target.Controller == source.Controller
				},
				Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
					c.Abilities = AppendKeywordAbility(c.Abilities, KeywordUnleash)
				},
			}}
		case testRiotRemovalOracle:
			return []StaticAbility{{
				Layer:            Layer6Ability,
				RemovesAbilities: true,
				AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
					return target.IsCreature()
				},
			}}
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { CatalogStaticAbilities = prev })
}

// pushStaticSource puts an enchantment carrying a stubbed static onto
// the battlefield under `controller`.
func pushStaticSource(g *Game, controller uuid.UUID, oracleID string) uuid.UUID {
	c := NewCard("Static "+oracleID, controller)
	c.TypeLine = "Enchantment"
	c.OracleID = oracleID
	return pushTypedTestCard(g, c)
}

func riotState(t *testing.T, g *Game, id uuid.UUID) (counters int, haste bool) {
	t.Helper()
	c := layeredBattlefieldCard(t, g, id)
	return c.Counters[CounterPlusOne], HasKeyword(&c, "haste")
}

// CR 702.136a, the counter: the question is asked BEFORE the permanent
// is on the battlefield (CR 614.12a), and the counter is on it as it
// lands.
func TestRiotCounterAnswerEntersWithACounter(t *testing.T) {
	g := newActiveGame(t)
	id := seedKeywordCreature(t, g, "Riot Goblin", KeywordRiot)
	castAndResolve(t, g, id)

	c := openChoice(t, g, PendingChoiceEntryRiot)
	if g.Battlefield.Contains(id) {
		t.Fatal("the creature is on the battlefield while its riot question is open")
	}
	if c.Chooser != g.Seats[g.Turn.ActiveSeat].ID || c.Source != id {
		t.Errorf("the question goes to the caster about the entering card: chooser %s source %s", c.Chooser, c.Source)
	}
	answerRiot(t, g, true)

	n, haste := riotState(t, g, id)
	if n != 1 || haste {
		t.Errorf("counter answer: %d counters, haste %v; want 1, false", n, haste)
	}
	if g.RiotHasteForEffect(id) {
		t.Error("a counter answer left a riot haste record")
	}
}

// CR 702.136a, haste: "if you don't, it gains haste" — no counter, and
// it can attack the turn it entered (CR 702.10b).
func TestRiotHasteAnswerGainsHaste(t *testing.T) {
	g := newActiveGame(t)
	id := seedKeywordCreature(t, g, "Riot Goblin", KeywordRiot)
	castAndResolve(t, g, id)
	answerRiot(t, g, false)

	n, haste := riotState(t, g, id)
	if n != 0 || !haste {
		t.Fatalf("haste answer: %d counters, haste %v; want 0, true", n, haste)
	}
	c := layeredBattlefieldCard(t, g, id)
	if HasSummoningSickness(&c) {
		t.Error("a creature that took riot's haste is summoning sick")
	}
	if !g.RiotHasteForEffect(id) {
		t.Error("the view's riot haste reader does not see the record")
	}
}

// The haste is the permanent's, as the object it is (CR 400.7): it does
// not follow the card back after a trip through the hand.
func TestRiotHasteEndsWithTheObject(t *testing.T) {
	g := newActiveGame(t)
	id := seedKeywordCreature(t, g, "Riot Goblin", KeywordRiot)
	castAndResolve(t, g, id)
	answerRiot(t, g, false)

	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, me.Hand, id); err != nil {
			t.Fatalf("bounce: %v", err)
		}
		g.layerVersion.Add(1)
	})
	g.WithWriteLock(func() { g.ClearExpiredScopedStaticsLocked() })
	castAndResolve(t, g, id)
	answerRiot(t, g, true)
	if n, haste := riotState(t, g, id); n != 1 || haste {
		t.Errorf("recast with the counter: %d counters, haste %v; want 1, false", n, haste)
	}
}

// CR 614.12: a creature entering under a static that grants riot has
// riot as it enters, so it is asked — the look-ahead's whole point.
func TestRiotGrantedByAStaticIsAskedAsItEnters(t *testing.T) {
	withRiotStatics(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	pushStaticSource(g, me.ID, testRiotGrantOracle)
	id := seedKeywordCreature(t, g, "Vanilla Bear")
	castAndResolve(t, g, id)
	answerRiot(t, g, false)
	if n, haste := riotState(t, g, id); n != 0 || !haste {
		t.Errorf("granted riot, haste answer: %d counters, haste %v", n, haste)
	}
}

// CR 702.136b: a printed riot and a granted one each work separately —
// two questions, and the two answers add up.
func TestPrintedAndGrantedRiotAskTwice(t *testing.T) {
	withRiotStatics(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	pushStaticSource(g, me.ID, testRiotGrantOracle)
	id := seedKeywordCreature(t, g, "Riot Goblin", KeywordRiot)
	castAndResolve(t, g, id)
	answerRiot(t, g, true)
	if g.Battlefield.Contains(id) {
		t.Fatal("the creature entered after one of its two riot questions")
	}
	answerRiot(t, g, false)
	if n, haste := riotState(t, g, id); n != 1 || !haste {
		t.Errorf("counter then haste: %d counters, haste %v; want 1, true", n, haste)
	}
}

// "Nontoken creatures YOU control": the grant is read against the
// player the permanent would enter under, so an opponent's Rhythm gives
// your creature nothing.
func TestAnOpponentsRiotGrantDoesNotApply(t *testing.T) {
	withRiotStatics(t)
	g := newActiveGame(t)
	pushStaticSource(g, g.Seats[1].ID, testRiotGrantOracle)
	id := seedKeywordCreature(t, g, "Vanilla Bear")
	castAndResolve(t, g, id)
	if len(g.PendingChoices) != 0 {
		t.Fatalf("an opponent's riot grant asked a question: %v", g.PendingChoices[0].Kind)
	}
	if !g.Battlefield.Contains(id) {
		t.Fatal("the creature did not enter")
	}
}

// CR 614.12 and CR 613.1f: a creature that would enter with no
// abilities has no riot to use, so it is not asked and gets neither.
func TestRiotUnderAnAbilityRemovalIsNotAsked(t *testing.T) {
	withRiotStatics(t)
	g := newActiveGame(t)
	pushStaticSource(g, g.Seats[1].ID, testRiotRemovalOracle)
	id := seedKeywordCreature(t, g, "Riot Goblin", KeywordRiot)
	castAndResolve(t, g, id)
	if len(g.PendingChoices) != 0 {
		t.Fatalf("a creature entering with no abilities was asked: %v", g.PendingChoices[0].Kind)
	}
	if n, haste := riotState(t, g, id); n != 0 || haste {
		t.Errorf("no riot: %d counters, haste %v", n, haste)
	}
}

// CR 400.7a: riot an effect gave the SPELL carries onto the permanent,
// so the look-ahead sees it (Domri, Chaos Bringer's shape).
func TestRiotGrantedToTheSpellIsAsked(t *testing.T) {
	g := newActiveGame(t)
	id := seedKeywordCreature(t, g, "Vanilla Bear")
	me := g.Seats[g.Turn.ActiveSeat]
	if err := g.CastSpell(me.ID, id, CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	g.WithWriteLock(func() {
		if !g.GrantKeywordsToSpellForEffect(uuid.Nil, id, []string{KeywordRiot}, "test: it gains riot") {
			t.Fatal("the grant to the spell was refused")
		}
	})
	resolveTop(t, g)
	answerRiot(t, g, true)
	if n, _ := riotState(t, g, id); n != 1 {
		t.Errorf("riot from the spell, counter answer: %d counters", n)
	}
}

// ADR 0109 §10 decision 3: an entry that cannot pause takes the counter,
// never neither.
func TestRiotOnAnEntryThatCannotPauseTakesTheCounter(t *testing.T) {
	g := newActiveGame(t)
	id := seedKeywordCreature(t, g, "Riot Goblin", KeywordRiot)
	me := g.Seats[g.Turn.ActiveSeat]
	var out *ReplacementEvent
	g.WithWriteLock(func() {
		ev := &ReplacementEvent{Kind: RepEventMove, CardID: id, OldZone: ZoneHand, NewZone: ZoneBattlefield, Actor: me.ID, mustSettleNow: true}
		var err error
		out, err = g.applyReplacementsLocked(ev)
		if err != nil {
			t.Fatalf("apply: %v", err)
		}
		g.clearReplacementEventLocked(ev.ID)
	})
	if len(g.PendingChoices) != 0 {
		t.Fatal("an entry that cannot pause asked riot's question")
	}
	if out == nil || out.EntersWithCounters[CounterPlusOne] != 1 || out.EntersWithHaste {
		t.Errorf("settled riot: %+v", out)
	}
}

// CR 702.98a: unleash is a "may" — yes, a counter and it can't block;
// no, neither.
func TestUnleashCounterAndItsBlockRule(t *testing.T) {
	g := newActiveGame(t)
	id := seedKeywordCreature(t, g, "Unleash Ogre", KeywordUnleash)
	castAndResolve(t, g, id)
	c := openChoice(t, g, PendingChoiceOptionalReplacement)
	if EntryKeywordOfReplacement(c.ReplacementEffectIDs[0]) != KeywordUnleash || c.Source != id {
		t.Errorf("the prompt does not say it is the entering card's unleash: %+v", c)
	}
	if err := g.ResolveOptionalReplacement(c.ID, c.Chooser, true); err != nil {
		t.Fatal(err)
	}
	card := layeredBattlefieldCard(t, g, id)
	if card.Counters[CounterPlusOne] != 1 || !Restricted(&card, CantBlock) {
		t.Fatalf("unleashed: %d counters, can't block %v", card.Counters[CounterPlusOne], Restricted(&card, CantBlock))
	}
	// "as long as it has a +1/+1 counter on it": without it, it blocks.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				delete(g.Battlefield.Cards[i].Counters, CounterPlusOne)
			}
		}
		g.layerVersion.Add(1)
	})
	if card := layeredBattlefieldCard(t, g, id); Restricted(&card, CantBlock) {
		t.Error("an unleash creature with no counter still can't block")
	}
}

func TestUnleashDeclinedEntersPlainAndBlocks(t *testing.T) {
	g := newActiveGame(t)
	id := seedKeywordCreature(t, g, "Unleash Ogre", KeywordUnleash)
	castAndResolve(t, g, id)
	c := openChoice(t, g, PendingChoiceOptionalReplacement)
	if err := g.ResolveOptionalReplacement(c.ID, c.Chooser, false); err != nil {
		t.Fatal(err)
	}
	card := layeredBattlefieldCard(t, g, id)
	if card.Counters[CounterPlusOne] != 0 || Restricted(&card, CantBlock) {
		t.Errorf("declined: %d counters, can't block %v", card.Counters[CounterPlusOne], Restricted(&card, CantBlock))
	}
}

// A +1/+1 counter from anywhere else stops an unleash creature blocking
// too: the rule reads the counter, not the entry.
func TestUnleashGrantedReadsAnyCounter(t *testing.T) {
	withRiotStatics(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := NewCard("Bear", me.ID)
	bear.TypeLine = "Creature — Bear"
	bear.Power, bear.Toughness = 2, 2
	bear.Counters = map[string]int{CounterPlusOne: 1}
	id := pushTypedTestCard(g, bear)
	if c := layeredBattlefieldCard(t, g, id); Restricted(&c, CantBlock) {
		t.Fatal("a creature without unleash can't block")
	}
	pushStaticSource(g, me.ID, testUnleashGrantOracle)
	if c := layeredBattlefieldCard(t, g, id); !Restricted(&c, CantBlock) {
		t.Error("a granted unleash with a +1/+1 counter can block")
	}
}

// The CR 616 ordering prompt names the keyword effects.
func TestEntryKeywordReplacementIDsNameTheirKeyword(t *testing.T) {
	g := newActiveGame(t)
	for _, tc := range []struct {
		kw string
		i  int
	}{{KeywordRiot, 0}, {KeywordRiot, 3}, {KeywordUnleash, 0}, {KeywordUnleash, 2}} {
		id := entryKeywordReplacementID(tc.kw, tc.i)
		if got := EntryKeywordOfReplacement(id); got != tc.kw {
			t.Errorf("%s #%d: decodes as %q", tc.kw, tc.i, got)
		}
		if id >= scopedReplacementIDBase || id < selfReplacementIDBase+2*MaxCatalogReplacementSlots {
			t.Errorf("%s #%d: id %d is outside its stride", tc.kw, tc.i, id)
		}
		if label, _ := g.ReplacementOptionMetaForEffect(id); label != entryKeywordLabel(tc.kw) {
			t.Errorf("%s #%d: label %q", tc.kw, tc.i, label)
		}
	}
	if EntryKeywordOfReplacement(selfReplacementIDBase) != "" {
		t.Error("a card's own slot 0 decodes as a keyword")
	}
}
