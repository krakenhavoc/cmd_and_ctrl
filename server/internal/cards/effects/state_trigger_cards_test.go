package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// state_trigger_cards_test.go — the CR 603.8 state-trigger cards of ADR
// 0107 delivery PR 1 (#1858). Each test makes the printed state arise
// the way play would, passes priority so the trigger goes on the stack
// and resolves, and checks the result.

const (
	stBarbarianOutcast  = "49d44e89-b94e-43f7-a996-f50651df9264"
	stCovetousDragon    = "bf7710cb-70cb-42fb-b840-cc4f385daa7a"
	stEmperorCrocodile  = "26eb7ee3-c0b6-4be0-bc80-92952f55a6f9"
	stSynodCenturion    = "fe6337dc-df02-4baa-9fe0-b14e8d34186a"
	stTetheredGriffin   = "909e1bff-237a-4259-9c0f-419185681782"
	stEndangeredArmodon = "68879437-bbe2-4e99-b17b-ddec30bfd3d0"
	stSkeletonShip      = "8c85887d-5935-46ec-a075-dc9f5133bb96"
	stTaskMageAssembly  = "dfbbfbe1-d6c1-4b5b-8d09-949d70b9ff42"
	stDeadlyDesigns     = "293c6f8f-9e6b-4267-a15d-1d5ce95fb5b4"
	stDarksteelReactor  = "bc483bab-14fb-498d-9310-9c070766c7ae"
)

// stCard is a template for a test permanent.
func stCard(name, oracle, typeLine string, power, toughness int) game.Card {
	return game.Card{Name: name, OracleID: oracle, TypeLine: typeLine, Power: power, Toughness: toughness}
}

// stStateItems counts `source`'s triggered items, queued and on the
// stack, whose label starts with `prefix`.
func stStateItems(g *game.Game, source uuid.UUID, prefix string) int {
	n := 0
	for _, it := range g.PendingTriggers {
		if it.SourceCardID == source && strings.HasPrefix(it.Label, prefix) {
			n++
		}
	}
	for _, it := range g.StackMeta {
		if it.SourceCardID == source && it.Kind == game.StackItemTriggered && strings.HasPrefix(it.Label, prefix) {
			n++
		}
	}
	return n
}

// stDestroy destroys a permanent the way an effect would.
func stDestroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
}

// The "when you control no <permanents>" family: with one of the named
// permanents there is no trigger; once the last one goes, the trigger
// goes on the stack and the creature is sacrificed.
func TestStateTriggerControlNoCardsSacrifice(t *testing.T) {
	cases := []struct {
		name, oracle, typeLine string
		needs                  game.Card
	}{
		{"Barbarian Outcast", stBarbarianOutcast, "Creature — Human Barbarian Beast",
			stCard("Swamp", "", "Basic Land — Swamp", 0, 0)},
		{"Covetous Dragon", stCovetousDragon, "Creature — Dragon",
			stCard("Ornithopter", "", "Artifact Creature — Thopter", 0, 2)},
		{"Tethered Griffin", stTetheredGriffin, "Creature — Griffin",
			stCard("Test Aura", "", "Enchantment", 0, 0)},
		{"Skeleton Ship", stSkeletonShip, "Legendary Creature — Skeleton",
			stCard("Island", "", "Basic Land — Island", 0, 0)},
		{"Emperor Crocodile", stEmperorCrocodile, "Creature — Crocodile",
			stCard("Bear", "", "Creature — Bear", 2, 2)},
		{"Synod Centurion", stSynodCenturion, "Artifact Creature — Construct",
			stCard("Mox", "", "Artifact", 0, 0)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			other := apaPush(g, me.ID, me.ID, tc.needs)
			card := apaPush(g, me.ID, me.ID, stCard(tc.name, tc.oracle, tc.typeLine, 4, 4))
			passPriorityAroundTable(t, g)
			if !onBattlefield(g, card) || stStateItems(g, card, tc.name) != 0 {
				t.Fatalf("%s triggered while its controller had what it needs", tc.name)
			}
			stDestroy(t, g, other)
			if n := stStateItems(g, card, tc.name); n != 1 {
				t.Fatalf("%s: %d state-trigger items after the last one went, want 1", tc.name, n)
			}
			passPriorityAroundTable(t, g)
			if onBattlefield(g, card) {
				t.Fatalf("%s is still on the battlefield after its trigger resolved", tc.name)
			}
		})
	}
}

