package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// energy_resolution_cards_b_test.go — ADR 0129 PR 3, the follow-up card
// batch on energy-resolution-payments.

// erPlay is how a test answers whatever the table asks `me` while it
// settles: the target to pick, whether to pay a fixed energy amount,
// the amount for a pay_amount and the option for an option pick.
type erPlay struct {
	target uuid.UUID
	pay    bool
	amount int
	option int
}

// settleEnergy passes priority until the table is quiet, answering each
// prompt owed by `me` from `play`. It returns how many energy prompts
// (either kind) it answered.
func settleEnergy(t *testing.T, g *game.Game, me uuid.UUID, play erPlay) int {
	t.Helper()
	answered := 0
	for i := 0; i < 20; i++ {
		passPriorityAroundTable(t, g)
		switch {
		case latestPickTarget(g, me) != nil:
			answerTriggerTargets(t, g, me, play.target)
		case energyPrompt(g, me) != nil:
			if err := g.ResolvePayUnless(energyPrompt(g, me).ID, me, play.pay); err != nil {
				t.Fatalf("answer energy: %v", err)
			}
			answered++
		case payAmountPrompt(g, me) != nil:
			if err := g.ResolvePayAmount(payAmountPrompt(g, me).ID, me, play.amount); err != nil {
				t.Fatalf("answer amount: %v", err)
			}
			answered++
		case latestOptionPickFor(g, me) != nil:
			answerOptionPick(t, g, me, play.option)
		default:
			return answered
		}
	}
	t.Fatal("the table did not settle")
	return answered
}

var erBatchB = []string{
	"c1a3eeef-7ea5-4d58-9e6f-91bbf1c2cc8c", // Consul's Shieldguard
	"748729bf-bfa0-4cbf-b643-ed5d7d007081", // Eddytrail Hawk
	"8bc30bd5-554e-4fcf-9dec-c7634921858a", // Smelted Chargebug
	"70a04e3f-e123-4925-9a02-9825488f1bc8", // Maulfist Doorbuster
	"06f09e5a-5cfb-437f-9dd1-1682d71f4e7f", // Aetherstorm Roc
	"ea34ec71-2f3b-4119-937d-4629cbf7ac5e", // Glint-Sleeve Siphoner
	"202091b2-9e08-47c1-9099-628b21d8ac12", // Electrozoa
	"b1bc8e58-ea4f-46a9-b833-d780cf6758b3", // Static Prison
	"b31026fe-faad-49b3-93b0-1324e32bb816", // Liberty Prime, Recharged
	"6f9390ec-08b5-4761-8cb1-5d60190b9661", // Robobrain War Mind
	"c5830515-9316-4b08-a990-3026ee15ffea", // Riddle Gate Gargoyle
	"10bca1b1-4f92-40e8-b9c0-d08c534903e1", // Cyclops Superconductor
	"7d05ffe2-e52b-42af-87ca-6aae488b9f41", // Behemoth of Vault 0
	"4f63d306-cc14-405c-8424-d8990c670c74", // Territorial Aetherkite
	"9b1f0780-072a-4fe6-91d8-e546a8aaea20", // Rampaging Aetherhood
	"810e1371-e028-47e7-b97e-a627a35932e0", // Localized Destruction
	"21efa6f2-3e7c-4576-b5a3-0b75b2843442", // Wrath of the Skies
	"c246cdb4-2fd8-487e-944e-3e52fbd1bbaa", // Confiscation Coup
	"4306a9e5-6845-4fd4-96ad-8888147612aa", // Jolted Awake
	"e4a85647-0b8c-40b2-a5d5-43c78877ce36", // Aether Spike
	"bbd569cc-bc21-46df-b8eb-5b5bcd8fe762", // Rush of Inspiration
	"2f82b232-21ff-44cc-bc35-0999a1d1c52f", // Lightning Runner
	"2c7818f4-a818-4603-9af1-05aee2d56ec8", // Assaultron Dominator
	"b304ac72-7f40-40de-b8d6-6392909b6029", // Guide of Souls
	"d0c78ee6-babb-410d-9d3a-0c1395596e56", // Sentry Bot
	"9e9cde32-0568-44bc-a00a-9709f15c1350", // Aether Refinery
	"370aced9-d8bc-4abe-b648-4c62b5aee5ba", // T-45 Power Armor
	"b592568b-11b0-4081-90a7-30cfb9c1ba80", // Suppression Ray
}

