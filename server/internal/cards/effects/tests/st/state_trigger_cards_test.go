package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The "when you control no <permanents>" family: with one of the named
// permanents there is no trigger; once the last one goes, the trigger
// goes on the stack and the creature is sacrificed.
func TestStateTriggerControlNoCardsSacrifice(t *testing.T) {
	cases := []struct {
		name, oracle, typeLine string
		needs                  game.Card
	}{
		{"Barbarian Outcast", effectstest.StBarbarianOutcast, "Creature — Human Barbarian Beast",
			effectstest.StCard("Swamp", "", "Basic Land — Swamp", 0, 0)},
		{"Covetous Dragon", effectstest.StCovetousDragon, "Creature — Dragon",
			effectstest.StCard("Ornithopter", "", "Artifact Creature — Thopter", 0, 2)},
		{"Tethered Griffin", effectstest.StTetheredGriffin, "Creature — Griffin",
			effectstest.StCard("Test Aura", "", "Enchantment", 0, 0)},
		{"Skeleton Ship", effectstest.StSkeletonShip, "Legendary Creature — Skeleton",
			effectstest.StCard("Island", "", "Basic Land — Island", 0, 0)},
		{"Emperor Crocodile", effectstest.StEmperorCrocodile, "Creature — Crocodile",
			effectstest.StCard("Bear", "", "Creature — Bear", 2, 2)},
		{"Synod Centurion", effectstest.StSynodCenturion, "Artifact Creature — Construct",
			effectstest.StCard("Mox", "", "Artifact", 0, 0)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			other := effectstest.ApaPush(g, me.ID, me.ID, tc.needs)
			card := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard(tc.name, tc.oracle, tc.typeLine, 4, 4))
			effectstest.PassPriorityAroundTable(t, g)
			if !effects.OnBattlefield(g, card) || effectstest.StStateItems(g, card, tc.name) != 0 {
				t.Fatalf("%s triggered while its controller had what it needs", tc.name)
			}
			effectstest.StDestroy(t, g, other)
			if n := effectstest.StStateItems(g, card, tc.name); n != 1 {
				t.Fatalf("%s: %d state-trigger items after the last one went, want 1", tc.name, n)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if effects.OnBattlefield(g, card) {
				t.Fatalf("%s is still on the battlefield after its trigger resolved", tc.name)
			}
		})
	}
}

// Another player's Swamp does not keep Barbarian Outcast: "you control".
func TestBarbarianOutcastCountsOnlyYourSwamps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	effectstest.ApaPush(g, bob.ID, bob.ID, effectstest.StCard("Swamp", "", "Basic Land — Swamp", 0, 0))
	outcast := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Barbarian Outcast", effectstest.StBarbarianOutcast, "Creature — Human Barbarian Beast", 2, 2))
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, outcast) {
		t.Fatal("Barbarian Outcast survived with only an opponent's Swamp")
	}
}

// Emperor Crocodile and its last companion dying in one wipe is one
// event: the Crocodile never controls no other creatures while it is on
// the battlefield, so nothing triggers.
func TestEmperorCrocodileDiesWithTheOthersWithoutTriggering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Bear", "", "Creature — Bear", 2, 2))
	croc := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Emperor Crocodile", effectstest.StEmperorCrocodile, "Creature — Crocodile", 5, 5))
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{Controller: me.ID})
		if err := (effects.DestroyAllMatching{Match: effects.Creature()}).Apply(ctx); err != nil {
			t.Fatal(err)
		}
	})
	if n := effectstest.StStateItems(g, croc, "Emperor Crocodile"); n != 0 {
		t.Fatalf("the Crocodile triggered during a wipe that took it too (%d items)", n)
	}
}

// Endangered Armodon is sacrificed when its controller controls a
// creature with toughness 2 or less.
func TestEndangeredArmodonSacrificedForASmallCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	armodon := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Endangered Armodon", effectstest.StEndangeredArmodon, "Creature — Elephant", 4, 5))
	effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Wall", "", "Creature — Wall", 0, 3))
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, armodon) {
		t.Fatal("a toughness-3 creature cost the Armodon")
	}
	effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Bear", "", "Creature — Bear", 2, 2))
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, armodon) {
		t.Fatal("Endangered Armodon survived a 2/2 under its controller's control")
	}
}

