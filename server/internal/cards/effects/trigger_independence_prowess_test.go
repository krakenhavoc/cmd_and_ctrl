package effects

import (
	"reflect"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_independence_prowess_test.go — ADR 0018's #3013 amendment,
// sub-PR 3, closing #1530 item 1. Prowess has a declared footprint (it
// reads and pumps its own source and emits nothing), so a Monastery
// Mentor board and a Sokka, Tenacious Tactician board, each prowess
// plus a token trigger, go on the stack with no prompt under "Ask only
// when the order matters". The batch still asks when the order can
// show: something watches the token enter, a trigger the check cannot
// read, or a token whose own keyword trigger watches what the batch
// does.

const (
	evolveTokenProbe = "test-3013-evolve-token-probe"
	monkTokenProbe   = "test-3013-monk-token-probe"
)

func init() {
	// "Whenever you cast a noncreature spell, create a 1/1 token with
	// evolve." No catalogued token has evolve; the shape is what
	// tokenTemplateInert used to miss.
	Register(Spec{
		OracleID: evolveTokenProbe,
		Name:     "Evolve Token Probe",
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Evolve Token Probe — create a 1/1 with evolve",
				Do(CreateToken{Template: game.Card{
					Name: "Evolver", TypeLine: "Token Creature — Ooze", Power: 1, Toughness: 1,
					Keywords: []string{game.KeywordEvolve},
				}, N: 1})),
		},
	})
	// The same with Monastery Mentor's own Monk, whose prowess watches a
	// cast: nothing in a batch casts a spell.
	Register(Spec{
		OracleID: monkTokenProbe,
		Name:     "Monk Token Probe",
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Monk Token Probe — create a 1/1 Monk with prowess",
				Do(CreateToken{Template: TokenCard("1/1 white Monk with prowess"), N: 1})),
		},
	})
}

func pushMentor(g *game.Game, me uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Monastery Mentor", OracleID: oracleMonasteryMentor,
		TypeLine: "Creature — Human Monk", Power: 2, Toughness: 2,
		Owner: me, Controller: me,
	})
}

func pushSokka(g *game.Game, me uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sokka, Tenacious Tactician", OracleID: oracleSokka,
		TypeLine: "Legendary Creature — Human Warrior Ally", Power: 3, Toughness: 3,
		Owner: me, Controller: me,
	})
}

func TestProwessHasAKeyedFootprint(t *testing.T) {
	want := []game.FootprintStep{{Kind: game.FootprintPumpSource, N: 1}}
	if got := game.KeyedBodyFootprint("prowess/pump"); !reflect.DeepEqual(got, want) {
		t.Fatalf("prowess/pump footprint %+v, want %+v", got, want)
	}
	// The token rows the batches pair it with already had one.
	for _, c := range []struct{ name, oracle string }{
		{"Monastery Mentor", oracleMonasteryMentor},
		{"Sokka, Tenacious Tactician", oracleSokka},
	} {
		rows := game.CatalogTriggers(c.oracle)
		if len(rows) != 1 || len(rows[0].Footprint) != 1 || rows[0].Footprint[0].Kind != game.FootprintCreateToken {
			t.Errorf("%s's token row: footprint %+v, want one token step", c.name, rows)
		}
	}
	// A token step carries its template's keywords, for what their
	// keyword triggers watch.
	rows := game.CatalogTriggers(evolveTokenProbe)
	if len(rows) != 1 || len(rows[0].Footprint) != 1 ||
		!slices.Equal(rows[0].Footprint[0].TokenKeywords, []string{game.KeywordEvolve}) {
		t.Errorf("the evolve probe's footprint %+v, want a token step with evolve", rows)
	}
}

// The printed-keyword row is the registry's: it marks the row it makes
// from PrintedKeywords and clears the mark on any row a card file
// wrote, so a hand-written static cannot pass for it.
func TestOnlyTheRegistryMarksThePrintedKeywordRow(t *testing.T) {
	const key = "test-3013-printed-keyword-row"
	registerForTest(t, Spec{
		OracleID: key, Name: "Printed Keyword Row Probe",
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability, AppliesTo: selfOnly, PrintedKeywordRow: true,
			Apply: func(*game.Characteristic, *game.Card, *game.Game, *game.Card) {},
		}},
	})
	d := lookupDef(key)
	if d == nil || len(d.Static) != 2 {
		t.Fatalf("definition statics %+v, want the card's row and the printed keywords", d)
	}
	if d.Static[0].PrintedKeywordRow || !d.Static[1].PrintedKeywordRow {
		t.Errorf("marks %v / %v, want false for the card's row and true for the registry's",
			d.Static[0].PrintedKeywordRow, d.Static[1].PrintedKeywordRow)
	}
	if got := game.CatalogLookup(oracleMonasteryMentor); got == nil || len(got.Static) != 1 || !got.Static[0].PrintedKeywordRow {
		t.Error("Monastery Mentor's only static should be its printed keywords")
	}
}

