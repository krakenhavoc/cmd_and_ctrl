package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// initiative_engine_test.go pins the engine half of ADR 0203: the three
// inherent triggers of CR 726.2, CR 726.4's hand-on, the venture walk
// through the Undercity, CR 704.5t, and restore points.

// initSettle resolves the stack, answering each prompt the Undercity
// asks: a trigger order in the order offered, a search with nothing, a
// scry keeping the order, a branch with the room `rooms` names next (in
// order). It stops at any other prompt, which the caller answers, and
// returns the rooms it was never asked for.
func initSettle(t *testing.T, g *game.Game, rooms ...string) []string {
	t.Helper()
	for i := 0; i < 128; i++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil {
				open = c
				break
			}
		}
		if open != nil {
			var err error
			switch open.Kind {
			case game.PendingChoiceTriggerOrder:
				err = g.ResolveTriggerOrder(open.ID, open.Chooser, append([]uuid.UUID(nil), open.TriggerOrderIDs...))
			case game.PendingChoiceSearchLibrary:
				err = g.ResolveSearchLibrary(open.ID, open.Chooser, nil)
			case game.PendingChoiceScry:
				err = g.ResolveScry(open.ID, open.Chooser, nil, append([]uuid.UUID(nil), open.ScryCards...))
			case game.PendingChoiceOptionPick:
				if len(rooms) == 0 || len(open.PickOptions) == 0 || open.PickOptions[0].Room == "" {
					return rooms
				}
				initPickRoom(t, g, open.Chooser, rooms[0])
				rooms = rooms[1:]
			default:
				return rooms
			}
			if err != nil {
				t.Fatalf("answering %s: %v", open.Kind, err)
			}
			continue
		}
		if stackFullyEmpty(g) {
			return rooms
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	t.Fatal("the stack did not settle")
	return rooms
}

// initVenture has `player` take the initiative again (CR 726.5: they
// venture) and settles, choosing `rooms` at the branches.
func initVenture(t *testing.T, g *game.Game, player uuid.UUID, rooms ...string) {
	t.Helper()
	initTake(t, g, player)
	if left := initSettle(t, g, rooms...); len(left) > 0 {
		t.Fatalf("rooms %v were never offered", left)
	}
}

// initCompleted reports `player`'s completed dungeons.
func initCompleted(g *game.Game, player uuid.UUID) []string {
	for _, p := range g.Seats {
		if p.ID == player {
			return append([]string(nil), p.CompletedDungeons...)
		}
	}
	return nil
}

// TestTakingTheInitiativeAgainVenturesAgain is CR 726.5: the holder
// told to take it again keeps one designation and ventures once more.
func TestTakingTheInitiativeAgainVenturesAgain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	initVenture(t, g, me.ID)
	if room := initRoom(g, me.ID); room != "secret-entrance" {
		t.Fatalf("room = %q, want secret-entrance", room)
	}
	initVenture(t, g, me.ID, "lost-well")
	if room := initRoom(g, me.ID); room != "lost-well" {
		t.Fatalf("room = %q, want lost-well", room)
	}
	if g.Initiative != me.ID {
		t.Fatalf("initiative = %v, want %v", g.Initiative, me.ID)
	}
}

