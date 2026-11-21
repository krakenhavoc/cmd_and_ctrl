package teststo

import (
	"reflect"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func pushMentor(g *game.Game, me uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Monastery Mentor", OracleID: effectstest.OracleMonasteryMentor,
		TypeLine: "Creature — Human Monk", Power: 2, Toughness: 2,
		Owner: me, Controller: me,
	})
}

func pushSokka(g *game.Game, me uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sokka, Tenacious Tactician", OracleID: effectstest.OracleSokka,
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
		{"Monastery Mentor", effectstest.OracleMonasteryMentor},
		{"Sokka, Tenacious Tactician", effectstest.OracleSokka},
	} {
		rows := game.CatalogTriggers(c.oracle)
		if len(rows) != 1 || len(rows[0].Footprint) != 1 || rows[0].Footprint[0].Kind != game.FootprintCreateToken {
			t.Errorf("%s's token row: footprint %+v, want one token step", c.name, rows)
		}
	}
	// A token step carries its template's keywords, for what their
	// keyword triggers watch.
	rows := game.CatalogTriggers(effectstest.EvolveTokenProbe)
	if len(rows) != 1 || len(rows[0].Footprint) != 1 ||
		!slices.Equal(rows[0].Footprint[0].TokenKeywords, []string{game.KeywordEvolve}) {
		t.Errorf("the evolve probe's footprint %+v, want a token step with evolve", rows)
	}
}

// Printed keywords are the object's own characteristics, so the registry
// makes no static of them (#4236): a card with PrintedKeywords and nothing
// else has no statics, and one with a static of its own keeps just that.
func TestPrintedKeywordsMakeNoStatic(t *testing.T) {
	const key = "test-4236-printed-keywords-no-static"
	own := game.StaticAbility{Layer: game.Layer6Ability, AppliesTo: effects.SelfOnly, Label: "probe"}
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: key, Name: "Printed Keywords Probe",
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{own},
	})
	d := effects.LookupDef(key)
	if d == nil || len(d.Static) != 1 || d.Static[0].Label != "probe" {
		t.Fatalf("definition statics %+v, want only the card's own", d)
	}
	if got := game.CatalogPrintedKeywords(key); len(got) != 1 || got[0] != "flying" {
		t.Errorf("printed keywords %v, want flying", got)
	}
	if got := game.CatalogLookup(effectstest.OracleMonasteryMentor); got == nil || len(got.Static) != 0 {
		t.Error("Monastery Mentor has no static: its prowess is a printed keyword")
	}
}

// Monastery Mentor: its prowess and its Monk trigger go on with no
// prompt. On the next spell the Monk's prowess joins them, and the
// three still go on with no prompt.
func TestAMentorBoardDoesNotAsk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mentor := pushMentor(g, me.ID)

	effectstest.CastNoncreature(t, g)
	effectstest.SettleWithoutPrompts(t, g)
	if got := effectstest.CountOnBattlefield(g, "Monk", me.ID); got != 1 {
		t.Fatalf("%d Monks after one spell, want 1", got)
	}
	if p := effectstest.EffectivePower(t, g, mentor); p != 3 {
		t.Errorf("Mentor: power %d after one spell, want 3", p)
	}
	first := effectstest.ProwessFirstNamed(t, g, "Monk", me.ID)

	effectstest.CastNoncreature(t, g)
	effectstest.SettleWithoutPrompts(t, g)
	if got := effectstest.CountOnBattlefield(g, "Monk", me.ID); got != 2 {
		t.Fatalf("%d Monks after two spells, want 2", got)
	}
	if p := effectstest.EffectivePower(t, g, mentor); p != 4 {
		t.Errorf("Mentor: power %d after two spells, want 4", p)
	}
	if p := effectstest.EffectivePower(t, g, first); p != 2 {
		t.Errorf("the first Monk: power %d after the second spell, want 2", p)
	}
}

// Sokka, Tenacious Tactician: his prowess, another Ally's granted
// prowess and his Ally trigger go on with no prompt, and so does the
// next spell's batch with the Ally token's granted prowess in it.
// Sokka's grants are statics, so pumping him could change what they do;
// no item in this batch reads anything they could change.
func TestASokkaBoardDoesNotAsk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	scout := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ally Scout", TypeLine: "Creature — Human Scout Ally",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	sokka := pushSokka(g, me.ID)

	effectstest.CastNoncreature(t, g)
	effectstest.SettleWithoutPrompts(t, g)
	if p, tg := effectstest.EffectivePower(t, g, sokka), effectstest.EffectiveToughness(t, g, sokka); p != 4 || tg != 4 {
		t.Errorf("Sokka: %d/%d, want 4/4", p, tg)
	}
	if p, tg := effectstest.EffectivePower(t, g, scout), effectstest.EffectiveToughness(t, g, scout); p != 2 || tg != 2 {
		t.Errorf("the Ally Scout: %d/%d, want 2/2 (granted prowess)", p, tg)
	}
	ally := effectstest.ProwessFirstNamed(t, g, "Ally", me.ID)

	effectstest.CastNoncreature(t, g)
	effectstest.SettleWithoutPrompts(t, g)
	if got := effectstest.CountOnBattlefield(g, "Ally", me.ID); got != 2 {
		t.Fatalf("%d Ally tokens after two spells, want 2", got)
	}
	if p := effectstest.EffectivePower(t, g, ally); p != 2 {
		t.Errorf("the first Ally token: power %d after the second spell, want 2", p)
	}
	if p := effectstest.EffectivePower(t, g, sokka); p != 5 {
		t.Errorf("Sokka: power %d after two spells, want 5", p)
	}
}

