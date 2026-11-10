package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const gossipsTalentOracle = "96e6d220-87db-4dab-abaa-e80c9577d89a"

func gtTalent(g *game.Game, owner uuid.UUID, level int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Gossip's Talent", TypeLine: "Enchantment — Class", OracleID: gossipsTalentOracle,
		Owner: owner, Controller: owner, ClassLevel: level,
	})
}

func TestGossipsTalentIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(gossipsTalentOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Gossip's Talent: %+v %v", spec, ok)
	}
}

func TestGossipsTalentLevelOneSurveilsWhenACreatureEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gtTalent(g, me.ID, 1)
	seedLibrary(me, "a", "b", "c")
	advanceToMain(t, g)
	wickCastRat(t, g, me)
	if c := surveilChoiceFor(g, me.ID); c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("surveil prompt %+v, want surveil 1", c)
	}
}

func TestGossipsTalentLevelTwoMakesAnAttackingCreatureWithPowerThreeOrLessUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gtTalent(g, me.ID, 2)
	small := b12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	big := b12Creature(g, me.ID, "Big", "Creature — Bear", 4, 4)
	blocker := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	declareAttack(t, g, opp.ID, small, big)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if len(pick.PickTargetCards) != 1 || pick.PickTargetCards[0] != small {
		t.Fatalf("offered %v, want only the attacking creature with power 3 or less", pick.PickTargetCards)
	}
	answerPickTarget(t, g, small)
	passPriorityAroundTable(t, g)
	assertRestrictions(t, g, small, game.CantBeBlocked)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, small); err == nil {
		t.Error("the unblockable creature was blocked")
	}
}

func TestGossipsTalentLevelThreeMayBlinkACreatureThatHitAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gtTalent(g, me.ID, 3)
	hitter := b12Creature(g, me.ID, "Hitter", "Creature — Bear", 3, 3)
	declareAttack(t, g, opp.ID, hitter)
	answerPickTarget(t, g, hitter) // level 2 is on too
	passPriorityAroundTable(t, g)
	life := opp.Life
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if life-opp.Life != 3 {
		t.Errorf("the defender lost %d, want 3", life-opp.Life)
	}
	if _, same := battlefieldCard(g, hitter); same {
		t.Error("the creature was not exiled and returned")
	}
	back := battlefieldIDsNamed(g, "Hitter")
	if len(back) != 1 {
		t.Fatalf("the battlefield has %d Hitters, want the one that came back", len(back))
	}
	if c, _ := battlefieldCard(g, back[0]); c.AttackingTarget != uuid.Nil {
		t.Error("the returned creature is still attacking")
	}
}

func TestGossipsTalentLevelThreeDecliningLeavesItAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gtTalent(g, me.ID, 3)
	hitter := b12Creature(g, me.ID, "Hitter", "Creature — Bear", 3, 3)
	declareAttack(t, g, opp.ID, hitter)
	answerPickTarget(t, g, hitter) // level 2 is on too
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if _, same := battlefieldCard(g, hitter); !same {
		t.Error("declining still exiled the creature")
	}
}

func TestGossipsTalentLowerLevelsLackTheHigherAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gtTalent(g, me.ID, 1)
	hitter := b12Creature(g, me.ID, "Hitter", "Creature — Bear", 3, 3)
	declareAttack(t, g, opp.ID, hitter)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("the level-2 ability worked at level 1")
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if _, same := battlefieldCard(g, hitter); !same {
		t.Error("the level-3 ability worked at level 1")
	}
}
