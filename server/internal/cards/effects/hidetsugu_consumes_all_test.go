package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hidetsugu_consumes_all_test.go — Hidetsugu Consumes All // Vessel of
// the All-Consuming (#5674), on the imported two-faced card.

func hidetsuguRow() cards.Card {
	row := transformRow(hidetsuguConsumesAllOracleID,
		"Hidetsugu Consumes All", "Enchantment — Saga", "{1}{B}{R}",
		"Vessel of the All-Consuming", "Enchantment Creature — Ogre Shaman",
		"3", "3", []string{"B", "R"})
	row.Keywords = []string{"Transform", "Trample"}
	return row
}

// vesselOnBattlefield puts the back face straight onto the battlefield,
// not summoning sick; chapter III's route is covered by the Saga test.
func vesselOnBattlefield(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	id := importToHand(hidetsuguRow(), me)
	var entered uuid.UUID
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == id {
				me.Hand.Cards[i].SetFace(1)
			}
		}
		var err error
		entered, err = g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{})
		if err != nil {
			t.Fatalf("put the back face onto the battlefield: %v", err)
		}
		findBattlefieldCardForTest(g, entered).SummonedThisTurn = false
		g.RecomputeLayersIfStaleLocked()
	})
	return entered
}

// vesselDeals has the Vessel deal n noncombat damage to a player and
// settles what that triggers.
func vesselDeals(t *testing.T, g *game.Game, vessel, player uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(vessel, player, n); err != nil {
			t.Fatalf("DealDamageToPlayerForEffect: %v", err)
		}
	})
	g.RunStateChecksForTest()
	b30SettleOrder(t, g)
}

func TestHidetsuguConsumesAllChapters(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]

	relic := apaPush(g, me.ID, me.ID, game.Card{Name: "One-Drop Relic", TypeLine: "Artifact", ManaCost: "{1}"})
	token := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1})
	bear := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	land := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	darksteel := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Darksteel Trinket", TypeLine: "Artifact", ManaCost: "{0}", Keywords: []string{"indestructible"}})
	mine := pushGraveyardCardForTest(me, "Old Spell")
	theirs := pushGraveyardCardForTest(opp, "Their Spell")

	saga := importAndCast(t, g, hidetsuguRow(), me)
	passPriorityAroundTable(t, g)

	// Chapter I: mana value 1 or less, nonland, every player's.
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want bool
	}{
		{"my mana value 1 artifact", relic, false},
		{"their token (mana value 0)", token, false},
		{"their mana value 2 creature", bear, true},
		{"their land", land, true},
		{"their indestructible mana value 0 artifact", darksteel, true},
		{"the Saga itself (mana value 3)", saga, true},
	} {
		if got := g.Battlefield.Contains(tc.id); got != tc.want {
			t.Errorf("after chapter I, %s on the battlefield = %v, want %v", tc.name, got, tc.want)
		}
	}
	if !inGraveyardOf(g, me.ID, relic) {
		t.Error("the relic was not destroyed into its owner's graveyard")
	}

	// Chapter II: every graveyard, mine included.
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if n := len(me.Graveyard.Cards) + len(opp.Graveyard.Cards); n != 0 {
		t.Errorf("after chapter II %d cards are left in the two graveyards", n)
	}
	for _, id := range []uuid.UUID{mine, theirs, relic} {
		if !g.Exile.Contains(id) {
			t.Errorf("card %s was not exiled by chapter II", id)
		}
	}

	// Chapter III: a new object, back face up, under my control.
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if battlefieldCardFor(g, saga) != nil {
		t.Fatal("the original object is still on the battlefield")
	}
	var back *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Vessel of the All-Consuming" {
			back = &g.Battlefield.Cards[i]
		}
	}
	if back == nil {
		t.Fatal("Vessel of the All-Consuming never reached the battlefield")
	}
	if back.ActiveFace != 1 || back.Controller != me.ID || back.Counters[game.CounterLore] != 0 {
		t.Errorf("came back face %d, controller %s, %d lore counters", back.ActiveFace, back.Controller, back.Counters[game.CounterLore])
	}
	if !back.IsCreature() || back.CurrentPower() != 3 || back.CurrentToughness() != 3 || !game.HasKeyword(back, "trample") {
		t.Errorf("came back as %q %d/%d trample=%v, want a 3/3 trampler",
			back.TypeLine, back.CurrentPower(), back.CurrentToughness(), game.HasKeyword(back, "trample"))
	}
	assertFaceInvariant(t, g)
}

