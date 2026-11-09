package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b61_test.go — Standard leftovers, batch 61 (#4452): Lorwyn Eclipsed.

const (
	stdB61LamenterOracle    = "ce009bd4-5ffb-46a4-a9dd-b9c060926467"
	stdB61ResponseOracle    = "c840b89b-81be-4644-8bb5-35d14b1bc620"
	stdB61DounguardOracle   = "162dcfa0-36e4-41cd-8322-08e810ff04d4"
	stdB61ReflexesOracle    = "0bdebad7-a05e-439a-b5a9-241a2d09f943"
	stdB61LurkerOracle      = "4288ba56-ddd0-4bd9-ba61-daf03e9f2d7f"
	stdB61CelebrantOracle   = "27c87371-697e-4ca6-a7f4-d99215d908bc"
	stdB61ImbuerOracle      = "9dee8c0b-71e0-4a94-b411-78b29b599847"
	stdB61ShieldbearerOracl = "8bef3511-ceeb-4d57-972a-ab363ecacd5b"
	stdB61VaulterOracle     = "991f4eec-cf3e-4554-ba88-c44bb0d255f0"
	stdB61PreacherOracle    = "4b4dcdb1-0017-494f-8d83-6706b188e575"
)

func TestStdB61EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		stdB61LamenterOracle:    "Moonlit Lamenter",
		stdB61ResponseOracle:    "Protective Response",
		stdB61DounguardOracle:   "Reluctant Dounguard",
		stdB61ReflexesOracle:    "Riverguard's Reflexes",
		stdB61LurkerOracle:      "Shore Lurker",
		stdB61CelebrantOracle:   "Sun-Dappled Celebrant",
		stdB61ImbuerOracle:      "Thoughtweft Imbuer",
		stdB61ShieldbearerOracl: "Timid Shieldbearer",
		stdB61VaulterOracle:     "Tributary Vaulter",
		stdB61PreacherOracle:    "Wanderbrine Preacher",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestStdB61LamenterEntersWithACounterAndTradesItForACard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B")
	id := castHoldCreature(t, g, "Moonlit Lamenter", "Creature — Treefolk Cleric", stdB61LamenterOracle, 2, 5)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, id, game.CounterMinusOne); n != 1 {
		t.Fatalf("%d -1/-1 counters, want 1", n)
	}
	hand := me.Hand.Size()
	// Sorcery speed and a real cost: no mana, no card.
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, CounterKind: game.CounterMinusOne}); err == nil {
		t.Fatal("activated with no mana")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{W}"); err != nil {
		t.Fatal(err)
	}
	b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{CounterKind: game.CounterMinusOne})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, me.Hand.Size())
	}
	if n := countersOn(g, id, game.CounterMinusOne); n != 0 {
		t.Errorf("%d -1/-1 counters left, want 0 (the cost removed it)", n)
	}
}

func TestStdB61ProtectiveResponseDestroysAnAttackerButNotABystander(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMain(t, g)
	raider := pushVanillaCreature(g, me.ID, "Raider", 2, 2)
	idle := pushVanillaCreature(g, opp.ID, "Idle", 2, 2)
	attackWith(t, g, opp.ID, raider)
	id := handCardForTest(me, "Protective Response", "Instant", stdB61ResponseOracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{W}{W}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: idle}}}); err == nil {
		t.Fatal("a creature that is not attacking or blocking was accepted")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: raider}}}); err != nil {
		t.Fatalf("cast on the attacker: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(raider) {
		t.Error("the attacking creature survived")
	}
}

func TestStdB61DounguardGrowsWhenAnotherCreatureEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	guard := castHoldCreature(t, g, "Reluctant Dounguard", "Creature — Kithkin Soldier", stdB61DounguardOracle, 4, 4)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, guard, game.CounterMinusOne); n != 2 {
		t.Fatalf("%d -1/-1 counters, want 2", n)
	}
	bear := game.NewCard("Bear", me.ID)
	bear.TypeLine, bear.Power, bear.Toughness = "Creature — Bear", 2, 2
	me.Hand.PushTop(bear)
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, guard, game.CounterMinusOne); n != 1 {
		t.Errorf("%d -1/-1 counters after another creature entered, want 1", n)
	}
}

func TestStdB61ReflexesPumpsGivesFirstStrikeAndUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	castCatalogSpell(t, g, "Riverguard's Reflexes", "Instant", stdB61ReflexesOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	c := gaLayered(t, g, bear)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("Bear = %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	if !hasKeywordOnBattlefield(t, g, bear, "first strike") {
		t.Error("no first strike")
	}
	if c.Tapped {
		t.Error("the creature is still tapped")
	}
}

func TestStdB61LurkerSurveilsOneWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Top A", "Top B")
	castHoldCreature(t, g, "Shore Lurker", "Creature — Merfolk Scout", stdB61LurkerOracle, 3, 3)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("surveil choice = %+v, want one card", c)
	}
}

func TestStdB61CelebrantIsAVigilantConvokeCreature(t *testing.T) {
	spec, _ := Lookup(stdB61CelebrantOracle)
	if spec.TapCost == nil {
		t.Error("no convoke tap cost")
	}
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "vigilance" {
		t.Errorf("printed keywords %v, want [vigilance]", spec.PrintedKeywords)
	}
}

func TestStdB61ImbuerPumpsALoneAttackerByTheKithkinYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Thoughtweft Imbuer", "Creature — Kithkin Advisor", stdB61ImbuerOracle, false)
	b12Creature(g, me.ID, "Kithkin Friend", "Creature — Kithkin", 1, 1)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, bear); c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("lone Bear = %d/%d, want 4/4 (two Kithkin)", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestStdB61ImbuerDoesNothingWhenTwoCreaturesAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Thoughtweft Imbuer", "Creature — Kithkin Advisor", stdB61ImbuerOracle, false)
	a := pushVanillaCreature(g, me.ID, "A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "B", 2, 2)
	declareAttack(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, a); c.CurrentPower() != 2 {
		t.Errorf("A = %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestStdB61ShieldbearerPumpsYourCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMain(t, g)
	sb := pushCatalogPermanent(g, me.ID, "Timid Shieldbearer", "Creature — Kithkin Soldier", stdB61ShieldbearerOracl, false)
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{W}{W}{W}{W}"); err != nil {
		t.Fatal(err)
	}
	b16Activate(t, g, me.ID, sb, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, mine); c.CurrentPower() != 3 {
		t.Errorf("Mine = %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if c := gaLayered(t, g, theirs); c.CurrentPower() != 2 {
		t.Errorf("an opponent's creature got the pump (%d/%d)", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestStdB61VaulterPumpsAnotherMerfolkWhenItBecomesTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vaulter := pushCatalogPermanent(g, me.ID, "Tributary Vaulter", "Creature — Merfolk Warrior", stdB61VaulterOracle, false)
	friend := b12Creature(g, me.ID, "Merfolk Friend", "Creature — Merfolk", 1, 1)
	declareAttack(t, g, opp.ID, vaulter)
	if latestPickTarget(g, me.ID) != nil {
		pickCard(t, g, me.ID, friend)
	}
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, friend); c.CurrentPower() != 3 {
		t.Errorf("Merfolk Friend = %d/%d, want 3/1", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestStdB61PreacherGainsTwoWhenItBecomesTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	preacher := pushCatalogPermanent(g, me.ID, "Wanderbrine Preacher", "Creature — Merfolk Cleric", stdB61PreacherOracle, false)
	life := me.Life
	declareAttack(t, g, opp.ID, preacher)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
}