// Another player's Swamp does not keep Barbarian Outcast: "you control".
func TestBarbarianOutcastCountsOnlyYourSwamps(t *testing.T) {
	g := newCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	apaPush(g, bob.ID, bob.ID, stCard("Swamp", "", "Basic Land — Swamp", 0, 0))
	outcast := apaPush(g, me.ID, me.ID, stCard("Barbarian Outcast", stBarbarianOutcast, "Creature — Human Barbarian Beast", 2, 2))
	passPriorityAroundTable(t, g)
	if onBattlefield(g, outcast) {
		t.Fatal("Barbarian Outcast survived with only an opponent's Swamp")
	}
}

// Emperor Crocodile and its last companion dying in one wipe is one
// event: the Crocodile never controls no other creatures while it is on
// the battlefield, so nothing triggers.
func TestEmperorCrocodileDiesWithTheOthersWithoutTriggering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	apaPush(g, me.ID, me.ID, stCard("Bear", "", "Creature — Bear", 2, 2))
	croc := apaPush(g, me.ID, me.ID, stCard("Emperor Crocodile", stEmperorCrocodile, "Creature — Crocodile", 5, 5))
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me.ID})
		if err := (DestroyAllMatching{Match: Creature()}).Apply(ctx); err != nil {
			t.Fatal(err)
		}
	})
	if n := stStateItems(g, croc, "Emperor Crocodile"); n != 0 {
		t.Fatalf("the Crocodile triggered during a wipe that took it too (%d items)", n)
	}
}

// Endangered Armodon is sacrificed when its controller controls a
// creature with toughness 2 or less.
func TestEndangeredArmodonSacrificedForASmallCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	armodon := apaPush(g, me.ID, me.ID, stCard("Endangered Armodon", stEndangeredArmodon, "Creature — Elephant", 4, 5))
	apaPush(g, me.ID, me.ID, stCard("Wall", "", "Creature — Wall", 0, 3))
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, armodon) {
		t.Fatal("a toughness-3 creature cost the Armodon")
	}
	apaPush(g, me.ID, me.ID, stCard("Bear", "", "Creature — Bear", 2, 2))
	passPriorityAroundTable(t, g)
	if onBattlefield(g, armodon) {
		t.Fatal("Endangered Armodon survived a 2/2 under its controller's control")
	}
}

// Skeleton Ship's tap ability puts a -1/-1 counter on its target.
func TestSkeletonShipTapsForAMinusCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	apaPush(g, me.ID, me.ID, stCard("Island", "", "Basic Land — Island", 0, 0))
	ship := apaPush(g, me.ID, me.ID, stCard("Skeleton Ship", stSkeletonShip, "Legendary Creature — Skeleton", 0, 3))
	bear := apaPush(g, bob.ID, bob.ID, stCard("Bear", "", "Creature — Bear", 2, 2))
	apaActivate(t, g, me, ship, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if c := apaLive(g, bear); c.Counters[game.CounterMinusOne] != 1 {
		t.Fatalf("the bear has %d -1/-1 counters, want 1", c.Counters[game.CounterMinusOne])
	}
}

// Task Mage Assembly: any player may ping a creature at sorcery speed,
// and the Assembly is sacrificed when there are no creatures at all.
func TestTaskMageAssemblyPingsAndLeavesWithTheLastCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	bear := apaPush(g, bob.ID, bob.ID, stCard("Bear", "", "Creature — Bear", 1, 1))
	tma := apaPush(g, me.ID, me.ID, stCard("Task Mage Assembly", stTaskMageAssembly, "Enchantment", 0, 0))
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, tma) {
		t.Fatal("Task Mage Assembly left while a creature was on the battlefield")
	}
	rows := game.ActivatedAbilitiesForCard(game.Card{OracleID: stTaskMageAssembly})
	if len(rows) != 1 || !rows[0].AnyPlayer || !rows[0].SorcerySpeed {
		t.Fatalf("Task Mage Assembly's row = %+v, want one any-player sorcery-speed row", rows)
	}
	stDestroy(t, g, bear)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, tma) {
		t.Fatal("Task Mage Assembly stayed with no creatures on the battlefield")
	}
}

