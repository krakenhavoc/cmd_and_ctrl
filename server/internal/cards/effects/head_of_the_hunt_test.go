package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// head_of_the_hunt_test.go — #4517: "If a creature an opponent controls
// would die, exile it instead. When you do, create a 2/2 green Wolf
// creature token." A reflexive trigger whose creator is the replacement
// (CR 603.12, CR 614.6), driven through the real destroy, wipe, CR 616
// and restore paths.

const headOfTheHuntOracle = "0523e271-5e0b-4e4b-9eac-86f915d18952"

func seedHeadOfTheHunt(g *game.Game, controller uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, controller, "Head of the Hunt", "Creature — Wolf", headOfTheHuntOracle, false)
}

// pendingHuntWolves counts the Wolf triggers waiting to go on the stack
// or already on it, and checks each is controlled by want.
func pendingHuntWolves(t *testing.T, g *game.Game, want uuid.UUID) int {
	t.Helper()
	n := 0
	check := func(it *game.StackItem) {
		if it == nil || !containsFold(it.Label, "Head of the Hunt — create") {
			return
		}
		n++
		if it.Controller != want {
			t.Errorf("the Wolf trigger is controlled by %v, want the replacement's controller %v", it.Controller, want)
		}
	}
	for _, it := range g.PendingTriggers {
		check(it)
	}
	for _, it := range g.StackMeta {
		check(it)
	}
	return n
}

func TestHeadOfTheHuntIsRegisteredFullWithFlash(t *testing.T) {
	spec, ok := Lookup(headOfTheHuntOracle)
	if !ok {
		t.Fatal("Head of the Hunt is not registered")
	}
	if spec.Completeness != CompletenessFull || !hasKeyword(spec.PrintedKeywords, "flash") {
		t.Fatalf("completeness %v, keywords %v", spec.Completeness, spec.PrintedKeywords)
	}
}

// Flash: an opponent of the active player casts it in the active
// player's main phase, and it resolves onto the battlefield.
func TestHeadOfTheHuntHasFlash(t *testing.T) {
	g := newCatalogGame(t)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	caster := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := uuid.New()
	card := game.Card{InstanceID: id, Name: "Head of the Hunt", TypeLine: "Creature — Wolf",
		OracleID: headOfTheHuntOracle, ManaCost: "{2}{B}{B}", Owner: caster.ID, Controller: caster.ID}
	var open bool
	g.ReadSnapshot(func() { open = g.CastTimingOpenLocked(caster.ID, card, game.ZoneHand, nil) })
	if !open {
		t.Fatal("flash should let a non-active player cast it in another player's main phase")
	}
	// The control: the same creature without flash may not be cast then.
	plain := card
	plain.OracleID = ""
	plain.Name = "A Plain Wolf"
	g.ReadSnapshot(func() { open = g.CastTimingOpenLocked(caster.ID, plain, game.ZoneHand, nil) })
	if open {
		t.Fatal("a creature without flash was castable on another player's turn")
	}
}

// One opponent creature destroyed: it is exiled, and the Wolf is a
// trigger that goes on the stack under Head of the Hunt's controller,
// not a token made inside the replacement.
func TestHeadOfTheHuntExilesAnOpponentsCreatureAndMakesAWolf(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedHeadOfTheHunt(g, me.ID)
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if !g.Exile.Contains(bear) || opp.Graveyard.Contains(bear) {
		t.Fatalf("the bear should be exiled, not in the graveyard (exile %v)", g.Exile.Contains(bear))
	}
	if got := tokensNamedControlledBy(g, "Wolf", me.ID); got != 0 {
		t.Fatalf("%d Wolves before the trigger resolved: it must use the stack", got)
	}
	if got := pendingHuntWolves(t, g, me.ID); got != 1 {
		t.Fatalf("%d Wolf triggers waiting, want 1", got)
	}
	passPriorityAroundTable(t, g)
	if got := tokensNamedControlledBy(g, "Wolf", me.ID); got != 1 {
		t.Fatalf("%d Wolves for Head of the Hunt's controller, want 1", got)
	}
}

// A creature that dies to lethal damage (a state-based action) is
// exiled and still makes its Wolf.
func TestHeadOfTheHuntOnAStateBasedDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedHeadOfTheHunt(g, me.ID)
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	if err := g.MarkDamage(bear, 2); err != nil {
		t.Fatalf("MarkDamage: %v", err)
	}
	if !g.Exile.Contains(bear) {
		t.Fatal("the creature that died to damage was not exiled")
	}
	passPriorityAroundTable(t, g)
	if got := tokensNamedControlledBy(g, "Wolf", me.ID); got != 1 {
		t.Fatalf("%d Wolves, want 1", got)
	}
}

