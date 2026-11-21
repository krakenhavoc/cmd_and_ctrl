package testsds

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// erBatchA is every card this PR registers, each Full.
var erBatchA = []string{
	"30ec09e4-82bf-4a6b-b6c8-ff267447e0a9", // Thriving Grubs
	"3f523026-beda-4549-96c0-e92108170261", // Thriving Ibex
	"0bb8500a-bc29-43f1-ab22-2b8adeb9c99d", // Thriving Rats
	effectstest.ErThrivingRhino,
	"d268143f-27f0-4f14-8cd3-481923aaff6d", // Thriving Skyclaw
	"85dda096-48b2-414e-8c4d-4ffa6e5dac11", // Thriving Turtle
	"4fa99dc1-c2d2-4b21-8511-e3d86626609d", // Scrapper Champion
	effectstest.ErAetherChaser,
	"be89a032-8ade-4048-b95d-76e958dca330", // Aether Herder
	"9373a176-a5e9-4fdc-906b-bc6aef657e5b", // Aether Inspector
	"73757d44-7889-416b-94d2-e730e601ace3", // Aether Poisoner
	"f7ed301a-dce7-49d0-a68a-d3a099201f65", // Aether Swooper
	effectstest.ErRiparianTiger,
	"8517e5f4-89c5-4879-9b4f-b35e268798df", // Aetherstream Leopard
	effectstest.ErVoltaicBrawler,
	effectstest.ErHexgoldSlith,
	effectstest.ErLathnuHellion,
	effectstest.ErHarnessed,
	effectstest.ErGalvanic,
	effectstest.ErDieYoung,
	effectstest.ErGreenbelt,
}

func TestEnergyResolutionCardsAreFull(t *testing.T) {
	for _, oracle := range erBatchA {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
	}
}

// attackAndAnswerEnergy attacks with `attacker`, resolves its attack
// trigger, and answers the energy prompt with `pay`.
func attackAndAnswerEnergy(t *testing.T, g *game.Game, attacker uuid.UUID, pay bool) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.EnergyPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no energy prompt after the attack trigger resolved")
	}
	if err := g.ResolvePayUnless(c.ID, me.ID, pay); err != nil {
		t.Fatalf("answer: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// The Thriving cycle: paid, the counter lands on the attacker and the
// energy comes off; declined, nothing changes.
func TestThrivingRhinoPaysForACounter(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		rhino := effectstest.PushDiesCreatureForTest(g, me.ID, "Thriving Rhino", effectstest.ErThrivingRhino, "Creature — Rhino", 2, 3)
		effectstest.SetEnergy(t, g, me, 3)
		attackAndAnswerEnergy(t, g, rhino, pay)
		c := effectstest.ApaLive(g, rhino)
		wantCounters, wantEnergy := 0, 3
		if pay {
			wantCounters, wantEnergy = 1, 1
		}
		if got := c.Counters[game.CounterPlusOne]; got != wantCounters {
			t.Errorf("pay=%v: +1/+1 counters = %d, want %d", pay, got, wantCounters)
		}
		if effectstest.EnergyOf(me) != wantEnergy {
			t.Errorf("pay=%v: energy = %d, want %d", pay, effectstest.EnergyOf(me), wantEnergy)
		}
	}
}

// Short of the energy, "Pay" is the decline (CR 118.3): no counter, and
// nothing comes off.
func TestThrivingRhinoShortPaysNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rhino := effectstest.PushDiesCreatureForTest(g, me.ID, "Thriving Rhino", effectstest.ErThrivingRhino, "Creature — Rhino", 2, 3)
	effectstest.SetEnergy(t, g, me, 1)
	attackAndAnswerEnergy(t, g, rhino, true)
	if got := effectstest.ApaLive(g, rhino).Counters[game.CounterPlusOne]; got != 0 {
		t.Errorf("+1/+1 counters = %d, want 0", got)
	}
	if effectstest.EnergyOf(me) != 1 {
		t.Errorf("energy = %d, want 1", effectstest.EnergyOf(me))
	}
}

// The prompt holds the declare attackers step: the table cannot reach
// blockers with the counter unbought.
func TestEnergyAttackPromptHoldsTheStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rhino := effectstest.PushDiesCreatureForTest(g, me.ID, "Thriving Rhino", effectstest.ErThrivingRhino, "Creature — Rhino", 2, 3)
	effectstest.SetEnergy(t, g, me, 2)
	effectstest.DeclareAttack(t, g, opp.ID, rhino)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.EnergyPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no energy prompt")
	}
	if !g.ChoicePromptBlocksTable(c) {
		t.Error("the attack trigger's energy prompt does not hold the step")
	}
}

// Aether Chaser's payment makes a Servo.
func TestAetherChaserPaysForAServo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chaser := effectstest.PushDiesCreatureForTest(g, me.ID, "Aether Chaser", effectstest.ErAetherChaser, "Creature — Human Artificer", 2, 1)
	effectstest.SetEnergy(t, g, me, 2)
	attackAndAnswerEnergy(t, g, chaser, true)
	servos := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Servo" && c.Controller == me.ID {
			servos++
		}
	}
	if servos != 1 || effectstest.EnergyOf(me) != 0 {
		t.Errorf("servos %d, energy %d; want 1 and 0", servos, effectstest.EnergyOf(me))
	}
}

