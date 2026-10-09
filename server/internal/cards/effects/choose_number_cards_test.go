package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// choose_number_cards_test.go — #1941, ADR 0129's amendment of
// 2026-10-09: the three cards the number prompt unblocked.

const (
	cnVolcanoHellion     = "12a88c08-11e3-4f51-a17c-ca06452bac6f"
	cnPhyrexianProcessor = "36c800cb-b1ca-4432-ad3c-4d8b90337f4c"
	cnNecrodominance     = "a10b3e35-8cc4-450e-9e30-0fce8df0fea4"
)

func TestChooseNumberCardsAreFull(t *testing.T) {
	for _, oid := range []string{cnVolcanoHellion, cnPhyrexianProcessor, cnNecrodominance} {
		s, ok := Lookup(oid)
		if !ok {
			t.Fatalf("%s is not in the catalog", oid)
		}
		if s.Completeness != CompletenessFull || len(s.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v; want Full with none", s.Name, s.Completeness, s.Caveats)
		}
	}
}

// castHellion casts Volcano Hellion as the active seat, a 6/5, and
// passes until its enters trigger asks for a target.
func castHellion(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Volcano Hellion", TypeLine: "Creature — Hellion",
		OracleID: cnVolcanoHellion, Power: 6, Toughness: 5, Owner: active.ID, Controller: active.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell Volcano Hellion: %v", err)
	}
	passPriorityAroundTable(t, g)
	return id
}

// Volcano Hellion deals the chosen amount to its controller and the
// target, with no ceiling, and the damage can't be prevented. The prompt
// offers a bot the target's lethal damage and the chooser's life total.
func TestVolcanoHellionDealsTheChosenAmountToYouAndTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	castHellion(t, g)
	answerTriggerTargets(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	c := payAmountPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no number prompt")
	}
	pa := c.PayAmount
	if pa.ResourceOrEnergy() != game.PayResourceNone || !pa.NoMax || !pa.SelfDamage || pa.Min != 0 || pa.Goal != 3 {
		t.Fatalf("prompt = %+v, want an unpaid number from 0 with no ceiling and goal 3", *pa)
	}
	if len(pa.Marks) != 2 || pa.Marks[0] != 3 || pa.Marks[1] != me.Life {
		t.Errorf("marks = %v, want [3 %d]", pa.Marks, me.Life)
	}
	life := me.Life
	if err := g.ResolvePayAmount(c.ID, me.ID, 4); err != nil {
		t.Fatalf("answer 4: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life-4 {
		t.Errorf("life %d, want %d", me.Life, life-4)
	}
	if _, ok := battlefieldCard(g, victim); ok {
		t.Error("the 3/3 survived 4 damage")
	}
}

// An amount past lethal is legal, and zero deals nothing to anyone.
func TestVolcanoHellionTakesAnyAmountAndZero(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := pushVanillaCreature(g, opp.ID, "Big", 20, 20)
	castHellion(t, g)
	answerTriggerTargets(t, g, me.ID, big)
	passPriorityAroundTable(t, g)
	life := me.Life
	if err := g.ResolvePayAmount(payAmountPrompt(g, me.ID).ID, me.ID, 0); err != nil {
		t.Fatalf("answer 0: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("zero cost %d life", life-me.Life)
	}
	if c, _ := battlefieldCard(g, big); c.DamageMarked != 0 {
		t.Errorf("zero marked %d damage", c.DamageMarked)
	}
}

// The lethal goal is offered for an opponent's creature only: aimed at
// your own creature, the bot's threshold is nothing.
func TestVolcanoHellionHasNoGoalOnYourOwnCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	castHellion(t, g)
	answerTriggerTargets(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	c := payAmountPrompt(g, me.ID)
	if c == nil || c.PayAmount.Goal != 0 {
		t.Fatalf("prompt = %+v, want goal 0", c)
	}
}

// Phyrexian Processor asks for life as it enters, up to its controller's
// life total, stores what was paid, and makes tokens that size — still
// that size when the Processor is gone before the ability resolves.
func TestPhyrexianProcessorPaysLifeAndMakesTokensThatSize(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Life = 40
	proc := castCatalogSpell(t, g, "Phyrexian Processor", "Artifact", cnPhyrexianProcessor, nil)
	passPriorityAroundTable(t, g)
	c := payAmountPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no life prompt as the Processor entered")
	}
	pa := c.PayAmount
	if pa.ResourceOrEnergy() != game.PayResourceLife || pa.Max != 40 || pa.Goal != 20 {
		t.Fatalf("prompt = %+v, want life up to 40 with goal 20", *pa)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 7); err != nil {
		t.Fatalf("pay 7: %v", err)
	}
	if me.Life != 33 {
		t.Errorf("life %d, want 33", me.Life)
	}
	if n := g.ChosenNumberOf(proc); n != 7 {
		t.Fatalf("stored %d, want 7", n)
	}

	floatForTest(g, me, "CCCC")
	if err := g.ActivateCatalogAbility(me.ID, proc, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// Sacrificed in response: the token is the size it last knew.
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(proc); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Phyrexian Minion" && c.Controller == me.ID {
			found = true
			if c.CurrentPower() != 7 || c.CurrentToughness() != 7 {
				t.Errorf("Minion is %d/%d, want 7/7", c.CurrentPower(), c.CurrentToughness())
			}
			if len(c.Colors) != 1 || c.Colors[0] != "B" {
				t.Errorf("Minion colours %v, want black", c.Colors)
			}
		}
	}
	if !found {
		t.Error("no Phyrexian Minion token")
	}
}

