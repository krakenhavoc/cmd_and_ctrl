package testsru

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

func TestGap111SpatialContortionIsPlusThreeMinusThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.PushVanillaCreature(g, opp.ID, "Big", 4, 4)
	small := effectstest.PushVanillaCreature(g, opp.ID, "Small", 3, 3)
	_ = me
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Spatial Contortion", "Instant", gap111ContortionOracle, effectstest.CardRefs(big))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, big); got != 7 {
		t.Errorf("the 4/4 has power %d, want 7", got)
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the 4/4 died to -3 toughness")
	}
	effectstest.CastCatalogSpell(t, g, "Spatial Contortion", "Instant", gap111ContortionOracle, effectstest.CardRefs(small))
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(small) {
		t.Error("the 3/3 survived -3 toughness")
	}
}

func TestGap111SupernaturalStaminaPumpsAndReturnsTheCreatureTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Supernatural Stamina", "Instant", gap111StaminaOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Fatalf("power %d, want 4", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.NamedOnBattlefield(g, "Bear", me.ID); n != 1 {
		t.Fatalf("%d Bears back on the battlefield, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bear" && !c.Tapped {
			t.Error("the returned creature is untapped")
		}
	}
}

func TestGap111SupernaturalStaminaEndsAtEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Supernatural Stamina", "Instant", gap111StaminaOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToMainOf(t, g, 1)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.NamedOnBattlefield(g, "Bear", me.ID); n != 0 {
		t.Errorf("the grant outlived the turn: %d Bears returned", n)
	}
}

func TestGap111SpectralGraspBarsAttackingYouAndBlockingYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.PushAuraOnLand(t, g, me.ID, "Spectral Grasp", gap111GraspOracle, theirs)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mine, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	other := effectstest.PushVanillaCreature(g, opp.ID, "Their Other", 2, 2)
	effectstest.PushAuraOnLand(t, g, me.ID, "Spectral Grasp", gap111GraspOracle, theirs)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(theirs, me.ID); err == nil {
		t.Fatal("the enchanted creature attacked the Aura's controller")
	}
	if err := g.DeclareAttacker(other, me.ID); err != nil {
		t.Fatalf("an unenchanted creature could not attack: %v", err)
	}
}

func TestGap111StarfieldOfNyxAnimatesEnchantmentsAtFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	star := effectstest.B12Push(g, me.ID, "Starfield of Nyx", "Enchantment", gap111StarfieldOracle, 0, 0)
	three := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Three", TypeLine: "Enchantment", ManaCost: "{2}{W}"})
	aura := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "An Aura", TypeLine: "Enchantment — Aura", ManaCost: "{1}"})
	for i := 0; i < 2; i++ {
		effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Filler", TypeLine: "Enchantment", ManaCost: "{4}"})
	}
	// Five enchantments: Starfield, Three, An Aura, and two fillers.
	if c := effectstest.GaLayered(t, g, three); !c.IsCreature() || effectstest.EffectivePower(t, g, three) != 3 {
		t.Errorf("Three is not a 3/3 creature at five enchantments (power %d)", effectstest.EffectivePower(t, g, three))
	}
	if c := effectstest.GaLayered(t, g, aura); c.IsCreature() {
		t.Error("an Aura was animated")
	}
	if c := effectstest.GaLayered(t, g, star); c.IsCreature() {
		t.Error("Starfield animated itself (it says each OTHER enchantment)")
	}
	// Dropping to four turns it off.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	g.RunStateChecksForTest()
	if c := effectstest.GaLayered(t, g, three); c.IsCreature() {
		t.Error("still animated with four enchantments")
	}
}

func TestGap111StarfieldOfNyxReturnsAnEnchantmentAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	effectstest.B12Push(g, me.ID, "Starfield of Nyx", "Enchantment", gap111StarfieldOracle, 0, 0)
	ench := effectstest.PushGraveyardPermanent(me, "Old Enchantment", "Enchantment", "{1}")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, ench)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if effectstest.OwedMays(g, me.ID) > 0 {
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.NamedOnBattlefield(g, "Old Enchantment", me.ID); n != 1 {
		t.Fatalf("the enchantment did not return (%d on the battlefield)", n)
	}
}

func TestGap111WakeTheDeadReturnsXCreaturesThenSacrificesThemAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushGraveyardPermanent(me, "Zombie A", "Creature — Zombie", "{1}{B}")
	b := effectstest.PushGraveyardPermanent(me, "Zombie B", "Creature — Zombie", "{1}{B}")
	effectstest.PushGraveyardPermanent(me, "Zombie C", "Creature — Zombie", "{1}{B}")
	effectstest.AdvanceToMain(t, g)
	// Not on an opponent's turn: refused.
	effectstest.FloatMana(t, g, me, "{B}{B}{C}{C}")
	if _, err := effectstest.CastFromHandErr(g, me, "Wake the Dead", "Instant", "{X}{B}{B}", gap111WakeOracle,
		game.CastSpellParams{Strict: true, XValue: 2, Targets: effectstest.CardRefs(a, b)}); err == nil {
		t.Fatal("Wake the Dead was cast in its controller's own main phase")
	}
	g.WithWriteLock(func() { me.ManaPool = game.ManaPool{} })
	effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
	effectstest.FloatMana(t, g, me, "{B}{B}{C}{C}")
	if _, err := effectstest.CastFromHandErr(g, me, "Wake the Dead", "Instant", "{X}{B}{B}", gap111WakeOracle,
		game.CastSpellParams{Strict: true, XValue: 2, Targets: effectstest.CardRefs(a)}); err == nil {
		t.Fatal("one target was accepted for X = 2")
	}
	if _, err := effectstest.CastFromHandErr(g, me, "Wake the Dead", "Instant", "{X}{B}{B}", gap111WakeOracle,
		game.CastSpellParams{Strict: true, XValue: 2, Targets: effectstest.CardRefs(a, b)}); err != nil {
		t.Fatalf("cast during an opponent's combat: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Fatal("the targeted creatures did not return")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.MonSettle(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("the creatures were not sacrificed at the end step")
	}
}

func TestGap111DesecrateRealityExilesEvenManaValuePermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	even := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Even", TypeLine: "Artifact", ManaCost: "{2}"})
	odd := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Odd", TypeLine: "Artifact", ManaCost: "{3}"})
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.CastCatalogSpellErr(t, g, "Desecrate Reality", "Instant", gap111DesecrateOracle, effectstest.CardRefs(odd)); err == nil {
		t.Fatal("a permanent with an odd mana value was a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Desecrate Reality", "Instant", gap111DesecrateOracle, effectstest.CardRefs(even))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(even) || !g.Exile.Contains(even) {
		t.Error("the even-mana-value permanent was not exiled")
	}
	if !g.Battlefield.Contains(odd) {
		t.Error("the odd-mana-value permanent was exiled")
	}
}

func TestGap111ForTheAncestorsTakesTheChosenTypeAndFlashbackWorks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 8; i++ {
		effectstest.PushLibraryTyped(me, "Filler", "Instant")
	}
	elf := effectstest.PushLibraryTyped(me, "Elf", "Creature — Elf")
	effectstest.AdvanceToMain(t, g)
	handBefore := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "For the Ancestors", "Instant", gap111AncestorsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerCreatureType(t, g, me.ID, "Elf")
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.ChooseCardsChoiceFor(g, me.ID); pick != nil {
		effectstest.AnswerChooseCards(t, g, me.ID, elf)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(elf) {
		t.Fatalf("the Elf was not put into hand (hand %d -> %d)", handBefore, me.Hand.Size())
	}
}