// Riparian Tiger gets +2/+2; Voltaic Brawler +1/+1 and trample.
func TestEnergyAttackPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tiger := effectstest.PushDiesCreatureForTest(g, me.ID, "Riparian Tiger", effectstest.ErRiparianTiger, "Creature — Cat", 4, 4)
	effectstest.SetEnergy(t, g, me, 2)
	attackAndAnswerEnergy(t, g, tiger, true)
	if c := effectstest.ApaLive(g, tiger); c.CurrentPower() != 6 || c.CurrentToughness() != 6 {
		t.Errorf("Tiger is %d/%d, want 6/6", c.CurrentPower(), c.CurrentToughness())
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	brawler := effectstest.PushDiesCreatureForTest(g, me.ID, "Voltaic Brawler", effectstest.ErVoltaicBrawler, "Creature — Human Warrior", 3, 2)
	effectstest.SetEnergy(t, g, me, 1)
	attackAndAnswerEnergy(t, g, brawler, true)
	c := effectstest.ApaLive(g, brawler)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 3 || !game.HasKeyword(c, "trample") {
		t.Errorf("Brawler is %d/%d trample=%v, want 4/3 with trample", c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(c, "trample"))
	}
}

// Lathnu Hellion: at the end step, declining sacrifices it, paying
// keeps it.
func TestLathnuHellionEndStep(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		hellion := effectstest.PushDiesCreatureForTest(g, me.ID, "Lathnu Hellion", effectstest.ErLathnuHellion, "Creature — Hellion", 4, 4)
		effectstest.SetEnergy(t, g, me, 2)
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.EnergyPrompt(g, me.ID)
		if c == nil {
			t.Fatalf("pay=%v: no prompt at the end step", pay)
		}
		if err := g.ResolvePayUnless(c.ID, me.ID, pay); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		alive := effectstest.FindBattlefieldCardByID(g, hellion) != nil
		if alive != pay {
			t.Errorf("pay=%v: Hellion on the battlefield = %v", pay, alive)
		}
	}
}

// Harnessed Lightning: three energy, then the stepper's goal is the
// target's toughness; paying it kills the creature.
func TestHarnessedLightningPaysForDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	effectstest.SetEnergy(t, g, me, 1)
	effectstest.CastCatalogSpell(t, g, "Harnessed Lightning", "Instant", effectstest.ErHarnessed,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.PayAmountPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no pay_amount prompt")
	}
	if !reflect.DeepEqual(*c.PayAmount, game.PayAmountPrompt{Min: 0, Max: 4, Goal: 3, Unit: game.PayAmountDamage, Resource: game.PayResourceEnergy}) {
		t.Errorf("prompt = %+v", *c.PayAmount)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 3); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardByID(g, victim) != nil {
		t.Error("the 3/3 survived three damage")
	}
	if effectstest.EnergyOf(me) != 1 {
		t.Errorf("energy = %d, want 1", effectstest.EnergyOf(me))
	}
}

// Galvanic Discharge paying nothing deals nothing, and keeps the energy.
func TestGalvanicDischargePayingNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Galvanic Discharge", "Instant", effectstest.ErGalvanic,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.PayAmountPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no pay_amount prompt")
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 0); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if v := effectstest.ApaLive(g, victim); v == nil || v.DamageMarked != 0 {
		t.Errorf("victim = %+v, want undamaged", v)
	}
	if effectstest.EnergyOf(me) != 3 {
		t.Errorf("energy = %d, want 3", effectstest.EnergyOf(me))
	}
}

// Die Young: -X/-X for the X paid.
func TestDieYoungShrinksByTheAmountPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Die Young", "Sorcery", effectstest.ErDieYoung,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.PayAmountPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no pay_amount prompt")
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 2); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if v := effectstest.ApaLive(g, victim); v == nil || v.CurrentPower() != 3 || v.CurrentToughness() != 3 {
		t.Errorf("victim = %+v, want 3/3", v)
	}
}

// Greenbelt Rampager: with two energy it pays and stays; without, it
// returns to hand and its controller gets {E}.
func TestGreenbeltRampager(t *testing.T) {
	for _, tc := range []struct {
		have, wantEnergy int
		stays            bool
	}{{2, 0, true}, {1, 2, false}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.SetEnergy(t, g, me, tc.have)
		id := effectstest.CastAndResolveCreature(t, g, "Greenbelt Rampager", "Creature — Elephant", effectstest.ErGreenbelt)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.FindBattlefieldCardByID(g, id) != nil; got != tc.stays {
			t.Errorf("have %d: on the battlefield = %v, want %v", tc.have, got, tc.stays)
		}
		if effectstest.EnergyOf(me) != tc.wantEnergy {
			t.Errorf("have %d: energy = %d, want %d", tc.have, effectstest.EnergyOf(me), tc.wantEnergy)
		}
	}
}
