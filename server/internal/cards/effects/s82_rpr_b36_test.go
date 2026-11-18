package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b36_test.go — S82 reprint pass, batch 36 (#5908).

const (
	rprSkyshipStalkerOracle     = "bece52d6-fc25-486a-9811-cfc3aeca285b"
	rprVanishIntoMemoryOracle   = "d3beb624-f589-4ed5-97a0-11ad61481250"
	rprViridianLongbowOracle    = "8b9119b0-9948-497f-aa92-2879d07bed1c"
	rprZhalfirinVoidOracle      = "13aab4fc-4c89-45e6-8275-b074b00d0ee9"
	rprBlinkmothNexusOracle     = "40d45c02-6416-4e19-8fe3-0ddadf5ba627"
	rprCathedralOfWarOracle     = "5ff647e4-730a-498f-8f2c-5bd64d5a9780"
	rprCelestialColonnadeOracle = "876ac6f6-74de-4666-84c6-83d81f054723"
	rprContestedWarZoneOracle   = "ed73de2b-d7f4-48d9-9be2-aa9d111b7aa7"
	rprDustBowlOracle           = "d3df7128-31dd-4d71-90be-87e2e9ff51b4"
	rprEndlessSandsOracle       = "c4033d97-769f-4811-8b11-f85b8817b7a2"
	rprHissingQuagmireOracle    = "7fff5224-c002-4a32-86fe-9a4b18d78b50"
)

func TestRprB36CardsAreRegisteredFull(t *testing.T) {
	for name, oracle := range map[string]string{
		"Skyship Stalker": rprSkyshipStalkerOracle, "Vanish into Memory": rprVanishIntoMemoryOracle,
		"Viridian Longbow": rprViridianLongbowOracle, "Zhalfirin Void": rprZhalfirinVoidOracle,
		"Blinkmoth Nexus": rprBlinkmothNexusOracle, "Cathedral of War": rprCathedralOfWarOracle,
		"Celestial Colonnade": rprCelestialColonnadeOracle, "Contested War Zone": rprContestedWarZoneOracle,
		"Dust Bowl": rprDustBowlOracle, "Endless Sands": rprEndlessSandsOracle,
		"Hissing Quagmire": rprHissingQuagmireOracle,
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: registered=%v name=%q completeness=%v", name, ok, spec.Name, spec.Completeness)
		}
	}
}

func TestRprB36SkyshipStalkerPumpsAndGainsKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dragon := b27Push(g, me.ID, "Skyship Stalker", "Creature — Cat Dragon", rprSkyshipStalkerOracle, "{2}{R}{R}", 3, 3, "R")
	advanceToMain(t, g)
	for i := 0; i < 2; i++ {
		fillPoolColored(me, "R", 1)
		b16Activate(t, g, me.ID, dragon, 0, game.ActivateAbilityParams{})
		passPriorityAroundTable(t, g)
	}
	if p := effectivePower(t, g, dragon); p != 5 {
		t.Errorf("power %d after two pumps, want 5", p)
	}
	for idx, kw := range map[int]string{1: "first strike", 2: "haste"} {
		if hasEffectiveKeyword(t, g, dragon, kw) {
			t.Fatalf("has %s before activating", kw)
		}
		fillPoolColored(me, "R", 1)
		b16Activate(t, g, me.ID, dragon, idx, game.ActivateAbilityParams{})
		passPriorityAroundTable(t, g)
		if !hasEffectiveKeyword(t, g, dragon, kw) {
			t.Errorf("did not gain %s", kw)
		}
	}
	// A pool without red can't pay.
	me.ManaPool.EmptyPool()
	fillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, dragon, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated without red mana")
	}
}

func TestRprB36ViridianLongbowGrantsATapPing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bow := b27Push(g, me.ID, "Viridian Longbow", "Artifact — Equipment", rprViridianLongbowOracle, "{1}", 0, 0)
	host := b16Creature(g, me.ID, "Host", "Creature — Bear", 2, 5, "G")
	advanceToMain(t, g)
	if idx, _ := grantedActivatedIndex(t, g, host); idx >= 0 {
		t.Fatal("the creature has the ability before it is equipped")
	}
	floatMana(t, g, me, "{C}{C}{C}")
	equipTo(t, g, me.ID, bow, host)
	idx, ref := grantedActivatedIndex(t, g, host)
	if idx < 0 {
		t.Fatal("the equipped creature has no granted ability")
	}
	before := opp.Life
	b16Activate(t, g, me.ID, host, idx, game.ActivateAbilityParams{
		Ref: ref, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("life %d → %d, want 1 damage", before, opp.Life)
	}
	if !layeredCard(t, g, host).Tapped {
		t.Error("the ability's {T} cost was not paid")
	}
	if err := g.ActivateCatalogAbility(me.ID, host, idx, game.ActivateAbilityParams{
		Ref: ref, Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err == nil {
		t.Error("a tapped creature activated the ability again")
	}
}

func TestRprB36ZhalfirinVoidScriesOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Bad", "Good")
	land := playLandFromHand(t, g, "Zhalfirin Void", rprZhalfirinVoidOracle)
	if isTapped(g, land) {
		t.Error("Zhalfirin Void entered tapped")
	}
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("scry choice %+v, want a scry 1", c)
	}
}

