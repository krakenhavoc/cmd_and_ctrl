package testsg

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// hobgoblin_bandit_lord_test.go — #811, CR 603.10 / CR 608.2h.
//
// "The number of Goblins that entered the battlefield under your
// control this turn" is a tally of ENTRY EVENTS, judged by each object
// as it entered. The three cases below are the three the live-board
// scan got wrong, and every one of them was wrong in a direction the
// printed card does not allow.
//
// The anthem and the ordinary count are pinned in
// TestB34HobgoblinBanditLordPumpsGoblinsAndCountsThisTurnsEntries.

const hobgoblinMaskwoodNexusOracle = "9b2cdbed-c733-409b-b0e4-2c8960c25111"

// A Goblin TOKEN that entered and has since died still counts, as
// printed. This is the case the card is played for — a Goblin board
// trading in combat — and the one the old scan lost outright: CR
// 704.5d takes the dead token out of the graveyard, so looking the
// entry up found no card at all.
func TestHobgoblinBanditLordCountsAGoblinThatHasSinceDied(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lord := effectstest.B12Push(g, me.ID, "Hobgoblin Bandit Lord", "Creature — Goblin Rogue", effectstest.B34HobgoblinBanditLordOracle, 2, 3)
	effectstest.AdvanceToMain(t, g)

	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1); err != nil {
			t.Fatalf("create the Goblin: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	token := effectstest.BattlefieldIDNamed(g, me.ID, "Goblin")
	if token == uuid.Nil {
		t.Fatal("the Goblin token did not enter")
	}
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(token); err != nil {
			t.Fatalf("kill the Goblin: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token) {
		t.Fatal("the Goblin token is still on the battlefield")
	}

	if got := effectstest.HobgoblinLordDamage(t, g, me, opp, lord); got != 1 {
		t.Errorf("a Goblin that entered and died: %d damage, want 1", got)
	}
}

// A creature that entered as something else and has since BECOME a
// Goblin does not count. Maskwood Nexus makes every creature its
// controller controls every creature type, and the printed card is
// about what entered.
func TestHobgoblinBanditLordIgnoresACreatureThatBecameAGoblinLater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lord := effectstest.B12Push(g, me.ID, "Hobgoblin Bandit Lord", "Creature — Goblin Rogue", effectstest.B34HobgoblinBanditLordOracle, 2, 3)
	effectstest.AdvanceToMain(t, g)

	// A Soldier enters first, then the Nexus arrives.
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, effects.B36WhiteHumanSoldierToken(), 1); err != nil {
			t.Fatalf("create the Soldier: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	soldier := effectstest.BattlefieldIDNamed(g, me.ID, "Human Soldier")
	if soldier == uuid.Nil {
		t.Fatal("the Soldier token did not enter")
	}
	effectstest.CastCatalogSpell(t, g, "Maskwood Nexus", "Artifact", hobgoblinMaskwoodNexusOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// The Nexus grants every creature type through the changeling
	// marker rather than by rewriting the type line, so the read the
	// old scan made — Card.HasSubtype, which folds that marker in —
	// now says the Soldier IS a Goblin.
	if nexused := effectstest.LayeredCard(t, g, soldier); !nexused.HasSubtype("Goblin") {
		t.Fatal("the Nexus did not make the Soldier a Goblin — the premise of this test is gone")
	}

	if got := effectstest.HobgoblinLordDamage(t, g, me, opp, lord); got != 0 {
		t.Errorf("a Soldier that became a Goblin afterwards: %d damage, want 0", got)
	}
}

// A Goblin that entered under an OPPONENT's control and that you have
// since gained control of does not count for you: "entered the
// battlefield under your control" is a fact about the entry, and the
// old scan read Card.Controller as it is now.
func TestHobgoblinBanditLordIgnoresAStolenGoblin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lord := effectstest.B12Push(g, me.ID, "Hobgoblin Bandit Lord", "Creature — Goblin Rogue", effectstest.B34HobgoblinBanditLordOracle, 2, 3)
	effectstest.AdvanceToMain(t, g)

	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 1); err != nil {
			t.Fatalf("create their Goblin: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	theirs := effectstest.BattlefieldIDNamed(g, opp.ID, "Goblin")
	if theirs == uuid.Nil {
		t.Fatal("the opponent's Goblin token did not enter")
	}
	effectstest.StealForTest(t, g, theirs, me.ID)
	if got := effectstest.LayeredCard(t, g, theirs).Controller; got != me.ID {
		t.Fatalf("the Goblin is controlled by %s, want %s", got, me.ID)
	}

	if got := effectstest.HobgoblinLordDamage(t, g, me, opp, lord); got != 0 {
		t.Errorf("a Goblin stolen after it entered: %d damage, want 0", got)
	}
}
