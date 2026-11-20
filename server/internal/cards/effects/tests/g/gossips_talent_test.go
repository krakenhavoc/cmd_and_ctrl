package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const gossipsTalentOracle = "96e6d220-87db-4dab-abaa-e80c9577d89a"

func gtTalent(g *game.Game, owner uuid.UUID, level int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Gossip's Talent", TypeLine: "Enchantment — Class", OracleID: gossipsTalentOracle,
		Owner: owner, Controller: owner, ClassLevel: level,
	})
}

func TestGossipsTalentIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(gossipsTalentOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Gossip's Talent: %+v %v", spec, ok)
	}
}

func TestGossipsTalentLevelOneSurveilsWhenACreatureEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gtTalent(g, me.ID, 1)
	effectstest.SeedLibrary(me, "a", "b", "c")
	effectstest.AdvanceToMain(t, g)
	effectstest.WickCastRat(t, g, me)
	if c := effectstest.SurveilChoiceFor(g, me.ID); c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("surveil prompt %+v, want surveil 1", c)
	}
}

func TestGossipsTalentLevelTwoMakesAnAttackingCreatureWithPowerThreeOrLessUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gtTalent(g, me.ID, 2)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 4, 4)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	effectstest.DeclareAttack(t, g, opp.ID, small, big)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if len(pick.PickTargetCards) != 1 || pick.PickTargetCards[0] != small {
		t.Fatalf("offered %v, want only the attacking creature with power 3 or less", pick.PickTargetCards)
	}
	effectstest.AnswerPickTarget(t, g, small)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, small, game.CantBeBlocked)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, small); err == nil {
		t.Error("the unblockable creature was blocked")
	}
}

func TestGossipsTalentLevelThreeMayBlinkACreatureThatHitAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gtTalent(g, me.ID, 3)
	hitter := effectstest.B12Creature(g, me.ID, "Hitter", "Creature — Bear", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, hitter)
	effectstest.AnswerPickTarget(t, g, hitter) // level 2 is on too
	effectstest.PassPriorityAroundTable(t, g)
	life := opp.Life
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if life-opp.Life != 3 {
		t.Errorf("the defender lost %d, want 3", life-opp.Life)
	}
	if _, same := effectstest.BattlefieldCard(g, hitter); same {
		t.Error("the creature was not exiled and returned")
	}
	back := effectstest.BattlefieldIDsNamed(g, "Hitter")
	if len(back) != 1 {
		t.Fatalf("the battlefield has %d Hitters, want the one that came back", len(back))
	}
	if c, _ := effectstest.BattlefieldCard(g, back[0]); c.AttackingTarget != uuid.Nil {
		t.Error("the returned creature is still attacking")
	}
}

func TestGossipsTalentLevelThreeDecliningLeavesItAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gtTalent(g, me.ID, 3)
	hitter := effectstest.B12Creature(g, me.ID, "Hitter", "Creature — Bear", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, hitter)
	effectstest.AnswerPickTarget(t, g, hitter) // level 2 is on too
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if _, same := effectstest.BattlefieldCard(g, hitter); !same {
		t.Error("declining still exiled the creature")
	}
}

func TestGossipsTalentLowerLevelsLackTheHigherAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gtTalent(g, me.ID, 1)
	hitter := effectstest.B12Creature(g, me.ID, "Hitter", "Creature — Bear", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, hitter)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("the level-2 ability worked at level 1")
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if _, same := effectstest.BattlefieldCard(g, hitter); !same {
		t.Error("the level-3 ability worked at level 1")
	}
}
