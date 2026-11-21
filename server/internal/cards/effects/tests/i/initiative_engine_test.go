package testsi

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// TestTakingTheInitiativeAgainVenturesAgain is CR 726.5: the holder
// told to take it again keeps one designation and ventures once more.
func TestTakingTheInitiativeAgainVenturesAgain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.InitVenture(t, g, me.ID)
	if room := effectstest.InitRoom(g, me.ID); room != "secret-entrance" {
		t.Fatalf("room = %q, want secret-entrance", room)
	}
	effectstest.InitVenture(t, g, me.ID, "lost-well")
	if room := effectstest.InitRoom(g, me.ID); room != "lost-well" {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	giantCard := game.Card{Name: "Hill Giant", TypeLine: "Creature — Giant", Power: 3, Toughness: 3}
	giantCard.InstanceID = uuid.New()
	giantCard.Owner, giantCard.Controller = me.ID, me.ID
	me.Library.PushTop(giantCard)

	effectstest.InitVenture(t, g, me.ID)
	effectstest.InitVenture(t, g, me.ID, "lost-well")
	effectstest.InitVenture(t, g, me.ID, "stash")
	if n := effectstest.MonTokens(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("Stash made %d Treasures, want 1", n)
	}
	effectstest.InitVenture(t, g, me.ID)
	if room := effectstest.InitRoom(g, me.ID); room != "catacombs" {
		t.Fatalf("room = %q, want catacombs (Stash leads only there)", room)
	}
	if n := effectstest.MonTokens(g, me.ID, "Skeleton"); n != 1 {
		t.Fatalf("Catacombs made %d Skeletons, want 1", n)
	}
	var first uuid.UUID
	g.WithWriteLock(func() { first = game.DungeonOf(g, me.ID).ID })

	effectstest.InitVenture(t, g, me.ID)
	giant := effectstest.FindBattlefieldCardForTest(g, giantCard.InstanceID)
	if giant == nil {
		t.Fatalf("the Throne did not put the Hill Giant onto the battlefield")
	}
	if got := giant.Counters["+1/+1"]; got != 3 {
		t.Errorf("Hill Giant has %d +1/+1 counters, want 3", got)
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, giantCard.InstanceID), "hexproof") {
		t.Errorf("Hill Giant does not have hexproof until your next turn")
	}
	if room := effectstest.InitRoom(g, me.ID); room != "" {
		t.Fatalf("the Undercity was not completed: marker in %q", room)
	}
	if got := effectstest.InitCompleted(g, me.ID); len(got) != 1 || got[0] != game.UndercityKey {
		t.Fatalf("completed = %v, want [undercity]", got)
	}

	effectstest.InitVenture(t, g, me.ID)
	if room := effectstest.InitRoom(g, me.ID); room != "secret-entrance" {
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
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.InitVenture(t, g, them.ID)
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	effectstest.AttackWith(t, g, them.ID, a, b)
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
	effectstest.InitSettle(t, g)
	if g.Initiative != me.ID {
		t.Fatalf("initiative = %v, want the attacker %v", g.Initiative, me.ID)
	}
	if room := effectstest.InitRoom(g, me.ID); room != "secret-entrance" {
		t.Fatalf("the attacker's room = %q, want secret-entrance", room)
	}
	if room := effectstest.InitRoom(g, them.ID); room != "secret-entrance" {
		t.Fatalf("the old holder's marker moved: %q", room)
	}
}

// TestTheHolderVenturesAtTheBeginningOfTheirUpkeep is CR 726.2's
// first ability, on the holder's own upkeep.
func TestTheHolderVenturesAtTheBeginningOfTheirUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	holder := g.Seats[1]
	effectstest.InitVenture(t, g, holder.ID)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.InitSettle(t, g, "lost-well")
	if room := effectstest.InitRoom(g, holder.ID); room != "lost-well" {
		t.Fatalf("room = %q, want lost-well after the upkeep venture", room)
	}
}