// TestTheUndercityWalkCompletesAndStartsAgain walks Secret Entrance,
// Lost Well, Stash, Catacombs and the Throne; the Throne puts the one
// creature among the top ten onto the battlefield with three +1/+1
// counters and hexproof; the dungeon is completed by CR 704.5t once
// the Throne's ability has left the stack; and the next venture starts
// a new Undercity at Secret Entrance.
func TestTheUndercityWalkCompletesAndStartsAgain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	giantCard := game.Card{Name: "Hill Giant", TypeLine: "Creature — Giant", Power: 3, Toughness: 3}
	giantCard.InstanceID = uuid.New()
	giantCard.Owner, giantCard.Controller = me.ID, me.ID
	me.Library.PushTop(giantCard)

	initVenture(t, g, me.ID)
	initVenture(t, g, me.ID, "lost-well")
	initVenture(t, g, me.ID, "stash")
	if n := monTokens(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("Stash made %d Treasures, want 1", n)
	}
	initVenture(t, g, me.ID)
	if room := initRoom(g, me.ID); room != "catacombs" {
		t.Fatalf("room = %q, want catacombs (Stash leads only there)", room)
	}
	if n := monTokens(g, me.ID, "Skeleton"); n != 1 {
		t.Fatalf("Catacombs made %d Skeletons, want 1", n)
	}
	var first uuid.UUID
	g.WithWriteLock(func() { first = game.DungeonOf(g, me.ID).ID })

	initVenture(t, g, me.ID)
	giant := findBattlefieldCardForTest(g, giantCard.InstanceID)
	if giant == nil {
		t.Fatalf("the Throne did not put the Hill Giant onto the battlefield")
	}
	if got := giant.Counters["+1/+1"]; got != 3 {
		t.Errorf("Hill Giant has %d +1/+1 counters, want 3", got)
	}
	if !hasAbility(effectiveAbilities(t, g, giantCard.InstanceID), "hexproof") {
		t.Errorf("Hill Giant does not have hexproof until your next turn")
	}
	if room := initRoom(g, me.ID); room != "" {
		t.Fatalf("the Undercity was not completed: marker in %q", room)
	}
	if got := initCompleted(g, me.ID); len(got) != 1 || got[0] != game.UndercityKey {
		t.Fatalf("completed = %v, want [undercity]", got)
	}

	initVenture(t, g, me.ID)
	if room := initRoom(g, me.ID); room != "secret-entrance" {
		t.Fatalf("after completing, room = %q, want a new Undercity at secret-entrance", room)
	}
	g.WithWriteLock(func() {
		if game.DungeonOf(g, me.ID).ID == first {
			t.Error("the new Undercity is the same object as the completed one")
		}
	})
}

// TestCombatDamageToTheHolderTakesTheInitiativeOncePerBatch: two
// creatures of one player connecting with the holder is one trigger
// (CR 726.2: "one or more creatures"), and their controller takes the
// initiative and ventures.
func TestCombatDamageToTheHolderTakesTheInitiativeOncePerBatch(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	initVenture(t, g, them.ID)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	attackWith(t, g, them.ID, a, b)
	steals := 0
	for _, it := range g.PendingTriggers {
		if it != nil && it.Body == "initiative/take" {
			steals++
		}
	}
	for _, it := range g.StackMeta {
		if it != nil && it.Body == "initiative/take" {
			steals++
		}
	}
	if steals != 1 {
		t.Fatalf("%d steal triggers, want 1 for the batch", steals)
	}
	initSettle(t, g)
	if g.Initiative != me.ID {
		t.Fatalf("initiative = %v, want the attacker %v", g.Initiative, me.ID)
	}
	if room := initRoom(g, me.ID); room != "secret-entrance" {
		t.Fatalf("the attacker's room = %q, want secret-entrance", room)
	}
	if room := initRoom(g, them.ID); room != "secret-entrance" {
		t.Fatalf("the old holder's marker moved: %q", room)
	}
}

// TestTheHolderVenturesAtTheBeginningOfTheirUpkeep is CR 726.2's
// first ability, on the holder's own upkeep.
func TestTheHolderVenturesAtTheBeginningOfTheirUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	holder := g.Seats[1]
	initVenture(t, g, holder.ID)
	advanceToUpkeepOf(t, g, 1)
	initSettle(t, g, "lost-well")
	if room := initRoom(g, holder.ID); room != "lost-well" {
		t.Fatalf("room = %q, want lost-well after the upkeep venture", room)
	}
}

