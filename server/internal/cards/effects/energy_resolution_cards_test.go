package effects

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// energy_resolution_cards_test.go — ADR 0129 PR 3: the cards that pay
// energy as a spell or ability resolves (CR 118.12, 118.12a).

const (
	erThrivingRhino  = "d8841f3a-f3ff-42ca-89a8-0cc1c3ca6a6c"
	erAetherChaser   = "8aab0b55-2779-43c7-a3ee-151b8e5c72f3"
	erRiparianTiger  = "dbb6b7f0-33fd-4530-a621-114148fb3929"
	erLathnuHellion  = "6dd71453-63ec-4f4c-87d2-34d207398b9a"
	erHarnessed      = "d8dd7e7f-0053-44bb-bac0-1a08843d4c49"
	erGalvanic       = "ef0f06f1-3991-4897-bf62-d4cab7ec79a0"
	erDieYoung       = "4bcc54d1-8ab6-4ad3-91b6-97fb02b88cc0"
	erGreenbelt      = "c63a17b8-d183-48c7-bb04-edbfffba1e03"
	erHexgoldSlith   = "0a467e75-f68d-4ad0-85ff-b8516de8d5bc"
	erVoltaicBrawler = "ab7609dd-f4c0-4636-8177-7a926c01e470"
)

// erBatchA is every card this PR registers, each Full.
var erBatchA = []string{
	"30ec09e4-82bf-4a6b-b6c8-ff267447e0a9", // Thriving Grubs
	"3f523026-beda-4549-96c0-e92108170261", // Thriving Ibex
	"0bb8500a-bc29-43f1-ab22-2b8adeb9c99d", // Thriving Rats
	erThrivingRhino,
	"d268143f-27f0-4f14-8cd3-481923aaff6d", // Thriving Skyclaw
	"85dda096-48b2-414e-8c4d-4ffa6e5dac11", // Thriving Turtle
	"4fa99dc1-c2d2-4b21-8511-e3d86626609d", // Scrapper Champion
	erAetherChaser,
	"be89a032-8ade-4048-b95d-76e958dca330", // Aether Herder
	"9373a176-a5e9-4fdc-906b-bc6aef657e5b", // Aether Inspector
	"73757d44-7889-416b-94d2-e730e601ace3", // Aether Poisoner
	"f7ed301a-dce7-49d0-a68a-d3a099201f65", // Aether Swooper
	erRiparianTiger,
	"8517e5f4-89c5-4879-9b4f-b35e268798df", // Aetherstream Leopard
	erVoltaicBrawler,
	erHexgoldSlith,
	erLathnuHellion,
	erHarnessed,
	erGalvanic,
	erDieYoung,
	erGreenbelt,
}

func TestEnergyResolutionCardsAreFull(t *testing.T) {
	for _, oracle := range erBatchA {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
	}
}

// energyPrompt is the open energy pay-unless owed by `p`, or nil.
func energyPrompt(g *game.Game, p uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == p {
			if a := c.PayAction(); a != nil && a.Kind == game.PayActionEnergy {
				return c
			}
		}
	}
	return nil
}

// payAmountPrompt is the open pay_amount owed by `p`, or nil.
func payAmountPrompt(g *game.Game, p uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayAmount && c.Chooser == p {
			return c
		}
	}
	return nil
}

// attackAndAnswerEnergy attacks with `attacker`, resolves its attack
// trigger, and answers the energy prompt with `pay`.
func attackAndAnswerEnergy(t *testing.T, g *game.Game, attacker uuid.UUID, pay bool) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	declareAttack(t, g, opp.ID, attacker)
	passPriorityAroundTable(t, g)
	c := energyPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no energy prompt after the attack trigger resolved")
	}
	if err := g.ResolvePayUnless(c.ID, me.ID, pay); err != nil {
		t.Fatalf("answer: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// The Thriving cycle: paid, the counter lands on the attacker and the
// energy comes off; declined, nothing changes.
func TestThrivingRhinoPaysForACounter(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		rhino := pushDiesCreatureForTest(g, me.ID, "Thriving Rhino", erThrivingRhino, "Creature — Rhino", 2, 3)
		setEnergy(t, g, me, 3)
		attackAndAnswerEnergy(t, g, rhino, pay)
		c := apaLive(g, rhino)
		wantCounters, wantEnergy := 0, 3
		if pay {
			wantCounters, wantEnergy = 1, 1
		}
		if got := c.Counters[game.CounterPlusOne]; got != wantCounters {
			t.Errorf("pay=%v: +1/+1 counters = %d, want %d", pay, got, wantCounters)
		}
		if energyOf(me) != wantEnergy {
			t.Errorf("pay=%v: energy = %d, want %d", pay, energyOf(me), wantEnergy)
		}
	}
}

// Short of the energy, "Pay" is the decline (CR 118.3): no counter, and
// nothing comes off.
func TestThrivingRhinoShortPaysNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rhino := pushDiesCreatureForTest(g, me.ID, "Thriving Rhino", erThrivingRhino, "Creature — Rhino", 2, 3)
	setEnergy(t, g, me, 1)
	attackAndAnswerEnergy(t, g, rhino, true)
	if got := apaLive(g, rhino).Counters[game.CounterPlusOne]; got != 0 {
		t.Errorf("+1/+1 counters = %d, want 0", got)
	}
	if energyOf(me) != 1 {
		t.Errorf("energy = %d, want 1", energyOf(me))
	}
}