// Aether Refinery doubles energy its controller gets, its own tap
// included, and pays for an X/X Aetherborn.
func TestAetherRefinery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ref := pushPermanentForTest(g, me.ID, "Aether Refinery", "9e9cde32-0568-44bc-a00a-9709f15c1350", "Artifact")
	setEnergy(t, g, me, 0)
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(me.ID, game.CounterEnergy, 3); err != nil {
			t.Fatal(err)
		}
	})
	if energyOf(me) != 6 {
		t.Fatalf("getting 3 gave %d, want 6", energyOf(me))
	}
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.ActivateCatalogAbility(me.ID, ref, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	settleEnergy(t, g, me.ID, erPlay{amount: 5})
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Aetherborn" && c.CurrentPower() == 5 && c.CurrentToughness() == 5 {
			found = true
		}
	}
	if !found {
		t.Error("no 5/5 Aetherborn")
	}
	if energyOf(me) != 3 {
		t.Errorf("energy = %d, want 3 (6 + 2 - 5)", energyOf(me))
	}
}

// Suppression Ray taps the player's creatures and stuns as many of the
// ones it tapped as the energy paid.
func TestSuppressionRay(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushVanillaCreature(g, opp.ID, "A", 2, 2)
	b := pushVanillaCreature(g, opp.ID, "B", 2, 2)
	setEnergy(t, g, me, 1)
	castCatalogSpell(t, g, "Suppression Ray", "Sorcery", "b592568b-11b0-4081-90a7-30cfb9c1ba80",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if c := payAmountPrompt(g, me.ID); c == nil || c.PayAmount.Goal != 1 {
		t.Fatalf("prompt = %+v, want goal 1 (the energy held)", c)
	}
	if err := g.ResolvePayAmount(payAmountPrompt(g, me.ID).ID, me.ID, 1); err != nil {
		t.Fatal(err)
	}
	pick := choiceOfKindFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick of the tapped creatures")
	}
	if err := g.ResolveTheirPermanents(pick.ID, me.ID, []uuid.UUID{a}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !apaLive(g, a).Tapped || !apaLive(g, b).Tapped {
		t.Error("the creatures were not tapped")
	}
	if apaLive(g, a).Counters[game.CounterStun] != 1 || apaLive(g, b).Counters[game.CounterStun] != 0 {
		t.Error("the stun counter is not on the one chosen")
	}
}

// choiceOfKindFor is the first open prompt owed by `p` that is neither
// an energy prompt nor a pay_amount.
func choiceOfKindFor(g *game.Game, p uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == p && c.Kind != game.PendingChoicePayUnless && c.Kind != game.PendingChoicePayAmount {
			return c
		}
	}
	return nil
}

// T-45 Power Armor: paying at upkeep untaps the equipped creature and
// gives it the chosen counter.
func TestT45PowerArmor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	armor := pushPermanentForTest(g, me.ID, "T-45 Power Armor", "370aced9-d8bc-4abe-b648-4c62b5aee5ba", "Artifact — Equipment")
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(armor, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
			t.Fatal(err)
		}
	})
	setEnergy(t, g, me, 1)
	apaLive(g, bear).Tapped = true
	for g.Turn.ActiveSeat != 0 || g.Turn.Step != game.StepUpkeep || g.Turn.Seq < 2 {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if !apaLive(g, bear).Tapped {
		t.Fatal("the equipped creature untapped in its untap step")
	}
	settleEnergy(t, g, me.ID, erPlay{pay: true, option: 2})
	c := apaLive(g, bear)
	if c.Tapped || c.Counters[game.CounterLifelink] != 1 {
		t.Errorf("tapped %v, lifelink counters %d; want untapped with one", c.Tapped, c.Counters[game.CounterLifelink])
	}
	if c.CurrentPower() != 5 {
		t.Errorf("power = %d, want 5", c.CurrentPower())
	}
}