// Skeleton Ship's tap ability puts a -1/-1 counter on its target.
func TestSkeletonShipTapsForAMinusCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Island", "", "Basic Land — Island", 0, 0))
	ship := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Skeleton Ship", effectstest.StSkeletonShip, "Legendary Creature — Skeleton", 0, 3))
	bear := effectstest.ApaPush(g, bob.ID, bob.ID, effectstest.StCard("Bear", "", "Creature — Bear", 2, 2))
	effectstest.ApaActivate(t, g, me, ship, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if c := effectstest.ApaLive(g, bear); c.Counters[game.CounterMinusOne] != 1 {
		t.Fatalf("the bear has %d -1/-1 counters, want 1", c.Counters[game.CounterMinusOne])
	}
}

// Task Mage Assembly: any player may ping a creature at sorcery speed,
// and the Assembly is sacrificed when there are no creatures at all.
func TestTaskMageAssemblyPingsAndLeavesWithTheLastCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	bear := effectstest.ApaPush(g, bob.ID, bob.ID, effectstest.StCard("Bear", "", "Creature — Bear", 1, 1))
	tma := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Task Mage Assembly", effectstest.StTaskMageAssembly, "Enchantment", 0, 0))
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, tma) {
		t.Fatal("Task Mage Assembly left while a creature was on the battlefield")
	}
	rows := game.ActivatedAbilitiesForCard(game.Card{OracleID: effectstest.StTaskMageAssembly})
	if len(rows) != 1 || !rows[0].AnyPlayer || !rows[0].SorcerySpeed {
		t.Fatalf("Task Mage Assembly's row = %+v, want one any-player sorcery-speed row", rows)
	}
	effectstest.StDestroy(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, tma) {
		t.Fatal("Task Mage Assembly stayed with no creatures on the battlefield")
	}
}

// Deadly Designs: the fifth plot counter, from any player, triggers it
// once; its controller targets up to two creatures, and the enchantment
// is sacrificed and they are destroyed.
func TestDeadlyDesignsFifthCounterDestroysTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	dd := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Deadly Designs", effectstest.StDeadlyDesigns, "Enchantment", 0, 0))
	b1 := effectstest.ApaPush(g, bob.ID, bob.ID, effectstest.StCard("Bear", "", "Creature — Bear", 2, 2))
	b2 := effectstest.ApaPush(g, bob.ID, bob.ID, effectstest.StCard("Bear", "", "Creature — Bear", 2, 2))
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(dd, "plot", 4); err != nil {
			t.Fatal(err)
		}
	})
	if n := effectstest.StStateItems(g, dd, "Deadly Designs"); n != 0 {
		t.Fatalf("four plot counters triggered it (%d items)", n)
	}
	// Bob pays for the fifth counter: the ability is any-player.
	effectstest.ApaMana(bob, "C", "C")
	if err := g.ActivateCatalogAbility(bob.ID, dd, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("Bob activates Deadly Designs: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	choice := effectstest.LatestPickTarget(g, me.ID)
	if choice == nil {
		t.Fatal("no target prompt for the enchantment's controller")
	}
	if err := g.ResolvePickTargets(choice.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: b1}, {Kind: game.TargetCard, ID: b2},
	}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, dd) || effects.OnBattlefield(g, b1) || effects.OnBattlefield(g, b2) {
		t.Fatalf("after resolving: designs %v, bears %v %v; want all three gone",
			effects.OnBattlefield(g, dd), effects.OnBattlefield(g, b1), effects.OnBattlefield(g, b2))
	}
}

