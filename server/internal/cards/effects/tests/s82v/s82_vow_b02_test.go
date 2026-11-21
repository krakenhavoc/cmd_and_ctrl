package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_vow_b02_test.go — Innistrad: Crimson Vow, batch 2 (#5495): thirteen
// simple creatures and one sorcery, each driven through the engine.

const (
	vowLightningWolf   = "d50f1ff7-d2bc-4e36-876d-77648f447f5b"
	vowEstwald         = "a9f111aa-5d34-4317-9a03-35cd94a888c7"
	vowMilitiaRallier  = "d5377597-3454-4636-8563-7e6a5544e36a"
	vowResistanceSquad = "a16eeef9-e03c-4ebb-b986-325252a9491f"
	vowDawnhartDisc    = "e86279ba-cb22-4754-90f4-38354657b8fa"
	vowDawnhartGeist   = "ed7a2ee0-289d-4e5f-92e0-1539395baf33"
	vowFrenziedDevils  = "4c912e9a-21f9-4058-b28c-43f319e35d09"
	vowCruelWitness    = "c1f1ca09-897a-434c-bbb1-e4ad4a1547aa"
	vowFlourishingHunt = "d714dc46-2db2-439f-b558-2b69d5c36b46"
	vowCourierBat      = "d8ff239c-cd71-4cd2-9e96-792a116989ca"
	vowSkulkingKiller  = "e711a2d6-008d-4e7b-986e-f34805165679"
	vowQuartermaster   = "14a555eb-9720-45b4-9eb5-77b739e1045c"
	vowArmTheCathars   = "c6a1a91f-46b9-4e41-a3d3-3423b02c42a5"
)

func TestS82VowB02EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		vowLightningWolf: "Lightning Wolf", vowEstwald: "Estwald Shieldbasher", vowMilitiaRallier: "Militia Rallier",
		vowResistanceSquad: "Resistance Squad", vowDawnhartDisc: "Dawnhart Disciple", vowDawnhartGeist: "Dawnhart Geist",
		vowFrenziedDevils: "Frenzied Devils", vowCruelWitness: "Cruel Witness", vowFlourishingHunt: "Flourishing Hunter",
		vowCourierBat: "Courier Bat", vowSkulkingKiller: "Skulking Killer", vowQuartermaster: "Angelic Quartermaster",
		vowArmTheCathars: "Arm the Cathars",
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

// vowCast puts a creature card in p's hand and casts it.
func vowCast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
	})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

func vowPick(t *testing.T, g *game.Game, chooser uuid.UUID, ids ...uuid.UUID) error {
	t.Helper()
	pick := effectstest.LatestPickTarget(g, chooser)
	if pick == nil {
		t.Fatal("no pick_target prompt")
	}
	return g.ResolvePickTargets(pick.ID, chooser, effectstest.CardRefs(ids...))
}

func vowPower(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := effectstest.FindBattlefieldCardForTest(g, id)
	if c == nil {
		t.Fatal("not on the battlefield")
	}
	return c.CurrentPower()
}

func vowToughness(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := effectstest.FindBattlefieldCardForTest(g, id)
	if c == nil {
		t.Fatal("not on the battlefield")
	}
	return c.CurrentToughness()
}

func TestS82VowB02LightningWolfGainsFirstStrikeAtSorcerySpeedOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wolf := effectstest.B21Push(g, me.ID, "Lightning Wolf", "Creature — Wolf", vowLightningWolf, 4, 3, "R")
	effectstest.AdvanceToMain(t, g)
	// A pool that cannot pay is refused.
	if err := g.ActivateCatalogAbility(me.ID, wolf, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with an empty pool")
	}
	effectstest.B06AddMana(me, "C", "R")
	if err := g.ActivateCatalogAbility(me.ID, wolf, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate in main phase: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, wolf, "first strike") {
		t.Error("no first strike after the ability resolved")
	}
	// Not in combat.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	wolf2 := effectstest.B21Push(g2, me2.ID, "Lightning Wolf", "Creature — Wolf", vowLightningWolf, 4, 3, "R")
	effectstest.AdvanceTo(t, g2, game.StepDeclareAttackers)
	effectstest.B06AddMana(me2, "C", "R")
	if err := g2.ActivateCatalogAbility(me2.ID, wolf2, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated during combat")
	}
}

func TestS82VowB02FrenziedDevilsGrowsOnNoncreatureSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	devils := effectstest.B21Push(g, me.ID, "Frenzied Devils", "Creature — Devil", vowFrenziedDevils, 3, 3, "R")
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := vowPower(t, g, devils), vowToughness(t, g, devils); p != 5 || th != 5 {
		t.Errorf("after a noncreature spell %d/%d, want 5/5", p, th)
	}
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := vowPower(t, g, devils); p != 5 {
		t.Errorf("a creature spell changed the power to %d", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, devils, "haste") {
		t.Error("no haste")
	}
}

