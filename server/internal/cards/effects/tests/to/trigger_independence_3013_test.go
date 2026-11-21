package teststo

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_independence_3013_test.go — ADR 0018's #3013 amendment,
// sub-PR 2: the owner's board. Marwyn, the Nurturer puts a +1/+1 counter
// on herself and Wolverine Riders gains life equal to the entering
// Elf's toughness; every order gives the same game, so "Ask only when
// the order matters" puts the batch on the stack with no prompt. It
// still asks when the order can show.

// marwynAndRiders seats the owner's pair for me and returns Marwyn.
func marwynAndRiders(g *game.Game, me uuid.UUID) uuid.UUID {
	marwyn := effectstest.PushCatalogPermanent(g, me, "Marwyn, the Nurturer", "Legendary Creature — Elf Druid", effectstest.B10MarwynOracle, false)
	effectstest.B21Push(g, me, "Wolverine Riders", "Creature — Elf Warrior", effectstest.B27WolverineRidersOracle, 4, 4)
	return marwyn
}

// enterTogether puts each card onto the battlefield as one event batch:
// every ETB is emitted before the next priority boundary harvests and
// drains, so their triggers are one CR 603.3b batch.
func enterTogether(g *game.Game, me uuid.UUID, ids ...uuid.UUID) {
	g.WithWriteLock(func() {
		for _, id := range ids {
			g.EmitEvent(game.Event{Kind: game.EventETB, CardID: id, Actor: me})
		}
	})
}

func TestTheRegistryDeclaresTheOwnersPairsFootprints(t *testing.T) {
	gainFromSubject := []game.FootprintStep{{Kind: game.FootprintGainLife, AmountFromSubject: true}}
	for _, c := range []struct {
		name   string
		oracle string
		row    int
		want   []game.FootprintStep
	}{
		{"Marwyn, the Nurturer", effectstest.B10MarwynOracle, 0, []game.FootprintStep{{Kind: game.FootprintCounterOnSource, N: 1, Counter: game.CounterPlusOne}}},
		{"Wolverine Riders' upkeep token", effectstest.B27WolverineRidersOracle, 0, nil},
		{"Wolverine Riders' lifegain", effectstest.B27WolverineRidersOracle, 1, gainFromSubject},
		{"Trostani, Selesnya's Voice", effectstest.TrostaniOracle, 0, gainFromSubject},
		{"Verdant Sun's Avatar", effectstest.B20VerdantSunsAvatarOracle, 0, gainFromSubject},
		{"Righteous Valkyrie", effectstest.B18RighteousValkyrieOracle, 0, gainFromSubject},
	} {
		rows := game.CatalogTriggers(c.oracle)
		if len(rows) <= c.row {
			t.Fatalf("%s: %d rows", c.name, len(rows))
		}
		got := rows[c.row].Footprint
		if c.want == nil {
			if len(got) != 1 || got[0].Kind != game.FootprintCreateToken {
				t.Errorf("%s: footprint %+v, want a token step", c.name, got)
			}
			continue
		}
		if len(got) != len(c.want) {
			t.Errorf("%s: footprint %+v, want %+v", c.name, got, c.want)
			continue
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: footprint %+v, want %+v", c.name, got, c.want)
		}
	}
}

// The owner's report: an Elf cast with Marwyn and Wolverine Riders out
// goes on with no prompt, Marwyn grows and the life is the Elf's
// toughness.
func TestMarwynAndRidersOnACastElfDoNotAsk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	marwyn := marwynAndRiders(g, me.ID)
	life := me.Life

	effectstest.B20CastCreature(t, g, me, "Llanowar Elves", "Creature — Elf Druid", "", 1, 2)
	if ch := effectstest.SettleUntilOrderPrompt(t, g); ch != nil {
		t.Fatalf("Marwyn and Wolverine Riders asked for an order: %+v", ch)
	}
	if got := effectstest.CounterCount(g, marwyn, game.CounterPlusOne); got != 1 {
		t.Errorf("Marwyn has %d counters, want 1", got)
	}
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2 (the Elf's toughness)", life, me.Life)
	}
}