// Darksteel Reactor wins the game for its controller at twenty charge
// counters.
func TestDarksteelReactorWinsAtTwenty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	reactor := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Darksteel Reactor", effectstest.StDarksteelReactor, "Artifact", 0, 0))
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(reactor, game.CounterCharge, 19); err != nil {
			t.Fatal(err)
		}
	})
	if n := effectstest.StStateItems(g, reactor, "Darksteel Reactor — you win"); n != 0 {
		t.Fatalf("nineteen counters triggered the win (%d items)", n)
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(reactor, game.CounterCharge, 1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
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
		{"Sea Serpent", "c16495fc-784d-4bac-9a68-ed437008df73", "Creature — Serpent", effectstest.StCard("Island", "", "Basic Land — Island", 0, 0)},
		{"Bog Serpent", "9a9877b5-9f75-4c83-b11b-f006aecb075b", "Creature — Serpent", effectstest.StCard("Swamp", "", "Basic Land — Swamp", 0, 0)},
		{"Dandân", "88929373-b2c8-4a81-a809-fed87fd5b0d7", "Creature — Fish", effectstest.StCard("Island", "", "Basic Land — Island", 0, 0)},
		{"Gorilla Pack", "f2c8814b-581b-483b-a7ae-d3d7b962aec1", "Creature — Ape", effectstest.StCard("Forest", "", "Basic Land — Forest", 0, 0)},
		{"Ronom Serpent", "ff35e480-8ea2-47bb-bb2a-24cefe9c2139", "Snow Creature — Serpent", effectstest.StCard("Snow-Covered Island", "", "Basic Snow Land — Island", 0, 0)},
		{"Slipstream Serpent", "aa1152cb-255f-43fa-81f5-430304ce4d98", "Creature — Serpent", effectstest.StCard("Island", "", "Basic Land — Island", 0, 0)},
		{"Pirate Ship", "c6b3f924-806d-47d3-b044-72b48470196c", "Creature — Human Pirate", effectstest.StCard("Island", "", "Basic Land — Island", 0, 0)},
		{"Vodalian Knights", "f6daa28f-e5ce-440c-8dc2-b36f59ae0d4f", "Creature — Merfolk Knight", effectstest.StCard("Island", "", "Basic Land — Island", 0, 0)},
		{"Manta Ray", "d5129531-e4b6-454e-9c67-dae925c8f2ee", "Creature — Fish", effectstest.StCard("Island", "", "Basic Land — Island", 0, 0)},
		{"Merchant Ship", "69556f6c-c05b-4902-bac7-012f0ed81b75", "Creature — Human", effectstest.StCard("Island", "", "Basic Land — Island", 0, 0)},
		{"Marjhan", "3fdee2ab-7ec6-4fc6-ad99-f04571f94583", "Creature — Serpent", effectstest.StCard("Island", "", "Basic Land — Island", 0, 0)},
		{"Island Fish Jasconius", "bb217f12-532f-4833-a27a-99e290aa47d0", "Creature — Fish", effectstest.StCard("Island", "", "Basic Land — Island", 0, 0)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			// A plain Island under Ronom Serpent is not a snow land.
			if tc.name == "Ronom Serpent" {
				effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Island", "", "Basic Land — Island", 0, 0))
			}
			land := effectstest.ApaPush(g, me.ID, me.ID, tc.land)
			serpent := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard(tc.name, tc.oracle, tc.typeLine, 4, 4))
			effectstest.PassPriorityAroundTable(t, g)
			if !effects.OnBattlefield(g, serpent) {
				t.Fatalf("%s left while its controller had the land", tc.name)
			}
			if c := effectstest.ApaLive(g, serpent); len(c.Effective().AttackTargetRestrictions) == 0 {
				t.Errorf("%s has no attack-target restriction", tc.name)
			}
			effectstest.StDestroy(t, g, land)
			effectstest.PassPriorityAroundTable(t, g)
			if effects.OnBattlefield(g, serpent) {
				t.Fatalf("%s survived the loss of its last land", tc.name)
			}
		})
	}
}

// Dark Depths: ten ice counters as it enters; the {3} row takes one off;
// with none left it is sacrificed for Marit Lage.
func TestDarkDepthsMakesMaritLage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	depths := effectstest.PlayLandFromHand(t, g, "Dark Depths", "c9b82110-7dfd-4617-9399-9510be449043")
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.ApaLive(g, depths); c == nil || c.Counters["ice"] != 10 {
		t.Fatalf("Dark Depths entered with %v ice counters, want 10", c)
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(depths, "ice", -9); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.ApaMana(me, "C", "C", "C")
	effectstest.ApaActivate(t, g, me, depths, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, depths) {
		t.Fatal("Dark Depths with no ice counters is still on the battlefield")
	}
	if n := effectstest.OnBattlefieldNamed(g, "Marit Lage"); n != 1 {
		t.Fatalf("%d Marit Lage tokens, want 1", n)
	}
}

// Endrek Sahr: a creature spell of mana value 3 makes three Thrulls;
// the seventh Thrull triggers the sacrifice.
func TestEndrekSahrBreedsThrullsUntilSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	endrek := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Endrek Sahr, Master Breeder", "47a0079f-3544-45bc-a32a-bd93844c8c43", "Legendary Creature — Human Wizard", 2, 2))
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{G}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "G", "C", "C")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.OnBattlefieldNamed(g, "Thrull"); n != 3 {
		t.Fatalf("%d Thrulls after a mana value 3 creature spell, want 3", n)
	}
	if !effects.OnBattlefield(g, endrek) {
		t.Fatal("Endrek Sahr left with three Thrulls")
	}
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 black Thrull"), 4); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, endrek) {
		t.Fatal("Endrek Sahr survived seven Thrulls")
	}
}

