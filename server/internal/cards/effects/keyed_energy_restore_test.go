package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyed_energy_restore_test.go — #4289: the energy payments ("you may pay
// {E}{E}. If you do, ...", "sacrifice it unless you pay {E}") and the
// "you may pay {X}{R}" of Flameblast Dragon, Leyline Tyrant and
// Tilonalli's Summoner carry their consequence as a registered key, the
// asking item as data and, for X, the number as data, so a table waiting
// on the payer is a restore point. Each test takes a STRICT restore while
// the prompt is open and answers it on the restored table.

// openEnergyPrompt is the energy pay-unless `chooser` owes, failing the
// test when there is none, with the asking item it carries.
func openEnergyPrompt(t *testing.T, g *game.Game, chooser uuid.UUID) (*game.PendingChoice, *game.AskedItem) {
	t.Helper()
	c := energyPrompt(g, chooser)
	if c == nil {
		t.Fatalf("no energy prompt for %s", chooser)
	}
	if c.Asked == nil {
		t.Fatalf("the energy prompt (key %q) carries no asking item", c.PickThen)
	}
	if c.PickThen == "" {
		t.Fatalf("the energy prompt holds no key: it is a closure prompt")
	}
	return c, c.Asked
}

// The Thriving cycle shares one key for "put a +1/+1 counter on it". The
// prompt holds the declare attackers step across a restore, and a
// restored "yes" pays the energy and lands the counter on the attacker.
func TestThrivingRhinoEnergyPromptIsARestorePoint(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		rhino := pushDiesCreatureForTest(g, me.ID, "Thriving Rhino", erThrivingRhino, "Creature — Rhino", 2, 3)
		setEnergy(t, g, me, 3)
		declareAttack(t, g, opp.ID, rhino)
		passPriorityAroundTable(t, g)
		c, asked := openEnergyPrompt(t, g, me.ID)
		if c.PickThen != energyPlusOneCounterOnThis.Key() {
			t.Fatalf("the prompt's key is %q, want the shared +1/+1 counter key", c.PickThen)
		}
		if asked.SourceObject == nil || asked.SourceObject.ID != rhino {
			t.Fatalf("the asking item names object %+v, want the Rhino", asked.SourceObject)
		}

		g = restoreRoundTrip(t, g, true)
		me = g.Seats[0]
		c, _ = openEnergyPrompt(t, g, me.ID)
		if c.PickThen != energyPlusOneCounterOnThis.Key() || c.PayAction() == nil || c.PayAction().Count != 2 {
			t.Fatalf("restored prompt: key %q payment %+v, want the shared key and {E}{E}", c.PickThen, c.PayAction())
		}
		if c.OwedInStep.Step != game.StepDeclareAttackers {
			t.Errorf("restored prompt is owed in %+v, want the declare attackers step", c.OwedInStep)
		}
		if !g.ChoicePromptBlocksTable(c) {
			t.Error("the restored energy prompt does not hold the step")
		}
		if err := g.ResolvePayUnless(c.ID, me.ID, pay); err != nil {
			t.Fatalf("answer: %v", err)
		}
		passPriorityAroundTable(t, g)
		wantCounters, wantEnergy := 0, 3
		if pay {
			wantCounters, wantEnergy = 1, 1
		}
		if got := apaLive(g, rhino).Counters[game.CounterPlusOne]; got != wantCounters {
			t.Errorf("pay=%v: +1/+1 counters = %d, want %d", pay, got, wantCounters)
		}
		if energyOf(me) != wantEnergy {
			t.Errorf("pay=%v: energy = %d, want %d", pay, energyOf(me), wantEnergy)
		}
	}
}

