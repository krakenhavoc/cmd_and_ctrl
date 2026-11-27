package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b539_test.go — Shadowmoor, batch 539 (#5591).

const (
	oldB539Merrow   = "7d5126a9-ab81-4de6-817d-3b4b84c7ca5f"
	oldB539Puca     = "7e93ea61-bc16-4d22-a5e5-e6ba0f7ae14f"
	oldB539Runner   = "9c0a8412-af9f-4a22-b462-e9b2d9bd04c6"
	oldB539Spinners = "b881abd7-1fe7-4404-b9ba-fb64869126fc"
	oldB539Somno    = "c819ebc8-feb5-48fa-81d2-a4f5aaa4fdbb"
	oldB539Titan    = "d8f2c12c-d463-42dc-95f1-2daf68f99873"
)

func TestOldB539EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB539Merrow: "Prismwake Merrow", oldB539Puca: "Cemetery Puca", oldB539Runner: "Medicine Runner",
		oldB539Spinners: "Glamer Spinners", oldB539Somno: "Somnomancer", oldB539Titan: "Deep-Slumber Titan",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB539SomnomancerMayTapTargetCreature(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		effectstest.Msc17Play(t, g, me, game.Card{Name: "Somnomancer", TypeLine: "Creature — Kithkin Wizard", OracleID: oldB539Somno, Power: 2, Toughness: 1})
		effectstest.B04WaitForPick(t, g, me.ID)
		effectstest.PickCard(t, g, me.ID, victim)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if c := effectstest.FindTestCard(g, victim); c == nil || c.Tapped != yes {
			t.Errorf("answered %v: tapped = %v", yes, c != nil && c.Tapped)
		}
	}
}

func TestOldB539DeepSlumberTitanEntersTappedStaysTappedAndUntapsWhenDamaged(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	titan := effectstest.Msc17Play(t, g, me, game.Card{Name: "Deep-Slumber Titan", TypeLine: "Creature — Giant Warrior", OracleID: oldB539Titan, Power: 7, Toughness: 7})
	if !effectstest.ArnTapped(t, g, titan) {
		t.Fatal("the Titan did not enter tapped")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if !effectstest.ArnTapped(t, g, titan) {
		t.Fatal("the Titan untapped during its controller's untap step")
	}
	src := effectstest.B12Creature(g, g.Seats[1].ID, "Pinger", "Creature — Bear", 1, 1)
	effectstest.DealDamageTo(t, g, src, titan, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ArnTapped(t, g, titan) {
		t.Error("being dealt damage did not untap the Titan")
	}
}

func TestOldB539MedicineRunnerRemovesOneCounterOfTheTargetPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.CounterBear(g, opp.ID, map[string]int{"+1/+1": 2})
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Medicine Runner", TypeLine: "Creature — Elf Cleric", OracleID: oldB539Runner, Power: 2, Toughness: 1})
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	// A one-kind permanent is one question: how many counters, up to one.
	if p := effectstest.PayAmountPrompt(g, me.ID); p == nil {
		t.Fatal("no \"how many counters\" question")
	}
	effectstest.AnswerNumber(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindTestCard(g, bear).Counters["+1/+1"]; got != 1 {
		t.Errorf("%d +1/+1 counters left, want 1", got)
	}
}

func TestOldB539PrismwakeMerrowRecoloursTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Elf", TypeLine: "Creature — Elf", Colors: []string{"G"},
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Prismwake Merrow", TypeLine: "Creature — Merfolk Wizard", OracleID: oldB539Merrow, Power: 2, Toughness: 1})
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, elf)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "WU")
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.EffectiveColors(t, g, elf)
	if len(got) != 2 {
		t.Errorf("colours %v, want exactly white and blue", got)
	}
}

func TestOldB539GlamerSpinnersMovesEveryAuraToTheChosenPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, opp.ID, "Host", "Creature — Bear", 2, 2)
	dest := effectstest.B12Creature(g, opp.ID, "Dest", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	a1 := effectstest.B12Permanent(g, me.ID, "Aura One", "Enchantment — Aura")
	a2 := effectstest.B12Permanent(g, opp.ID, "Aura Two", "Enchantment — Aura")
	effectstest.AttachForTest(t, g, a1, host)
	effectstest.AttachForTest(t, g, a2, host)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Glamer Spinners", TypeLine: "Creature — Faerie Wizard", OracleID: oldB539Spinners, Power: 2, Toughness: 4})
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, host)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choice of the other permanent")
	}
	for _, id := range pick.ChooseCards {
		if id == mine || id == host {
			t.Errorf("offered a permanent that is not another one the target's controller controls: %v", id)
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{dest}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, a := range []uuid.UUID{a1, a2} {
		if !effectstest.AttachedToNow(t, g, a, dest) {
			t.Errorf("an Aura did not move to the chosen permanent")
		}
	}
}

func TestOldB539CemeteryPucaPaidBecomesACopyOfTheDeadCreatureAndKeepsTheAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	puca := effectstest.B12Push(g, me.ID, "Cemetery Puca", "Creature — Shapeshifter", oldB539Puca, 1, 2)
	victim := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hill Giant", TypeLine: "Creature — Giant", Colors: []string{"R"},
		Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID,
	})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, puca); p != 3 || th != 3 {
		t.Errorf("the Puca is %d/%d, want a 3/3 copy", p, th)
	}
	if name := effectstest.FindTestCard(g, puca).Effective().Name; name != "Hill Giant" {
		t.Errorf("the Puca is named %q, want Hill Giant", name)
	}
	// It has the ability still: another death offers the payment again.
	other := effectstest.B12Creature(g, opp.ID, "Fodder", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(other) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("the copy lost the Puca's triggered ability")
	}
}

func TestOldB539CemeteryPucaDeclinedStaysAPuca(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	puca := effectstest.B12Push(g, me.ID, "Cemetery Puca", "Creature — Shapeshifter", oldB539Puca, 1, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, puca); p != 1 || th != 2 {
		t.Errorf("the Puca is %d/%d, want 1/2", p, th)
	}
}

func TestOldB539CemeteryPucaCopiesADyingToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	puca := effectstest.B12Push(g, me.ID, "Cemetery Puca", "Creature — Shapeshifter", oldB539Puca, 1, 2)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(g.Seats[1].ID, effects.TokenCard("3/3 green Beast"), 1) })
	var beast uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Beast" {
				beast = c.InstanceID
			}
		}
	})
	if beast == uuid.Nil {
		t.Fatal("no Beast token")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(beast) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("a dying token did not trigger the Puca")
	}
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, puca); p != 3 || th != 3 {
		t.Errorf("the Puca is %d/%d, want a 3/3 copy of the token", p, th)
	}
}