func TestRprB36CathedralOfWarEntersTappedWithExalted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := b12PlayFromHand(t, g, "Cathedral of War", "Land", rprCathedralOfWarOracle, game.CastSpellParams{})
	if !isTapped(g, land) {
		t.Error("Cathedral of War entered untapped")
	}
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 3 {
		t.Errorf("a lone attacker has power %d, want 3 (exalted)", p)
	}
}

func TestRprB36TheTwoManlandsAnimateAndEnterTapped(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, kw string
		power, tough     int
		cost             []string
		colours          []string
	}{
		{"Celestial Colonnade", rprCelestialColonnadeOracle, "flying", 4, 4, []string{"C", "C", "C", "W", "U"}, []string{"W", "U"}},
		{"Hissing Quagmire", rprHissingQuagmireOracle, "deathtouch", 2, 2, []string{"C", "B", "G"}, []string{"B", "G"}},
	} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		if id := b12PlayFromHand(t, g, tc.name, "Land", tc.oracle, game.CastSpellParams{}); !isTapped(g, id) {
			t.Errorf("%s entered untapped", tc.name)
		}
		land := s58p6Push(g, me.ID, tc.name, "Land", tc.oracle, "", 0, 0)
		advanceTo(t, g, game.StepPrecombatMain)
		if s58p6View(t, g, land).IsCreature() {
			t.Fatalf("%s: an idle land is a creature", tc.name)
		}
		if err := cmdr2B13Activate(t, g, me, land, game.ActivateAbilityParams{}, tc.cost...); err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		c := s58p6View(t, g, land)
		if !c.IsCreature() || !c.IsLand() || c.CurrentPower() != tc.power || c.CurrentToughness() != tc.tough {
			t.Errorf("%s: creature=%v land=%v %d/%d", tc.name, c.IsCreature(), c.IsLand(), c.CurrentPower(), c.CurrentToughness())
		}
		if !slices.Contains(c.Effective().Subtypes, "Elemental") || !slices.Equal(c.EffectiveColors(), tc.colours) {
			t.Errorf("%s: subtypes %v colours %v", tc.name, c.Effective().Subtypes, c.EffectiveColors())
		}
		assertEffectiveKeyword(t, g, land, tc.kw)
		if tc.name == "Celestial Colonnade" {
			assertEffectiveKeyword(t, g, land, "vigilance")
		}
		// Not enough mana: refused.
		me.ManaPool.EmptyPool()
		if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
			t.Errorf("%s: activated with an empty pool", tc.name)
		}
	}
}

func TestRprB36BlinkmothNexusAnimatesAndPumpsABlinkmoth(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	nexus := s58p6Push(g, me.ID, "Blinkmoth Nexus", "Land", rprBlinkmothNexusOracle, "", 0, 0)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	advanceToMain(t, g)
	// A non-Blinkmoth is not a legal target of the pump.
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, nexus, 1, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("a Bear was accepted as a Blinkmoth target")
	}
	me.ManaPool.EmptyPool()
	floatMana(t, g, me, "{C}")
	b16Activate(t, g, me.ID, nexus, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	c := s58p6View(t, g, nexus)
	if !c.IsCreature() || !c.IsLand() || !c.IsArtifact() || c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Fatalf("creature=%v land=%v artifact=%v %d/%d", c.IsCreature(), c.IsLand(), c.IsArtifact(), c.CurrentPower(), c.CurrentToughness())
	}
	if !slices.Contains(c.Effective().Subtypes, "Blinkmoth") {
		t.Errorf("subtypes %v, want Blinkmoth", c.Effective().Subtypes)
	}
	assertEffectiveKeyword(t, g, nexus, "flying")
	floatMana(t, g, me, "{C}")
	b16Activate(t, g, me.ID, nexus, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: nexus}}})
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, nexus); p != 2 {
		t.Errorf("power %d after the pump, want 2", p)
	}
}

