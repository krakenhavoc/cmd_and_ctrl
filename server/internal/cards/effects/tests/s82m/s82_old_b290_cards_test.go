package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b290_cards_test.go — Shadows over Innistrad creatures, S82 OLD
// batch 290 (#5543), played through the real catalog entries.

const (
	oldB290MissionariesOracle = "5b356a01-900c-4fe5-93bd-45629528c484"
	oldB290KillerOracle       = "c1801b5c-1a87-46d0-83d3-e8dfd94211f4"
	oldB290HunterOracle       = "c99834dc-755f-407c-ad0e-550f8ae12198"
	oldB290ArsonistsOracle    = "cde622fa-a634-409d-962c-1e4c5ddf9b3b"
	oldB290GloomwidowOracle   = "e5139376-cdb1-4f1e-b417-b956d6b713b9"
	oldB290PacifistOracle     = "e7d6ef32-616f-4c21-bcdd-467ce8296c31"
)

var (
	oldB290Killer = effectstest.IwCard{Oracle: oldB290KillerOracle, Front: "Convicted Killer", FrontType: "Creature — Human Werewolf", Cost: "{2}{R}",
		FrontText: effectstest.IwFrontText, Fp: "2", Ft: "2", Back: "Branded Howler", BackText: effectstest.IwBackText, Bp: "4", Bt: "4", Colors: []string{"R"}}
	oldB290Hunter = effectstest.IwCard{Oracle: oldB290HunterOracle, Front: "Solitary Hunter", FrontType: "Creature — Human Warrior Werewolf", Cost: "{3}{G}",
		FrontText: effectstest.IwFrontText, Fp: "3", Ft: "4", Back: "One of the Pack", BackText: effectstest.IwBackText, Bp: "5", Bt: "6", Colors: []string{"G"}}
	oldB290Arsonists = effectstest.IwCard{Oracle: oldB290ArsonistsOracle, Front: "Gatstaf Arsonists", FrontType: "Creature — Human Werewolf", Cost: "{4}{R}",
		FrontText: effectstest.IwFrontText, Fp: "5", Ft: "4", Back: "Gatstaf Ravagers", BackText: "Menace\n" + effectstest.IwBackText, Bp: "6", Bt: "5", Colors: []string{"R"}, Keywords: []string{"Menace"}}
	oldB290Pacifist = effectstest.IwCard{Oracle: oldB290PacifistOracle, Front: "Lambholt Pacifist", FrontType: "Creature — Human Shaman Werewolf", Cost: "{1}{G}",
		FrontText: "This creature can't attack unless you control a creature with power 4 or greater.\n" + effectstest.IwFrontText, Fp: "3", Ft: "3",
		Back: "Lambholt Butcher", BackText: effectstest.IwBackText, Bp: "4", Bt: "4", Colors: []string{"G"}}
	oldB290Missionaries = effectstest.IwCard{Oracle: oldB290MissionariesOracle, Front: "Avacynian Missionaries", FrontType: "Creature — Human Cleric", Cost: "{3}{W}",
		FrontText: "At the beginning of your end step, if this creature is equipped, transform it.", Fp: "3", Ft: "3",
		Back: "Lunarch Inquisitors", BackText: "When this creature transforms into Lunarch Inquisitors, you may exile another target creature until this creature leaves the battlefield.", Bp: "4", Bt: "4", Colors: []string{"W"}}
)

