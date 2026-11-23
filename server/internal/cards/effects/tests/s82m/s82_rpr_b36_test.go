package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: registered=%v name=%q completeness=%v", name, ok, spec.Name, spec.Completeness)
		}
	}
}

func TestRprB36SkyshipStalkerPumpsAndGainsKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dragon := effectstest.B27Push(g, me.ID, "Skyship Stalker", "Creature — Cat Dragon", rprSkyshipStalkerOracle, "{2}{R}{R}", 3, 3, "R")
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 2; i++ {
		effectstest.FillPoolColored(me, "R", 1)
		effectstest.B16Activate(t, g, me.ID, dragon, 0, game.ActivateAbilityParams{})
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p := effectstest.EffectivePower(t, g, dragon); p != 5 {
		t.Errorf("power %d after two pumps, want 5", p)
	}
	for idx, kw := range map[int]string{1: "first strike", 2: "haste"} {
		if effectstest.HasEffectiveKeyword(t, g, dragon, kw) {
			t.Fatalf("has %s before activating", kw)
		}
		effectstest.FillPoolColored(me, "R", 1)
		effectstest.B16Activate(t, g, me.ID, dragon, idx, game.ActivateAbilityParams{})
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasEffectiveKeyword(t, g, dragon, kw) {
			t.Errorf("did not gain %s", kw)
		}
	}
	// A pool without red can't pay.
	me.ManaPool.EmptyPool()
	effectstest.FillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, dragon, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated without red mana")
	}
}

func TestRprB36ViridianLongbowGrantsATapPing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bow := effectstest.B27Push(g, me.ID, "Viridian Longbow", "Artifact — Equipment", rprViridianLongbowOracle, "{1}", 0, 0)
	host := effectstest.B16Creature(g, me.ID, "Host", "Creature — Bear", 2, 5, "G")
	effectstest.AdvanceToMain(t, g)
	if idx, _ := effectstest.GrantedActivatedIndex(t, g, host); idx >= 0 {
		t.Fatal("the creature has the ability before it is equipped")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	effectstest.EquipTo(t, g, me.ID, bow, host)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, host)
	if idx < 0 {
		t.Fatal("the equipped creature has no granted ability")
	}
	before := opp.Life
	effectstest.B16Activate(t, g, me.ID, host, idx, game.ActivateAbilityParams{
		Ref: ref, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("life %d → %d, want 1 damage", before, opp.Life)
	}
	if !effectstest.LayeredCard(t, g, host).Tapped {
		t.Error("the ability's {T} cost was not paid")
	}
	if err := g.ActivateCatalogAbility(me.ID, host, idx, game.ActivateAbilityParams{
		Ref: ref, Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err == nil {
		t.Error("a tapped creature activated the ability again")
	}
}

func TestRprB36ZhalfirinVoidScriesOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Bad", "Good")
	land := effectstest.PlayLandFromHand(t, g, "Zhalfirin Void", rprZhalfirinVoidOracle)
	if effectstest.IsTapped(g, land) {
		t.Error("Zhalfirin Void entered tapped")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("scry choice %+v, want a scry 1", c)
	}
}

func TestRprB36CathedralOfWarEntersTappedWithExalted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.B12PlayFromHand(t, g, "Cathedral of War", "Land", rprCathedralOfWarOracle, game.CastSpellParams{})
	if !effectstest.IsTapped(g, land) {
		t.Error("Cathedral of War entered untapped")
	}
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
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
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		if id := effectstest.B12PlayFromHand(t, g, tc.name, "Land", tc.oracle, game.CastSpellParams{}); !effectstest.IsTapped(g, id) {
			t.Errorf("%s entered untapped", tc.name)
		}
		land := effectstest.S58p6Push(g, me.ID, tc.name, "Land", tc.oracle, "", 0, 0)
		effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
		if effectstest.S58p6View(t, g, land).IsCreature() {
			t.Fatalf("%s: an idle land is a creature", tc.name)
		}
		if err := effectstest.Cmdr2B13Activate(t, g, me, land, game.ActivateAbilityParams{}, tc.cost...); err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		c := effectstest.S58p6View(t, g, land)
		if !c.IsCreature() || !c.IsLand() || c.CurrentPower() != tc.power || c.CurrentToughness() != tc.tough {
			t.Errorf("%s: creature=%v land=%v %d/%d", tc.name, c.IsCreature(), c.IsLand(), c.CurrentPower(), c.CurrentToughness())
		}
		if !slices.Contains(c.Effective().Subtypes, "Elemental") || !slices.Equal(c.EffectiveColors(), tc.colours) {
			t.Errorf("%s: subtypes %v colours %v", tc.name, c.Effective().Subtypes, c.EffectiveColors())
		}
		effectstest.AssertEffectiveKeyword(t, g, land, tc.kw)
		if tc.name == "Celestial Colonnade" {
			effectstest.AssertEffectiveKeyword(t, g, land, "vigilance")
		}
		// Not enough mana: refused.
		me.ManaPool.EmptyPool()
		if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
			t.Errorf("%s: activated with an empty pool", tc.name)
		}
	}
}

