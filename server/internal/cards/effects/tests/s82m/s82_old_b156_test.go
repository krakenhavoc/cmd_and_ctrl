package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b156_test.go — Core Set 2020 creatures, S82 set OLD batch 156
// (#5518). Every test plays the real catalog entry through the engine.

const (
	oldB156Spider     = "37f3733e-cc4e-4d84-b29b-d474f6e254a2"
	oldB156Troll      = "b7d231eb-255d-4505-89a9-0c6824622209"
	oldB156Awakener   = "c0171343-9c0f-4001-8f4e-a5af9256c4d2"
	oldB156Embodiment = "9d6686a0-1648-42f6-a606-16d2a84138a7"
	oldB156Knight     = "3b4eaa50-dff6-4825-97da-42b2f5554218"
	oldB156Boreal     = "68a2e371-ccf0-4c2a-aeef-c66852811e1a"
	oldB156Overgrowth = "3271e3b9-1710-4dfe-a381-00a3ee1032b2"
	oldB156Pegasus    = "3b7b862a-7edd-4e96-b4b9-f7113b4c93f9"
	oldB156Beetle     = "b4cd7e4e-4b63-41a1-bfad-2840a96f84bb"
	oldB156Diamond    = "29ae4471-c28f-4e43-8afb-fee5023e4d08"
)

func TestOldB156CardsAreRegisteredWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB156Spider: "Canopy Spider", oldB156Troll: "Barkhide Troll", oldB156Awakener: "Thunderkin Awakener",
		oldB156Embodiment: "Embodiment of Agonies", oldB156Knight: "Knight of the Ebon Legion",
		oldB156Boreal: "Boreal Elemental", oldB156Overgrowth: "Overgrowth Elemental",
		oldB156Pegasus: "Loyal Pegasus", oldB156Beetle: "Blightbeetle", oldB156Diamond: "Diamond Knight",
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

func oldB156Seats(g *game.Game) (*game.Player, *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func TestOldB156PrintedKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB156Seats(g)
	for _, tc := range []struct {
		name, typeLine, oracle string
		kws                    []string
	}{
		{"Canopy Spider", "Creature — Spider", oldB156Spider, []string{"reach"}},
		{"Loyal Pegasus", "Creature — Pegasus", oldB156Pegasus, []string{"flying"}},
		{"Boreal Elemental", "Creature — Elemental", oldB156Boreal, []string{"flying"}},
		{"Thunderkin Awakener", "Creature — Elemental Shaman", oldB156Awakener, []string{"haste"}},
		{"Embodiment of Agonies", "Creature — Demon", oldB156Embodiment, []string{"flying", "deathtouch"}},
		{"Blightbeetle", "Creature — Insect", oldB156Beetle, []string{"protection from green"}},
		{"Diamond Knight", "Artifact Creature — Knight", oldB156Diamond, []string{"vigilance"}},
	} {
		id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
		for _, kw := range tc.kws {
			if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s lacks %s", tc.name, kw)
			}
		}
	}
}

// --- Loyal Pegasus ------------------------------------------------------------

func TestOldB156LoyalPegasusCantAttackAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB156Seats(g)
	peg := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Loyal Pegasus", oldB156Pegasus, 2, 1))
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(peg, opp.ID); err != nil {
		t.Fatalf("staging the Pegasus: %v", err)
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackLimit) {
		t.Fatalf("passing with the Pegasus alone: %v, want ErrAttackLimit", err)
	}
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("the companion was refused: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, peg); c == nil || c.AttackingTarget == uuid.Nil {
		t.Error("the Pegasus is not attacking beside the Bear")
	}
}

func TestOldB156LoyalPegasusCantBlockAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB156Seats(g)
	a1 := effectstest.PushVanillaCreature(g, me.ID, "Raider One", 2, 2)
	a2 := effectstest.PushVanillaCreature(g, me.ID, "Raider Two", 2, 2)
	peg := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Loyal Pegasus", oldB156Pegasus, 2, 1))
	bear := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Bear", "", 2, 2))
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{{Attacker: a1, Target: opp.ID}, {Attacker: a2, Target: opp.ID}}); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	err := g.DeclareBlocker(peg, a1)
	var re *game.BlockRefusedError
	if !errors.As(err, &re) || re.Reason != game.BlockReasonCantBlockAlone {
		t.Fatalf("blocking alone: %v, want cant_block_alone", err)
	}
	if err := g.DeclareBlocker(bear, a1); err != nil {
		t.Fatalf("the Bear block: %v", err)
	}
	if err := g.DeclareBlocker(peg, a2); err != nil {
		t.Fatalf("the Pegasus beside the Bear: %v", err)
	}
}

// --- Boreal Elemental ---------------------------------------------------------