// TestForgePutsTwoCountersOnTheTargetedCreature: a room ability that
// targets uses the ordinary trigger targeting.
func TestForgePutsTwoCountersOnTheTargetedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	initVenture(t, g, me.ID)
	initTake(t, g, me.ID)
	initSettle(t, g, "forge")
	c := initPending(g, game.PendingChoicePickTarget, me.ID)
	if c == nil {
		t.Fatalf("Forge asked for no target")
	}
	if err := g.ResolvePickTargets(c.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	initSettle(t, g)
	if got := findBattlefieldCardForTest(g, bear).Counters["+1/+1"]; got != 2 {
		t.Fatalf("Bear has %d +1/+1 counters, want 2", got)
	}
}

// TestTrapCostsTheTargetedPlayerFiveLife.
func TestTrapCostsTheTargetedPlayerFiveLife(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	initVenture(t, g, me.ID)
	initTake(t, g, me.ID)
	initSettle(t, g, "forge")
	c := initPending(g, game.PendingChoicePickTarget, me.ID)
	if c == nil {
		t.Fatalf("Forge asked for no target")
	}
	if err := g.ResolvePickTargets(c.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	initSettle(t, g)
	before := them.Life
	initTake(t, g, me.ID)
	initSettle(t, g, "trap")
	c = initPending(g, game.PendingChoicePickTarget, me.ID)
	if c == nil {
		t.Fatalf("Trap! asked for no target")
	}
	if err := g.ResolvePickTargets(c.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	initSettle(t, g)
	if them.Life != before-5 {
		t.Fatalf("their life %d, want %d", them.Life, before-5)
	}
}

// TestTheInitiativePassesToTheActivePlayerWhenItsHolderLeaves is CR
// 726.4: the active player takes it at once, and ventures.
func TestTheInitiativePassesToTheActivePlayerWhenItsHolderLeaves(t *testing.T) {
	g := newCatalogGame(t)
	active, holder := g.Seats[0], g.Seats[2]
	initVenture(t, g, holder.ID)
	if err := g.Concede(holder.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if g.Initiative != active.ID {
		t.Fatalf("initiative = %v, want the active player %v", g.Initiative, active.ID)
	}
	initSettle(t, g)
	if room := initRoom(g, active.ID); room != "secret-entrance" {
		t.Fatalf("the new holder's room = %q, want secret-entrance", room)
	}
	if room := initRoom(g, holder.ID); room != "" {
		t.Fatalf("the departed player's dungeon stayed: %q", room)
	}
}

// TestTheManualInitiativeVentures: the sandbox's set_initiative is the
// same take a card makes, and clearing it moves no marker.
func TestTheManualInitiativeVentures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if err := g.SetInitiative(me.ID); err != nil {
		t.Fatalf("SetInitiative: %v", err)
	}
	initSettle(t, g)
	if room := initRoom(g, me.ID); room != "secret-entrance" {
		t.Fatalf("room = %q, want secret-entrance", room)
	}
	if err := g.SetInitiative(uuid.Nil); err != nil {
		t.Fatalf("SetInitiative clear: %v", err)
	}
	initSettle(t, g)
	if g.Initiative != uuid.Nil || initRoom(g, me.ID) != "secret-entrance" {
		t.Fatalf("clear: initiative %v room %q", g.Initiative, initRoom(g, me.ID))
	}
}

// TestAVentureBranchPromptIsARestorePoint: the room choice is keyed,
// so a table waiting on it survives a strict snapshot round trip and
// the restored answer moves the marker.
func TestAVentureBranchPromptIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	initVenture(t, g, me.ID)
	initTake(t, g, me.ID)
	initSettle(t, g)
	if c := initPending(g, game.PendingChoiceOptionPick, me.ID); c == nil {
		t.Fatalf("no branch prompt open")
	}
	restored := restoreRoundTrip(t, g, true)
	initSettle(t, restored, "lost-well")
	if room := initRoom(restored, me.ID); room != "lost-well" {
		t.Fatalf("restored room = %q, want lost-well", room)
	}
}

// TestARoomAbilityOnTheStackIsARestorePoint: a waiting room ability
// names its catalog row, so it restores and resolves.
func TestARoomAbilityOnTheStackIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	initVenture(t, g, me.ID)
	initVenture(t, g, me.ID, "lost-well")
	initVenture(t, g, me.ID, "stash")
	// Stash leads only to the Catacombs: venture once more and stop
	// with the Catacombs' ability waiting.
	initTake(t, g, me.ID)
	for i := 0; i < 8 && initRoom(g, me.ID) != "catacombs"; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if initRoom(g, me.ID) != "catacombs" {
		t.Fatalf("never reached the Catacombs")
	}
	if monTokens(g, me.ID, "Skeleton") != 0 {
		t.Fatalf("the Catacombs resolved before the round trip")
	}
	restored := restoreRoundTrip(t, g, true)
	initSettle(t, restored)
	if n := monTokens(restored, me.ID, "Skeleton"); n != 1 {
		t.Fatalf("restored Catacombs made %d Skeletons, want 1", n)
	}
}

