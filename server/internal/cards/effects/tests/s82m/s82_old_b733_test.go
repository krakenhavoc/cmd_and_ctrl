package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b733_test.go — Portal Second Age, batch 733 (#5625): other
// creatures. Each test plays the real catalog entry.

const (
	oldB733Giant     = "4f94ce97-5b2d-42fb-ac8e-c64075fe1af2"
	oldB733Engine    = "128b6751-4b47-4d7d-895e-c5d85776f6c4"
	oldB733Prowling  = "6279078a-5c77-4a9e-8e9b-e99310363b6f"
	oldB733Wurm      = "48408f60-56bc-402f-ba19-c9dc53836765"
	oldB733Ox        = "1871af55-f7cd-4b8e-affb-5d010df811b1"
	oldB733Cavalier  = "f8273146-f2e5-40ec-ba96-c0a4192caec8"
	oldB733Riders    = "1816b87c-73ba-4bfc-a51e-c982381c067d"
	oldB733Zealot    = "0d0d97dd-2653-4d37-b475-ac3b50c5ee53"
	oldB733BearType  = "Creature — Bear"
	oldB733BlackType = "Creature — Nightstalker"
)

func TestOldB733EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB733Giant: "Cunning Giant", oldB733Engine: "Nightstalker Engine",
		oldB733Prowling: "Prowling Nightstalker", oldB733Wurm: "Deathcoil Wurm",
		oldB733Ox: "Ironhoof Ox", oldB733Cavalier: "Alaborn Cavalier",
		oldB733Riders: "Norwood Riders", oldB733Zealot: "Alaborn Zealot",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v; want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB733CunningGiantAssignsItsDamageToACreatureWhenUnblocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	giant := effectstest.B12Push(g, me.ID, "Cunning Giant", "Creature — Giant", oldB733Giant, 4, 4)
	bear := effectstest.B12Creature(g, them.ID, "Their Bear", oldB733BearType, 2, 2)
	effectstest.AttackIntoBlocks(t, g, them.ID, giant)
	effectstest.LockInBlocks(t, g)
	effectstest.AdvanceToStepInTurn(t, g, game.StepCombatDamage)
	var choice *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceDamageAssignment {
			choice = c
		}
	}
	if choice == nil || choice.Chooser != me.ID {
		t.Fatalf("pending choice = %+v, want the Giant's controller asked for its damage", choice)
	}
	if err := g.ResolveDamageAssignment(choice.ID, me.ID, []game.DamageAssignmentEntry{{BlockerID: bear, Amount: 4}}, 0); err != nil {
		t.Fatalf("all of it to the Bear: %v", err)
	}
	if onBattlefield(g, bear) {
		t.Error("the Bear took 4 and is still on the battlefield")
	}
	if them.Life != game.StartingLife {
		t.Errorf("the defender lost %d life, want none", game.StartingLife-them.Life)
	}
}

func TestOldB733NightstalkerEnginePowerCountsYourGraveyardCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	engine := effectstest.B12Push(g, me.ID, "Nightstalker Engine", oldB733BlackType, oldB733Engine, 0, 3)
	if p, tough := effectstest.EffectivePTOf(t, g, engine); p != 0 || tough != 3 {
		t.Fatalf("Engine is %d/%d with an empty graveyard, want 0/3", p, tough)
	}
	bury := func(owner *game.Player, typeLine string) {
		id := effectstest.B12Push(g, owner.ID, "Dead Card", typeLine, "", 2, 2)
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: owner.ID}, id); err != nil {
			t.Fatalf("MoveCardByID: %v", err)
		}
	}
	bury(me, oldB733BearType)
	bury(me, oldB733BearType)
	bury(me, "Artifact")
	bury(opp, oldB733BearType)
	if live := effectstest.ApaLive(g, engine).Effective(); live.Power != 2 || live.Toughness != 3 {
		t.Errorf("Engine is %d/%d with two of your creature cards in the graveyard, want 2/3", live.Power, live.Toughness)
	}
}