// Deadly Designs: the fifth plot counter, from any player, triggers it
// once; its controller targets up to two creatures, and the enchantment
// is sacrificed and they are destroyed.
func TestDeadlyDesignsFifthCounterDestroysTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	dd := apaPush(g, me.ID, me.ID, stCard("Deadly Designs", stDeadlyDesigns, "Enchantment", 0, 0))
	b1 := apaPush(g, bob.ID, bob.ID, stCard("Bear", "", "Creature — Bear", 2, 2))
	b2 := apaPush(g, bob.ID, bob.ID, stCard("Bear", "", "Creature — Bear", 2, 2))
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(dd, "plot", 4); err != nil {
			t.Fatal(err)
		}
	})
	if n := stStateItems(g, dd, "Deadly Designs"); n != 0 {
		t.Fatalf("four plot counters triggered it (%d items)", n)
	}
	// Bob pays for the fifth counter: the ability is any-player.
	apaMana(bob, "C", "C")
	if err := g.ActivateCatalogAbility(bob.ID, dd, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("Bob activates Deadly Designs: %v", err)
	}
	passPriorityAroundTable(t, g)
	choice := latestPickTarget(g, me.ID)
	if choice == nil {
		t.Fatal("no target prompt for the enchantment's controller")
	}
	if err := g.ResolvePickTargets(choice.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: b1}, {Kind: game.TargetCard, ID: b2},
	}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, dd) || onBattlefield(g, b1) || onBattlefield(g, b2) {
		t.Fatalf("after resolving: designs %v, bears %v %v; want all three gone",
			onBattlefield(g, dd), onBattlefield(g, b1), onBattlefield(g, b2))
	}
}

// Darksteel Reactor wins the game for its controller at twenty charge
// counters.
func TestDarksteelReactorWinsAtTwenty(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	reactor := apaPush(g, me.ID, me.ID, stCard("Darksteel Reactor", stDarksteelReactor, "Artifact", 0, 0))
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(reactor, game.CounterCharge, 19); err != nil {
			t.Fatal(err)
		}
	})
	if n := stStateItems(g, reactor, "Darksteel Reactor — you win"); n != 0 {
		t.Fatalf("nineteen counters triggered the win (%d items)", n)
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(reactor, game.CounterCharge, 1); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if g.State == game.StateActive || g.Outcome == nil || g.Outcome.Winner != me.ID {
		t.Fatalf("state %v outcome %+v, want the game over and won by the Reactor's controller", g.State, g.Outcome)
	}
}

// The serpents (ADR 0107 §1 with §2): each carries the attack
// restriction and is sacrificed when its controller's last land of the
// type goes.
func TestStateTriggerSerpentsNeedTheirLand(t *testing.T) {
	cases := []struct {
		name, oracle, typeLine string
		land                   game.Card
	}{
		{"Sea Serpent", "c16495fc-784d-4bac-9a68-ed437008df73", "Creature — Serpent", stCard("Island", "", "Basic Land — Island", 0, 0)},
		{"Bog Serpent", "9a9877b5-9f75-4c83-b11b-f006aecb075b", "Creature — Serpent", stCard("Swamp", "", "Basic Land — Swamp", 0, 0)},
		{"Dandân", "88929373-b2c8-4a81-a809-fed87fd5b0d7", "Creature — Fish", stCard("Island", "", "Basic Land — Island", 0, 0)},
		{"Gorilla Pack", "f2c8814b-581b-483b-a7ae-d3d7b962aec1", "Creature — Ape", stCard("Forest", "", "Basic Land — Forest", 0, 0)},
		{"Ronom Serpent", "ff35e480-8ea2-47bb-bb2a-24cefe9c2139", "Snow Creature — Serpent", stCard("Snow-Covered Island", "", "Basic Snow Land — Island", 0, 0)},
		{"Slipstream Serpent", "aa1152cb-255f-43fa-81f5-430304ce4d98", "Creature — Serpent", stCard("Island", "", "Basic Land — Island", 0, 0)},
		{"Pirate Ship", "c6b3f924-806d-47d3-b044-72b48470196c", "Creature — Human Pirate", stCard("Island", "", "Basic Land — Island", 0, 0)},
		{"Vodalian Knights", "f6daa28f-e5ce-440c-8dc2-b36f59ae0d4f", "Creature — Merfolk Knight", stCard("Island", "", "Basic Land — Island", 0, 0)},
		{"Manta Ray", "d5129531-e4b6-454e-9c67-dae925c8f2ee", "Creature — Fish", stCard("Island", "", "Basic Land — Island", 0, 0)},
		{"Merchant Ship", "69556f6c-c05b-4902-bac7-012f0ed81b75", "Creature — Human", stCard("Island", "", "Basic Land — Island", 0, 0)},
		{"Marjhan", "3fdee2ab-7ec6-4fc6-ad99-f04571f94583", "Creature — Serpent", stCard("Island", "", "Basic Land — Island", 0, 0)},
		{"Island Fish Jasconius", "bb217f12-532f-4833-a27a-99e290aa47d0", "Creature — Fish", stCard("Island", "", "Basic Land — Island", 0, 0)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			// A plain Island under Ronom Serpent is not a snow land.
			if tc.name == "Ronom Serpent" {
				apaPush(g, me.ID, me.ID, stCard("Island", "", "Basic Land — Island", 0, 0))
			}
			land := apaPush(g, me.ID, me.ID, tc.land)
			serpent := apaPush(g, me.ID, me.ID, stCard(tc.name, tc.oracle, tc.typeLine, 4, 4))
			passPriorityAroundTable(t, g)
			if !onBattlefield(g, serpent) {
				t.Fatalf("%s left while its controller had the land", tc.name)
			}
			if c := apaLive(g, serpent); len(c.Effective().AttackTargetRestrictions) == 0 {
				t.Errorf("%s has no attack-target restriction", tc.name)
			}
			stDestroy(t, g, land)
			passPriorityAroundTable(t, g)
			if onBattlefield(g, serpent) {
				t.Fatalf("%s survived the loss of its last land", tc.name)
			}
		})
	}
}

