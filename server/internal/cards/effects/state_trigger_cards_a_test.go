package effects

import (
	"testing"

	"github.com/google/uuid"

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
	g := newCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	// The creatures first: an enchantment that enters onto an empty
	// board is sacrificed at once.
	small := apaPush(g, bob.ID, bob.ID, stCard("Elf", "", "Creature — Elf", 1, 1))
	big := apaPush(g, me.ID, me.ID, stCard("Bear", "", "Creature — Bear", 2, 2))
	drop := apaPush(g, me.ID, me.ID, stCard("Drop of Honey", stDropOfHoney, "Enchantment", 0, 0))
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, small) || !onBattlefield(g, big) {
		t.Fatalf("after the upkeep: 1-power %v, 2-power %v; want only the 1-power creature gone",
			onBattlefield(g, small), onBattlefield(g, big))
	}
	if !onBattlefield(g, drop) {
		t.Fatal("Drop of Honey left while a creature remained")
	}
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, big) || onBattlefield(g, drop) {
		t.Fatalf("after the second upkeep: bear %v, Drop of Honey %v; want both gone",
			onBattlefield(g, big), onBattlefield(g, drop))
	}
}

// A tie for least power is broken by the controller, and only the
// tied creatures are offered.
func TestPorphyryNodesControllerBreaksATie(t *testing.T) {
	g := newCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	a := apaPush(g, bob.ID, bob.ID, stCard("Elf", "", "Creature — Elf", 1, 1))
	b := apaPush(g, me.ID, me.ID, stCard("Squire", "", "Creature — Squire", 1, 2))
	big := apaPush(g, bob.ID, bob.ID, stCard("Bear", "", "Creature — Bear", 2, 2))
	apaPush(g, me.ID, me.ID, stCard("Porphyry Nodes", stPorphyryNodes, "Enchantment", 0, 0))
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no tie-break prompt for the controller")
	}
	if len(pick.ChooseCards) != 2 || !hasID(pick.ChooseCards, a) || !hasID(pick.ChooseCards, b) {
		t.Fatalf("offered %v, want exactly the two 1-power creatures", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, a)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, a) || !onBattlefield(g, b) || !onBattlefield(g, big) {
		t.Fatalf("after the pick: chosen %v, other tied %v, bear %v; want only the chosen one gone",
			onBattlefield(g, a), onBattlefield(g, b), onBattlefield(g, big))
	}
}

// stIsCreatureNotEnchantment reports whether the permanent is now a
// creature with the given P/T and no longer an enchantment.
func stIsCreatureNotEnchantment(t *testing.T, g *game.Game, id uuid.UUID, power, toughness int) {
	t.Helper()
	c := apaLive(g, id)
	if c == nil {
		t.Fatal("the permanent is gone")
	}
	if !c.IsCreature() || c.HasCardType("enchantment") || c.CurrentPower() != power || c.CurrentToughness() != toughness {
		t.Fatalf("creature %v enchantment %v %d/%d; want a %d/%d creature that is not an enchantment",
			c.IsCreature(), c.HasCardType("enchantment"), c.CurrentPower(), c.CurrentToughness(), power, toughness)
	}
}

func stSetLife(t *testing.T, g *game.Game, p *game.Player, life int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, life-p.Life); err != nil {
			t.Fatal(err)
		}
	})
}

// Opal Avenger becomes a 3/5 Soldier when its controller is at 10 life
// or less, once.
func TestOpalAvengerBecomesASoldierAtTenLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opal := apaPush(g, me.ID, me.ID, stCard("Opal Avenger", stOpalAvenger, "Enchantment", 0, 0))
	stSetLife(t, g, me, 11)
	if n := stStateItems(g, opal, "Opal Avenger"); n != 0 {
		t.Fatalf("11 life triggered it (%d items)", n)
	}
	stSetLife(t, g, me, 10)
	passPriorityAroundTable(t, g)
	stIsCreatureNotEnchantment(t, g, opal, 3, 5)
	if c := apaLive(g, opal); !c.HasSubtype("Soldier") {
		t.Errorf("subtypes %v, want Soldier", c.Effective().Subtypes)
	}
	stSetLife(t, g, me, 5)
	if n := stStateItems(g, opal, "Opal Avenger"); n != 0 {
		t.Fatalf("the creature, no longer an enchantment, triggered again (%d items)", n)
	}
}