// Last Laugh: a permanent going to a graveyard from the battlefield deals
// 1 to each creature and each player; with no creatures left it goes.
func TestLastLaughPingsAndLeavesWithTheLastCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	bear := effectstest.ApaPush(g, bob.ID, bob.ID, effectstest.StCard("Bear", "", "Creature — Bear", 2, 2))
	wall := effectstest.ApaPush(g, bob.ID, bob.ID, effectstest.StCard("Wall", "", "Creature — Wall", 0, 1))
	laugh := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Last Laugh", "5facb256-b993-43f8-b971-b3a59a7434bf", "Enchantment", 0, 0))
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, laugh) {
		t.Fatal("Last Laugh left with creatures on the battlefield")
	}
	life := me.Life
	effectstest.StDestroy(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Fatalf("life %d, want %d after one death", me.Life, life-1)
	}
	if effects.OnBattlefield(g, wall) {
		t.Fatal("the 0/1 Wall survived Last Laugh's 1 damage")
	}
	// The Wall's death triggers the damage again, and the empty board
	// triggers the sacrifice: two triggers of one controller to order.
	for i := 0; i < 4 && effects.OnBattlefield(g, laugh); i++ {
		effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if effects.OnBattlefield(g, laugh) {
		t.Fatal("Last Laugh stayed with no creatures on the battlefield")
	}
}

// Lurebound Scarecrow: the colour is chosen as it enters, and it is
// sacrificed when its controller has nothing of that colour.
func TestLureboundScarecrowWatchesItsColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	red := game.Card{Name: "Goblin", TypeLine: "Creature — Goblin", Power: 1, Toughness: 1, Colors: []string{"R"}}
	goblin := effectstest.ApaPush(g, me.ID, me.ID, red)
	scarecrow := effectstest.StCard("Lurebound Scarecrow", "8353f834-678a-4fa1-857a-8bfdfb8d6378", "Artifact Creature — Scarecrow", 4, 4)
	scarecrow.ChosenColor = "R" // as if it had chosen red as it entered
	sc := effectstest.ApaPush(g, me.ID, me.ID, scarecrow)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, sc) {
		t.Fatal("the Scarecrow left while its controller had a red permanent")
	}
	effectstest.StDestroy(t, g, goblin)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, sc) {
		t.Fatal("the Scarecrow survived the loss of its controller's last red permanent")
	}
}

// Mazemind Tome: the fourth page counter is part of the cost, so it
// triggers the exile at once; the exile gains 4 life.
func TestMazemindTomeExilesOnTheFourthPage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tome := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Mazemind Tome", "800fb917-8898-4eba-9f94-aec6e9d75236", "Artifact — Book", 0, 0))
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(tome, "page", 3); err != nil {
			t.Fatal(err)
		}
	})
	life := me.Life
	effectstest.ApaMana(me, "C", "C")
	effectstest.ApaActivate(t, g, me, tome, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, tome) {
		t.Fatal("Mazemind Tome with four page counters is still on the battlefield")
	}
	if me.Life != life+4 {
		t.Fatalf("life %d, want %d", me.Life, life+4)
	}
}

// Plague Boiler: the third plague counter sacrifices it and destroys
// every nonland permanent; lands stay.
func TestPlagueBoilerWipesOnTheThirdCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	boiler := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Plague Boiler", "fef502af-6e79-4c55-a86a-b45adb3fc64a", "Artifact", 0, 0))
	bear := effectstest.ApaPush(g, bob.ID, bob.ID, effectstest.StCard("Bear", "", "Creature — Bear", 2, 2))
	land := effectstest.ApaPush(g, bob.ID, bob.ID, effectstest.StCard("Forest", "", "Basic Land — Forest", 0, 0))
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(boiler, "plague", 3); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, boiler) || effects.OnBattlefield(g, bear) || !effects.OnBattlefield(g, land) {
		t.Fatalf("boiler %v bear %v land %v; want the boiler and the bear gone and the land kept",
			effects.OnBattlefield(g, boiler), effects.OnBattlefield(g, bear), effects.OnBattlefield(g, land))
	}
}

// Afiya Grove is sacrificed once its last +1/+1 counter is gone.
func TestAfiyaGroveSacrificedWhenEmpty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Bear", "", "Creature — Bear", 2, 2))
	tmpl := effectstest.StCard("Afiya Grove", "056a98aa-f3f7-4ac9-9755-3e8a60a59abb", "Enchantment", 0, 0)
	tmpl.Counters = map[string]int{game.CounterPlusOne: 1}
	grove := effectstest.ApaPush(g, me.ID, me.ID, tmpl)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, grove) {
		t.Fatal("a Grove with a counter left the battlefield")
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(grove, game.CounterPlusOne, -1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, grove) {
		t.Fatal("a Grove with no +1/+1 counters is still on the battlefield")
	}
}