// Lathnu Hellion, Liberty Prime and Static Prison share one decline,
// "sacrifice it unless you pay". The restored prompt sacrifices the
// permanent that triggered, and a permanent that left and came back is a
// new object and is not sacrificed (CR 400.7).
func TestLathnuHellionEnergyPromptIsARestorePoint(t *testing.T) {
	for _, tc := range []struct {
		name    string
		pay     bool
		flicker bool
		stays   bool
	}{
		{"paid", true, false, true},
		{"declined", false, false, false},
		{"declined after it left and came back", false, true, true},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		hellion := pushDiesCreatureForTest(g, me.ID, "Lathnu Hellion", erLathnuHellion, "Creature — Hellion", 4, 4)
		setEnergy(t, g, me, 2)
		advanceTo(t, g, game.StepEnd)
		passPriorityAroundTable(t, g)
		c, asked := openEnergyPrompt(t, g, me.ID)
		if c.PickThen != energySacrificeThis.Key() {
			t.Fatalf("%s: the prompt's key is %q, want the shared sacrifice key", tc.name, c.PickThen)
		}
		if asked.SourceObject == nil || asked.SourceObject.ID != hellion {
			t.Fatalf("%s: the asking item names object %+v, want the Hellion", tc.name, asked.SourceObject)
		}

		g = restoreRoundTrip(t, g, true)
		me = g.Seats[0]
		c, _ = openEnergyPrompt(t, g, me.ID)
		if c.PickThen != energySacrificeThis.Key() {
			t.Fatalf("%s: restored key %q", tc.name, c.PickThen)
		}
		if tc.flicker {
			g.BeginEventBatchForTest()
			flickerInResponse(t, g, hellion)
		}
		if err := g.ResolvePayUnless(c.ID, me.ID, tc.pay); err != nil {
			t.Fatalf("%s: answer: %v", tc.name, err)
		}
		if _, there := battlefieldCard(g, hellion); there != tc.stays {
			t.Errorf("%s: Lathnu Hellion on the battlefield = %v, want %v", tc.name, there, tc.stays)
		}
		if want := map[bool]int{true: 0, false: 2}[tc.pay]; energyOf(me) != want {
			t.Errorf("%s: energy = %d, want %d", tc.name, energyOf(me), want)
		}
	}
}