func TestS82VowB02CruelWitnessSurveilsOnNoncreatureSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B21Push(g, me.ID, "Cruel Witness", "Creature — Bird Horror", vowCruelWitness, 3, 3, "U")
	top := effectstest.SeedLibrary(me, "Top Card")[0]
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SurveilChoiceFor(g, me.ID) != nil {
		t.Fatal("a creature spell surveilled")
	}
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("a noncreature spell did not surveil")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{top}, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if !me.Graveyard.Contains(top) {
		t.Error("the surveilled card did not reach the graveyard")
	}
}

func TestS82VowB02DawnhartGeistGainsLifeOnEnchantmentSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B21Push(g, me.ID, "Dawnhart Geist", "Creature — Spirit Warlock", vowDawnhartGeist, 1, 3, "W")
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before {
		t.Fatalf("a creature spell gained life: %d → %d", before, me.Life)
	}
	effectstest.CastCatalogSpell(t, g, "Pacifism", "Enchantment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+2 {
		t.Errorf("an enchantment spell: %d → %d, want +2", before, me.Life)
	}
}

func TestS82VowB02DawnhartDiscipleGrowsWhenAnotherHumanEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	disc := effectstest.B21Push(g, me.ID, "Dawnhart Disciple", "Creature — Human Warlock", vowDawnhartDisc, 2, 2, "G")
	effectstest.ToMain(t, g)
	vowCast(t, g, me, "Villager", "Creature — Human Peasant", "", 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if p := vowPower(t, g, disc); p != 3 {
		t.Fatalf("another Human entered: power %d, want 3", p)
	}
	vowCast(t, g, me, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if p := vowPower(t, g, disc); p != 3 {
		t.Errorf("a non-Human entered: power %d, want 3", p)
	}
	effectstest.B21Push(g, opp.ID, "Their Human", "Creature — Human", "", 1, 1)
	if p := vowPower(t, g, disc); p != 3 {
		t.Errorf("an opponent's Human changed the power to %d", p)
	}
}

func TestS82VowB02ResistanceSquadDrawsOnlyWithAnotherHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B")
	effectstest.B21Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	before := len(me.Hand.Cards)
	vowCast(t, g, me, "Resistance Squad", "Creature — Human Soldier", vowResistanceSquad, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != before {
		t.Fatalf("with no other Human the hand went %d → %d", before, got)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.ToMain(t, g2)
	effectstest.SeedLibrary(me2, "A", "B")
	effectstest.B21Push(g2, me2.ID, "Villager", "Creature — Human Peasant", "", 1, 1)
	before = len(me2.Hand.Cards)
	vowCast(t, g2, me2, "Resistance Squad", "Creature — Human Soldier", vowResistanceSquad, 3, 2)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := len(me2.Hand.Cards); got != before+1 {
		t.Errorf("with another Human the hand went %d → %d, want +1", before, got)
	}
}

func vowGainLife(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, n) })
}

func TestS82VowB02CourierBatReturnsACreatureOnlyAfterGainingLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	vowCast(t, g, me, "Courier Bat", "Creature — Bat", vowCourierBat, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || !me.Graveyard.Contains(dead) {
		t.Fatal("the Bat triggered without any life gained")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.ToMain(t, g2)
	dead2 := uuid.New()
	me2.Graveyard.PushTop(game.Card{InstanceID: dead2, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me2.ID, Controller: me2.ID})
	land := uuid.New()
	me2.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Dead Land", TypeLine: "Land", Owner: me2.ID, Controller: me2.ID})
	vowGainLife(g2, me2, 1)
	vowCast(t, g2, me2, "Courier Bat", "Creature — Bat", vowCourierBat, 2, 2)
	effectstest.PassPriorityAroundTable(t, g2)
	if err := vowPick(t, g2, me2.ID, land); err == nil {
		t.Fatal("a land card was accepted")
	}
	if err := vowPick(t, g2, me2.ID, dead2); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if !me2.Hand.Contains(dead2) {
		t.Error("the creature card was not returned to hand")
	}
}

func TestS82VowB02FlourishingHunterGainsLifeEqualToGreatestOtherToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	effectstest.B21Push(g, me.ID, "Wall", "Creature — Wall", "", 0, 5)
	effectstest.B21Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.B21Push(g, opp.ID, "Their Wall", "Creature — Wall", "", 0, 9)
	before := me.Life
	vowCast(t, g, me, "Flourishing Hunter", "Creature — Wolf Spirit", vowFlourishingHunt, 6, 6)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+5 {
		t.Errorf("life %d → %d, want +5 (its own 6 and their 9 do not count)", before, me.Life)
	}
}