// Ty Lee, Chi Blocker's "when this enters" watches entering, and the
// check does not ask whether a watcher would fire (owner decision 3 of
// the #3013 amendment), so beside Sokka's Ally token the batch asks.
func TestTyLeeBesideSokkaStillAsks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ty Lee, Chi Blocker", OracleID: effectstest.OracleTyLee,
		TypeLine: "Legendary Creature — Human Performer Ally", Power: 2, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	pushSokka(g, me.ID)
	before := promptsByReason(t, game.TriggerOrderReasonReactor)
	effectstest.CastNoncreature(t, g)
	if ch := effectstest.SettleUntilOrderPrompt(t, g); ch == nil {
		t.Fatal("Ty Lee's ETB watcher beside Sokka's token must keep the prompt")
	}
	if got := promptsByReason(t, game.TriggerOrderReasonReactor) - before; got != 1 {
		t.Errorf("counted %v under reactor, want 1", got)
	}
	effectstest.SettleProwess(t, g)
}

// The batch still asks when its order can show.
func TestAProwessBoardStillAsksWhenTheOrderCanShow(t *testing.T) {
	t.Run("something watches the token enter", func(t *testing.T) {
		// Soul Warden gains a life for the Monk. Where its trigger lands
		// among the rest depends on where the Monk's trigger is.
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		pushMentor(g, me.ID)
		effectstest.PushCatalogPermanent(g, me.ID, "Soul Warden", "Creature — Human Cleric", effectstest.B04SoulWardenOracle, false)
		before := promptsByReason(t, game.TriggerOrderReasonReactor)
		effectstest.CastNoncreature(t, g)
		if ch := effectstest.SettleUntilOrderPrompt(t, g); ch == nil {
			t.Fatal("a Mentor board with an ETB watcher must keep the prompt")
		}
		if got := promptsByReason(t, game.TriggerOrderReasonReactor) - before; got != 1 {
			t.Errorf("counted %v under reactor, want 1", got)
		}
		effectstest.SettleProwess(t, g)
	})
	t.Run("Caldera Pyremaw reads its own power", func(t *testing.T) {
		// Granted prowess pumps it, and its own trigger deals damage
		// equal to its power. Pyremaw's row is a closure the check
		// cannot read; described, it would tie on the source.
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.PushProwessCreature(g, me.ID, "Monk", 1, 1)
		effectstest.B43Catalog(g, me.ID, "Caldera Pyremaw", "Creature — Dragon", effectstest.B43CalderaPyremawOracle, 3, 3)
		before := promptsByReason(t, game.TriggerOrderReasonUndescribed)
		effectstest.CastCatalogSpell(t, g, "Opt", "Instant", "", nil)
		effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
		if ch := effectstest.TriggerOrderPrompt(g); ch == nil {
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
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		pushSokka(g, me.ID)
		effectstest.PushCatalogPermanent(g, me.ID, "Firebrand Archer", "Creature — Human Archer", effectstest.FirebrandArcherOracle, false)
		before := promptsByReason(t, game.TriggerOrderReasonRules)
		effectstest.CastNoncreature(t, g)
		if ch := effectstest.SettleUntilOrderPrompt(t, g); ch == nil {
			t.Fatal("a pump on Sokka beside a damage trigger must keep the prompt")
		}
		if got := promptsByReason(t, game.TriggerOrderReasonRules) - before; got != 1 {
			t.Errorf("counted %v under rules, want 1", got)
		}
		effectstest.SettleProwess(t, g)

		g = effectstest.NewCatalogGame(t)
		me = g.Seats[g.Turn.ActiveSeat]
		pushMentor(g, me.ID)
		effectstest.PushCatalogPermanent(g, me.ID, "Firebrand Archer", "Creature — Human Archer", effectstest.FirebrandArcherOracle, false)
		effectstest.CastNoncreature(t, g)
		effectstest.SettleWithoutPrompts(t, g)
	})
	t.Run("always ask", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		if err := g.SetTriggerOrderPreference(me.ID, game.TriggerOrderAlways); err != nil {
			t.Fatal(err)
		}
		pushMentor(g, me.ID)
		effectstest.CastNoncreature(t, g)
		if ch := effectstest.SettleUntilOrderPrompt(t, g); ch == nil {
			t.Fatal("always ask must ask on a Mentor board")
		}
		effectstest.SettleProwess(t, g)
	})
}

// A token whose own keyword trigger watches what the batch does is
// refused: an evolve token made partway through sees the creatures that
// enter after it and not those before. A Monk token's prowess watches a
// cast, which no item in a batch does, so it passes.
func TestATokenWithAWatchingKeywordTriggerAsks(t *testing.T) {
	t.Run("evolve", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.PushCatalogPermanent(g, me.ID, "Evolve Token Probe", "Enchantment", effectstest.EvolveTokenProbe, false)
		effectstest.PushCatalogPermanent(g, me.ID, "Independence Token Probe", "Enchantment", effectstest.IndependenceTokenProbe, false)
		before := promptsByReason(t, game.TriggerOrderReasonReactor)
		effectstest.CastNoncreature(t, g)
		if ch := effectstest.SettleUntilOrderPrompt(t, g); ch == nil {
			t.Fatal("an evolve token beside another token maker must keep the prompt")
		}
		if got := promptsByReason(t, game.TriggerOrderReasonReactor) - before; got != 1 {
			t.Errorf("counted %v under reactor, want 1", got)
		}
	})
	t.Run("prowess", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.PushCatalogPermanent(g, me.ID, "Monk Token Probe", "Enchantment", effectstest.MonkTokenProbe, false)
		effectstest.PushCatalogPermanent(g, me.ID, "Independence Token Probe", "Enchantment", effectstest.IndependenceTokenProbe, false)
		effectstest.CastNoncreature(t, g)
		effectstest.SettleWithoutPrompts(t, g)
	})
}