func TestOldB156BorealElementalTaxesOpponentsSpellsThatTargetIt(t *testing.T) {
	g, me, opp := effectstest.P7Table(t)
	boreal := effectstest.B12Push(g, me.ID, "Boreal Elemental", "Creature — Elemental", oldB156Boreal, 3, 4)
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < len(g.Seats); i++ {
		if g.Turn.ActiveSeat == 1 {
			break
		}
		effectstest.EndTurn(t, g)
	}
	effectstest.AdvanceToMain(t, g)
	bolt := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: effectstest.LightningBoltOracle, Owner: opp.ID, Controller: opp.ID})
	effectstest.FlashIfFund(g, opp, "R")
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(boreal)}); err == nil {
		t.Fatal("cast a {R} spell at the Elemental with only {R}")
	}
	effectstest.FlashIfFund(g, opp, "R", "C", "C")
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(boreal)}); err != nil {
		t.Fatalf("with {R}{C}{C}: %v", err)
	}
}

// --- Barkhide Troll -----------------------------------------------------------

func TestOldB156BarkhideTrollEntersWithACounterAndSpendsItForHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB156Seats(g)
	trollCard := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: trollCard, Name: "Barkhide Troll", TypeLine: "Creature — Troll", ManaCost: "{G}{G}",
			OracleID: oldB156Troll, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, trollCard, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var troll uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == oldB156Troll {
			troll = c.InstanceID
		}
	}
	if troll == uuid.Nil {
		t.Fatal("the Troll did not enter")
	}
	if got := effectstest.CounterCount(g, troll, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d +1/+1 counters on entry, want 1", got)
	}
	if p, tough := effectstest.MkmB1PT(t, g, troll); p != 3 || tough != 3 {
		t.Errorf("the Troll is %d/%d, want 3/3", p, tough)
	}
	g.WithWriteLock(func() { me.ManaPool = nil })
	if err := g.ActivateCatalogAbility(me.ID, troll, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated {1} with no mana")
	}
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, troll, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := effectstest.CounterCount(g, troll, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters after paying the cost, want 0", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, troll, "hexproof") {
		t.Error("the Troll has no hexproof")
	}
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, troll, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with no counter left to remove")
	}
}

// --- Knight of the Ebon Legion ------------------------------------------------

func TestOldB156KnightPumpsAndGainsDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB156Seats(g)
	knight := effectstest.PushCatalogPermanent(g, me.ID, "Knight of the Ebon Legion", "Creature — Vampire Knight", oldB156Knight, false)
	effectstest.FloatForTest(g, me, "CCB")
	if err := g.ActivateCatalogAbility(me.ID, knight, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.MkmB1PT(t, g, knight); p != 4 || tough != 4 {
		t.Errorf("the Knight is %d/%d, want 4/4 (1/1 plus 3/3)", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, knight, "deathtouch") {
		t.Error("no deathtouch")
	}
}

func TestOldB156KnightGrowsOnlyIfAPlayerLostFourLife(t *testing.T) {
	for _, tc := range []struct {
		loss int
		want int
	}{{3, 0}, {4, 1}} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp := oldB156Seats(g)
		knight := effectstest.PushCatalogPermanent(g, me.ID, "Knight of the Ebon Legion", "Creature — Vampire Knight", oldB156Knight, false)
		g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -tc.loss) })
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.CounterCount(g, knight, game.CounterPlusOne); got != tc.want {
			t.Errorf("a player lost %d: %d counters, want %d", tc.loss, got, tc.want)
		}
	}
}

// --- Embodiment of Agonies ----------------------------------------------------

func TestOldB156EmbodimentCountsDifferentManaCostsAmongNonlandCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB156Seats(g)
	g.WithWriteLock(func() {
		for _, c := range []game.Card{
			{Name: "A", TypeLine: "Creature — Bear", ManaCost: "{2}{B}"},
			{Name: "B", TypeLine: "Instant", ManaCost: "{2}{B}"},
			{Name: "C", TypeLine: "Sorcery", ManaCost: "{1}{B}{B}"},
			{Name: "D", TypeLine: "Artifact", ManaCost: "{3}"},
			{Name: "Swamp", TypeLine: "Basic Land — Swamp"},
			{Name: "Cursed Land", TypeLine: "Land", ManaCost: "{9}"},
		} {
			c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
			me.Graveyard.PushTop(c)
		}
	})
	effectstest.CastFromHandForTest(t, g, me, "Embodiment of Agonies", "Creature — Demon", "{1}{B}{B}", oldB156Embodiment, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	var id uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == oldB156Embodiment {
			id = c.InstanceID
		}
	}
	if id == uuid.Nil {
		t.Fatal("the Embodiment did not survive: it should enter with counters")
	}
	if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 3 {
		t.Errorf("%d counters, want 3 (three different costs: {2}{B}, {1}{B}{B}, {3})", got)
	}
}

// --- Thunderkin Awakener ------------------------------------------------------