// Wolverine Riders' own Elf, made at the beginning of every upkeep at
// the table, makes the same batch.
func TestMarwynAndRidersOnTheUpkeepElfDoNotAsk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	marwyn := marwynAndRiders(g, me.ID)
	life := me.Life

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if ch := effectstest.SettleUntilOrderPrompt(t, g); ch != nil {
		t.Fatalf("the upkeep Elf's batch asked for an order: %+v", ch)
	}
	if got := effectstest.CounterCount(g, marwyn, game.CounterPlusOne); got != 1 {
		t.Errorf("Marwyn has %d counters, want 1", got)
	}
	if me.Life != life+1 {
		t.Errorf("life %d → %d, want +1 (the 1/1 token)", life, me.Life)
	}
}

// The batch still asks when its order can show.
func TestMarwynAndRidersStillAskWhenTheOrderCanShow(t *testing.T) {
	t.Run("a lifegain payoff watches the gain", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		marwynAndRiders(g, me.ID)
		effectstest.PushCatalogPermanent(g, me.ID, "Ajani's Pridemate", "Creature — Cat Soldier", effectstest.B24AjanisPridemateOracle, false)
		before := promptsByReason(t, game.TriggerOrderReasonReactor)
		effectstest.B20CastCreature(t, g, me, "Llanowar Elves", "Creature — Elf Druid", "", 1, 1)
		if ch := effectstest.SettleUntilOrderPrompt(t, g); ch == nil {
			t.Fatal("a lifegain payoff on the board must keep the prompt")
		}
		if got := promptsByReason(t, game.TriggerOrderReasonReactor) - before; got != 1 {
			t.Errorf("counted %v under reactor, want 1", got)
		}
	})
	t.Run("a lord's static could read what the counter changes", func(t *testing.T) {
		// Elvish Archdruid's "other Elves you control get +1/+1" is a
		// layer-7 static. Its output cannot change here, but the check
		// cannot see what a static reads, so a characteristic read
		// beside a write asks.
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		marwynAndRiders(g, me.ID)
		effectstest.PushCatalogPermanent(g, me.ID, "Elvish Archdruid", "Creature — Elf Druid", effectstest.ElvishArchdruidOracle, false)
		before := promptsByReason(t, game.TriggerOrderReasonCharacteristic)
		effectstest.B20CastCreature(t, g, me, "Llanowar Elves", "Creature — Elf Druid", "", 1, 1)
		if ch := effectstest.SettleUntilOrderPrompt(t, g); ch == nil {
			t.Fatal("a P/T static on the board must keep the prompt")
		}
		if got := promptsByReason(t, game.TriggerOrderReasonCharacteristic) - before; got != 1 {
			t.Errorf("counted %v under characteristic, want 1", got)
		}
	})
	t.Run("Marwyn enters with another Elf", func(t *testing.T) {
		// Riders triggers for both Elves and Marwyn for the other one.
		// Riders' trigger for Marwyn gains 1 before her counter and 2
		// after it.
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.B21Push(g, me.ID, "Wolverine Riders", "Creature — Elf Warrior", effectstest.B27WolverineRidersOracle, 4, 4)
		effectstest.ToMain(t, g)
		marwyn := effectstest.PushCatalogPermanent(g, me.ID, "Marwyn, the Nurturer", "Legendary Creature — Elf Druid", effectstest.B10MarwynOracle, false)
		elf := effectstest.B21Push(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", "", 1, 1)
		before := promptsByReason(t, game.TriggerOrderReasonObjectTie)
		enterTogether(g, me.ID, marwyn, elf)
		ch := effectstest.SettleUntilOrderPrompt(t, g)
		if ch == nil || len(ch.TriggerOrderIDs) != 3 {
			t.Fatalf("Marwyn entering beside another Elf must ask for all three, got %+v", ch)
		}
		if got := promptsByReason(t, game.TriggerOrderReasonObjectTie) - before; got != 1 {
			t.Errorf("counted %v under object_tie, want 1", got)
		}
	})
}