func TestEnergyResolutionBatchBCardsAreFull(t *testing.T) {
	for _, oracle := range erBatchB {
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

// Consul's Shieldguard: another attacking creature, chosen as the
// trigger goes on the stack, gains indestructible when the energy is
// paid.
func TestConsulsShieldguardGrantsIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guard := pushDiesCreatureForTest(g, me.ID, "Consul's Shieldguard", "c1a3eeef-7ea5-4d58-9e6f-91bbf1c2cc8c", "Creature — Dwarf Soldier", 3, 4)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	setEnergy(t, g, me, 1)
	declareAttack(t, g, opp.ID, guard, bear)
	if n := settleEnergy(t, g, me.ID, erPlay{target: bear, pay: true}); n != 1 {
		t.Fatalf("energy prompts answered = %d, want 1", n)
	}
	if !game.HasKeyword(apaLive(g, bear), "indestructible") {
		t.Error("the other attacker did not gain indestructible")
	}
	if game.HasKeyword(apaLive(g, guard), "indestructible") {
		t.Error("the Shieldguard itself gained indestructible")
	}
	if energyOf(me) != 0 {
		t.Errorf("energy = %d, want 0", energyOf(me))
	}
}

// Maulfist Doorbuster: the target can't block this turn.
func TestMaulfistDoorbusterStopsABlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	door := pushDiesCreatureForTest(g, me.ID, "Maulfist Doorbuster", "70a04e3f-e123-4925-9a02-9825488f1bc8", "Creature — Human Warrior", 4, 2)
	wall := pushVanillaCreature(g, opp.ID, "Wall", 0, 5)
	setEnergy(t, g, me, 1)
	declareAttack(t, g, opp.ID, door)
	settleEnergy(t, g, me.ID, erPlay{target: wall, pay: true})
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, door); err == nil {
		t.Error("the Wall blocked after the Doorbuster's energy was paid")
	}
}

// Electrozoa: declining at the first main phase taps it.
func TestElectrozoaTapsUnlessPaid(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		zoa := pushDiesCreatureForTest(g, me.ID, "Electrozoa", "202091b2-9e08-47c1-9099-628b21d8ac12", "Creature — Jellyfish", 3, 1)
		setEnergy(t, g, me, 1)
		advanceTo(t, g, game.StepPrecombatMain)
		settleEnergy(t, g, me.ID, erPlay{pay: pay})
		if got := apaLive(g, zoa).Tapped; got == pay {
			t.Errorf("pay=%v: tapped = %v", pay, got)
		}
	}
}

// Static Prison: the exile and the energy; declining its upkeep
// sacrifices it and the exiled card comes back.
func TestStaticPrison(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	castCatalogSpell(t, g, "Static Prison", "Enchantment", "b1bc8e58-ea4f-46a9-b833-d780cf6758b3", nil)
	settleEnergy(t, g, me.ID, erPlay{target: victim})
	if findBattlefieldCardByID(g, victim) != nil {
		t.Fatal("the target was not exiled")
	}
	if energyOf(me) != 2 {
		t.Errorf("energy = %d, want 2", energyOf(me))
	}
	// Next turn of mine: the first main phase asks for {E}. Decline.
	for g.Turn.ActiveSeat != 0 || g.Turn.Step != game.StepPrecombatMain || g.Turn.Seq < 2 {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if energyPrompt(g, me.ID) != nil {
			break
		}
	}
	if n := settleEnergy(t, g, me.ID, erPlay{pay: false}); n != 1 {
		t.Fatalf("energy prompts at the first main phase = %d, want 1 (turn %d seat %d step %s)", n, g.Turn.Seq, g.Turn.ActiveSeat, g.Turn.Step)
	}
	prisons, victims := 0, 0
	for _, c := range g.Battlefield.Cards {
		switch c.Name {
		case "Static Prison":
			prisons++
		case "Victim":
			victims++
		}
	}
	if prisons != 0 || victims != 1 {
		t.Errorf("prisons %d, victims %d on the battlefield; want 0 and 1 (sacrificing the Prison returns the card)", prisons, victims)
	}
}

// Cyclops Superconductor: dies, pay {E}{E}{E}, and the reflexive
// trigger deals its power to the target.
func TestCyclopsSuperconductorDiesAndShoots(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cyclops := pushDiesCreatureForTest(g, me.ID, "Cyclops Superconductor", "10bca1b1-4f92-40e8-b9c0-d08c534903e1", "Creature — Cyclops Wizard", 2, 2)
	victim := pushVanillaCreature(g, opp.ID, "Victim", 2, 2)
	setEnergy(t, g, me, 3)
	advanceTo(t, g, game.StepPrecombatMain)
	killCreature(t, g, me.ID, cyclops)
	settleEnergy(t, g, me.ID, erPlay{target: victim, pay: true})
	if findBattlefieldCardByID(g, victim) != nil {
		t.Error("the 2/2 survived two damage")
	}
	if energyOf(me) != 0 {
		t.Errorf("energy = %d, want 0", energyOf(me))
	}
}