// TestForgePutsTwoCountersOnTheTargetedCreature: a room ability that
// targets uses the ordinary trigger targeting.
func TestForgePutsTwoCountersOnTheTargetedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.InitVenture(t, g, me.ID)
	effectstest.InitTake(t, g, me.ID)
	effectstest.InitSettle(t, g, "forge")
	c := effectstest.InitPending(g, game.PendingChoicePickTarget, me.ID)
	if c == nil {
		t.Fatalf("Forge asked for no target")
	}
	if err := g.ResolvePickTargets(c.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.InitSettle(t, g)
	if got := effectstest.FindBattlefieldCardForTest(g, bear).Counters["+1/+1"]; got != 2 {
		t.Fatalf("Bear has %d +1/+1 counters, want 2", got)
	}
}

// TestTrapCostsTheTargetedPlayerFiveLife.
func TestTrapCostsTheTargetedPlayerFiveLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.InitVenture(t, g, me.ID)
	effectstest.InitTake(t, g, me.ID)
	effectstest.InitSettle(t, g, "forge")
	c := effectstest.InitPending(g, game.PendingChoicePickTarget, me.ID)
	if c == nil {
		t.Fatalf("Forge asked for no target")
	}
	if err := g.ResolvePickTargets(c.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.InitSettle(t, g)
	before := them.Life
	effectstest.InitTake(t, g, me.ID)
	effectstest.InitSettle(t, g, "trap")
	c = effectstest.InitPending(g, game.PendingChoicePickTarget, me.ID)
	if c == nil {
		t.Fatalf("Trap! asked for no target")
	}
	if err := g.ResolvePickTargets(c.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.InitSettle(t, g)
	if them.Life != before-5 {
		t.Fatalf("their life %d, want %d", them.Life, before-5)
	}
}

// TestTheInitiativePassesToTheActivePlayerWhenItsHolderLeaves is CR
// 726.4: the active player takes it at once, and ventures.
func TestTheInitiativePassesToTheActivePlayerWhenItsHolderLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active, holder := g.Seats[0], g.Seats[2]
	effectstest.InitVenture(t, g, holder.ID)
	if err := g.Concede(holder.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if g.Initiative != active.ID {
		t.Fatalf("initiative = %v, want the active player %v", g.Initiative, active.ID)
	}
	effectstest.InitSettle(t, g)
	if room := effectstest.InitRoom(g, active.ID); room != "secret-entrance" {
		t.Fatalf("the new holder's room = %q, want secret-entrance", room)
	}
	if room := effectstest.InitRoom(g, holder.ID); room != "" {
		t.Fatalf("the departed player's dungeon stayed: %q", room)
	}
}

// TestTheManualInitiativeVentures: the sandbox's set_initiative is the
// same take a card makes, and clearing it moves no marker.
func TestTheManualInitiativeVentures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if err := g.SetInitiative(me.ID); err != nil {
		t.Fatalf("SetInitiative: %v", err)
	}
	effectstest.InitSettle(t, g)
	if room := effectstest.InitRoom(g, me.ID); room != "secret-entrance" {
		t.Fatalf("room = %q, want secret-entrance", room)
	}
	if err := g.SetInitiative(uuid.Nil); err != nil {
		t.Fatalf("SetInitiative clear: %v", err)
	}
	effectstest.InitSettle(t, g)
	if g.Initiative != uuid.Nil || effectstest.InitRoom(g, me.ID) != "secret-entrance" {
		t.Fatalf("clear: initiative %v room %q", g.Initiative, effectstest.InitRoom(g, me.ID))
	}
}

