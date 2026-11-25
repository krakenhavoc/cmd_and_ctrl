package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b233_test.go — Commander 2017, batch 233 (#5533): each test plays the
// real catalog entry through the real cast, activation and trigger paths.

const (
	oldB233SpelltwineOracle   = "307e68e0-5095-474c-af02-bfcb2e463763"
	oldB233NecromancerOracle  = "d1a6aa5f-2ad6-4bc7-82f0-4262f66af730"
	oldB233CauldronOracle     = "e9784eb4-96f2-409e-a7b5-eb6478644943"
	oldB233ThornsOracle       = "363b0504-27bf-40df-96ec-9e3cb1e47588"
	oldB233CrushOracle        = "92381398-94f6-471f-874a-b5894e6f9487"
	oldB233FewOracle          = "d732d9b8-dea0-41c4-8cff-67b2b3d36ecb"
	oldB233DranaOracle        = "67391f99-739a-490a-ba16-7fa1d758ec34"
	oldB233SilumgarOracle     = "b26c91d9-87de-45f7-a24c-bbe0cef3a35c"
	oldB233KolaghanOracle     = "f6e160c1-b29f-4f3d-ab55-43ffe3ab4acf"
	oldB233BaronOracle        = "326adf88-3f6d-4bc2-aafa-dd55c6c95784"
	oldB233BloodTributeOracle = "aa2da79c-b36c-4120-8d7d-f5571e87e45e"
)

func TestOldB233CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB233SpelltwineOracle:   "Spelltwine",
		oldB233NecromancerOracle:  "Apprentice Necromancer",
		oldB233CauldronOracle:     "Cauldron Dance",
		oldB233ThornsOracle:       "Rain of Thorns",
		oldB233CrushOracle:        "Relic Crush",
		oldB233FewOracle:          "Fortunate Few",
		oldB233DranaOracle:        "Drana, Kalastria Bloodchief",
		oldB233SilumgarOracle:     "Silumgar, the Drifting Death",
		oldB233KolaghanOracle:     "Kolaghan, the Storm's Fury",
		oldB233BaronOracle:        "Blood Baron of Vizkopa",
		oldB233BloodTributeOracle: "Blood Tribute",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

func oldB233Card(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func oldB233Count(g *game.Game, name string) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, z := range [][]game.Card{g.Exile.Cards, g.Battlefield.Cards, g.Seats[0].Graveyard.Cards, g.Seats[1].Graveyard.Cards} {
			for _, c := range z {
				if c.Name == name {
					n++
				}
			}
		}
	})
	return n
}

func TestOldB233SpelltwineCopiesBothCardsAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine, theirs := uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: mine, Name: "My Cantrip", TypeLine: "Instant", ManaCost: "{1}{U}", Owner: me.ID, Controller: me.ID})
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Cantrip", TypeLine: "Sorcery", ManaCost: "{2}{U}", Owner: opp.ID, Controller: opp.ID})
	twine := effectstest.CastCatalogSpell(t, g, "Spelltwine", "Sorcery", oldB233SpelltwineOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(mine) || !g.Exile.Contains(theirs) {
		t.Error("both cards were not exiled")
	}
	if !g.Exile.Contains(twine) {
		t.Error("Spelltwine did not exile itself")
	}
	if n := oldB233Count(g, "My Cantrip"); n != 2 {
		t.Errorf("My Cantrip count (original + copy) = %d, want 2", n)
	}
	if n := oldB233Count(g, "Their Cantrip"); n != 2 {
		t.Errorf("Their Cantrip count (original + copy) = %d, want 2", n)
	}
}

func TestOldB233SpelltwineRefusesAnOpponentsCardInTheFirstSlot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a, b := uuid.New(), uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: a, Name: "A", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	opp.Graveyard.PushTop(game.Card{InstanceID: b, Name: "B", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	_ = me
	id := uuid.New()
	g.Seats[0].Hand.PushTop(game.Card{InstanceID: id, Name: "Spelltwine", TypeLine: "Sorcery", OracleID: oldB233SpelltwineOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}}); err == nil {
		t.Error("two opponent cards were accepted for the your-graveyard slot")
	}
}

