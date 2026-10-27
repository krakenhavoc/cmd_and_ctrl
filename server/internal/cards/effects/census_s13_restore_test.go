package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// census_s13_restore_test.go — #4361 S13, owner decision D7: a cast or an
// activation parked before anything is paid is written into the snapshot
// (game/parked_announcement.go). Each test parks a real catalog card's
// announcement, restores the table STRICTLY (the boot path) while the
// prompt is open, answers it on the restored table and checks the
// announcement completed there:
//
//   - the CR 903.9b question a cost asks a commander's owner: Ninja of
//     the Deep Hours' ninjutsu returning an attacking commander (an
//     activation), and Daze's alternative cost returning a commander
//     Island (a cast);
//   - "of an opponent's choice": Nova Pentacle's controller choosing the
//     opponent and that opponent's pick (an activation, both prompts
//     restored), and Evangelize's sole opponent's pick (a cast).

// s13StrictRestore checks the open prompt was written under `marker` and
// restores the table strictly through JSON.
func s13StrictRestore(t *testing.T, g *game.Game, marker string) *game.Game {
	t.Helper()
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a table waiting on a parked announcement is not a restore point: %+v", snap.Continuations)
	}
	found := false
	for _, c := range snap.PendingChoices {
		if c.PickThen == marker {
			found = true
		}
	}
	if !found {
		t.Fatalf("no prompt in the snapshot is marked %q", marker)
	}
	return restoreRoundTrip(t, g, true)
}

// Ninja of the Deep Hours' ninjutsu spent on an attacking commander:
// restored while the commander's owner is asked, the "command zone"
// answer re-makes the activation on the restored table — the commander
// goes to the command zone and the Ninja enters attacking the player the
// commander was attacking.
func TestANinjutsuCommanderQuestionIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmd := pushBattlefieldCardWithTimestamp(g, costCommander(me.ID, "My Commander"))
	ninja := pushNinjaToHand(me, "Ninja of the Deep Hours", "Creature — Human Ninja", ninjaOfTheDeepHoursOracle, 2, 2)
	declareAttack(t, g, opp.ID, cmd)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{ReturnIDs: []uuid.UUID{cmd}}); err != nil {
		t.Fatalf("ninjutsu: %v", err)
	}

	g = s13StrictRestore(t, g, game.CostCommanderAnnounceKey)
	me, opp = g.Seats[0], g.Seats[1]
	if got := g.PendingChoices[0].CommanderHeadedFor(); got != game.ZoneHand {
		t.Errorf("restored CommanderHeadedFor() = %q, want %q", got, game.ZoneHand)
	}
	if !g.Battlefield.Contains(cmd) || len(g.StackMeta) != 0 {
		t.Fatal("ninjutsu was paid before the restored question was answered")
	}
	answerCostCommander(t, g, me.ID, true)
	if !me.Command.Contains(cmd) || me.Hand.Contains(cmd) {
		t.Fatal("the returned commander did not take the command zone on the restored table")
	}
	passPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(ninja)
	if !ok || !g.Battlefield.Contains(ninja) {
		t.Fatal("the ninja did not enter on the restored table")
	}
	if c.AttackingTarget != opp.ID {
		t.Errorf("the ninja is attacking %s, want %s", c.AttackingTarget, opp.ID)
	}
}