// TestAVentureBranchPromptIsARestorePoint: the room choice is keyed,
// so a table waiting on it survives a strict snapshot round trip and
// the restored answer moves the marker.
func TestAVentureBranchPromptIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.InitVenture(t, g, me.ID)
	effectstest.InitTake(t, g, me.ID)
	effectstest.InitSettle(t, g)
	if c := effectstest.InitPending(g, game.PendingChoiceOptionPick, me.ID); c == nil {
		t.Fatalf("no branch prompt open")
	}
	restored := effectstest.RestoreRoundTrip(t, g, true)
	effectstest.InitSettle(t, restored, "lost-well")
	if room := effectstest.InitRoom(restored, me.ID); room != "lost-well" {
		t.Fatalf("restored room = %q, want lost-well", room)
	}
}

// TestARoomAbilityOnTheStackIsARestorePoint: a waiting room ability
// names its catalog row, so it restores and resolves.
func TestARoomAbilityOnTheStackIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.InitVenture(t, g, me.ID)
	effectstest.InitVenture(t, g, me.ID, "lost-well")
	effectstest.InitVenture(t, g, me.ID, "stash")
	// Stash leads only to the Catacombs: venture once more and stop
	// with the Catacombs' ability waiting.
	effectstest.InitTake(t, g, me.ID)
	for i := 0; i < 8 && effectstest.InitRoom(g, me.ID) != "catacombs"; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if effectstest.InitRoom(g, me.ID) != "catacombs" {
		t.Fatalf("never reached the Catacombs")
	}
	if effectstest.MonTokens(g, me.ID, "Skeleton") != 0 {
		t.Fatalf("the Catacombs resolved before the round trip")
	}
	restored := effectstest.RestoreRoundTrip(t, g, true)
	effectstest.InitSettle(t, restored)
	if n := effectstest.MonTokens(restored, me.ID, "Skeleton"); n != 1 {
		t.Fatalf("restored Catacombs made %d Skeletons, want 1", n)
	}
}

// TestTheDungeonSurvivesUndo: Clone / RestoreFrom carry the marker.
func TestTheDungeonSurvivesUndo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.InitVenture(t, g, me.ID)
	before := g.Clone()
	effectstest.InitVenture(t, g, me.ID, "lost-well")
	g.RestoreFrom(before)
	if room := effectstest.InitRoom(g, me.ID); room != "secret-entrance" {
		t.Fatalf("after undo room = %q, want secret-entrance", room)
	}
}

// TestTheEnumeratorOffersEachRoomOfABranch: the bot seat's move list
// answers the venture prompt with one move per arrow, and each one
// moves the marker into the room it names.
func TestTheEnumeratorOffersEachRoomOfABranch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.InitVenture(t, g, me.ID)
	effectstest.InitTake(t, g, me.ID)
	effectstest.InitSettle(t, g)
	c := effectstest.InitPending(g, game.PendingChoiceOptionPick, me.ID)
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
	if room := effectstest.InitRoom(g, me.ID); room != want {
		t.Fatalf("room = %q, want %q", room, want)
	}
}

// TestTheLastRoomStaysUntilItsAbilityLeavesTheStack is CR 704.5t's
// exception: the marker on the Throne does not complete the Undercity
// while the Throne's own ability is still waiting to resolve.
func TestTheLastRoomStaysUntilItsAbilityLeavesTheStack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.InitVenture(t, g, me.ID)
	effectstest.InitVenture(t, g, me.ID, "lost-well")
	effectstest.InitVenture(t, g, me.ID, "stash")
	effectstest.InitVenture(t, g, me.ID)
	effectstest.InitTake(t, g, me.ID)
	for i := 0; i < 8 && effectstest.InitRoom(g, me.ID) != "throne"; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if effectstest.InitRoom(g, me.ID) != "throne" {
		t.Fatalf("never reached the Throne")
	}
	// One more pass puts the Throne's ability on the stack and gives
	// priority: state-based actions have run with it waiting.
	if len(effectstest.InitCompleted(g, me.ID)) != 0 {
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
	effectstest.InitSettle(t, g)
	if got := effectstest.InitCompleted(g, me.ID); len(got) != 1 {
		t.Fatalf("completed = %v after the Throne resolved, want [undercity]", got)
	}
}
