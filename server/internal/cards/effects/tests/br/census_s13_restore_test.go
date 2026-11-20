package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ninja of the Deep Hours' ninjutsu spent on an attacking commander:
// restored while the commander's owner is asked, the "command zone"
// answer re-makes the activation on the restored table — the commander
// goes to the command zone and the Ninja enters attacking the player the
// commander was attacking.
func TestANinjutsuCommanderQuestionIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmd := effectstest.PushBattlefieldCardWithTimestamp(g, effectstest.CostCommander(me.ID, "My Commander"))
	ninja := effectstest.PushNinjaToHand(me, "Ninja of the Deep Hours", "Creature — Human Ninja", effectstest.NinjaOfTheDeepHoursOracle, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, cmd)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{ReturnIDs: []uuid.UUID{cmd}}); err != nil {
		t.Fatalf("ninjutsu: %v", err)
	}

	g = effectstest.S13StrictRestore(t, g, game.CostCommanderAnnounceKey)
	me, opp = g.Seats[0], g.Seats[1]
	if got := g.PendingChoices[0].CommanderHeadedFor(); got != game.ZoneHand {
		t.Errorf("restored CommanderHeadedFor() = %q, want %q", got, game.ZoneHand)
	}
	if !g.Battlefield.Contains(cmd) || len(g.StackMeta) != 0 {
		t.Fatal("ninjutsu was paid before the restored question was answered")
	}
	effectstest.AnswerCostCommander(t, g, me.ID, true)
	if !me.Command.Contains(cmd) || me.Hand.Contains(cmd) {
		t.Fatal("the returned commander did not take the command zone on the restored table")
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	them := g.Seats[1]
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	daze := effectstest.HandCardFull(them, "Daze", "Instant", "{1}{U}", effectstest.DazeOracle, []string{"U"})
	island := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
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

	g = effectstest.S13StrictRestore(t, g, game.CostCommanderAnnounceKey)
	them = g.Seats[1]
	effectstest.AnswerCostCommander(t, g, them.ID, false)
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
	g := effectstest.NewCatalogGame(t)
	me, a := g.Seats[0], g.Seats[1]
	pentacle := effectstest.PushCatalogPermanent(g, me.ID, "Nova Pentacle", "Artifact", effectstest.OtNovaPentacle, false)
	mine := effectstest.Pr7Creature(g, me.ID, "My Bear", 2, "G")
	effectstest.Pr7Creature(g, a.ID, "Their Bear", 2, "G")
	if err := g.ActivateCatalogAbility(me.ID, pentacle, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}

	// The controller's "which opponent" question.
	g = effectstest.S13StrictRestore(t, g, "option-pick/opponent-target-chooser")
	me, a = g.Seats[0], g.Seats[1]
	if effectstest.FindBattlefieldCardForTest(g, pentacle).Tapped {
		t.Fatal("the Pentacle was tapped before the target was chosen")
	}
	effectstest.OtChooseOpponent(t, g, me.ID, a.ID)

	// The opponent's pick.
	effectstest.OtPickFor(t, g, a.ID, me.ID)
	g = effectstest.S13StrictRestore(t, g, game.OpponentTargetAnnounceKey)
	me, a = g.Seats[0], g.Seats[1]
	effectstest.OtPickFor(t, g, a.ID, me.ID)
	effectstest.OtAnswer(t, g, a.ID, effectstest.OtCard(mine))
	item := effectstest.OtItemFrom(g, pentacle)
	if item == nil {
		t.Fatal("no ability on the stack after the restored pick")
	}
	if len(item.TargetChoosers) != 1 || item.TargetChoosers[0].Player != a.ID {
		t.Fatalf("TargetChoosers %+v, want the chosen opponent", item.TargetChoosers)
	}
	if len(item.Targets) != 1 || item.Targets[0].ID != mine {
		t.Fatalf("targets %+v, want the opponent's pick", item.Targets)
	}
	if !effectstest.FindBattlefieldCardForTest(g, pentacle).Tapped {
		t.Fatal("the cost was not paid once the restored announcement completed")
	}
}

// Evangelize: with one opponent able to choose, that opponent is asked at
// once. Restored while they pick, the answer casts the spell on the
// restored table, and it resolves: the caster gains control of the pick.
func TestAnEvangelizeOpponentPickIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	a := g.Seats[1]
	theirs := effectstest.Pr7Creature(g, a.ID, "Their Bear", 2, "G")
	evangelize := effectstest.CastCatalogSpell(t, g, "Evangelize", "Sorcery", effectstest.OtEvangelize, nil)
	if g.Stack.Contains(evangelize) {
		t.Fatal("Evangelize was cast before the opponent chose")
	}

	g = effectstest.S13StrictRestore(t, g, game.OpponentTargetAnnounceKey)
	me := g.Seats[0]
	a = g.Seats[1]
	effectstest.OtPickFor(t, g, a.ID, me.ID)
	effectstest.OtAnswer(t, g, a.ID, effectstest.OtCard(theirs))
	if !g.Stack.Contains(evangelize) {
		t.Fatal("Evangelize is not on the stack after the restored pick")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, theirs); c == nil || c.Controller != me.ID {
		t.Fatal("the caster did not gain control of the opponent's pick on the restored table")
	}
}