// Behemoth of Vault 0: dies targeting a permanent; paying its mana
// value destroys it.
func TestBehemothOfVault0DestroysForItsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	behemoth := pushDiesCreatureForTest(g, me.ID, "Behemoth of Vault 0", "7d05ffe2-e52b-42af-87ca-6aae488b9f41", "Artifact Creature — Robot", 6, 6)
	relic := pushTypedCard(g, opp.ID, "Relic", "Artifact", "{3}")
	setEnergy(t, g, me, 3)
	advanceTo(t, g, game.StepPrecombatMain)
	killCreature(t, g, me.ID, behemoth)
	settleEnergy(t, g, me.ID, erPlay{target: relic, pay: true})
	if findBattlefieldCardByID(g, relic) != nil {
		t.Error("the {3} artifact was not destroyed")
	}
	if energyOf(me) != 0 {
		t.Errorf("energy = %d, want 0", energyOf(me))
	}
}

// Territorial Aetherkite: pay two, and two damage to each other creature.
func TestTerritorialAetherkiteSweeps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushVanillaCreature(g, opp.ID, "Victim", 2, 2)
	survivor := pushVanillaCreature(g, opp.ID, "Survivor", 3, 3)
	kite := castAndResolveCreature(t, g, "Territorial Aetherkite", "Creature — Cat Dragon", "4f63d306-cc14-405c-8424-d8990c670c74")
	settleEnergy(t, g, me.ID, erPlay{amount: 2})
	if findBattlefieldCardByID(g, victim) != nil {
		t.Error("the 2/2 survived")
	}
	if s := apaLive(g, survivor); s == nil || s.DamageMarked != 2 {
		t.Errorf("survivor = %+v, want 2 damage", s)
	}
	if k := apaLive(g, kite); k == nil || k.DamageMarked != 0 {
		t.Error("the Aetherkite damaged itself")
	}
}

// Rampaging Aetherhood: energy equal to its power, then that many
// counters for the energy paid.
func TestRampagingAetherhoodGrows(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hydra := pushDiesCreatureForTest(g, me.ID, "Rampaging Aetherhood", "9b1f0780-072a-4fe6-91d8-e546a8aaea20", "Creature — Snake Hydra", 4, 4)
	for g.Turn.ActiveSeat != 0 || g.Turn.Step != game.StepUpkeep || g.Turn.Seq < 2 {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	c := func() *game.PendingChoice { passPriorityAroundTable(t, g); return payAmountPrompt(g, me.ID) }()
	if c == nil || c.PayAmount.Max != 4 || c.PayAmount.Goal != 4 {
		t.Fatalf("prompt = %+v, want max 4 goal 4", c)
	}
	settleEnergy(t, g, me.ID, erPlay{amount: 3})
	if got := apaLive(g, hydra).Counters[game.CounterPlusOne]; got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3", got)
	}
	if energyOf(me) != 1 {
		t.Errorf("energy = %d, want 1", energyOf(me))
	}
}

// Localized Destruction: the creatures you control with the paid power
// survive; everything else dies.
func TestLocalizedDestruction(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine2 := pushVanillaCreature(g, me.ID, "Mine Two", 2, 2)
	mine3 := pushVanillaCreature(g, me.ID, "Mine Three", 3, 3)
	theirs2 := pushVanillaCreature(g, opp.ID, "Theirs Two", 2, 2)
	setEnergy(t, g, me, 1)
	castCatalogSpell(t, g, "Localized Destruction", "Sorcery", "810e1371-e028-47e7-b97e-a627a35932e0", nil)
	settleEnergy(t, g, me.ID, erPlay{amount: 2})
	if findBattlefieldCardByID(g, mine2) == nil {
		t.Error("my power-2 creature died")
	}
	if findBattlefieldCardByID(g, mine3) != nil || findBattlefieldCardByID(g, theirs2) != nil {
		t.Error("a creature that should have died survived")
	}
}