// TestTheDungeonSurvivesUndo: Clone / RestoreFrom carry the marker.
func TestTheDungeonSurvivesUndo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	initVenture(t, g, me.ID)
	before := g.Clone()
	initVenture(t, g, me.ID, "lost-well")
	g.RestoreFrom(before)
	if room := initRoom(g, me.ID); room != "secret-entrance" {
		t.Fatalf("after undo room = %q, want secret-entrance", room)
	}
}

// TestTheEnumeratorOffersEachRoomOfABranch: the bot seat's move list
// answers the venture prompt with one move per arrow, and each one
// moves the marker into the room it names.
func TestTheEnumeratorOffersEachRoomOfABranch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	initVenture(t, g, me.ID)
	initTake(t, g, me.ID)
	initSettle(t, g)
	c := initPending(g, game.PendingChoiceOptionPick, me.ID)
	if c == nil {
		t.Fatalf("no branch prompt open")
	}
	var answers []legal.Move
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Kind == legal.KindChoice && m.Type == legal.TypeResolveChoice {
			answers = append(answers, m)
		}
	}
	if len(answers) != 2 {
		t.Fatalf("%d answers, want one per room (Forge, Lost Well): %+v", len(answers), answers)
	}
	var p struct {
		Index *int `json:"option_index"`
	}
	if err := json.Unmarshal(answers[1].Params, &p); err != nil || p.Index == nil {
		t.Fatalf("params %s: %v", answers[1].Params, err)
	}
	want := c.PickOptions[*p.Index].Room
	if err := g.ResolveOptionPick(c.ID, me.ID, *p.Index); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if room := initRoom(g, me.ID); room != want {
		t.Fatalf("room = %q, want %q", room, want)
	}
}

// TestTheLastRoomStaysUntilItsAbilityLeavesTheStack is CR 704.5t's
// exception: the marker on the Throne does not complete the Undercity
// while the Throne's own ability is still waiting to resolve.
func TestTheLastRoomStaysUntilItsAbilityLeavesTheStack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	initVenture(t, g, me.ID)
	initVenture(t, g, me.ID, "lost-well")
	initVenture(t, g, me.ID, "stash")
	initVenture(t, g, me.ID)
	initTake(t, g, me.ID)
	for i := 0; i < 8 && initRoom(g, me.ID) != "throne"; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if initRoom(g, me.ID) != "throne" {
		t.Fatalf("never reached the Throne")
	}
	// One more pass puts the Throne's ability on the stack and gives
	// priority: state-based actions have run with it waiting.
	if len(initCompleted(g, me.ID)) != 0 {
		t.Fatalf("the Undercity was completed with the Throne's ability still waiting")
	}
	waiting := false
	for _, it := range g.StackMeta {
		if it != nil && it.Label == "Undercity — Throne of the Dead Three" {
			waiting = true
		}
	}
	for _, it := range g.PendingTriggers {
		if it != nil && it.Label == "Undercity — Throne of the Dead Three" {
			waiting = true
		}
	}
	if !waiting {
		t.Fatalf("the Throne's ability is not waiting")
	}
	initSettle(t, g)
	if got := initCompleted(g, me.ID); len(got) != 1 {
		t.Fatalf("completed = %v after the Throne resolved, want [undercity]", got)
	}
}