func TestOldB156AwakenerReturnsALesserToughnessElementalAttackingAndSacrificesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB156Seats(g)
	awakener := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Thunderkin Awakener", oldB156Awakener, 1, 2))
	small := uuid.New()
	big := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: small, Name: "Small Elemental", TypeLine: "Creature — Elemental", Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: big, Name: "Big Elemental", TypeLine: "Creature — Elemental", Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(awakener, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.AnswerTriggerTargets(t, g, me.ID, small)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// "What does it attack?": the first option is an opponent.
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, small)
	if !ok {
		t.Fatal("the lesser-toughness Elemental did not return")
	}
	if !c.Tapped || c.AttackingTarget == uuid.Nil {
		t.Errorf("it entered tapped %v, attacking %v; want both", c.Tapped, c.AttackingTarget != uuid.Nil)
	}
	if _, ok := effectstest.BattlefieldCard(g, big); ok || !effectstest.InGraveyard(me, big) {
		t.Error("the equal-toughness Elemental should have stayed in the graveyard")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, small); ok {
		t.Error("the Elemental was not sacrificed at the end step")
	}
}

// --- Overgrowth Elemental -----------------------------------------------------

func TestOldB156OvergrowthElementalCountersAnotherElementalWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB156Seats(g)
	other := effectstest.B12Creature(g, me.ID, "Fellow Elemental", "Creature — Elemental", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastFromHandForTest(t, g, me, "Overgrowth Elemental", "Creature — Elemental", "{2}{G}", oldB156Overgrowth, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.AnswerTriggerTargets(t, g, me.ID, other)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.CounterCount(g, other, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters on the other Elemental, want 1", got)
	}
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters on the Bear, want 0", got)
	}
}

func TestOldB156OvergrowthElementalGainsLifeAndGrowsForAnElemental(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB156Seats(g)
	og := effectstest.PushCatalogPermanent(g, me.ID, "Overgrowth Elemental", "Creature — Elemental", oldB156Overgrowth, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	elem := effectstest.B12Creature(g, me.ID, "Fellow Elemental", "Creature — Elemental", 2, 2)
	life := me.Life
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 || effectstest.CounterCount(g, og, game.CounterPlusOne) != 0 {
		t.Fatalf("a Bear died: life %d → %d, counters %d; want +1 life and no counter", life, me.Life, effectstest.CounterCount(g, og, game.CounterPlusOne))
	}
	effectstest.KillCreature(t, g, me.ID, elem)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 || effectstest.CounterCount(g, og, game.CounterPlusOne) != 1 {
		t.Errorf("an Elemental died: life %d → %d, counters %d; want +2 life and one counter", life, me.Life, effectstest.CounterCount(g, og, game.CounterPlusOne))
	}
}

func TestOldB156OvergrowthElementalIgnoresAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB156Seats(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Overgrowth Elemental", "Creature — Elemental", oldB156Overgrowth, false)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Elemental", 2, 2)
	life := me.Life
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d → %d for an opponent's creature", life, me.Life)
	}
}

// --- Blightbeetle -------------------------------------------------------------

func TestOldB156BlightbeetleStopsPlusOneCountersOnOpposingCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB156Seats(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Blightbeetle", "Creature — Insect", oldB156Beetle, false)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	g.RunStateChecksForTest()
	_ = g.AddCounter(theirs, game.CounterPlusOne, 2)
	_ = g.AddCounter(theirs, game.CounterMinusOne, 1)
	_ = g.AddCounter(mine, game.CounterPlusOne, 2)
	if got := effectstest.CounterCount(g, theirs, game.CounterPlusOne); got != 0 {
		t.Errorf("an opposing creature got %d +1/+1 counters, want 0", got)
	}
	if got := effectstest.CounterCount(g, theirs, game.CounterMinusOne); got != 1 {
		t.Errorf("an opposing creature got %d -1/-1 counters, want 1 (only +1/+1 is barred)", got)
	}
	if got := effectstest.CounterCount(g, mine, game.CounterPlusOne); got != 2 {
		t.Errorf("my creature got %d +1/+1 counters, want 2", got)
	}
}

func TestOldB156BlightbeetleLeavingLiftsTheBar(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB156Seats(g)
	beetle := effectstest.PushCatalogPermanent(g, me.ID, "Blightbeetle", "Creature — Insect", oldB156Beetle, false)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, beetle)
	_ = g.AddCounter(theirs, game.CounterPlusOne, 1)
	if got := effectstest.CounterCount(g, theirs, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after the Beetle left, want 1", got)
	}
}

// --- Diamond Knight -----------------------------------------------------------

func TestOldB156DiamondKnightGrowsForSpellsOfTheChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB156Seats(g)
	effectstest.CastFromHandForTest(t, g, me, "Diamond Knight", "Artifact Creature — Knight", "{3}", oldB156Diamond, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "R")
	effectstest.PassPriorityAroundTable(t, g)
	var knight uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == oldB156Diamond {
			knight = c.InstanceID
		}
	}
	if knight == uuid.Nil {
		t.Fatal("the Diamond Knight did not enter")
	}
	effectstest.CastFromHandForTest(t, g, me, "Red Spell", "Sorcery", "{R}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, knight, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after a red spell, want 1", got)
	}
	effectstest.CastFromHandForTest(t, g, me, "Blue Spell", "Sorcery", "{U}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, knight, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after a blue spell, want still 1", got)
	}
}