func TestOldB290EveryCardIsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB290MissionariesOracle: "Avacynian Missionaries", oldB290MissionariesOracle + "#1": "Lunarch Inquisitors",
		oldB290KillerOracle: "Convicted Killer", oldB290KillerOracle + "#1": "Branded Howler",
		oldB290HunterOracle: "Solitary Hunter", oldB290HunterOracle + "#1": "One of the Pack",
		oldB290ArsonistsOracle: "Gatstaf Arsonists", oldB290ArsonistsOracle + "#1": "Gatstaf Ravagers",
		oldB290GloomwidowOracle: "Gloomwidow",
		oldB290PacifistOracle:   "Lambholt Pacifist", oldB290PacifistOracle + "#1": "Lambholt Butcher",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func TestOldB290WerewolvesTransformAndBack(t *testing.T) {
	for _, w := range []effectstest.IwCard{oldB290Killer, oldB290Hunter, oldB290Arsonists, oldB290Pacifist} {
		g, me := effectstest.IwTable(t)
		id := effectstest.IwPut(t, g, w, me)
		effectstest.IwNextUpkeep(t, g, map[int]int{g.Turn.ActiveSeat: 1})
		if effectstest.IwFace(g, id) != 0 {
			t.Fatalf("%s transformed although a spell was cast last turn", w.Front)
		}
		effectstest.IwNextUpkeep(t, g, nil)
		if effectstest.IwFace(g, id) != 1 {
			t.Fatalf("%s did not transform after a quiet turn", w.Front)
		}
		if w.Front == "Gatstaf Arsonists" {
			effectstest.AssertKeywords(t, g, id, "menace")
		}
		effectstest.IwNextUpkeep(t, g, map[int]int{g.Turn.ActiveSeat: 2})
		if effectstest.IwFace(g, id) != 0 {
			t.Errorf("%s did not turn back after a player cast two spells", w.Back)
		}
	}
}

func TestOldB290PacifistCantAttackWithoutAPowerFourCreature(t *testing.T) {
	g, me := effectstest.IwTable(t)
	id := effectstest.IwPut(t, g, oldB290Pacifist, me)
	effectstest.AssertRestrictions(t, g, id, game.CantAttack)
	effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 3, 3)
	effectstest.AssertRestrictions(t, g, id, game.CantAttack)
	effectstest.B12Creature(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Their Big", "Creature — Bear", 5, 5)
	effectstest.AssertRestrictions(t, g, id, game.CantAttack) // only your own creatures count
	effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 4, 4)
	effectstest.AssertRestrictions(t, g, id, 0)
}

func TestOldB290MissionariesTransformOnlyWhenEquipped(t *testing.T) {
	g, me := effectstest.IwTable(t)
	id := effectstest.IwPut(t, g, oldB290Missionaries, me)
	seat := g.Turn.ActiveSeat
	effectstest.AdvanceToEndStepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IwFace(g, id) != 0 {
		t.Fatal("the Missionaries transformed without Equipment")
	}

	g, me = effectstest.IwTable(t)
	id = effectstest.IwPut(t, g, oldB290Missionaries, me)
	sword := effectstest.SeedEquipment(g, me.ID, "Sword", "")
	effectstest.AttachForTest(t, g, sword, id)
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IwFace(g, id) != 1 {
		t.Fatal("the equipped Missionaries did not transform at the end step")
	}
}

func TestOldB290InquisitorsExileAnotherCreatureUntilTheyLeave(t *testing.T) {
	g, me := effectstest.IwTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := effectstest.IwPut(t, g, oldB290Missionaries, me)
	effectstest.IwTransform(t, g, id)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, id) {
		t.Error("the Inquisitors may target themselves")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) {
		t.Fatal("the target was not exiled")
	}
	effectstest.KillCreature(t, g, me.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.BattlefieldIDsNamed(g, "Their Bear")) != 1 {
		t.Error("the exiled creature did not return when the Inquisitors left")
	}
}

func TestOldB290InquisitorsMayDeclineTheExile(t *testing.T) {
	g, me := effectstest.IwTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := effectstest.IwPut(t, g, oldB290Missionaries, me)
	effectstest.IwTransform(t, g, id)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(theirs) {
		t.Error("the target was exiled although the player declined")
	}
}

func TestOldB290GloomwidowBlocksOnlyFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	widow := effectstest.B12Push(g, them.ID, "Gloomwidow", "Creature — Spider", oldB290GloomwidowOracle, 3, 3)
	if !effectstest.HasEffectiveKeyword(t, g, widow, "reach") {
		t.Fatal("Gloomwidow does not have reach")
	}
	ground := effectstest.B12Creature(g, me.ID, "Ground attacker", "Creature — Bear", 2, 2)
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flying attacker", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AttackIntoBlocks(t, g, them.ID, ground, flyer)
	if effectstest.BlockRefusal(t, g, widow, ground) == "" {
		t.Error("Gloomwidow blocked a creature without flying")
	}
	if got := effectstest.BlockRefusal(t, g, widow, flyer); got != "" {
		t.Errorf("Gloomwidow was refused a flyer: %q", got)
	}
}