// Monastery Mentor: its prowess and its Monk trigger go on with no
// prompt. On the next spell the Monk's prowess joins them, and the
// three still go on with no prompt.
func TestAMentorBoardDoesNotAsk(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mentor := pushMentor(g, me.ID)

	castNoncreature(t, g)
	settleWithoutPrompts(t, g)
	if got := countOnBattlefield(g, "Monk", me.ID); got != 1 {
		t.Fatalf("%d Monks after one spell, want 1", got)
	}
	if p := effectivePower(t, g, mentor); p != 3 {
		t.Errorf("Mentor: power %d after one spell, want 3", p)
	}
	first := prowessFirstNamed(t, g, "Monk", me.ID)

	castNoncreature(t, g)
	settleWithoutPrompts(t, g)
	if got := countOnBattlefield(g, "Monk", me.ID); got != 2 {
		t.Fatalf("%d Monks after two spells, want 2", got)
	}
	if p := effectivePower(t, g, mentor); p != 4 {
		t.Errorf("Mentor: power %d after two spells, want 4", p)
	}
	if p := effectivePower(t, g, first); p != 2 {
		t.Errorf("the first Monk: power %d after the second spell, want 2", p)
	}
}

// Sokka, Tenacious Tactician: his prowess, another Ally's granted
// prowess and his Ally trigger go on with no prompt, and so does the
// next spell's batch with the Ally token's granted prowess in it.
// Sokka's grants are statics, so pumping him could change what they do;
// no item in this batch reads anything they could change.
func TestASokkaBoardDoesNotAsk(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	scout := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ally Scout", TypeLine: "Creature — Human Scout Ally",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	sokka := pushSokka(g, me.ID)

	castNoncreature(t, g)
	settleWithoutPrompts(t, g)
	if p, tg := effectivePower(t, g, sokka), effectiveToughness(t, g, sokka); p != 4 || tg != 4 {
		t.Errorf("Sokka: %d/%d, want 4/4", p, tg)
	}
	if p, tg := effectivePower(t, g, scout), effectiveToughness(t, g, scout); p != 2 || tg != 2 {
		t.Errorf("the Ally Scout: %d/%d, want 2/2 (granted prowess)", p, tg)
	}
	ally := prowessFirstNamed(t, g, "Ally", me.ID)

	castNoncreature(t, g)
	settleWithoutPrompts(t, g)
	if got := countOnBattlefield(g, "Ally", me.ID); got != 2 {
		t.Fatalf("%d Ally tokens after two spells, want 2", got)
	}
	if p := effectivePower(t, g, ally); p != 2 {
		t.Errorf("the first Ally token: power %d after the second spell, want 2", p)
	}
	if p := effectivePower(t, g, sokka); p != 5 {
		t.Errorf("Sokka: power %d after two spells, want 5", p)
	}
}

// Ty Lee, Chi Blocker's "when this enters" watches entering, and the
// check does not ask whether a watcher would fire (owner decision 3 of
// the #3013 amendment), so beside Sokka's Ally token the batch asks.
func TestTyLeeBesideSokkaStillAsks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ty Lee, Chi Blocker", OracleID: oracleTyLee,
		TypeLine: "Legendary Creature — Human Performer Ally", Power: 2, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	pushSokka(g, me.ID)
	before := promptsByReason(t, game.TriggerOrderReasonReactor)
	castNoncreature(t, g)
	if ch := settleUntilOrderPrompt(t, g); ch == nil {
		t.Fatal("Ty Lee's ETB watcher beside Sokka's token must keep the prompt")
	}
	if got := promptsByReason(t, game.TriggerOrderReasonReactor) - before; got != 1 {
		t.Errorf("counted %v under reactor, want 1", got)
	}
	settleProwess(t, g)
}