func TestRprB36ContestedWarZonePumpsAttackersAndChangesHands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zone := b12Push(g, opp.ID, "Contested War Zone", "Land", rprContestedWarZoneOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	attackInto(t, g, map[uuid.UUID]uuid.UUID{bear: opp.ID})
	floatMana(t, g, opp, "{C}")
	b16Activate(t, g, opp.ID, zone, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 3 {
		t.Errorf("attacker power %d, want 3", p)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	var controller uuid.UUID
	g.ReadSnapshot(func() { controller = findBattlefieldCardForTest(g, zone).Controller })
	if controller != me.ID {
		t.Errorf("the War Zone is controlled by %s, want the damaging creature's controller %s", controller, me.ID)
	}
}

func TestRprB36DustBowlSacrificesAnyLandAndDestroysANonbasic(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bowl := b12Push(g, me.ID, "Dust Bowl", "Land", rprDustBowlOracle, 0, 0)
	forest := b12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	basic := b12Push(g, opp.ID, "Island", "Basic Land — Island", "", 0, 0)
	nonbasic := b12Push(g, opp.ID, "Strip Mine", "Land", "", 0, 0)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, bowl, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{forest}, Targets: cardRefs(basic),
	}); err == nil {
		t.Fatal("a basic land was accepted as the target")
	}
	me.ManaPool.EmptyPool()
	floatMana(t, g, me, "{C}{C}{C}")
	b16Activate(t, g, me.ID, bowl, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{forest}, Targets: cardRefs(nonbasic)})
	passPriorityAroundTable(t, g)
	if onBattlefield(g, nonbasic) {
		t.Error("the nonbasic land survived")
	}
	if onBattlefield(g, forest) {
		t.Error("the sacrificed land is still on the battlefield")
	}
	if !onBattlefield(g, bowl) || !onBattlefield(g, basic) {
		t.Error("Dust Bowl and the basic should remain")
	}
}

func TestRprB36EndlessSandsExilesThenReturnsEachCreatureUnderItsOwner(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sands := b12Push(g, me.ID, "Endless Sands", "Land — Desert", rprEndlessSandsOracle, 0, 0)
	// A creature I control but the opponent owns.
	stolen := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stolen Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3,
		Owner: opp.ID, Controller: me.ID,
	})
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	// Exiling takes the Sands' tap, so untap between the two exiles.
	if err := g.ActivateCatalogAbility(me.ID, sands, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err == nil {
		t.Fatal("a player was accepted as the target")
	}
	for _, id := range []uuid.UUID{stolen, mine} {
		floatMana(t, g, me, "{C}{C}")
		b16Activate(t, g, me.ID, sands, 0, game.ActivateAbilityParams{Targets: cardRefs(id)})
		passPriorityAroundTable(t, g)
		g.WithWriteLock(func() { findBattlefieldCardByID(g, sands).Tapped = false })
	}
	if onBattlefield(g, stolen) || onBattlefield(g, mine) {
		t.Fatal("the creatures were not exiled")
	}
	floatMana(t, g, me, "{C}{C}{C}{C}")
	b16Activate(t, g, me.ID, sands, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if onBattlefield(g, sands) {
		t.Error("the Sands was not sacrificed")
	}
	var back []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Stolen Ogre" || c.Name == "Bear" {
				back = append(back, c)
			}
		}
	})
	if len(back) != 2 {
		t.Fatalf("%d creatures came back, want 2", len(back))
	}
	for _, c := range back {
		want := me.ID
		if c.Name == "Stolen Ogre" {
			want = opp.ID
		}
		if c.Controller != want {
			t.Errorf("%s returned under %s, want its owner %s", c.Name, c.Controller, want)
		}
	}
}

func TestRprB36VanishIntoMemoryDrawsExilesReturnsAndDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "A", "B", "C", "D", "E", "F")
	ogre := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 2,
		Owner: opp.ID, Controller: opp.ID,
	})
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Vanish into Memory", "Instant", rprVanishIntoMemoryOracle, cardRefs(ogre))
	passPriorityAroundTable(t, g)
	if onBattlefield(g, ogre) {
		t.Fatal("the creature was not exiled")
	}
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d → %d, want +3 (the creature's power)", hand, got)
	}
	// My next upkeep: the card returns under its owner's control, then I
	// discard cards equal to its toughness.
	b12ToMyNextUpkeepAfterTurn(t, g)
	passPriorityAroundTable(t, g)
	var owner uuid.UUID
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Ogre" {
				owner, found = c.Controller, true
			}
		}
	})
	if !found || owner != opp.ID {
		t.Fatalf("the Ogre is back=%v under %s, want its owner %s", found, owner, opp.ID)
	}
	c := discardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no discard prompt after the card returned")
	}
	if c.ChooseMax != 2 {
		t.Errorf("discard %d, want 2 (the creature's toughness)", c.ChooseMax)
	}
}

// b12ToMyNextUpkeepAfterTurn walks past the current turn into seat 0's next
// upkeep.
func b12ToMyNextUpkeepAfterTurn(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 600; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == 0 && g.Turn.Round > 1 {
			return
		}
	}
	t.Fatal("never reached seat 0's next upkeep")
}
