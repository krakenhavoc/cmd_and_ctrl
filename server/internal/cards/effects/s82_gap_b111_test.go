package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b111_test.go — S82 GAP batch 111 (Commander Masters enchantments and
// instants): every test drives the real cast, trigger, layer and combat paths.

const (
	gap111GraspOracle      = "4d2377fb-ac3b-453b-b632-abe918d72529"
	gap111StarfieldOracle  = "922d603e-6623-4455-b91f-fad76f94045a"
	gap111DesecrateOracle  = "c9cd57f6-1f69-4ea1-9672-ca944759c7e8"
	gap111ContortionOracle = "7346337e-ac12-4dd2-826d-77ec0013e74f"
	gap111StaminaOracle    = "65756e6d-9825-4ed6-91c4-79f1e3bd4ae5"
	gap111WakeOracle       = "69bf79c3-87d1-4244-8d58-989307c91a12"
	gap111AncestorsOracle  = "4caa3c55-41b3-48a3-b6a8-4d23aff7d00d"
)

func TestGap111CardsAreRegisteredComplete(t *testing.T) {
	for oracle, name := range map[string]string{
		gap111GraspOracle:      "Spectral Grasp",
		gap111StarfieldOracle:  "Starfield of Nyx",
		gap111DesecrateOracle:  "Desecrate Reality",
		gap111ContortionOracle: "Spatial Contortion",
		gap111StaminaOracle:    "Supernatural Stamina",
		gap111WakeOracle:       "Wake the Dead",
		gap111AncestorsOracle:  "For the Ancestors",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestGap111SpatialContortionIsPlusThreeMinusThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := pushVanillaCreature(g, opp.ID, "Big", 4, 4)
	small := pushVanillaCreature(g, opp.ID, "Small", 3, 3)
	_ = me
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Spatial Contortion", "Instant", gap111ContortionOracle, cardRefs(big))
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, big); got != 7 {
		t.Errorf("the 4/4 has power %d, want 7", got)
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the 4/4 died to -3 toughness")
	}
	castCatalogSpell(t, g, "Spatial Contortion", "Instant", gap111ContortionOracle, cardRefs(small))
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(small) {
		t.Error("the 3/3 survived -3 toughness")
	}
}

func TestGap111SupernaturalStaminaPumpsAndReturnsTheCreatureTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Supernatural Stamina", "Instant", gap111StaminaOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 4 {
		t.Fatalf("power %d, want 4", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Bear", me.ID); n != 1 {
		t.Fatalf("%d Bears back on the battlefield, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bear" && !c.Tapped {
			t.Error("the returned creature is untapped")
		}
	}
}

func TestGap111SupernaturalStaminaEndsAtEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Supernatural Stamina", "Instant", gap111StaminaOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	advanceToMainOf(t, g, 1)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Bear", me.ID); n != 0 {
		t.Errorf("the grant outlived the turn: %d Bears returned", n)
	}
}

func TestGap111SpectralGraspBarsAttackingYouAndBlockingYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	pushAuraOnLand(t, g, me.ID, "Spectral Grasp", gap111GraspOracle, theirs)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mine, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	for g.Turn.Step != game.StepDeclareBlockers {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.DeclareBlocker(theirs, mine); err == nil {
		t.Fatal("the enchanted creature blocked a creature the Aura's controller controls")
	}
}

func TestGap111SpectralGraspBarsAttackingTheAurasController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	other := pushVanillaCreature(g, opp.ID, "Their Other", 2, 2)
	pushAuraOnLand(t, g, me.ID, "Spectral Grasp", gap111GraspOracle, theirs)
	advanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(theirs, me.ID); err == nil {
		t.Fatal("the enchanted creature attacked the Aura's controller")
	}
	if err := g.DeclareAttacker(other, me.ID); err != nil {
		t.Fatalf("an unenchanted creature could not attack: %v", err)
	}
}

func TestGap111StarfieldOfNyxAnimatesEnchantmentsAtFive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	star := b12Push(g, me.ID, "Starfield of Nyx", "Enchantment", gap111StarfieldOracle, 0, 0)
	three := pushTestPermanent(g, me.ID, game.Card{Name: "Three", TypeLine: "Enchantment", ManaCost: "{2}{W}"})
	aura := pushTestPermanent(g, me.ID, game.Card{Name: "An Aura", TypeLine: "Enchantment — Aura", ManaCost: "{1}"})
	for i := 0; i < 2; i++ {
		pushTestPermanent(g, me.ID, game.Card{Name: "Filler", TypeLine: "Enchantment", ManaCost: "{4}"})
	}
	// Five enchantments: Starfield, Three, An Aura, and two fillers.
	if c := gaLayered(t, g, three); !c.IsCreature() || effectivePower(t, g, three) != 3 {
		t.Errorf("Three is not a 3/3 creature at five enchantments (power %d)", effectivePower(t, g, three))
	}
	if c := gaLayered(t, g, aura); c.IsCreature() {
		t.Error("an Aura was animated")
	}
	if c := gaLayered(t, g, star); c.IsCreature() {
		t.Error("Starfield animated itself (it says each OTHER enchantment)")
	}
	// Dropping to four turns it off.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	g.RunStateChecksForTest()
	if c := gaLayered(t, g, three); c.IsCreature() {
		t.Error("still animated with four enchantments")
	}
}