// Assaultron Dominator reads which creature attacked off the asking
// trigger's event, which the prompt carries: after a restore the counter
// still goes on that attacker.
func TestAssaultronDominatorEnergyPromptIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dom := pushDiesCreatureForTest(g, me.ID, "Assaultron Dominator", "2c7818f4-a818-4603-9af1-05aee2d56ec8", "Artifact Creature — Robot", 2, 2)
	setEnergy(t, g, me, 1)
	declareAttack(t, g, opp.ID, dom)
	passPriorityAroundTable(t, g)
	_, asked := openEnergyPrompt(t, g, me.ID)
	if asked.Trigger == nil || asked.Trigger.Event.CardID != dom {
		t.Fatalf("the asking item's trigger is %+v, want the Dominator's attack", asked.Trigger)
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c, asked := openEnergyPrompt(t, g, me.ID)
	if asked.Trigger == nil || asked.Trigger.Event.CardID != dom {
		t.Fatalf("restored: the asking item's trigger is %+v, want the Dominator's attack", asked.Trigger)
	}
	if err := g.ResolvePayUnless(c.ID, me.ID, true); err != nil {
		t.Fatalf("answer: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 2)
	if got := apaLive(g, dom).Counters[game.CounterTrample]; got != 1 {
		t.Errorf("trample counters = %d, want 1", got)
	}
	if energyOf(me) != 0 {
		t.Errorf("energy = %d, want 0", energyOf(me))
	}
}

// Confiscation Coup is a spell: the target it chose rides the prompt as
// the asking item's data, and a restored "yes" pays its mana value in
// energy and takes control of that permanent.
func TestConfiscationCoupEnergyPromptIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	relic := pushTypedCard(g, opp.ID, "Relic", "Artifact", "{4}")
	castCatalogSpell(t, g, "Confiscation Coup", "Sorcery", "c246cdb4-2fd8-487e-944e-3e52fbd1bbaa",
		[]game.TargetRef{{Kind: game.TargetCard, ID: relic}})
	passPriorityAroundTable(t, g)
	_, asked := openEnergyPrompt(t, g, me.ID)
	if len(asked.Targets) != 1 || asked.Targets[0].ID != relic {
		t.Fatalf("the asking item's targets are %+v, want the relic", asked.Targets)
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c, asked := openEnergyPrompt(t, g, me.ID)
	if len(asked.Targets) != 1 || asked.Targets[0].ID != relic {
		t.Fatalf("restored: the asking item's targets are %+v, want the relic", asked.Targets)
	}
	if a := c.PayAction(); a == nil || a.Count != 4 {
		t.Fatalf("restored payment %+v, want {E}{E}{E}{E}", a)
	}
	if err := g.ResolvePayUnless(c.ID, me.ID, true); err != nil {
		t.Fatalf("answer: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := findBattlefieldCardByID(g, relic); got == nil || got.Controller != me.ID {
		t.Errorf("the relic is %+v, want it under my control", got)
	}
	if energyOf(me) != 0 {
		t.Errorf("energy = %d, want 0", energyOf(me))
	}
}

// Electrozoa's decline is a registered key that taps the Jellyfish only
// while it is still the creature that triggered.
func TestElectrozoaEnergyPromptIsARestorePoint(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		zoa := pushDiesCreatureForTest(g, me.ID, "Electrozoa", "202091b2-9e08-47c1-9099-628b21d8ac12", "Creature — Jellyfish", 3, 1)
		setEnergy(t, g, me, 1)
		advanceTo(t, g, game.StepPrecombatMain)
		passPriorityAroundTable(t, g)
		if c, _ := openEnergyPrompt(t, g, me.ID); c.PickThen != electrozoaTap.Key() {
			t.Fatalf("the prompt's key is %q, want Electrozoa's", c.PickThen)
		}
		g = restoreRoundTrip(t, g, true)
		me = g.Seats[0]
		c, _ := openEnergyPrompt(t, g, me.ID)
		if err := g.ResolvePayUnless(c.ID, me.ID, pay); err != nil {
			t.Fatalf("answer: %v", err)
		}
		if got := apaLive(g, zoa).Tapped; got == pay {
			t.Errorf("pay=%v: tapped = %v", pay, got)
		}
	}
}

// A binary that does not know an energy key refuses the restore point.
func TestEnergyPromptWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rhino := pushDiesCreatureForTest(g, me.ID, "Thriving Rhino", erThrivingRhino, "Creature — Rhino", 2, 3)
	setEnergy(t, g, me, 3)
	declareAttack(t, g, opp.ID, rhino)
	passPriorityAroundTable(t, g)
	refuseUnknownKey(t, g, energyPlusOneCounterOnThis.Key())
}

// A MayPayEnergy with no registered branches is a programming error, not
// a closure prompt: it queues nothing.
func TestMayPayEnergyWithoutBranchesQueuesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	setEnergy(t, g, me, 3)
	var err error
	g.WithWriteLock(func() {
		err = MayPayEnergy{N: 1, Question: "pay?"}.Apply(NewContext(g, &game.StackItem{
			Kind: game.StackItemTriggered, Controller: me.ID, Owner: me.ID,
		}))
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("err = %v, want ErrInvalidParam", err)
	}
	if energyPrompt(g, me.ID) != nil {
		t.Error("a prompt was queued with no branches")
	}
}

// --- MayPayX ---------------------------------------------------------

// payXPrompt is the keyed "choose X" `p` owes: a pay_amount that holds
// no frame, naming the branches the payment runs.
func payXPrompt(t *testing.T, g *game.Game, p uuid.UUID) *game.PendingChoice {
	t.Helper()
	c := payAmountPrompt(g, p)
	if c == nil {
		t.Fatalf("no number prompt for %s", p)
	}
	if c.MayPayX == nil || c.PickThen == "" || c.PayAmount == nil {
		t.Fatalf("the number prompt is not keyed: %+v", c)
	}
	return c
}

// Tilonalli's Summoner: both prompts of "you may pay {X}{R}" are restore
// points. The restored number prompt keeps its bounds and goal; the
// restored payment keeps X, so the Elementals made are the X chosen, and
// the attacked opponent comes off the carried trigger event.
func TestTilonallisSummonerPayXPromptsAreRestorePoints(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := summonerAttacks(t, g, 3, 1)
	c := payXPrompt(t, g, me.ID)
	if c.PickThen != tilonallisSummonerPaid.Key() {
		t.Fatalf("the X prompt's key is %q, want Tilonalli's", c.PickThen)
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c = payXPrompt(t, g, me.ID)
	if pa := c.PayAmount; pa.Min != 0 || pa.Max != 3 || pa.Goal != 3 || pa.ResourceOrEnergy() != game.PayResourceNone {
		t.Fatalf("restored X prompt = %+v, want 0..3 with goal 3, not paid", pa)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 2); err != nil {
		t.Fatalf("answer X: %v", err)
	}
	pay := payUnlessFor(g, me.ID)
	if pay == nil || pay.PayCost != "{2}{R}" || pay.PickNumber != 2 {
		t.Fatalf("payment prompt = %+v, want {2}{R} carrying X = 2", pay)
	}
	if pay.OwedInStep.Step != game.StepDeclareAttackers {
		t.Errorf("payment owed in %+v, want the declare attackers step", pay.OwedInStep)
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	pay = payUnlessFor(g, me.ID)
	if pay == nil || pay.PayCost != "{2}{R}" || pay.PickNumber != 2 || pay.PickThen != tilonallisSummonerPaid.Key() {
		t.Fatalf("restored payment prompt = %+v, want {2}{R} carrying X = 2", pay)
	}
	if err := g.ResolvePayUnless(pay.ID, me.ID, true); err != nil {
		t.Fatalf("pay: %v", err)
	}
	// The attacked opponent is asked first about the split.
	answerNumber(t, g, me.ID, 2)
	toks := elementals(g, me.ID)
	if len(toks) != 2 {
		t.Fatalf("%d Elementals, want X = 2", len(toks))
	}
	for _, tok := range toks {
		if !tok.Tapped || tok.AttackingTarget != opp.ID {
			t.Errorf("Elemental tapped=%v attacking %s, want tapped and attacking %s", tok.Tapped, tok.AttackingTarget, opp.ID)
		}
	}
	if n := me.ManaPool.Total(); n != 1 {
		t.Errorf("%d mana left, want 1 ({2}{R} paid from 4)", n)
	}
}

// Declining the restored payment makes nothing, and X = 0 on a restored
// number prompt asks for {R} alone (the X settled at 0).
func TestTilonallisSummonerRestoredPayXDeclineAndZero(t *testing.T) {
	g := newCatalogGame(t)
	summonerAttacks(t, g, 3, 1)
	g = restoreRoundTrip(t, g, true)
	me := g.Seats[0]
	c := payXPrompt(t, g, me.ID)
	if err := g.ResolvePayAmount(c.ID, me.ID, 0); err != nil {
		t.Fatalf("answer X: %v", err)
	}
	pay := payUnlessFor(g, me.ID)
	if pay == nil || pay.PayCost != "{R}" || pay.PickNumber != 0 {
		t.Fatalf("payment prompt = %+v, want {R} carrying X = 0", pay)
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	pay = payUnlessFor(g, me.ID)
	if pay == nil {
		t.Fatal("the restored table lost the payment prompt")
	}
	if err := g.ResolvePayUnless(pay.ID, me.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := len(elementals(g, me.ID)); n != 0 {
		t.Errorf("%d Elementals after declining, want none", n)
	}
}

// Flameblast Dragon: the target chosen as the trigger went on the stack
// rides both prompts as the asking item's data, so the restored
// payment deals X damage to it.
func TestFlameblastDragonPayXPromptsAreRestorePoints(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	dragon := pushDiesCreatureForTest(g, me.ID, "Flameblast Dragon", flameblastDragonOracle, "Creature — Dragon", 5, 5)
	victim := pushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(dragon, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	answerTriggerTargets(t, g, me.ID, victim)
	fillPool(me, 5)
	fillPoolColored(me, "R", 1)
	passPriorityAroundTable(t, g)
	c := payXPrompt(t, g, me.ID)
	if c.PickThen != flameblastDragonPaid.Key() || c.Asked == nil || len(c.Asked.Targets) != 1 || c.Asked.Targets[0].ID != victim {
		t.Fatalf("the X prompt = %+v, want Flameblast's key asking about the victim", c)
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	c = payXPrompt(t, g, me.ID)
	if pa := c.PayAmount; pa.Max != 5 || pa.Goal != 3 || pa.Unit != game.PayAmountDamage {
		t.Fatalf("restored X prompt = %+v, want 0..5 with goal 3 in damage", pa)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 3); err != nil {
		t.Fatalf("answer X: %v", err)
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	pay := payUnlessFor(g, me.ID)
	if pay == nil || pay.PayCost != "{3}{R}" || pay.PickNumber != 3 {
		t.Fatalf("restored payment prompt = %+v, want {3}{R} carrying X = 3", pay)
	}
	if err := g.ResolvePayUnless(pay.ID, me.ID, true); err != nil {
		t.Fatalf("pay: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, victim); ok {
		t.Error("the 3/3 survived 3 damage")
	}
	if n := me.ManaPool.Total(); n != 2 {
		t.Errorf("pool %d, want 2 left", n)
	}
}

// Leyline Tyrant: X is the amount of {R} paid, carried as data to the
// reflexive trigger's damage.
func TestLeylineTyrantPayXPromptsAreRestorePoints(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	tyrant := pushCatalogPermanent(g, me.ID, "Leyline Tyrant", "Creature — Dragon", mkTyrantOracle, false)
	fillPoolColored(me, "R", 3)
	before := opp.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tyrant) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if c := payXPrompt(t, g, me.ID); c.PickThen != leylineTyrantPaid.Key() || c.MayPayX.Symbol != "{R}" {
		t.Fatalf("the X prompt = %+v, want Leyline Tyrant's, paying in {R}", c)
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	c := payXPrompt(t, g, me.ID)
	if pa := c.PayAmount; pa.Min != 0 || pa.Max != 3 {
		t.Fatalf("restored X prompt = %+v, want 0..3", pa)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 2); err != nil {
		t.Fatalf("answer X: %v", err)
	}
	g = restoreRoundTrip(t, g, true)
	me, opp = g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pay := payUnlessFor(g, me.ID)
	if pay == nil || pay.PayCost != "{R}{R}" || pay.PickNumber != 2 {
		t.Fatalf("restored payment prompt = %+v, want {R}{R} carrying X = 2", pay)
	}
	if err := g.ResolvePayUnless(pay.ID, me.ID, true); err != nil {
		t.Fatalf("pay: %v", err)
	}
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 2 {
		t.Errorf("damage = %d, want 2", got)
	}
	if got := redInPool(me); got != 1 {
		t.Errorf("red left in the pool = %d, want 1", got)
	}
}

// The "choose X" prompt names a pay-unless key, and a binary that lacks it
// refuses the file rather than ask a question it cannot answer. The
// payment prompt after it is the pay-unless registry's own.
func TestPayXPromptWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := summonerAttacks(t, g, 3, 1)
	payXPrompt(t, g, me.ID)
	refuseUnknownKey(t, g, tilonallisSummonerPaid.Key())
}

// A chooser who leaves while the restored X prompt is open pays nothing
// (CR 800.4f): X settles at 0, the payment is not asked of a player who
// has left, and nothing is made.
func TestRestoredPayXPromptDropsWithItsChooser(t *testing.T) {
	g := newCatalogGame(t)
	summonerAttacks(t, g, 3, 1)
	g = restoreRoundTrip(t, g, true)
	me := g.Seats[0]
	payXPrompt(t, g, me.ID)
	if err := g.Concede(me.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if payAmountPrompt(g, me.ID) != nil || payUnlessFor(g, me.ID) != nil {
		t.Error("a prompt outlived its chooser")
	}
	if n := len(elementals(g, me.ID)); n != 0 {
		t.Errorf("%d Elementals made for a chooser who left", n)
	}
}
