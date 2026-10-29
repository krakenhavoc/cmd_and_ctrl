package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_target_walk_restore_test.go — #4361 S14b: real catalog cards
// whose triggers pause in their CR 603.3d target walk, restored strictly
// mid-prompt through JSON, and answered on the restored table.

// Manglehorn's "destroy target artifact": a one-step walk. The table
// waiting on the pick is a restore point, and the restored pick destroys
// the artifact.
func TestManglehornTargetPickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myRock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	theirRock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	castAndResolveCreature(t, g, "Manglehorn", "Creature — Beast", b21ManglehornOracle)
	b04WaitForPick(t, g, me.ID)

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	p := latestPickTarget(g, me.ID)
	if p == nil || !hasID(p.PickTargetCards, theirRock) || !hasID(p.PickTargetCards, myRock) {
		t.Fatalf("the restored pick is %+v, want both artifacts", p)
	}
	pickCard(t, g, me.ID, theirRock)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirRock) || !g.Battlefield.Contains(myRock) {
		t.Error("the restored pick did not destroy the chosen artifact")
	}
}

// Charming Prince's exile bullet, restored while its target is open: the
// written walk carries the CR 603.3c mode, and the restored pick exiles.
func TestCharmingPrinceTargetPickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pal := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pal", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	castAndResolveCreature(t, g, "Charming Prince", "Creature — Human Noble", charmingPrinceOracle)
	c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatal("no mode_pick")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{2}); err != nil {
		t.Fatalf("ResolveModePick(exile): %v", err)
	}
	if latestPickTarget(g, me.ID) == nil {
		t.Fatal("the exile bullet asks for no target")
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	pickCard(t, g, me.ID, pal)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(pal) || !g.Exile.Contains(pal) {
		t.Error("the restored pick did not run the exile bullet")
	}
}

// Ureni's division, restored while it is being announced: the amount is
// the lands counted as the trigger was put on the stack (CR 601.2d via
// CR 603.3d), not the restored board's.
func TestUreniDivisionPromptIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	d1657UreniEnters(t, g, 3)

	g = restoreRoundTrip(t, g, true)
	me := g.Seats[0]
	acLands(g, me.ID, 2, "Forest")
	p := latestPickTarget(g, me.ID)
	if d := game.PickTargetDivideForEffect(p); d == nil || d.TotalFor(0) != 3 {
		t.Fatalf("the restored prompt divides %+v, want the 3 fixed at announce", d)
	}
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, cardRefs(a, b), map[uuid.UUID]int{a: 3, b: 2}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a 5-point split on the restored table: %v, want ErrInvalidParam", err)
	}
	b17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{a: 1, b: 2}, a, b)
	passPriorityAroundTable(t, g)
	if d1, d2 := e2Card(t, g, a).DamageMarked, e2Card(t, g, b).DamageMarked; d1 != 1 || d2 != 2 {
		t.Errorf("Ureni dealt %d/%d after the restore, want 1/2", d1, d2)
	}
}

// Mausoleum Turnkey's "of an opponent's choice": the controller's
// "which opponent" question and that opponent's pick are each a restore
// point, and the card the opponent picks comes back.
func TestMausoleumTurnkeyOpponentLegIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	weak, strong := uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: weak, Name: "Weak", TypeLine: "Creature — Test", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: strong, Name: "Strong", TypeLine: "Creature — Test", Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Mausoleum Turnkey", "Creature — Ogre Rogue", otMausoleumTurnkey, nil)
	passPriorityAroundTable(t, g)
	otOptionPick(t, g, me.ID)

	g = restoreRoundTrip(t, g, true)
	me, a := g.Seats[0], g.Seats[1]
	otChooseOpponent(t, g, me.ID, a.ID)
	otPickFor(t, g, a.ID, me.ID)

	g = restoreRoundTrip(t, g, true)
	me, a = g.Seats[0], g.Seats[1]
	p := otPickFor(t, g, a.ID, me.ID)
	if !hasID(p.PickTargetCards, weak) || !hasID(p.PickTargetCards, strong) {
		t.Fatalf("the restored opponent's pick offers %v, want both creature cards", p.PickTargetCards)
	}
	otAnswer(t, g, a.ID, otCard(weak))
	passPriorityAroundTable(t, g)
	if !inHandOf(me, weak) || inHandOf(me, strong) {
		t.Fatal("the controller's hand does not hold exactly the opponent's pick")
	}
}