func TestOldB233ApprenticeNecromancerReturnsWithHasteAndSacrificesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	necro := effectstest.PushCatalogPermanent(g, me.ID, "Apprentice Necromancer", "Creature — Zombie Wizard", oldB233NecromancerOracle, false)
	dead := effectstest.GyCard(me, "Dead Bear", "Creature — Bear", "", "")
	effectstest.FloatMana(t, g, me, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, necro, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB233Card(dead)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(necro) {
		t.Error("the Necromancer was not sacrificed as a cost")
	}
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the creature card did not return")
	}
	c := effectstest.EffectiveCardForTest(t, g, dead)
	if !effectstest.EffectiveHasKeyword(&c, "haste") {
		t.Error("the returned creature has no haste")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dead) {
		t.Error("the returned creature was not sacrificed at the end step")
	}
	if effectstest.ZoneOf(g, dead) != game.ZoneGraveyard {
		t.Errorf("the sacrificed creature is in %v, want the graveyard", effectstest.ZoneOf(g, dead))
	}
}

func TestOldB233ApprenticeNecromancerTargetsOnlyYourCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	necro := effectstest.PushCatalogPermanent(g, me.ID, "Apprentice Necromancer", "Creature — Zombie Wizard", oldB233NecromancerOracle, false)
	theirs := effectstest.GyCard(opp, "Their Bear", "Creature — Bear", "", "")
	effectstest.FloatMana(t, g, me, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, necro, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB233Card(theirs)}); err == nil {
		t.Error("an opponent's graveyard card was a legal target")
	}
}

func TestOldB233CauldronDanceReturnsBothAndTheirFatesDiffer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dead := effectstest.GyCard(me, "Dead Bear", "Creature — Bear", "", "")
	fresh := effectstest.HandCard(me, "Hand Ogre", "Creature — Ogre")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Cauldron Dance", TypeLine: "Instant", OracleID: oldB233CauldronOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: oldB233Card(dead)}); err != nil {
		t.Fatalf("cast during combat: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, fresh)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) || !g.Battlefield.Contains(fresh) {
		t.Fatalf("returned = %v, put from hand = %v, want both on the battlefield", g.Battlefield.Contains(dead), g.Battlefield.Contains(fresh))
	}
	for _, c := range []uuid.UUID{dead, fresh} {
		card := effectstest.EffectiveCardForTest(t, g, c)
		if !effectstest.EffectiveHasKeyword(&card, "haste") {
			t.Errorf("%s has no haste", card.Name)
		}
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	// Both delayed triggers are yours: keep the order the engine proposes.
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Errorf("the returned creature is in %v, want your hand", effectstest.ZoneOf(g, dead))
	}
	if effectstest.ZoneOf(g, fresh) != game.ZoneGraveyard {
		t.Errorf("the creature put from hand is in %v, want the graveyard (sacrificed)", effectstest.ZoneOf(g, fresh))
	}
}

func TestOldB233CauldronDanceCanOnlyBeCastDuringCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dead := effectstest.GyCard(me, "Dead Bear", "Creature — Bear", "", "")
	id := effectstest.HandCardAtMain(t, g, "Cauldron Dance", "Instant", oldB233CauldronOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: oldB233Card(dead)}); err == nil {
		t.Error("Cauldron Dance was cast in the main phase")
	}
}

func TestOldB233RainOfThornsDestroysEachChosenKind(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.B12Push(g, opp.ID, "Rock", "Artifact", "", 0, 0)
	ench := effectstest.B12Push(g, opp.ID, "Aura", "Enchantment", "", 0, 0)
	land := effectstest.B12Push(g, opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpellWithModesTargets(t, g, oldB233ThornsOracle, []int{0, 2},
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}, {Kind: game.TargetCard, ID: land}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(land) {
		t.Error("a chosen artifact or land survived")
	}
	if !g.Battlefield.Contains(ench) || !g.Battlefield.Contains(bear) {
		t.Error("the unchosen enchantment mode, or a creature, was destroyed")
	}
}

func TestOldB233RelicCrushDestroysBothTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.B12Push(g, opp.ID, "Rock", "Artifact", "", 0, 0)
	ench := effectstest.B12Push(g, opp.ID, "Aura", "Enchantment", "", 0, 0)
	bystander := effectstest.B12Push(g, opp.ID, "Idol", "Artifact", "", 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Relic Crush", "Instant", oldB233CrushOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}, {Kind: game.TargetCard, ID: ench}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(ench) {
		t.Error("a target survived")
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("an untargeted artifact was destroyed")
	}
}