// Dark Depths: ten ice counters as it enters; the {3} row takes one off;
// with none left it is sacrificed for Marit Lage.
func TestDarkDepthsMakesMaritLage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	depths := playLandFromHand(t, g, "Dark Depths", "c9b82110-7dfd-4617-9399-9510be449043")
	passPriorityAroundTable(t, g)
	if c := apaLive(g, depths); c == nil || c.Counters["ice"] != 10 {
		t.Fatalf("Dark Depths entered with %v ice counters, want 10", c)
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(depths, "ice", -9); err != nil {
			t.Fatal(err)
		}
	})
	apaMana(me, "C", "C", "C")
	apaActivate(t, g, me, depths, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if onBattlefield(g, depths) {
		t.Fatal("Dark Depths with no ice counters is still on the battlefield")
	}
	if n := onBattlefieldNamed(g, "Marit Lage"); n != 1 {
		t.Fatalf("%d Marit Lage tokens, want 1", n)
	}
}

// Endrek Sahr: a creature spell of mana value 3 makes three Thrulls;
// the seventh Thrull triggers the sacrifice.
func TestEndrekSahrBreedsThrullsUntilSeven(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	endrek := apaPush(g, me.ID, me.ID, stCard("Endrek Sahr, Master Breeder", "47a0079f-3544-45bc-a32a-bd93844c8c43", "Legendary Creature — Human Wizard", 2, 2))
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{G}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	apaMana(me, "G", "C", "C")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := onBattlefieldNamed(g, "Thrull"); n != 3 {
		t.Fatalf("%d Thrulls after a mana value 3 creature spell, want 3", n)
	}
	if !onBattlefield(g, endrek) {
		t.Fatal("Endrek Sahr left with three Thrulls")
	}
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, TokenCard("1/1 black Thrull"), 4); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if onBattlefield(g, endrek) {
		t.Fatal("Endrek Sahr survived seven Thrulls")
	}
}

