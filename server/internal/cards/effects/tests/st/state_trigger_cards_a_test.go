package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// state_trigger_cards_a_test.go — more CR 603.8 state-trigger cards of
// ADR 0107 delivery PR 1 (#1858): the least-power upkeeps, the hidden
// enchantments that become creatures, Transcendence and Mana Vortex.

const (
	stDropOfHoney     = "383e9005-5869-4d1d-917d-30e5f214fbd9"
	stPorphyryNodes   = "7cfeb732-8a35-4a1f-a417-11db4c1499fd"
	stManaVortex      = "98b88734-09e2-4209-a787-9b5b345391f9"
	stLurkingJackals  = "e9e5e17a-95fc-41b7-af4f-aec4802960a0"
	stOpalAvenger     = "13a85c9f-f653-4d90-bea6-94022ee82527"
	stVeiledCrocodile = "bdc3ff35-f8d8-477d-b042-94a9f8ed7243"
	stTranscendence   = "f9279f8f-3d19-4f5b-b194-cb8c65313b7e"
)

// Drop of Honey destroys the one creature with the least power at its
// controller's upkeep, asks its controller to break a tie, and is
// sacrificed once no creature is left.
func TestDropOfHoneyDestroysTheLeastPowerAndLeavesWithTheLastCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	// The creatures first: an enchantment that enters onto an empty
	// board is sacrificed at once.
	small := effectstest.ApaPush(g, bob.ID, bob.ID, effectstest.StCard("Elf", "", "Creature — Elf", 1, 1))
	big := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Bear", "", "Creature — Bear", 2, 2))
	drop := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Drop of Honey", stDropOfHoney, "Enchantment", 0, 0))
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, small) || !effects.OnBattlefield(g, big) {
		t.Fatalf("after the upkeep: 1-power %v, 2-power %v; want only the 1-power creature gone",
			effects.OnBattlefield(g, small), effects.OnBattlefield(g, big))
	}
	if !effects.OnBattlefield(g, drop) {
		t.Fatal("Drop of Honey left while a creature remained")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, big) || effects.OnBattlefield(g, drop) {
		t.Fatalf("after the second upkeep: bear %v, Drop of Honey %v; want both gone",
			effects.OnBattlefield(g, big), effects.OnBattlefield(g, drop))
	}
}

// A tie for least power is broken by the controller, and only the
// tied creatures are offered.
func TestPorphyryNodesControllerBreaksATie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	a := effectstest.ApaPush(g, bob.ID, bob.ID, effectstest.StCard("Elf", "", "Creature — Elf", 1, 1))
	b := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Squire", "", "Creature — Squire", 1, 2))
	big := effectstest.ApaPush(g, bob.ID, bob.ID, effectstest.StCard("Bear", "", "Creature — Bear", 2, 2))
	effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Porphyry Nodes", stPorphyryNodes, "Enchantment", 0, 0))
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no tie-break prompt for the controller")
	}
	if len(pick.ChooseCards) != 2 || !effectstest.HasID(pick.ChooseCards, a) || !effectstest.HasID(pick.ChooseCards, b) {
		t.Fatalf("offered %v, want exactly the two 1-power creatures", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, a) || !effects.OnBattlefield(g, b) || !effects.OnBattlefield(g, big) {
		t.Fatalf("after the pick: chosen %v, other tied %v, bear %v; want only the chosen one gone",
			effects.OnBattlefield(g, a), effects.OnBattlefield(g, b), effects.OnBattlefield(g, big))
	}
}

// stIsCreatureNotEnchantment reports whether the permanent is now a
// creature with the given P/T and no longer an enchantment.
func stIsCreatureNotEnchantment(t *testing.T, g *game.Game, id uuid.UUID, power, toughness int) {
	t.Helper()
	c := effectstest.ApaLive(g, id)
	if c == nil {
		t.Fatal("the permanent is gone")
	}
	if !c.IsCreature() || c.HasCardType("enchantment") || c.CurrentPower() != power || c.CurrentToughness() != toughness {
		t.Fatalf("creature %v enchantment %v %d/%d; want a %d/%d creature that is not an enchantment",
			c.IsCreature(), c.HasCardType("enchantment"), c.CurrentPower(), c.CurrentToughness(), power, toughness)
	}
}