func TestOldB233RelicCrushWorksWithOneTargetAndRefusesTheSameTwice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Push(g, opp.ID, "Rock", "Artifact", "", 0, 0)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Relic Crush", TypeLine: "Instant", OracleID: oldB233CrushOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}, {Kind: game.TargetCard, ID: rock}}}); err == nil {
		t.Error("the same permanent was accepted for both slots")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: oldB233Card(rock)}); err != nil {
		t.Fatalf("one target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the single target survived")
	}
}

func TestOldB233FortunateFewEveryPlayerSavesOnePermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, p1, p2, p3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	a := effectstest.B12Creature(g, p1.ID, "A Bear", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, p2.ID, "B Bear", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, p3.ID, "C Bear", "Creature — Bear", 2, 2)
	c2 := effectstest.B12Creature(g, p3.ID, "C Other", "Creature — Bear", 2, 2)
	land := effectstest.B12Push(g, p1.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Fortunate Few", "Sorcery", oldB233FewOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// You choose first, from permanents you don't control.
	effectstest.SlChoose(t, g, me.ID, a)
	effectstest.SlChoose(t, g, p1.ID, b)
	effectstest.SlChoose(t, g, p2.ID, c)
	effectstest.SlChoose(t, g, p3.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"A Bear": a, "B Bear": b, "C Bear": c, "My Bear": mine, "Forest": land} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s did not survive", name)
		}
	}
	if g.Battlefield.Contains(c2) {
		t.Error("an unchosen nonland permanent survived")
	}
}

func TestOldB233DranaShrinksTheTargetAndGrowsByX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	drana := effectstest.PushCatalogPermanent(g, me.ID, "Drana, Kalastria Bloodchief", "Legendary Creature — Vampire Shaman", oldB233DranaOracle, false)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 4)
	p0, t0 := effectstest.EffectivePower(t, g, drana), effectstest.EffectiveToughness(t, g, drana)
	effectstest.FloatMana(t, g, me, "{B}{B}{B}{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, drana, 0, game.ActivateAbilityParams{Strict: true, XValue: 3, Targets: oldB233Card(victim)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectiveToughness(t, g, victim); got != 1 {
		t.Errorf("the victim's toughness is %d, want 1", got)
	}
	if got := effectstest.EffectivePower(t, g, drana); got != p0+3 {
		t.Errorf("Drana's power is %d, want %d", got, p0+3)
	}
	if got := effectstest.EffectiveToughness(t, g, drana); got != t0 {
		t.Errorf("Drana's toughness changed to %d", got)
	}
}

func TestOldB233SilumgarShrinksDefendersCreaturesPerAttackingDragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	silumgar := effectstest.B12Push(g, me.ID, "Silumgar, the Drifting Death", "Legendary Creature — Dragon", oldB233SilumgarOracle, 3, 7)
	whelp := effectstest.B12Creature(g, me.ID, "Whelp", "Creature — Dragon", 2, 2)
	human := effectstest.B12Creature(g, me.ID, "Human", "Creature — Human", 2, 2)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 4, 4)
	other := effectstest.B12Creature(g, g.Seats[2].ID, "Bystander", "Creature — Bear", 4, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range []uuid.UUID{silumgar, whelp, human} {
		if err := g.DeclareAttacker(a, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, foe); p != 2 || tt != 2 {
		t.Errorf("the defender's creature is %d/%d, want 2/2 (two Dragons attacked)", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, other); p != 4 || tt != 4 {
		t.Errorf("another player's creature is %d/%d, want 4/4", p, tt)
	}
	if p, _ := effectstest.PtOf(t, g, human); p != 2 {
		t.Errorf("your own creature changed to power %d", p)
	}
}

func TestOldB233KolaghanPumpsYourCreaturesPerAttackingDragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kolaghan := effectstest.B12Push(g, me.ID, "Kolaghan, the Storm's Fury", "Legendary Creature — Dragon", oldB233KolaghanOracle, 4, 5)
	human := effectstest.B12Creature(g, me.ID, "Human", "Creature — Human", 2, 2)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 4, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(kolaghan, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, human); p != 3 {
		t.Errorf("the idle Human has power %d, want 3", p)
	}
	if p, _ := effectstest.PtOf(t, g, kolaghan); p != 5 {
		t.Errorf("Kolaghan has power %d, want 5", p)
	}
	if p, _ := effectstest.PtOf(t, g, foe); p != 4 {
		t.Errorf("the opponent's creature changed to power %d", p)
	}
}

func TestOldB233KolaghanIgnoresANonDragonAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Kolaghan, the Storm's Fury", "Legendary Creature — Dragon", oldB233KolaghanOracle, 4, 5)
	human := effectstest.B12Creature(g, me.ID, "Human", "Creature — Human", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(human, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, human); p != 2 {
		t.Errorf("a non-Dragon attacker triggered Kolaghan: power %d", p)
	}
}

func TestOldB233KolaghanDashesWithHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastWithAltCost(t, g, "Kolaghan, the Storm's Fury", "Legendary Creature — Dragon", oldB233KolaghanOracle, "dash")
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatalf("the dashed Kolaghan is in %v", effectstest.ZoneOf(g, id))
	}
	c := effectstest.EffectiveCardForTest(t, g, id)
	if !effectstest.EffectiveHasKeyword(&c, "haste") {
		t.Error("the dashed Kolaghan has no haste")
	}
}