// Last Laugh: a permanent going to a graveyard from the battlefield deals
// 1 to each creature and each player; with no creatures left it goes.
func TestLastLaughPingsAndLeavesWithTheLastCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	bear := apaPush(g, bob.ID, bob.ID, stCard("Bear", "", "Creature — Bear", 2, 2))
	wall := apaPush(g, bob.ID, bob.ID, stCard("Wall", "", "Creature — Wall", 0, 1))
	laugh := apaPush(g, me.ID, me.ID, stCard("Last Laugh", "5facb256-b993-43f8-b971-b3a59a7434bf", "Enchantment", 0, 0))
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, laugh) {
		t.Fatal("Last Laugh left with creatures on the battlefield")
	}
	life := me.Life
	stDestroy(t, g, bear)
	passPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Fatalf("life %d, want %d after one death", me.Life, life-1)
	}
	if onBattlefield(g, wall) {
		t.Fatal("the 0/1 Wall survived Last Laugh's 1 damage")
	}
	// The Wall's death triggers the damage again, and the empty board
	// triggers the sacrifice: two triggers of one controller to order.
	for i := 0; i < 4 && onBattlefield(g, laugh); i++ {
		answerAnyTriggerOrderPrompt(t, g, me.ID)
		passPriorityAroundTable(t, g)
	}
	if onBattlefield(g, laugh) {
		t.Fatal("Last Laugh stayed with no creatures on the battlefield")
	}
}

// Lurebound Scarecrow: the colour is chosen as it enters, and it is
// sacrificed when its controller has nothing of that colour.
func TestLureboundScarecrowWatchesItsColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	red := game.Card{Name: "Goblin", TypeLine: "Creature — Goblin", Power: 1, Toughness: 1, Colors: []string{"R"}}
	goblin := apaPush(g, me.ID, me.ID, red)
	scarecrow := stCard("Lurebound Scarecrow", "8353f834-678a-4fa1-857a-8bfdfb8d6378", "Artifact Creature — Scarecrow", 4, 4)
	scarecrow.ChosenColor = "R" // as if it had chosen red as it entered
	sc := apaPush(g, me.ID, me.ID, scarecrow)
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, sc) {
		t.Fatal("the Scarecrow left while its controller had a red permanent")
	}
	stDestroy(t, g, goblin)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, sc) {
		t.Fatal("the Scarecrow survived the loss of its controller's last red permanent")
	}
}

// Mazemind Tome: the fourth page counter is part of the cost, so it
// triggers the exile at once; the exile gains 4 life.
func TestMazemindTomeExilesOnTheFourthPage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tome := apaPush(g, me.ID, me.ID, stCard("Mazemind Tome", "800fb917-8898-4eba-9f94-aec6e9d75236", "Artifact — Book", 0, 0))
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(tome, "page", 3); err != nil {
			t.Fatal(err)
		}
	})
	life := me.Life
	apaMana(me, "C", "C")
	apaActivate(t, g, me, tome, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if onBattlefield(g, tome) {
		t.Fatal("Mazemind Tome with four page counters is still on the battlefield")
	}
	if me.Life != life+4 {
		t.Fatalf("life %d, want %d", me.Life, life+4)
	}
}

// Plague Boiler: the third plague counter sacrifices it and destroys
// every nonland permanent; lands stay.
func TestPlagueBoilerWipesOnTheThirdCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	boiler := apaPush(g, me.ID, me.ID, stCard("Plague Boiler", "fef502af-6e79-4c55-a86a-b45adb3fc64a", "Artifact", 0, 0))
	bear := apaPush(g, bob.ID, bob.ID, stCard("Bear", "", "Creature — Bear", 2, 2))
	land := apaPush(g, bob.ID, bob.ID, stCard("Forest", "", "Basic Land — Forest", 0, 0))
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(boiler, "plague", 3); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if onBattlefield(g, boiler) || onBattlefield(g, bear) || !onBattlefield(g, land) {
		t.Fatalf("boiler %v bear %v land %v; want the boiler and the bear gone and the land kept",
			onBattlefield(g, boiler), onBattlefield(g, bear), onBattlefield(g, land))
	}
}

// Afiya Grove is sacrificed once its last +1/+1 counter is gone.
func TestAfiyaGroveSacrificedWhenEmpty(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	apaPush(g, me.ID, me.ID, stCard("Bear", "", "Creature — Bear", 2, 2))
	tmpl := stCard("Afiya Grove", "056a98aa-f3f7-4ac9-9755-3e8a60a59abb", "Enchantment", 0, 0)
	tmpl.Counters = map[string]int{game.CounterPlusOne: 1}
	grove := apaPush(g, me.ID, me.ID, tmpl)
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, grove) {
		t.Fatal("a Grove with a counter left the battlefield")
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(grove, game.CounterPlusOne, -1); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if onBattlefield(g, grove) {
		t.Fatal("a Grove with no +1/+1 counters is still on the battlefield")
	}
}