func TestRprB36BlinkmothNexusAnimatesAndPumpsABlinkmoth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	nexus := effectstest.S58p6Push(g, me.ID, "Blinkmoth Nexus", "Land", rprBlinkmothNexusOracle, "", 0, 0)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	// A non-Blinkmoth is not a legal target of the pump.
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, nexus, 1, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("a Bear was accepted as a Blinkmoth target")
	}
	me.ManaPool.EmptyPool()
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.B16Activate(t, g, me.ID, nexus, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.S58p6View(t, g, nexus)
	if !c.IsCreature() || !c.IsLand() || !c.IsArtifact() || c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Fatalf("creature=%v land=%v artifact=%v %d/%d", c.IsCreature(), c.IsLand(), c.IsArtifact(), c.CurrentPower(), c.CurrentToughness())
	}
	if !slices.Contains(c.Effective().Subtypes, "Blinkmoth") {
		t.Errorf("subtypes %v, want Blinkmoth", c.Effective().Subtypes)
	}
	effectstest.AssertEffectiveKeyword(t, g, nexus, "flying")
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.B16Activate(t, g, me.ID, nexus, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: nexus}}})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, nexus); p != 2 {
		t.Errorf("power %d after the pump, want 2", p)
	}
}

func TestRprB36ContestedWarZonePumpsAttackersAndChangesHands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zone := effectstest.B12Push(g, opp.ID, "Contested War Zone", "Land", rprContestedWarZoneOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AttackInto(t, g, map[uuid.UUID]uuid.UUID{bear: opp.ID})
	effectstest.FloatMana(t, g, opp, "{C}")
	effectstest.B16Activate(t, g, opp.ID, zone, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("attacker power %d, want 3", p)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	var controller uuid.UUID
	g.ReadSnapshot(func() { controller = effectstest.FindBattlefieldCardForTest(g, zone).Controller })
	if controller != me.ID {
		t.Errorf("the War Zone is controlled by %s, want the damaging creature's controller %s", controller, me.ID)
	}
}

func TestRprB36DustBowlSacrificesAnyLandAndDestroysANonbasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bowl := effectstest.B12Push(g, me.ID, "Dust Bowl", "Land", rprDustBowlOracle, 0, 0)
	forest := effectstest.B12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	basic := effectstest.B12Push(g, opp.ID, "Island", "Basic Land — Island", "", 0, 0)
	nonbasic := effectstest.B12Push(g, opp.ID, "Strip Mine", "Land", "", 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, bowl, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{forest}, Targets: effectstest.CardRefs(basic),
	}); err == nil {
		t.Fatal("a basic land was accepted as the target")
	}
	me.ManaPool.EmptyPool()
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	effectstest.B16Activate(t, g, me.ID, bowl, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{forest}, Targets: effectstest.CardRefs(nonbasic)})
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, nonbasic) {
		t.Error("the nonbasic land survived")
	}
	if effects.OnBattlefield(g, forest) {
		t.Error("the sacrificed land is still on the battlefield")
	}
	if !effects.OnBattlefield(g, bowl) || !effects.OnBattlefield(g, basic) {
		t.Error("Dust Bowl and the basic should remain")
	}
}

func TestRprB36EndlessSandsExilesThenReturnsEachCreatureUnderItsOwner(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sands := effectstest.B12Push(g, me.ID, "Endless Sands", "Land — Desert", rprEndlessSandsOracle, 0, 0)
	// A creature I control but the opponent owns.
	stolen := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stolen Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3,
		Owner: opp.ID, Controller: me.ID,
	})
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	// Exiling takes the Sands' tap, so untap between the two exiles.
	if err := g.ActivateCatalogAbility(me.ID, sands, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err == nil {
		t.Fatal("a player was accepted as the target")
	}
	for _, id := range []uuid.UUID{stolen, mine} {
		effectstest.FloatMana(t, g, me, "{C}{C}")
		effectstest.B16Activate(t, g, me.ID, sands, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(id)})
		effectstest.PassPriorityAroundTable(t, g)
		g.WithWriteLock(func() { effectstest.FindBattlefieldCardByID(g, sands).Tapped = false })
	}
	if effects.OnBattlefield(g, stolen) || effects.OnBattlefield(g, mine) {
		t.Fatal("the creatures were not exiled")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	effectstest.B16Activate(t, g, me.ID, sands, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, sands) {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	ogre := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 2,
		Owner: opp.ID, Controller: opp.ID,
	})
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Vanish into Memory", "Instant", rprVanishIntoMemoryOracle, effectstest.CardRefs(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, ogre) {
		t.Fatal("the creature was not exiled")
	}
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d → %d, want +3 (the creature's power)", hand, got)
	}
	// My next upkeep: the card returns under its owner's control, then I
	// discard cards equal to its toughness.
	b12ToMyNextUpkeepAfterTurn(t, g)
	effectstest.PassPriorityAroundTable(t, g)
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
	c := effectstest.DiscardChoiceFor(g, me.ID)
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