// Daze cast by returning a commander Island: restored while its owner is
// asked, the "no" re-makes the cast — the Island goes to its owner's hand
// and Daze is on the stack, its mana cost not paid.
func TestADazeCommanderQuestionIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	them := g.Seats[1]
	toMainForCost(t, g)
	victim := castCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	daze := handCardFull(them, "Daze", "Instant", "{1}{U}", dazeOracle, []string{"U"})
	island := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Commander Island", TypeLine: "Legendary Land — Island",
		Owner: them.ID, Controller: them.ID, IsCommander: true,
	})
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority to the responder: %v", err)
	}
	if err := g.CastSpell(them.ID, daze, game.CastSpellParams{
		Strict:          true,
		AlternativeCost: "return",
		AltCostIDs:      []uuid.UUID{island},
		Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("casting Daze: %v", err)
	}
	if g.Stack.Contains(daze) || !g.Battlefield.Contains(island) {
		t.Fatal("Daze was cast before the commander's owner answered")
	}

	g = s13StrictRestore(t, g, game.CostCommanderAnnounceKey)
	them = g.Seats[1]
	answerCostCommander(t, g, them.ID, false)
	if !them.Hand.Contains(island) || g.Battlefield.Contains(island) {
		t.Fatal("the declined commander Island did not go to its owner's hand on the restored table")
	}
	if !g.Stack.Contains(daze) {
		t.Fatal("Daze is not on the stack after the restored answer")
	}
}

// Nova Pentacle: restored while the controller is asked which opponent
// chooses, and again while that opponent picks. The pick completes the
// activation on the restored table: the Pentacle is tapped and its
// ability is on the stack with the opponent recorded as the chooser.
func TestANovaPentacleOpponentChoiceIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, a := g.Seats[0], g.Seats[1]
	pentacle := pushCatalogPermanent(g, me.ID, "Nova Pentacle", "Artifact", otNovaPentacle, false)
	mine := pr7Creature(g, me.ID, "My Bear", 2, "G")
	pr7Creature(g, a.ID, "Their Bear", 2, "G")
	if err := g.ActivateCatalogAbility(me.ID, pentacle, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}

	// The controller's "which opponent" question.
	g = s13StrictRestore(t, g, "option-pick/opponent-target-chooser")
	me, a = g.Seats[0], g.Seats[1]
	if findBattlefieldCardForTest(g, pentacle).Tapped {
		t.Fatal("the Pentacle was tapped before the target was chosen")
	}
	otChooseOpponent(t, g, me.ID, a.ID)

	// The opponent's pick.
	otPickFor(t, g, a.ID, me.ID)
	g = s13StrictRestore(t, g, game.OpponentTargetAnnounceKey)
	me, a = g.Seats[0], g.Seats[1]
	otPickFor(t, g, a.ID, me.ID)
	otAnswer(t, g, a.ID, otCard(mine))
	item := otItemFrom(g, pentacle)
	if item == nil {
		t.Fatal("no ability on the stack after the restored pick")
	}
	if len(item.TargetChoosers) != 1 || item.TargetChoosers[0].Player != a.ID {
		t.Fatalf("TargetChoosers %+v, want the chosen opponent", item.TargetChoosers)
	}
	if len(item.Targets) != 1 || item.Targets[0].ID != mine {
		t.Fatalf("targets %+v, want the opponent's pick", item.Targets)
	}
	if !findBattlefieldCardForTest(g, pentacle).Tapped {
		t.Fatal("the cost was not paid once the restored announcement completed")
	}
}

// Evangelize: with one opponent able to choose, that opponent is asked at
// once. Restored while they pick, the answer casts the spell on the
// restored table, and it resolves: the caster gains control of the pick.
func TestAnEvangelizeOpponentPickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	a := g.Seats[1]
	theirs := pr7Creature(g, a.ID, "Their Bear", 2, "G")
	evangelize := castCatalogSpell(t, g, "Evangelize", "Sorcery", otEvangelize, nil)
	if g.Stack.Contains(evangelize) {
		t.Fatal("Evangelize was cast before the opponent chose")
	}

	g = s13StrictRestore(t, g, game.OpponentTargetAnnounceKey)
	me := g.Seats[0]
	a = g.Seats[1]
	otPickFor(t, g, a.ID, me.ID)
	otAnswer(t, g, a.ID, otCard(theirs))
	if !g.Stack.Contains(evangelize) {
		t.Fatal("Evangelize is not on the stack after the restored pick")
	}
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardForTest(g, theirs); c == nil || c.Controller != me.ID {
		t.Fatal("the caster did not gain control of the opponent's pick on the restored table")
	}
}