// A wipe of three opponent creatures: three exiles, three triggers,
// three Wolves (CR 603.12a).
func TestHeadOfTheHuntOneWolfPerCreatureInAWipe(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedHeadOfTheHunt(g, me.ID)
	ids := []uuid.UUID{
		pushVanillaCreature(g, g.Seats[1].ID, "Bear A", 2, 2),
		pushVanillaCreature(g, g.Seats[2].ID, "Bear B", 2, 2),
		pushVanillaCreature(g, g.Seats[3].ID, "Bear C", 2, 2),
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentsKeyedForEffect(ids, game.KeyedTail{}); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	for _, id := range ids {
		if !g.Exile.Contains(id) {
			t.Errorf("a creature of the wipe is not in exile")
		}
	}
	if got := pendingHuntWolves(t, g, me.ID); got != 3 {
		t.Fatalf("%d Wolf triggers, want 3", got)
	}
	passPriorityAroundTable(t, g)
	if got := tokensNamedControlledBy(g, "Wolf", me.ID); got != 3 {
		t.Fatalf("%d Wolves, want 3", got)
	}
}

// Its controller's own creature is not an opponent's: it dies normally
// and makes no Wolf.
func TestHeadOfTheHuntIgnoresYourOwnCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedHeadOfTheHunt(g, me.ID)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	if !me.Graveyard.Contains(mine) {
		t.Fatal("your own creature should go to your graveyard")
	}
	if got := pendingHuntWolves(t, g, me.ID); got != 0 {
		t.Fatalf("%d Wolf triggers for your own creature", got)
	}
	passPriorityAroundTable(t, g)
	if got := tokensNamedControlledBy(g, "Wolf", me.ID); got != 0 {
		t.Fatalf("%d Wolves for your own creature", got)
	}
}

// Head of the Hunt has left before the creature dies: no replacement,
// no Wolf.
func TestHeadOfTheHuntGoneFirstMeansNoWolf(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	head := seedHeadOfTheHunt(g, me.ID)
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, head); err != nil {
		t.Fatalf("bounce Head of the Hunt: %v", err)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if !opp.Graveyard.Contains(bear) {
		t.Fatal("with Head of the Hunt gone the creature dies normally")
	}
	passPriorityAroundTable(t, g)
	if got := tokensNamedControlledBy(g, "Wolf", me.ID); got != 0 {
		t.Fatalf("%d Wolves with Head of the Hunt gone", got)
	}
}

// Rest in Peace is a second replacement on the same death. The dying
// creature's controller chooses (CR 616.1). Rest in Peace first: the
// creature is exiled by Rest in Peace, Head of the Hunt never applies,
// and there is no Wolf, though the creature did end up in exile. Head of
// the Hunt first: it exiles the creature and the Wolf comes.
func TestHeadOfTheHuntWithRestInPeace(t *testing.T) {
	for _, tc := range []struct {
		first  string
		wolves int
	}{
		{"Rest in Peace", 0},
		{"Head of the Hunt", 1},
	} {
		t.Run(tc.first, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			seedEnchantment(g, restInPeaceOracle, "Rest in Peace", "Enchantment", g.Seats[2].ID)
			seedHeadOfTheHunt(g, me.ID)
			bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
			answerNexusOrder(t, g, tc.first, opp.ID)
			if !g.Exile.Contains(bear) {
				t.Fatal("either way the creature is exiled")
			}
			passPriorityAroundTable(t, g)
			if got := tokensNamedControlledBy(g, "Wolf", me.ID); got != tc.wolves {
				t.Fatalf("%s first: %d Wolves, want %d", tc.first, got, tc.wolves)
			}
		})
	}
}

// A restore while the reflexive trigger waits to go on the stack: the
// restored table still makes the Wolf, for the same controller.
func TestHeadOfTheHuntTriggerSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedHeadOfTheHunt(g, me.ID)
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if got := pendingHuntWolves(t, g, me.ID); got != 1 {
		t.Fatalf("%d Wolf triggers before the restore, want 1", got)
	}

	restored := restoreThroughJSON(t, g)
	rme := restored.Seats[0]
	if got := pendingHuntWolves(t, restored, rme.ID); got != 1 {
		t.Fatalf("%d Wolf triggers after the restore, want 1", got)
	}
	passPriorityAroundTable(t, restored)
	if got := tokensNamedControlledBy(restored, "Wolf", rme.ID); got != 1 {
		t.Fatalf("%d Wolves on the restored table, want 1", got)
	}
}