// Trample damage split between a blocker and the player is dealt at the
// same time: one occurrence of "deals damage", one counter (CR 603.2c,
// 510.2).
func TestVesselGetsOneCounterForDamageDealtAtTheSameTime(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vessel := vesselOnBattlefield(t, g, me)
	blocker := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 1})
	life := opp.Life

	spm20Combat(t, g, vessel, opp, []uuid.UUID{blocker}, 1, 2, false)
	passPriorityAroundTable(t, g)

	if got := opp.Life - life; got != -2 {
		t.Fatalf("opponent life changed by %d, want -2 trampled over", got)
	}
	if got := counterOn(g, vessel, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1 for one combat damage step", got)
	}
	if opp.Eliminated {
		t.Error("2 damage is not 10")
	}
}

// Each separate time it deals damage is a counter, and the tenth point
// of damage to one player this turn, combat or not, loses them the game.
func TestVesselMakesAPlayerLoseAtTenDamageThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	vessel := vesselOnBattlefield(t, g, me)

	vesselDeals(t, g, vessel, opp.ID, 6)
	if got := counterOn(g, vessel, game.CounterPlusOne); got != 1 {
		t.Fatalf("+1/+1 counters after the first damage = %d, want 1", got)
	}
	if opp.Eliminated {
		t.Fatal("6 damage is not 10")
	}
	// The count is per player: 9 to someone else is nobody's 10.
	vesselDeals(t, g, vessel, third.ID, 9)
	if opp.Eliminated || third.Eliminated {
		t.Fatal("6 to one player and 9 to another is no player's 10")
	}
	// Combat damage counts with the noncombat damage before it.
	dealCombatDamageToPlayer(g, vessel, opp.ID, 4)
	g.RunStateChecksForTest()
	b30SettleOrder(t, g)
	if !opp.Eliminated {
		t.Error("6 noncombat and 4 combat damage to one player this turn: they lose the game")
	}
	if third.Eliminated || me.Eliminated {
		t.Error("only the damaged player loses")
	}
	if got := counterOn(g, vessel, game.CounterPlusOne); got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3 for three separate times it dealt damage", got)
	}
}

// Last turn's damage does not count toward this turn's 10.
func TestVesselCountsOnlyThisTurnsDamage(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	vessel := vesselOnBattlefield(t, g, me)

	vesselDeals(t, g, vessel, opp.ID, 9)
	// Into my main phase, then round the table to my next one.
	advanceTo(t, g, game.StepPrecombatMain)
	advanceToPrecombatMainOf(t, g, seat)
	vesselDeals(t, g, vessel, opp.ID, 9)
	if opp.Eliminated {
		t.Error("9 last turn and 9 this turn is not 10 this turn")
	}
}

// The trigger is the damaged player's loss even if the Vessel has left
// before it resolves: what that object dealt is still on the record
// (CR 608.2h).
func TestVesselTriggerResolvesAfterItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vessel := vesselOnBattlefield(t, g, me)

	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(vessel, opp.ID, 10); err != nil {
			t.Fatalf("DealDamageToPlayerForEffect: %v", err)
		}
	})
	g.RunStateChecksForTest()
	// Two triggers of one controller at once: they order them (CR 603.3b).
	if triggerOrderPrompt(g) == nil {
		t.Fatal("the counter and the loss triggered together and were not offered for ordering")
	}
	answerTriggerOrderInOfferedOrder(t, g)
	if n := triggersOnStackFrom(g, vessel); n != 2 {
		t.Fatalf("%d Vessel triggers on the stack, want the counter and the loss", n)
	}
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(vessel); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if !opp.Eliminated {
		t.Error("the Vessel dealt 10 to that player this turn; its leaving does not undo that")
	}
}

// The record is carried by a snapshot: 6 before a restore and 4 after
// are 10 this turn.
func TestVesselDamageRecordSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vessel := vesselOnBattlefield(t, g, me)
	vesselDeals(t, g, vessel, opp.ID, 6)

	restored := restoreRoundTrip(t, g, true)
	vesselDeals(t, restored, vessel, opp.ID, 3)
	if restored.Seats[1].Eliminated {
		t.Fatal("6 and 3 is 9")
	}
	vesselDeals(t, restored, vessel, opp.ID, 1)
	if !restored.Seats[1].Eliminated {
		t.Error("6 before the restore and 4 after it are 10 this turn: they lose the game")
	}
	if got := counterOn(restored, vessel, game.CounterPlusOne); got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3", got)
	}
}

// Both of the Vessel's one-sentence triggers are rows the engine can
// order without asking where it does not matter.
func TestVesselCounterTriggerDeclaresAFootprint(t *testing.T) {
	rows := game.CatalogTriggers(hidetsuguConsumesAllOracleID + "#1")
	if len(rows) != 2 {
		t.Fatalf("%d triggered rows, want 2", len(rows))
	}
	if len(rows[0].Footprint) == 0 {
		t.Error("the +1/+1 counter row has no footprint")
	}
}