// Righteous Valkyrie's and Trostani's gains for one entering Angel can
// cross Valkyrie's +7 line between them, and her anthem changes the
// Angel's toughness for the second: the static reads a life total, so
// even a batch that only gains life asks. (Two Valkyrie gains for two
// creatures are one source and one label; the sub-PR 8 skip used to
// pass them unread and asks the same question since #3021, see
// trigger_independence_3021_test.go.)
func TestTwoSubjectGainsAskBesideALifeKeyedAnthem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Righteous Valkyrie", "Creature — Angel Cleric", effectstest.B18RighteousValkyrieOracle, 2, 4)
	effectstest.B12Push(g, me.ID, "Trostani, Selesnya's Voice", "Legendary Creature — Dryad", effectstest.TrostaniOracle, 2, 5)
	effectstest.ToMain(t, g)
	angel := effectstest.B21Push(g, me.ID, "An Angel", "Creature — Angel", "", 2, 2)
	before := promptsByReason(t, game.TriggerOrderReasonCharacteristic)
	enterTogether(g, me.ID, angel)
	if ch := effectstest.SettleUntilOrderPrompt(t, g); ch == nil {
		t.Fatal("two gains beside a life-keyed anthem must ask")
	}
	if got := promptsByReason(t, game.TriggerOrderReasonCharacteristic) - before; got != 1 {
		t.Errorf("counted %v under characteristic, want 1", got)
	}
}

// Toughness is read at resolution, and as the creature last existed on
// the battlefield once it has left (CR 608.2h).
func TestRidersReadsTheElfAtResolutionAndAfterItLeaves(t *testing.T) {
	t.Run("grown in response", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.B21Push(g, me.ID, "Wolverine Riders", "Creature — Elf Warrior", effectstest.B27WolverineRidersOracle, 4, 4)
		effectstest.ToMain(t, g)
		elf := effectstest.B21Push(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", "", 1, 1)
		life := me.Life
		enterTogether(g, me.ID, elf)
		g.RunStateChecksForTest()
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(elf, game.CounterPlusOne, 2) })
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != life+3 {
			t.Errorf("life %d → %d, want +3 (1 toughness and two counters)", life, me.Life)
		}
	})
	t.Run("gone before resolution", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.B21Push(g, me.ID, "Wolverine Riders", "Creature — Elf Warrior", effectstest.B27WolverineRidersOracle, 4, 4)
		effectstest.ToMain(t, g)
		elf := effectstest.B21Push(g, me.ID, "Elvish Visionary", "Creature — Elf Shaman", "", 1, 3)
		life := me.Life
		enterTogether(g, me.ID, elf)
		g.RunStateChecksForTest()
		g.WithWriteLock(func() {
			_ = g.AddCounterForEffect(elf, game.CounterPlusOne, 1)
			_ = g.DestroyPermanentForEffect(elf)
		})
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != life+4 {
			t.Errorf("life %d → %d, want +4 (its last-known toughness, counter included)", life, me.Life)
		}
	})
}

// An item a binary from before #3013 queued carries the trigger-time
// toughness in Params.Amount. It still resolves. (The check cannot read
// Params, so triggerAccessLocked refuses the item and its batch keeps
// asking.)
func TestAnOlderRidersItemStillResolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	riders := effectstest.B21Push(g, me.ID, "Wolverine Riders", "Creature — Elf Warrior", effectstest.B27WolverineRidersOracle, 4, 4)
	effectstest.ToMain(t, g)
	var source game.Card
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == riders {
			source = c
		}
	}
	item := game.NewTriggeredItem(&source, "Wolverine Riders — gain life equal to the Elf's toughness")
	item.Params.Amount = 3
	life := me.Life
	g.WithWriteLock(func() {
		if err := (effects.GainLifeEqualToSubject{Stat: effects.SubjectToughness}).Apply(effects.NewContext(g, item)); err != nil {
			t.Fatal(err)
		}
	})
	if me.Life != life+3 {
		t.Errorf("an item with no subject record gains its Params.Amount: life %d → %d, want +3", life, me.Life)
	}
}