func TestOldB733ProwlingNightstalkerCanOnlyBeBlockedByBlackCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	stalker := effectstest.B12Push(g, me.ID, "Prowling Nightstalker", oldB733BlackType, oldB733Prowling, 2, 2)
	bear := effectstest.B12Creature(g, them.ID, "Their Bear", oldB733BearType, 2, 2)
	black := effectstest.B12Push(g, them.ID, "Black Bear", oldB733BlackType, "", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == black {
				g.Battlefield.Cards[i].Colors = []string{"B"}
			}
		}
	})
	effectstest.AttackIntoBlocks(t, g, them.ID, stalker)
	if err := g.Clone().DeclareBlocker(bear, stalker); err == nil {
		t.Error("a non-black creature blocked it")
	}
	if err := g.DeclareBlocker(black, stalker); err != nil {
		t.Errorf("a black creature was refused: %v", err)
	}
}

func TestOldB733DeathcoilWurmMayAssignAsThoughUnblocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wurm := effectstest.B12Push(g, me.ID, "Deathcoil Wurm", "Creature — Wurm", oldB733Wurm, 7, 6)
	if !effectstest.ApaLive(g, wurm).Effective().AssignsAsThoughUnblocked {
		t.Error("the Wurm lacks the as-though-unblocked flag")
	}
	other := effectstest.B12Creature(g, me.ID, "Bear", oldB733BearType, 2, 2)
	if effectstest.ApaLive(g, other).Effective().AssignsAsThoughUnblocked {
		t.Error("another creature has the flag")
	}
}

func TestOldB733OxAndRidersCantBeBlockedByMoreThanOneCreature(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{{"Ironhoof Ox", oldB733Ox}, {"Norwood Riders", oldB733Riders}} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, them := effectstest.SeatsForCombat(g)
		attacker := effectstest.B12Push(g, me.ID, tc.name, "Creature — Ox", tc.oracle, 4, 4)
		plain := effectstest.B12Creature(g, me.ID, "Plain Bear", oldB733BearType, 2, 2)
		b1 := effectstest.B12Creature(g, them.ID, "Blocker 1", oldB733BearType, 2, 2)
		b2 := effectstest.B12Creature(g, them.ID, "Blocker 2", oldB733BearType, 2, 2)
		effectstest.AdvanceToStepInTurn(t, g, game.StepDeclareAttackers)
		for _, a := range []uuid.UUID{attacker, plain} {
			if err := g.DeclareAttacker(a, them.ID); err != nil {
				t.Fatalf("DeclareAttacker: %v", err)
			}
		}
		effectstest.AdvanceToStepInTurn(t, g, game.StepDeclareBlockers)
		g2 := g.Clone()
		if err := g2.DeclareBlocker(b1, attacker); err != nil {
			t.Fatalf("%s: first block refused: %v", tc.name, err)
		}
		err := g2.DeclareBlocker(b2, attacker)
		if err == nil {
			err = g2.FinishBlocks(them.ID)
		}
		if err == nil {
			t.Errorf("%s was double-blocked", tc.name)
		}
		g3 := g.Clone()
		if err := g3.DeclareBlocker(b1, plain); err != nil {
			t.Fatalf("plain block refused: %v", err)
		}
		if err := g3.DeclareBlocker(b2, plain); err != nil {
			t.Errorf("a plain creature's second blocker was refused: %v", err)
		}
	}
}

func TestOldB733AlabornCavalierMayTapATargetWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	cav := effectstest.B12Push(g, me.ID, "Alaborn Cavalier", "Creature — Human Knight", oldB733Cavalier, 2, 2)
	bear := effectstest.B12Creature(g, them.ID, "Their Bear", oldB733BearType, 2, 2)
	effectstest.DeclareAttack(t, g, them.ID, cav)
	effectstest.AnswerTriggerTargets(t, g, me.ID, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, bear) {
		t.Error("the target was not tapped")
	}
}

func TestOldB733AlabornZealotDestroysItselfAndTheCreatureItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Creature(g, me.ID, "Attacking Bear", oldB733BearType, 2, 2)
	zealot := effectstest.B12Push(g, them.ID, "Alaborn Zealot", "Creature — Human Soldier", oldB733Zealot, 1, 1)
	bystander := effectstest.B12Creature(g, them.ID, "Bystander", oldB733BearType, 2, 2)
	effectstest.AttackIntoBlocks(t, g, them.ID, bear)
	if err := g.DeclareBlocker(zealot, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if onBattlefield(g, bear) {
		t.Error("the blocked creature survived")
	}
	if onBattlefield(g, zealot) {
		t.Error("the Zealot survived")
	}
	if !onBattlefield(g, bystander) {
		t.Error("an unrelated creature was destroyed")
	}
}
