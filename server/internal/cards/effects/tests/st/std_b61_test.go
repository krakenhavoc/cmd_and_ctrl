package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestStdB61LamenterEntersWithACounterAndTradesItForACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B")
	id := effectstest.CastHoldCreature(t, g, "Moonlit Lamenter", "Creature — Treefolk Cleric", stdB61LamenterOracle, 2, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, id, game.CounterMinusOne); n != 1 {
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
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{CounterKind: game.CounterMinusOne})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, me.Hand.Size())
	}
	if n := effectstest.CountersOn(g, id, game.CounterMinusOne); n != 0 {
		t.Errorf("%d -1/-1 counters left, want 0 (the cost removed it)", n)
	}
}

func TestStdB61ProtectiveResponseDestroysAnAttackerButNotABystander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	raider := effectstest.PushVanillaCreature(g, me.ID, "Raider", 2, 2)
	idle := effectstest.PushVanillaCreature(g, opp.ID, "Idle", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, raider)
	id := effectstest.HandCardForTest(me, "Protective Response", "Instant", stdB61ResponseOracle)
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
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(raider) {
		t.Error("the attacking creature survived")
	}
}

func TestStdB61DounguardGrowsWhenAnotherCreatureEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	guard := effectstest.CastHoldCreature(t, g, "Reluctant Dounguard", "Creature — Kithkin Soldier", stdB61DounguardOracle, 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, guard, game.CounterMinusOne); n != 2 {
		t.Fatalf("%d -1/-1 counters, want 2", n)
	}
	bear := game.NewCard("Bear", me.ID)
	bear.TypeLine, bear.Power, bear.Toughness = "Creature — Bear", 2, 2
	me.Hand.PushTop(bear)
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, guard, game.CounterMinusOne); n != 1 {
		t.Errorf("%d -1/-1 counters after another creature entered, want 1", n)
	}
}

func TestStdB61ReflexesPumpsGivesFirstStrikeAndUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.CastCatalogSpell(t, g, "Riverguard's Reflexes", "Instant", stdB61ReflexesOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.GaLayered(t, g, bear)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("Bear = %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "first strike") {
		t.Error("no first strike")
	}
	if c.Tapped {
		t.Error("the creature is still tapped")
	}
}

func TestStdB61LurkerSurveilsOneWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Top A", "Top B")
	effectstest.CastHoldCreature(t, g, "Shore Lurker", "Creature — Merfolk Scout", stdB61LurkerOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("surveil choice = %+v, want one card", c)
	}
}

func TestStdB61CelebrantIsAVigilantConvokeCreature(t *testing.T) {
	spec, _ := effects.Lookup(stdB61CelebrantOracle)
	if spec.TapCost == nil {
		t.Error("no convoke tap cost")
	}
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "vigilance" {
		t.Errorf("printed keywords %v, want [vigilance]", spec.PrintedKeywords)
	}
}

func TestStdB61ImbuerPumpsALoneAttackerByTheKithkinYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Thoughtweft Imbuer", "Creature — Kithkin Advisor", stdB61ImbuerOracle, false)
	effectstest.B12Creature(g, me.ID, "Kithkin Friend", "Creature — Kithkin", 1, 1)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.GaLayered(t, g, bear); c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("lone Bear = %d/%d, want 4/4 (two Kithkin)", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestStdB61ImbuerDoesNothingWhenTwoCreaturesAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Thoughtweft Imbuer", "Creature — Kithkin Advisor", stdB61ImbuerOracle, false)
	a := effectstest.PushVanillaCreature(g, me.ID, "A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "B", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.GaLayered(t, g, a); c.CurrentPower() != 2 {
		t.Errorf("A = %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestStdB61ShieldbearerPumpsYourCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	sb := effectstest.PushCatalogPermanent(g, me.ID, "Timid Shieldbearer", "Creature — Kithkin Soldier", stdB61ShieldbearerOracl, false)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{W}{W}{W}{W}"); err != nil {
		t.Fatal(err)
	}
	effectstest.B16Activate(t, g, me.ID, sb, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.GaLayered(t, g, mine); c.CurrentPower() != 3 {
		t.Errorf("Mine = %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if c := effectstest.GaLayered(t, g, theirs); c.CurrentPower() != 2 {
		t.Errorf("an opponent's creature got the pump (%d/%d)", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestStdB61VaulterPumpsAnotherMerfolkWhenItBecomesTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vaulter := effectstest.PushCatalogPermanent(g, me.ID, "Tributary Vaulter", "Creature — Merfolk Warrior", stdB61VaulterOracle, false)
	friend := effectstest.B12Creature(g, me.ID, "Merfolk Friend", "Creature — Merfolk", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, vaulter)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, friend)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.GaLayered(t, g, friend); c.CurrentPower() != 3 {
		t.Errorf("Merfolk Friend = %d/%d, want 3/1", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestStdB61PreacherGainsTwoWhenItBecomesTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	preacher := effectstest.PushCatalogPermanent(g, me.ID, "Wanderbrine Preacher", "Creature — Merfolk Cleric", stdB61PreacherOracle, false)
	life := me.Life
	effectstest.DeclareAttack(t, g, opp.ID, preacher)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
}