// Lurking Jackals watches its controller's opponents, not its controller.
func TestLurkingJackalsBecomesAJackalWhenAnOpponentIsLow(t *testing.T) {
	g := newCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	jackals := apaPush(g, me.ID, me.ID, stCard("Lurking Jackals", stLurkingJackals, "Enchantment", 0, 0))
	stSetLife(t, g, me, 3)
	if n := stStateItems(g, jackals, "Lurking Jackals"); n != 0 {
		t.Fatalf("its controller's own low life triggered it (%d items)", n)
	}
	stSetLife(t, g, bob, 10)
	passPriorityAroundTable(t, g)
	stIsCreatureNotEnchantment(t, g, jackals, 3, 2)
}

// Veiled Crocodile triggers on any player's empty hand.
func TestVeiledCrocodileBecomesACrocodileOnAnEmptyHand(t *testing.T) {
	g := newCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	croc := apaPush(g, me.ID, me.ID, stCard("Veiled Crocodile", stVeiledCrocodile, "Enchantment", 0, 0))
	passPriorityAroundTable(t, g)
	if n := stStateItems(g, croc, "Veiled Crocodile"); n != 0 || !apaLive(g, croc).HasCardType("enchantment") {
		t.Fatalf("it triggered with every hand full (%d items)", n)
	}
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(bob.ID, bob.Hand.Size()); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	stIsCreatureNotEnchantment(t, g, croc, 4, 4)
}

// Transcendence: 0 or less life does not end the game, every life lost
// comes back twice, and 20 or more life does.
func TestTranscendenceLifeRules(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Life = 5
	tr := apaPush(g, me.ID, me.ID, stCard("Transcendence", stTranscendence, "Enchantment", 0, 0))
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(uuid.Nil, me.ID, 10); err != nil {
			t.Fatal(err)
		}
	})
	if me.Eliminated {
		t.Fatal("its controller lost at -5 life")
	}
	passPriorityAroundTable(t, g)
	if me.Eliminated || me.Life != 15 {
		t.Fatalf("life %d eliminated %v; want 15 (-5 plus 2 for each of the 10 lost) and still in the game", me.Life, me.Eliminated)
	}
	if n := stStateItems(g, tr, "Transcendence — you lose"); n != 0 {
		t.Fatalf("15 life triggered the loss (%d items)", n)
	}
	stSetLife(t, g, me, 20)
	passPriorityAroundTable(t, g)
	if !me.Eliminated {
		t.Fatalf("its controller is still in the game at %d life", me.Life)
	}
}

// Mana Vortex: with no land to sacrifice, the cast is countered.
func TestManaVortexWithoutALandIsCountered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vortex := castCatalogSpell(t, g, "Mana Vortex", "Enchantment", stManaVortex, nil)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, vortex) {
		t.Fatal("Mana Vortex resolved though its caster had no land to sacrifice")
	}
	if !me.Graveyard.Contains(vortex) {
		t.Error("the countered Mana Vortex is not in its owner's graveyard")
	}
}

// Mana Vortex: the caster sacrifices a land to keep it; each upkeep its
// player sacrifices a land; it leaves when no land is left anywhere.
func TestManaVortexSacrificesLandsAndLeavesWithTheLast(t *testing.T) {
	g := newCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	myLand := apaPush(g, me.ID, me.ID, stCard("Island", "", "Basic Land — Island", 0, 0))
	bobLand := apaPush(g, bob.ID, bob.ID, stCard("Forest", "", "Basic Land — Forest", 0, 0))
	vortex := castCatalogSpell(t, g, "Mana Vortex", "Enchantment", stManaVortex, nil)
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 0) // sacrifice a land
	answerChooseCards(t, g, me.ID, myLand)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, myLand) || !onBattlefield(g, vortex) {
		t.Fatalf("after paying: land %v, vortex %v; want the land gone and the Vortex in play",
			onBattlefield(g, myLand), onBattlefield(g, vortex))
	}
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, bob.ID, bobLand)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, bobLand) {
		t.Fatal("Bob kept his land through his upkeep")
	}
	if onBattlefield(g, vortex) {
		t.Fatal("Mana Vortex stayed with no lands on the battlefield")
	}
}