func TestOldB233BloodBaronNeedsBothLifeTotals(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	baron := effectstest.B12Push(g, me.ID, "Blood Baron of Vizkopa", "Creature — Vampire", oldB233BaronOracle, 4, 4)
	check := func(label string, wantP int, wantFlying bool) {
		t.Helper()
		g.RunStateChecksForTest()
		if p, _ := effectstest.PtOf(t, g, baron); p != wantP {
			t.Errorf("%s: power %d, want %d", label, p, wantP)
		}
		c := effectstest.EffectiveCardForTest(t, g, baron)
		if effectstest.EffectiveHasKeyword(&c, "flying") != wantFlying {
			t.Errorf("%s: flying = %v, want %v", label, !wantFlying, wantFlying)
		}
	}
	setLife := func(p *game.Player, to int) {
		g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, to-p.Life) })
	}
	setLife(me, 40)
	setLife(opp, 40)
	check("opponents healthy", 4, false)
	setLife(me, 29)
	setLife(opp, 10)
	check("you at 29", 4, false)
	setLife(me, 30)
	check("you at 30, opponent at 10", 10, true)
	setLife(opp, 11)
	check("opponent at 11", 4, false)
	for _, kw := range []string{"lifelink", "protection from white", "protection from black"} {
		c := effectstest.EffectiveCardForTest(t, g, baron)
		if !effectstest.EffectiveHasKeyword(&c, kw) {
			t.Errorf("the Baron lacks %q", kw)
		}
	}
}

func oldB233TributeTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vamp := effectstest.B12Push(g, me.ID, "Test Vampire", "Creature — Vampire", "", 2, 2)
	id := effectstest.HandCardAtMain(t, g, "Blood Tribute", "Sorcery", oldB233BloodTributeOracle)
	g.WithWriteLock(func() { opp.Life = 21 })
	return g, me, opp, id, vamp
}

func TestOldB233BloodTributeLosesHalfRoundedUpWithoutKicker(t *testing.T) {
	g, me, opp, id, _ := oldB233TributeTable(t)
	mine := me.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 10 {
		t.Errorf("the opponent is at %d, want 10 (21 loses 11)", opp.Life)
	}
	if me.Life != mine {
		t.Errorf("you gained life without the kicker: %d", me.Life)
	}
}

func TestOldB233BloodTributeKickedTapsAVampireAndGainsTheLifeLost(t *testing.T) {
	g, me, opp, id, vamp := oldB233TributeTable(t)
	mine := me.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, TeamworkIDs: []uuid.UUID{vamp},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("kicked cast: %v", err)
	}
	if !effectstest.AlliTapped(t, g, vamp) {
		t.Error("the Vampire was not tapped to pay the kicker")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 10 {
		t.Errorf("the opponent is at %d, want 10", opp.Life)
	}
	if me.Life != mine+11 {
		t.Errorf("you are at %d, want %d (the 11 life lost)", me.Life, mine+11)
	}
}

func TestOldB233BloodTributeKickerRefusesANonVampire(t *testing.T) {
	g, me, opp, id, _ := oldB233TributeTable(t)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, TeamworkIDs: []uuid.UUID{bear},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Error("a Bear paid the Vampire kicker")
	}
}

func TestOldB233BloodTributeKickerRefusesATappedVampire(t *testing.T) {
	g, me, opp, id, vamp := oldB233TributeTable(t)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == vamp {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, TeamworkIDs: []uuid.UUID{vamp},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Error("a tapped Vampire paid the kicker")
	}
}