// The prompt holds the declare attackers step: the table cannot reach
// blockers with the counter unbought.
func TestEnergyAttackPromptHoldsTheStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rhino := pushDiesCreatureForTest(g, me.ID, "Thriving Rhino", erThrivingRhino, "Creature — Rhino", 2, 3)
	setEnergy(t, g, me, 2)
	declareAttack(t, g, opp.ID, rhino)
	passPriorityAroundTable(t, g)
	c := energyPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no energy prompt")
	}
	if !g.ChoicePromptBlocksTable(c) {
		t.Error("the attack trigger's energy prompt does not hold the step")
	}
}

// Aether Chaser's payment makes a Servo.
func TestAetherChaserPaysForAServo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	chaser := pushDiesCreatureForTest(g, me.ID, "Aether Chaser", erAetherChaser, "Creature — Human Artificer", 2, 1)
	setEnergy(t, g, me, 2)
	attackAndAnswerEnergy(t, g, chaser, true)
	servos := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Servo" && c.Controller == me.ID {
			servos++
		}
	}
	if servos != 1 || energyOf(me) != 0 {
		t.Errorf("servos %d, energy %d; want 1 and 0", servos, energyOf(me))
	}
}

// Riparian Tiger gets +2/+2; Voltaic Brawler +1/+1 and trample.
func TestEnergyAttackPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tiger := pushDiesCreatureForTest(g, me.ID, "Riparian Tiger", erRiparianTiger, "Creature — Cat", 4, 4)
	setEnergy(t, g, me, 2)
	attackAndAnswerEnergy(t, g, tiger, true)
	if c := apaLive(g, tiger); c.CurrentPower() != 6 || c.CurrentToughness() != 6 {
		t.Errorf("Tiger is %d/%d, want 6/6", c.CurrentPower(), c.CurrentToughness())
	}

	g = newCatalogGame(t)
	me = g.Seats[0]
	brawler := pushDiesCreatureForTest(g, me.ID, "Voltaic Brawler", erVoltaicBrawler, "Creature — Human Warrior", 3, 2)
	setEnergy(t, g, me, 1)
	attackAndAnswerEnergy(t, g, brawler, true)
	c := apaLive(g, brawler)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 3 || !game.HasKeyword(c, "trample") {
		t.Errorf("Brawler is %d/%d trample=%v, want 4/3 with trample", c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(c, "trample"))
	}
}

// Lathnu Hellion: at the end step, declining sacrifices it, paying
// keeps it.
func TestLathnuHellionEndStep(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		hellion := pushDiesCreatureForTest(g, me.ID, "Lathnu Hellion", erLathnuHellion, "Creature — Hellion", 4, 4)
		setEnergy(t, g, me, 2)
		advanceTo(t, g, game.StepEnd)
		passPriorityAroundTable(t, g)
		c := energyPrompt(g, me.ID)
		if c == nil {
			t.Fatalf("pay=%v: no prompt at the end step", pay)
		}
		if err := g.ResolvePayUnless(c.ID, me.ID, pay); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		alive := findBattlefieldCardByID(g, hellion) != nil
		if alive != pay {
			t.Errorf("pay=%v: Hellion on the battlefield = %v", pay, alive)
		}
	}
}

// Harnessed Lightning: three energy, then the stepper's goal is the
// target's toughness; paying it kills the creature.
func TestHarnessedLightningPaysForDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	setEnergy(t, g, me, 1)
	castCatalogSpell(t, g, "Harnessed Lightning", "Instant", erHarnessed,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)
	c := payAmountPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no pay_amount prompt")
	}
	if !reflect.DeepEqual(*c.PayAmount, game.PayAmountPrompt{Min: 0, Max: 4, Goal: 3, Unit: game.PayAmountDamage, Resource: game.PayResourceEnergy}) {
		t.Errorf("prompt = %+v", *c.PayAmount)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 3); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldCardByID(g, victim) != nil {
		t.Error("the 3/3 survived three damage")
	}
	if energyOf(me) != 1 {
		t.Errorf("energy = %d, want 1", energyOf(me))
	}
}

// Galvanic Discharge paying nothing deals nothing, and keeps the energy.
func TestGalvanicDischargePayingNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushVanillaCreature(g, opp.ID, "Victim", 2, 2)
	castCatalogSpell(t, g, "Galvanic Discharge", "Instant", erGalvanic,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)
	c := payAmountPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no pay_amount prompt")
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 0); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if v := apaLive(g, victim); v == nil || v.DamageMarked != 0 {
		t.Errorf("victim = %+v, want undamaged", v)
	}
	if energyOf(me) != 3 {
		t.Errorf("energy = %d, want 3", energyOf(me))
	}
}

// Die Young: -X/-X for the X paid.
func TestDieYoungShrinksByTheAmountPaid(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushVanillaCreature(g, opp.ID, "Victim", 5, 5)
	castCatalogSpell(t, g, "Die Young", "Sorcery", erDieYoung,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)
	c := payAmountPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no pay_amount prompt")
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 2); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if v := apaLive(g, victim); v == nil || v.CurrentPower() != 3 || v.CurrentToughness() != 3 {
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
		g := newCatalogGame(t)
		me := g.Seats[0]
		setEnergy(t, g, me, tc.have)
		id := castAndResolveCreature(t, g, "Greenbelt Rampager", "Creature — Elephant", erGreenbelt)
		passPriorityAroundTable(t, g)
		if got := findBattlefieldCardByID(g, id) != nil; got != tc.stays {
			t.Errorf("have %d: on the battlefield = %v, want %v", tc.have, got, tc.stays)
		}
		if energyOf(me) != tc.wantEnergy {
			t.Errorf("have %d: energy = %d, want %d", tc.have, energyOf(me), tc.wantEnergy)
		}
	}
}
