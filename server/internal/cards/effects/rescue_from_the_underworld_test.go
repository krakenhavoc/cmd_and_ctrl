package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rescue_from_the_underworld_test.go — #2099: the delayed return of the
// card a cost sacrificed, with a targeted card, at the caster's next
// upkeep (ADR 0026 amendment 2026-10-16).

const rescueFromTheUnderworldOracle = "7217ea53-3685-48d6-9df1-58a551f8c979"

// castRescue casts Rescue from the Underworld from the active seat
// (seat 0) targeting `target` in a graveyard and sacrificing `victim`.
func castRescue(t *testing.T, g *game.Game, target, victim uuid.UUID) uuid.UUID {
	t.Helper()
	return paCast(t, g, "Rescue from the Underworld", "Instant", "{4}{B}", rescueFromTheUnderworldOracle, game.CastSpellParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		SacrificeIDs: []uuid.UUID{victim},
	})
}

// rescueReachYourNextUpkeep walks from the caster's turn through the
// next seat's upkeep, where the trigger must still be waiting, to the
// caster's own next upkeep, and resolves what is on the stack there.
func rescueReachYourNextUpkeep(t *testing.T, g *game.Game) {
	t.Helper()
	advanceToUpkeepOf(t, g, 1)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("an opponent's upkeep fired the trigger: %d still queued, want 1", len(g.DelayedTriggers))
	}
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
}

func TestRescueFromTheUnderworldReturnsBothAtYourNextUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	target := pushGraveyardCardTyped(me, "Old Friend", "Creature — Human")
	bear := paPermanent(g, me.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2)

	rescue := castRescue(t, g, target, bear)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(rescue) {
		t.Error("Rescue from the Underworld exiles itself as it resolves")
	}
	if !me.Graveyard.Contains(bear) || !me.Graveyard.Contains(target) {
		t.Fatal("both cards wait in the graveyard until the upkeep")
	}
	if len(g.DelayedTriggers) != 1 || len(g.DelayedTriggers[0].Params.Objects) != 2 {
		t.Fatalf("one delayed trigger naming two objects, got %+v", g.DelayedTriggers)
	}

	rescueReachYourNextUpkeep(t, g)
	for _, id := range []uuid.UUID{target, bear} {
		c := findBattlefieldCardForTest(g, id)
		if c == nil {
			t.Fatalf("%v did not return at the beginning of your next upkeep", id)
		}
		if c.Controller != me.ID {
			t.Error("it returns under your control")
		}
	}
	// Returned together: one event, so each sees the other enter (CR 603.6a).
	batches := map[uuid.UUID]uint64{}
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventETB && (ev.CardID == target || ev.CardID == bear) {
			batches[ev.CardID] = ev.Batch
		}
	}
	if len(batches) != 2 || batches[target] == 0 || batches[target] != batches[bear] {
		t.Errorf("the two cards enter as one event: batches %v", batches)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Error("the delayed trigger fires once")
	}
}

// A creature you control but don't own goes to its owner's graveyard,
// and comes back under your control.
func TestRescueFromTheUnderworldReturnsAStolenCreatureUnderYourControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := pushGraveyardCardTyped(me, "Old Friend", "Creature — Human")
	stolen := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: me.ID,
	})

	castRescue(t, g, target, stolen)
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(stolen) {
		t.Fatal("the sacrificed creature goes to its owner's graveyard")
	}
	rescueReachYourNextUpkeep(t, g)
	c := findBattlefieldCardForTest(g, stolen)
	if c == nil || c.Controller != me.ID {
		t.Fatalf("the sacrificed card returns from its owner's graveyard under your control: %+v", c)
	}
}

// A sacrificed token has ceased to exist; the target still returns (the
// 2013-09-15 ruling).
func TestRescueFromTheUnderworldReturnsOnlyTheTargetForAToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	target := pushGraveyardCardTyped(me, "Old Friend", "Creature — Human")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	goblin := findBattlefieldByName(g, "Goblin")

	castRescue(t, g, target, goblin)
	passPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 || len(g.DelayedTriggers[0].Params.Objects) != 1 {
		t.Fatalf("the trigger names the target alone, got %+v", g.DelayedTriggers)
	}
	rescueReachYourNextUpkeep(t, g)
	if !g.Battlefield.Contains(target) {
		t.Error("the target returns")
	}
	if findBattlefieldByName(g, "Goblin") != uuid.Nil {
		t.Error("a token does not come back")
	}
}