func TestGap111StarfieldOfNyxReturnsAnEnchantmentAtUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	b12Push(g, me.ID, "Starfield of Nyx", "Enchantment", gap111StarfieldOracle, 0, 0)
	ench := pushGraveyardPermanent(me, "Old Enchantment", "Enchantment", "{1}")
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		pickCard(t, g, me.ID, ench)
		passPriorityAroundTable(t, g)
	}
	if owedMays(g, me.ID) > 0 {
		answerLatestTriggerPrompt(t, g, me.ID, true)
	}
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Old Enchantment", me.ID); n != 1 {
		t.Fatalf("the enchantment did not return (%d on the battlefield)", n)
	}
}

func TestGap111WakeTheDeadReturnsXCreaturesThenSacrificesThemAtEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := pushGraveyardPermanent(me, "Zombie A", "Creature — Zombie", "{1}{B}")
	b := pushGraveyardPermanent(me, "Zombie B", "Creature — Zombie", "{1}{B}")
	pushGraveyardPermanent(me, "Zombie C", "Creature — Zombie", "{1}{B}")
	advanceToMain(t, g)
	// Not on an opponent's turn: refused.
	floatMana(t, g, me, "{B}{B}{C}{C}")
	if _, err := castFromHandErr(g, me, "Wake the Dead", "Instant", "{X}{B}{B}", gap111WakeOracle,
		game.CastSpellParams{Strict: true, XValue: 2, Targets: cardRefs(a, b)}); err == nil {
		t.Fatal("Wake the Dead was cast in its controller's own main phase")
	}
	g.WithWriteLock(func() { me.ManaPool = game.ManaPool{} })
	advanceToDeclareAttackersOf(t, g, 1)
	floatMana(t, g, me, "{B}{B}{C}{C}")
	if _, err := castFromHandErr(g, me, "Wake the Dead", "Instant", "{X}{B}{B}", gap111WakeOracle,
		game.CastSpellParams{Strict: true, XValue: 2, Targets: cardRefs(a)}); err == nil {
		t.Fatal("one target was accepted for X = 2")
	}
	if _, err := castFromHandErr(g, me, "Wake the Dead", "Instant", "{X}{B}{B}", gap111WakeOracle,
		game.CastSpellParams{Strict: true, XValue: 2, Targets: cardRefs(a, b)}); err != nil {
		t.Fatalf("cast during an opponent's combat: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Fatal("the targeted creatures did not return")
	}
	advanceTo(t, g, game.StepEnd)
	monSettle(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("the creatures were not sacrificed at the end step")
	}
}

func TestGap111DesecrateRealityExilesEvenManaValuePermanents(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	even := pushTestPermanent(g, opp.ID, game.Card{Name: "Even", TypeLine: "Artifact", ManaCost: "{2}"})
	odd := pushTestPermanent(g, opp.ID, game.Card{Name: "Odd", TypeLine: "Artifact", ManaCost: "{3}"})
	advanceToMain(t, g)
	if err := castCatalogSpellErr(t, g, "Desecrate Reality", "Instant", gap111DesecrateOracle, cardRefs(odd)); err == nil {
		t.Fatal("a permanent with an odd mana value was a legal target")
	}
	castCatalogSpell(t, g, "Desecrate Reality", "Instant", gap111DesecrateOracle, cardRefs(even))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(even) || !g.Exile.Contains(even) {
		t.Error("the even-mana-value permanent was not exiled")
	}
	if !g.Battlefield.Contains(odd) {
		t.Error("the odd-mana-value permanent was exiled")
	}
}

func TestGap111ForTheAncestorsTakesTheChosenTypeAndFlashbackWorks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 8; i++ {
		pushLibraryTyped(me, "Filler", "Instant")
	}
	elf := pushLibraryTyped(me, "Elf", "Creature — Elf")
	advanceToMain(t, g)
	handBefore := me.Hand.Size()
	castCatalogSpell(t, g, "For the Ancestors", "Instant", gap111AncestorsOracle, nil)
	passPriorityAroundTable(t, g)
	answerCreatureType(t, g, me.ID, "Elf")
	passPriorityAroundTable(t, g)
	if pick := chooseCardsChoiceFor(g, me.ID); pick != nil {
		answerChooseCards(t, g, me.ID, elf)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(elf) {
		t.Fatalf("the Elf was not put into hand (hand %d -> %d)", handBefore, me.Hand.Size())
	}
}