// Wrath of the Skies: X energy, pay 2, destroy mana value 2 or less.
func TestWrathOfTheSkies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cheap := pushTypedCard(g, opp.ID, "Cheap", "Artifact", "{2}")
	dear := pushTypedCard(g, opp.ID, "Dear", "Enchantment", "{3}")
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Wrath of the Skies", TypeLine: "Sorcery",
		OracleID: "21efa6f2-3e7c-4576-b5a3-0b75b2843442", ManaCost: "{X}{W}{W}", Owner: me.ID, Controller: me.ID})
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := payAmountPrompt(g, me.ID)
	if c == nil || c.PayAmount.Max != 3 {
		t.Fatalf("prompt = %+v, want max 3", c)
	}
	settleEnergy(t, g, me.ID, erPlay{amount: 2})
	if findBattlefieldCardByID(g, cheap) != nil {
		t.Error("the mana value 2 artifact survived")
	}
	if findBattlefieldCardByID(g, dear) == nil {
		t.Error("the mana value 3 enchantment died")
	}
}

// Confiscation Coup: four energy, pay the mana value, gain control.
func TestConfiscationCoup(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	relic := pushTypedCard(g, opp.ID, "Relic", "Artifact", "{4}")
	castCatalogSpell(t, g, "Confiscation Coup", "Sorcery", "c246cdb4-2fd8-487e-944e-3e52fbd1bbaa",
		[]game.TargetRef{{Kind: game.TargetCard, ID: relic}})
	settleEnergy(t, g, me.ID, erPlay{pay: true})
	if c := findBattlefieldCardByID(g, relic); c == nil || c.Controller != me.ID {
		t.Errorf("relic = %+v, want it under my control", c)
	}
	if energyOf(me) != 0 {
		t.Errorf("energy = %d, want 0", energyOf(me))
	}
}

// Jolted Awake with no target still gets the energy.
func TestJoltedAwakeWithNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Jolted Awake", "Sorcery", "4306a9e5-6845-4fd4-96ad-8888147612aa", nil)
	if n := settleEnergy(t, g, me.ID, erPlay{}); n != 0 {
		t.Errorf("energy prompts = %d, want 0", n)
	}
	if energyOf(me) != 2 {
		t.Errorf("energy = %d, want 2", energyOf(me))
	}
}

// Rush of Inspiration: declining the {E}{E} discards at random.
func TestRushOfInspirationDiscardsUnlessPaid(t *testing.T) {
	for _, tc := range []struct {
		energy int
		pay    bool
		delta  int
	}{{2, true, 2}, {2, false, 1}} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		setEnergy(t, g, me, tc.energy)
		castCatalogSpell(t, g, "Rush of Inspiration", "Instant", "bbd569cc-bc21-46df-b8eb-5b5bcd8fe762", nil)
		hand := me.Hand.Size()
		settleEnergy(t, g, me.ID, erPlay{pay: tc.pay})
		if got := me.Hand.Size() - hand; got != tc.delta {
			t.Errorf("pay=%v: hand delta %d, want %d", tc.pay, got, tc.delta)
		}
	}
}

// Assaultron Dominator: an attacking artifact creature gets the chosen
// counter.
func TestAssaultronDominatorChoosesACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dom := pushDiesCreatureForTest(g, me.ID, "Assaultron Dominator", "2c7818f4-a818-4603-9af1-05aee2d56ec8", "Artifact Creature — Robot", 2, 2)
	setEnergy(t, g, me, 1)
	declareAttack(t, g, opp.ID, dom)
	settleEnergy(t, g, me.ID, erPlay{pay: true, option: 2})
	if got := apaLive(g, dom).Counters[game.CounterTrample]; got != 1 {
		t.Errorf("trample counters = %d, want 1", got)
	}
}

// Guide of Souls: paying makes the target an Angel with counters.
func TestGuideOfSoulsMakesAnAngel(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushDiesCreatureForTest(g, me.ID, "Guide of Souls", "b304ac72-7f40-40de-b8d6-6392909b6029", "Creature — Human Cleric", 1, 2)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	setEnergy(t, g, me, 3)
	declareAttack(t, g, opp.ID, bear)
	settleEnergy(t, g, me.ID, erPlay{target: bear, pay: true})
	c := apaLive(g, bear)
	if c.Counters[game.CounterPlusOne] != 2 || c.Counters[game.CounterFlying] != 1 {
		t.Errorf("counters = %v, want two +1/+1 and a flying", c.Counters)
	}
	if !HasSubtype("Angel")(g, me.ID, *c) {
		t.Error("the creature is not an Angel")
	}
}