// Either card that leaves the graveyard before the trigger resolves,
// even to come straight back, is a new object and stays (CR 400.7,
// 603.7c; the 2013-09-15 ruling). The other still returns.
func TestRescueFromTheUnderworldLeavesACardThatLeftTheGraveyard(t *testing.T) {
	for _, moveTarget := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		target := pushGraveyardCardTyped(me, "Old Friend", "Creature — Human")
		bear := paPermanent(g, me.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2)
		castRescue(t, g, target, bear)
		passPriorityAroundTable(t, g)

		moved, stays := bear, target
		if moveTarget {
			moved, stays = target, bear
		}
		g.WithWriteLock(func() {
			if _, err := game.MoveCard(me.Graveyard, g.Exile, moved); err != nil {
				t.Fatalf("MoveCard: %v", err)
			}
			if _, err := game.MoveCard(g.Exile, me.Graveyard, moved); err != nil {
				t.Fatalf("MoveCard: %v", err)
			}
		})
		rescueReachYourNextUpkeep(t, g)
		if !me.Graveyard.Contains(moved) || g.Battlefield.Contains(moved) {
			t.Errorf("moveTarget=%v: the card that left and came back is a new object and stays", moveTarget)
		}
		if !g.Battlefield.Contains(stays) {
			t.Errorf("moveTarget=%v: the other card still returns", moveTarget)
		}
	}
}

// A sacrificed creature a replacement sends to exile instead is found
// there and returns from exile (CR 400.7j; the 2013-09-15 ruling).
func TestRescueFromTheUnderworldFollowsTheSacrificedCardIntoExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castRestInPeace(t, g)
	target := pushGraveyardCardTyped(me, "Old Friend", "Creature — Human")
	bear := paPermanent(g, me.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2)

	castRescue(t, g, target, bear)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) {
		t.Fatal("Rest in Peace exiles the sacrificed creature")
	}
	rescueReachYourNextUpkeep(t, g)
	if g.Exile.Contains(bear) || findBattlefieldByName(g, "Bear") == uuid.Nil {
		t.Error("the sacrificed card returns from exile")
	}
	if !g.Battlefield.Contains(target) {
		t.Error("the target returns from the graveyard")
	}
}

// A sacrificed commander: left in the graveyard it returns; moved to the
// command zone by CR 903.9a's state-based action it has changed zones a
// second time, is a new object, and stays (CR 400.7).
func TestRescueFromTheUnderworldAndASacrificedCommander(t *testing.T) {
	for _, toCommandZone := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		target := pushGraveyardCardTyped(me, "Old Friend", "Creature — Human")
		commander := game.NewCommander("Test Commander", me.ID)
		commander.TypeLine = "Legendary Creature — Bear"
		commander.Power, commander.Toughness = 2, 2
		commander.Controller = me.ID
		g.Battlefield.PushTop(commander)

		castRescue(t, g, target, commander.InstanceID)
		choices := commanderReturnChoicesForTest(g)
		if len(choices) != 1 {
			t.Fatalf("toCommandZone=%v: CR 903.9a asks once, got %d", toCommandZone, len(choices))
		}
		if err := g.ResolveCommanderReturn(choices[0].ID, me.ID, toCommandZone); err != nil {
			t.Fatalf("ResolveCommanderReturn: %v", err)
		}
		passPriorityAroundTable(t, g)
		rescueReachYourNextUpkeep(t, g)
		if !g.Battlefield.Contains(target) {
			t.Errorf("toCommandZone=%v: the target returns", toCommandZone)
		}
		back := g.Battlefield.Contains(commander.InstanceID)
		if back == toCommandZone {
			t.Errorf("toCommandZone=%v: commander returned = %v", toCommandZone, back)
		}
		if toCommandZone && !me.Command.Contains(commander.InstanceID) {
			t.Error("the commander stays in the command zone")
		}
	}
}

// An illegal target as the spell resolves: it does not resolve (CR
// 608.2b), so no trigger, the sacrificed card stays, and the spell goes
// to the graveyard rather than exile.
func TestRescueFromTheUnderworldWithAnIllegalTargetDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	target := pushGraveyardCardTyped(me, "Old Friend", "Creature — Human")
	bear := paPermanent(g, me.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2)

	rescue := castRescue(t, g, target, bear)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(target); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 0 {
		t.Error("a spell that does not resolve creates no delayed trigger")
	}
	if !me.Graveyard.Contains(rescue) {
		t.Error("the spell goes to the graveyard, not exile")
	}
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(bear) {
		t.Error("the sacrificed card stays in the graveyard")
	}
}