func TestS82VowB02SkulkingKillerShrinksOnlyAnOpponentsLoneCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	lone := effectstest.B21Push(g, opp.ID, "Lone Bear", "Creature — Bear", "", 3, 3)
	vowCast(t, g, me, "Skulking Killer", "Creature — Vampire Assassin", vowSkulkingKiller, 4, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if err := vowPick(t, g, me.ID, lone); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := vowPower(t, g, lone); p != 1 {
		t.Errorf("the lone creature is %d power, want 1", p)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	effectstest.ToMain(t, g2)
	a := effectstest.B21Push(g2, opp2.ID, "Bear A", "Creature — Bear", "", 3, 3)
	effectstest.B21Push(g2, opp2.ID, "Bear B", "Creature — Bear", "", 3, 3)
	vowCast(t, g2, me2, "Skulking Killer", "Creature — Vampire Assassin", vowSkulkingKiller, 4, 2)
	effectstest.PassPriorityAroundTable(t, g2)
	if err := vowPick(t, g2, me2.ID, a); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if p := vowPower(t, g2, a); p != 3 {
		t.Errorf("with another creature the target is %d power, want 3", p)
	}
}

func TestS82VowB02MilitiaRallierCantAttackAloneAndUntapsATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rallier := effectstest.B12Push(g, me.ID, "Militia Rallier", "Creature — Human Soldier", vowMilitiaRallier, 3, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(rallier, opp.ID); err == nil {
		t.Fatal("the Rallier attacked alone")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	rallier2 := effectstest.B12Push(g2, me2.ID, "Militia Rallier", "Creature — Human Soldier", vowMilitiaRallier, 3, 3)
	bear := effectstest.B12Push(g2, me2.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.DeclareAttack(t, g2, opp2.ID, rallier2, bear)
	if effectstest.FindBattlefieldCardForTest(g2, bear).Tapped == false {
		t.Fatal("the attacking Bear is not tapped")
	}
	if err := vowPick(t, g2, me2.ID, bear); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if effectstest.FindBattlefieldCardForTest(g2, bear).Tapped {
		t.Error("the Bear is still tapped")
	}
}

func TestS82VowB02AngelicQuartermasterCountersUpToTwoOtherCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	bear := effectstest.B21Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B21Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	third := effectstest.B21Push(g, me.ID, "Elf", "Creature — Elf", "", 1, 1)
	q := vowCast(t, g, me, "Angelic Quartermaster", "Creature — Angel Soldier", vowQuartermaster, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if err := vowPick(t, g, me.ID, q); err == nil {
		t.Fatal("the Quartermaster itself was accepted")
	}
	if err := vowPick(t, g, me.ID, bear, theirs, third); err == nil {
		t.Fatal("three targets were accepted")
	}
	if err := vowPick(t, g, me.ID, bear, theirs); err != nil {
		t.Fatalf("two other creatures, one an opponent's: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, theirs} {
		if effectstest.B12Counter(t, g, id, game.CounterPlusOne) != 1 {
			t.Error("a target did not get a +1/+1 counter")
		}
	}
	if effectstest.B12Counter(t, g, third, game.CounterPlusOne) != 0 || effectstest.B12Counter(t, g, q, game.CounterPlusOne) != 0 {
		t.Error("a non-target got a counter")
	}
}

func TestS82VowB02EstwaldShieldbasherGainsIndestructibleOnlyWhenYouPay(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		sb := effectstest.B12Push(g, me.ID, "Estwald Shieldbasher", "Creature — Human Soldier", vowEstwald, 4, 2)
		effectstest.DeclareAttack(t, g, opp.ID, sb)
		effectstest.B06AddMana(me, "W")
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.HasEffectiveKeyword(t, g, sb, "indestructible"); got != pay {
			t.Errorf("pay=%v: indestructible=%v", pay, got)
		}
	}
	// A pool that cannot pay gets nothing.
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sb := effectstest.B12Push(g, me.ID, "Estwald Shieldbasher", "Creature — Human Soldier", vowEstwald, 4, 2)
	effectstest.DeclareAttack(t, g, opp.ID, sb)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, sb, "indestructible") {
		t.Error("indestructible without paying")
	}
}

func TestS82VowB02ArmTheCatharsPumpsThreeDistinctCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	a := effectstest.B12Creature(g, g.Seats[0].ID, "A", "Creature — Bear", 1, 1)
	c := effectstest.B12Creature(g, g.Seats[1].ID, "C", "Creature — Bear", 1, 1)
	if err := effectstest.CastCatalogSpellErr(t, g, "Arm the Cathars", "Sorcery", vowArmTheCathars, effectstest.CardRefs(a, a, c)); err == nil {
		t.Fatal("the same creature was accepted for two clauses")
	}

	g = effectstest.NewCatalogGame(t)
	a = effectstest.B12Creature(g, g.Seats[0].ID, "A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, g.Seats[0].ID, "B", "Creature — Bear", 1, 1)
	c = effectstest.B12Creature(g, g.Seats[1].ID, "C", "Creature — Bear", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Arm the Cathars", "Sorcery", vowArmTheCathars, effectstest.CardRefs(a, b, c))
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{a: 4, b: 3, c: 2} {
		if got := vowPower(t, g, id); got != want {
			t.Errorf("power %d, want %d", got, want)
		}
		if !effectstest.HasEffectiveKeyword(t, g, id, "vigilance") {
			t.Error("a target has no vigilance")
		}
	}

	g = effectstest.NewCatalogGame(t)
	a = effectstest.B12Creature(g, g.Seats[0].ID, "A", "Creature — Bear", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Arm the Cathars", "Sorcery", vowArmTheCathars, effectstest.CardRefs(a))
	effectstest.PassPriorityAroundTable(t, g)
	if got := vowPower(t, g, a); got != 4 {
		t.Errorf("a lone target has power %d, want 4", got)
	}
}