// Opal Avenger becomes a 3/5 Soldier when its controller is at 10 life
// or less, once.
func TestOpalAvengerBecomesASoldierAtTenLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opal := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Opal Avenger", stOpalAvenger, "Enchantment", 0, 0))
	effectstest.StSetLife(t, g, me, 11)
	if n := effectstest.StStateItems(g, opal, "Opal Avenger"); n != 0 {
		t.Fatalf("11 life triggered it (%d items)", n)
	}
	effectstest.StSetLife(t, g, me, 10)
	effectstest.PassPriorityAroundTable(t, g)
	stIsCreatureNotEnchantment(t, g, opal, 3, 5)
	if c := effectstest.ApaLive(g, opal); !c.HasSubtype("Soldier") {
		t.Errorf("subtypes %v, want Soldier", c.Effective().Subtypes)
	}
	effectstest.StSetLife(t, g, me, 5)
	if n := effectstest.StStateItems(g, opal, "Opal Avenger"); n != 0 {
		t.Fatalf("the creature, no longer an enchantment, triggered again (%d items)", n)
	}
}

// Lurking Jackals watches its controller's opponents, not its controller.
func TestLurkingJackalsBecomesAJackalWhenAnOpponentIsLow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	jackals := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Lurking Jackals", stLurkingJackals, "Enchantment", 0, 0))
	effectstest.StSetLife(t, g, me, 3)
	if n := effectstest.StStateItems(g, jackals, "Lurking Jackals"); n != 0 {
		t.Fatalf("its controller's own low life triggered it (%d items)", n)
	}
	effectstest.StSetLife(t, g, bob, 10)
	effectstest.PassPriorityAroundTable(t, g)
	stIsCreatureNotEnchantment(t, g, jackals, 3, 2)
}

// Veiled Crocodile triggers on any player's empty hand.
func TestVeiledCrocodileBecomesACrocodileOnAnEmptyHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	croc := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Veiled Crocodile", stVeiledCrocodile, "Enchantment", 0, 0))
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.StStateItems(g, croc, "Veiled Crocodile"); n != 0 || !effectstest.ApaLive(g, croc).HasCardType("enchantment") {
		t.Fatalf("it triggered with every hand full (%d items)", n)
	}
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(bob.ID, bob.Hand.Size()); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	stIsCreatureNotEnchantment(t, g, croc, 4, 4)
}

// Transcendence: 0 or less life does not end the game, every life lost
// comes back twice, and 20 or more life does.
func TestTranscendenceLifeRules(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Life = 5
	tr := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Transcendence", stTranscendence, "Enchantment", 0, 0))
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(uuid.Nil, me.ID, 10); err != nil {
			t.Fatal(err)
		}
	})
	if me.Eliminated {
		t.Fatal("its controller lost at -5 life")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Eliminated || me.Life != 15 {
		t.Fatalf("life %d eliminated %v; want 15 (-5 plus 2 for each of the 10 lost) and still in the game", me.Life, me.Eliminated)
	}
	if n := effectstest.StStateItems(g, tr, "Transcendence — you lose"); n != 0 {
		t.Fatalf("15 life triggered the loss (%d items)", n)
	}
	effectstest.StSetLife(t, g, me, 20)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Eliminated {
		t.Fatalf("its controller is still in the game at %d life", me.Life)
	}
}

// Mana Vortex: with no land to sacrifice, the cast is countered.
func TestManaVortexWithoutALandIsCountered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vortex := effectstest.CastCatalogSpell(t, g, "Mana Vortex", "Enchantment", stManaVortex, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, vortex) {
		t.Fatal("Mana Vortex resolved though its caster had no land to sacrifice")
	}
	if !me.Graveyard.Contains(vortex) {
		t.Error("the countered Mana Vortex is not in its owner's graveyard")
	}
}

// Mana Vortex: the caster sacrifices a land to keep it; each upkeep its
// player sacrifices a land; it leaves when no land is left anywhere.
func TestManaVortexSacrificesLandsAndLeavesWithTheLast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	myLand := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Island", "", "Basic Land — Island", 0, 0))
	bobLand := effectstest.ApaPush(g, bob.ID, bob.ID, effectstest.StCard("Forest", "", "Basic Land — Forest", 0, 0))
	vortex := effectstest.CastCatalogSpell(t, g, "Mana Vortex", "Enchantment", stManaVortex, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0) // sacrifice a land
	effectstest.AnswerChooseCards(t, g, me.ID, myLand)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, myLand) || !effects.OnBattlefield(g, vortex) {
		t.Fatalf("after paying: land %v, vortex %v; want the land gone and the Vortex in play",
			effects.OnBattlefield(g, myLand), effects.OnBattlefield(g, vortex))
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, bob.ID, bobLand)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, bobLand) {
		t.Fatal("Bob kept his land through his upkeep")
	}
	if effects.OnBattlefield(g, vortex) {
		t.Fatal("Mana Vortex stayed with no lands on the battlefield")
	}
}