// Paying nothing is a legal answer, and makes a 0/0 that dies.
func TestPhyrexianProcessorPayingNothingMakesAZeroZero(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	proc := castCatalogSpell(t, g, "Phyrexian Processor", "Artifact", cnPhyrexianProcessor, nil)
	passPriorityAroundTable(t, g)
	life := me.Life
	if err := g.ResolvePayAmount(payAmountPrompt(g, me.ID).ID, me.ID, 0); err != nil {
		t.Fatalf("pay 0: %v", err)
	}
	if me.Life != life || g.ChosenNumberOf(proc) != 0 {
		t.Fatalf("life %d (want %d), stored %d", me.Life, life, g.ChosenNumberOf(proc))
	}
	floatForTest(g, me, "CCCC")
	if err := g.ActivateCatalogAbility(me.ID, proc, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Phyrexian Minion" {
			t.Error("a 0/0 Minion survived the state-based actions")
		}
	}
}

// Necrodominance: at your end step, pay any amount of life and draw that
// many; your maximum hand size is five; your cards go to exile instead
// of your graveyard.
func TestNecrodominancePaysLifeToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "Necrodominance", cnNecrodominance, "Legendary Enchantment")
	me.Life = 30
	for len(me.Hand.Cards) > 2 {
		if _, err := game.MoveCard(me.Hand, me.Library, me.Hand.Cards[0].InstanceID); err != nil {
			t.Fatal(err)
		}
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	c := payAmountPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no end-step life prompt")
	}
	if pa := c.PayAmount; pa.ResourceOrEnergy() != game.PayResourceLife || pa.Max != 30 || pa.Goal != 3 {
		t.Fatalf("prompt = %+v, want life up to 30, goal 3 (fill a hand of 2 to five)", *pa)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 4); err != nil {
		t.Fatalf("pay 4: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != 26 || len(me.Hand.Cards) != 6 {
		t.Errorf("life %d, hand %d; want 26 and 6", me.Life, len(me.Hand.Cards))
	}
	var limit int
	g.WithWriteLock(func() { limit = g.EffectiveMaxHandSizeLocked(me) })
	if limit != 5 {
		t.Errorf("maximum hand size %d, want 5", limit)
	}
}

func TestNecrodominanceExilesYourCardsInsteadOfTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Necrodominance", cnNecrodominance, "Legendary Enchantment")
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{mine, theirs} {
			if err := g.DestroyPermanentForEffect(id); err != nil {
				t.Fatal(err)
			}
		}
	})
	if z := g.FindCardZoneForEffect(mine); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("your creature went to %v, want exile", z)
	}
	if z := g.FindCardZoneForEffect(theirs); z == nil || z.Kind != game.ZoneGraveyard {
		t.Errorf("an opponent's creature went to %v, want their graveyard", z)
	}
}