// The batch still asks when its order can show.
func TestAProwessBoardStillAsksWhenTheOrderCanShow(t *testing.T) {
	t.Run("something watches the token enter", func(t *testing.T) {
		// Soul Warden gains a life for the Monk. Where its trigger lands
		// among the rest depends on where the Monk's trigger is.
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		pushMentor(g, me.ID)
		pushCatalogPermanent(g, me.ID, "Soul Warden", "Creature — Human Cleric", b04SoulWardenOracle, false)
		before := promptsByReason(t, game.TriggerOrderReasonReactor)
		castNoncreature(t, g)
		if ch := settleUntilOrderPrompt(t, g); ch == nil {
			t.Fatal("a Mentor board with an ETB watcher must keep the prompt")
		}
		if got := promptsByReason(t, game.TriggerOrderReasonReactor) - before; got != 1 {
			t.Errorf("counted %v under reactor, want 1", got)
		}
		settleProwess(t, g)
	})
	t.Run("Caldera Pyremaw reads its own power", func(t *testing.T) {
		// Granted prowess pumps it, and its own trigger deals damage
		// equal to its power. Pyremaw's row is a closure the check
		// cannot read; described, it would tie on the source.
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		pushProwessCreature(g, me.ID, "Monk", 1, 1)
		b43Catalog(g, me.ID, "Caldera Pyremaw", "Creature — Dragon", b43CalderaPyremawOracle, 3, 3)
		before := promptsByReason(t, game.TriggerOrderReasonUndescribed)
		castCatalogSpell(t, g, "Opt", "Instant", "", nil)
		b16PickPlayer(t, g, me.ID, opp.ID)
		if ch := triggerOrderPrompt(g); ch == nil {
			t.Fatal("prowess beside Caldera Pyremaw must keep the prompt")
		}
		if got := promptsByReason(t, game.TriggerOrderReasonUndescribed) - before; got != 1 {
			t.Errorf("counted %v under undescribed, want 1", got)
		}
	})
	t.Run("a pumped permanent's statics beside a damage source", func(t *testing.T) {
		// Sokka's statics are closures. One could read his power and
		// grant lifelink, so a pump on him asks beside an item that
		// reads a damage source's abilities. Monastery Mentor has no
		// static, and beside the same damage his board does not ask.
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		pushSokka(g, me.ID)
		pushCatalogPermanent(g, me.ID, "Firebrand Archer", "Creature — Human Archer", firebrandArcherOracle, false)
		before := promptsByReason(t, game.TriggerOrderReasonRules)
		castNoncreature(t, g)
		if ch := settleUntilOrderPrompt(t, g); ch == nil {
			t.Fatal("a pump on Sokka beside a damage trigger must keep the prompt")
		}
		if got := promptsByReason(t, game.TriggerOrderReasonRules) - before; got != 1 {
			t.Errorf("counted %v under rules, want 1", got)
		}
		settleProwess(t, g)

		g = newCatalogGame(t)
		me = g.Seats[g.Turn.ActiveSeat]
		pushMentor(g, me.ID)
		pushCatalogPermanent(g, me.ID, "Firebrand Archer", "Creature — Human Archer", firebrandArcherOracle, false)
		castNoncreature(t, g)
		settleWithoutPrompts(t, g)
	})
	t.Run("always ask", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		if err := g.SetTriggerOrderPreference(me.ID, game.TriggerOrderAlways); err != nil {
			t.Fatal(err)
		}
		pushMentor(g, me.ID)
		castNoncreature(t, g)
		if ch := settleUntilOrderPrompt(t, g); ch == nil {
			t.Fatal("always ask must ask on a Mentor board")
		}
		settleProwess(t, g)
	})
}

// A token whose own keyword trigger watches what the batch does is
// refused: an evolve token made partway through sees the creatures that
// enter after it and not those before. A Monk token's prowess watches a
// cast, which no item in a batch does, so it passes.
func TestATokenWithAWatchingKeywordTriggerAsks(t *testing.T) {
	t.Run("evolve", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		pushCatalogPermanent(g, me.ID, "Evolve Token Probe", "Enchantment", evolveTokenProbe, false)
		pushCatalogPermanent(g, me.ID, "Independence Token Probe", "Enchantment", independenceTokenProbe, false)
		before := promptsByReason(t, game.TriggerOrderReasonReactor)
		castNoncreature(t, g)
		if ch := settleUntilOrderPrompt(t, g); ch == nil {
			t.Fatal("an evolve token beside another token maker must keep the prompt")
		}
		if got := promptsByReason(t, game.TriggerOrderReasonReactor) - before; got != 1 {
			t.Errorf("counted %v under reactor, want 1", got)
		}
	})
	t.Run("prowess", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		pushCatalogPermanent(g, me.ID, "Monk Token Probe", "Enchantment", monkTokenProbe, false)
		pushCatalogPermanent(g, me.ID, "Independence Token Probe", "Enchantment", independenceTokenProbe, false)
		castNoncreature(t, g)
		settleWithoutPrompts(t, g)
	})
}
