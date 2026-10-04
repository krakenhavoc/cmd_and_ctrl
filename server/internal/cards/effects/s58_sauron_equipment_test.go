package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s58_sauron_equipment_test.go — S58 PR 7: Luxior, The Key to the Vault,
// Dawnsire and Vincent Valentine // Galian Beast.

const (
	luxiorOracle       = "2f9263d7-916f-4535-95d2-888ab73cf339"
	keyToTheVaultOracl = "7a75d565-b0a7-46bf-94a0-c9255c8abd6d"
	dawnsireOracle     = "afc9436b-8cad-4916-929d-ff33a37b42d5"
	vincentOracle      = "f40e6bbf-1fed-4ec5-869d-18c2dd396d14"
)

// sauronPT is a battlefield permanent's power and toughness with its
// counters, after a fresh layer pass.
func sauronPT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := findBattlefieldCardByID(g, id)
	if c == nil {
		t.Fatalf("%v is not on the battlefield", id)
	}
	return c.CurrentPower(), c.CurrentToughness()
}

func sauronEquip(t *testing.T, g *game.Game, controller, equipment uuid.UUID, ability int, target uuid.UUID) error {
	t.Helper()
	err := g.ActivateCatalogAbility(controller, equipment, ability, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	})
	if err == nil {
		passPriorityAroundTable(t, g)
	}
	return err
}

// --- Luxior -----------------------------------------------------------

func TestLuxiorMakesAnEquippedPlaneswalkerACreatureThatKeepsItsLoyalty(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	fillPool(me, 4)
	luxior := pushCatalogPermanent(g, me.ID, "Luxior, Giada's Gift", "Legendary Artifact — Equipment", luxiorOracle, false)
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Walker", TypeLine: "Legendary Planeswalker — Test",
		StartingLoyalty: 4, Counters: map[string]int{game.CounterLoyalty: 4},
		Owner: me.ID, Controller: me.ID,
	})
	bear := pushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)

	// "Equip planeswalker" names a planeswalker, not a creature.
	if err := sauronEquip(t, g, me.ID, luxior, 0, bear); err == nil {
		t.Error("Equip planeswalker should not accept a Bear")
	}
	if err := sauronEquip(t, g, me.ID, luxior, 0, walker); err != nil {
		t.Fatalf("Equip planeswalker: %v", err)
	}

	c := findBattlefieldCardByID(g, walker)
	if c == nil {
		t.Fatal("the planeswalker left the battlefield")
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c.IsPlaneswalker() {
		t.Error("equipped permanent isn't a planeswalker")
	}
	if !c.IsCreature() {
		t.Error("equipped permanent is a creature")
	}
	if lux := findBattlefieldCardByID(g, luxior); lux == nil || !lux.IsAttachedTo(walker) {
		t.Fatal("Luxior must stay attached: the equipped permanent is a creature, so the equipment state-based action leaves it alone")
	}
	if p, tough := sauronPT(t, g, walker); p != 4 || tough != 4 {
		t.Errorf("the equipped walker is %d/%d, want 4/4 (+1/+1 for each of its 4 loyalty counters)", p, tough)
	}

	// Its counters are the whole size, and they move with it.
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(walker, game.CounterLoyalty, 2); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	if p, tough := sauronPT(t, g, walker); p != 6 || tough != 6 {
		t.Errorf("after +2 loyalty the walker is %d/%d, want 6/6", p, tough)
	}
}

func TestLuxiorPumpsAnOrdinaryCreatureByItsCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	fillPool(me, 3)
	luxior := pushCatalogPermanent(g, me.ID, "Luxior, Giada's Gift", "Legendary Artifact — Equipment", luxiorOracle, false)
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, PrintedPTKnown: true,
		Counters: map[string]int{game.CounterPlusOne: 2, "charge": 1},
		Owner:    me.ID, Controller: me.ID,
	})
	if err := sauronEquip(t, g, me.ID, luxior, 1, bear); err != nil {
		t.Fatalf("Equip {3}: %v", err)
	}
	// 2/2 base, two +1/+1 counters make it 4/4, and each of its three
	// counters (two +1/+1 and a charge) adds another +1/+1.
	if p, tough := sauronPT(t, g, bear); p != 7 || tough != 7 {
		t.Errorf("the equipped Bear is %d/%d, want 7/7", p, tough)
	}
}

// --- The Key to the Vault --------------------------------------------

func TestKeyToTheVaultLooksAtThatManyCardsAndCastsOneFree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	key := pushCatalogPermanent(g, me.ID, "The Key to the Vault", "Legendary Artifact — Equipment", keyToTheVaultOracl, false)
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 3, Toughness: 3, PrintedPTKnown: true,
		Owner: me.ID, Controller: me.ID,
	})
	fillPoolColored(me, "U", 3)
	if err := sauronEquip(t, g, me.ID, key, 0, bear); err != nil {
		t.Fatalf("equip: %v", err)
	}
	ids := sunbirdLibrary(me,
		game.Card{Name: "A Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Free Wurm", TypeLine: "Creature — Wurm", ManaCost: "{4}{G}", Power: 5, Toughness: 5},
		game.Card{Name: "Another Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Too Deep", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
	)
	forest, wurm, bear2, deep := ids[0], ids[1], ids[2], ids[3]

	dealCombatDamageToPlayer(g, bear, opp.ID, 3)
	passPriorityUntilChoice(t, g)
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the Key should offer the nonland cards among the top three")
	}
	if !hasID(c.ChooseCards, wurm) || !hasID(c.ChooseCards, bear2) || hasID(c.ChooseCards, forest) || hasID(c.ChooseCards, deep) {
		t.Errorf("offered %v: want the two nonland cards of the top three only", c.ChooseCards)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{wurm}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !inExile(g, wurm) {
		t.Fatal("the chosen card is exiled")
	}
	for _, id := range []uuid.UUID{forest, bear2} {
		if i := libraryIndex(me, id); i < 0 || i > 1 {
			t.Errorf("the rest go on the bottom of the library, found at index %d", i)
		}
	}
	if i := libraryIndex(me, deep); i != len(me.Library.Cards)-1 {
		t.Errorf("a card beyond the damage dealt stays on top, found at index %d of %d", i, len(me.Library.Cards))
	}
	if err := castFromExile(t, g, me.ID, wurm); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wurm) {
		t.Error("the free spell should resolve")
	}
}

func TestKeyToTheVaultIgnoresUnequippedAndNoncombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "The Key to the Vault", "Legendary Artifact — Equipment", keyToTheVaultOracl, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 3, 3)
	sunbirdLibrary(me, game.Card{Name: "Free Wurm", TypeLine: "Creature — Wurm", ManaCost: "{4}{G}", Power: 5, Toughness: 5})

	dealCombatDamageToPlayer(g, bear, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("an unattached Key does nothing")
	}
}

// --- Dawnsire, Sunstar Dreadnought -----------------------------------

func TestDawnsireDealsOneHundredAtTenChargeCountersWhenYouAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dawnsire := pushCatalogPermanent(g, me.ID, "Dawnsire, Sunstar Dreadnought", "Legendary Artifact — Spacecraft", dawnsireOracle, false)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(dawnsire, game.CounterCharge, 10); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	attacker := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	victim := pushVanillaCreature(g, opp.ID, "Wall", 0, 50)

	declareAttack(t, g, opp.ID, attacker)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("100 damage should have destroyed the 0/50")
	}
	if c := findBattlefieldCardByID(g, dawnsire); c == nil || c.IsCreature() {
		t.Error("at 10 charge counters Dawnsire is still not a creature")
	}
}

func TestDawnsireDoesNothingBelowTenChargeCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dawnsire := pushCatalogPermanent(g, me.ID, "Dawnsire, Sunstar Dreadnought", "Legendary Artifact — Spacecraft", dawnsireOracle, false)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(dawnsire, game.CounterCharge, 9); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	attacker := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	victim := pushVanillaCreature(g, opp.ID, "Wall", 0, 50)

	declareAttack(t, g, opp.ID, attacker)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("at 9 charge counters the 10+ trigger does not exist")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Error("the Wall should be untouched")
	}
}

func TestDawnsireIsATwentyTwentyFlyerAtTwentyChargeCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dawnsire := pushCatalogPermanent(g, me.ID, "Dawnsire, Sunstar Dreadnought", "Legendary Artifact — Spacecraft", dawnsireOracle, false)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(dawnsire, game.CounterCharge, 20); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
		g.RecomputeLayersIfStaleLocked()
	})
	c := findBattlefieldCardByID(g, dawnsire)
	if c == nil || !c.IsCreature() || !game.HasKeyword(c, "flying") {
		t.Fatalf("at 20 counters Dawnsire is a flying creature (creature=%v)", c != nil && c.IsCreature())
	}
	// 20/20 base: 20 charge counters are not +1/+1 counters.
	if p, tough := sauronPT(t, g, dawnsire); p != 20 || tough != 20 {
		t.Errorf("Dawnsire is %d/%d, want 20/20", p, tough)
	}
}

// --- Vincent Valentine // Galian Beast --------------------------------

func TestVincentGrowsByTheDeadCreaturesPowerAndTransformsOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vincent := pushVincent(g, me.ID)
	victim := pushDiesCreatureForTest(g, opp.ID, "Pumped Bear", "", "Creature — Bear", 4, 4)
	mine := pushDiesCreatureForTest(g, me.ID, "Mine", "", "Creature — Bear", 3, 3)

	killCreature(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, vincent, game.CounterPlusOne); got != 0 {
		t.Errorf("your own creature dying gave Vincent %d counters, want 0", got)
	}
	killCreature(t, g, opp.ID, victim)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, vincent, game.CounterPlusOne); got != 4 {
		t.Errorf("Vincent has %d counters, want 4 (the dead creature's power)", got)
	}
}

// pushVincent puts Vincent Valentine // Galian Beast, front face up, on
// the battlefield.
func pushVincent(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Vincent Valentine", OracleID: vincentOracle,
		Layout: game.LayoutTransform, TypeLine: "Legendary Creature — Assassin",
		Power: 2, Toughness: 2, PrintedPTKnown: true,
		Faces: []game.Face{
			{Name: "Vincent Valentine", TypeLine: "Legendary Creature — Assassin", ManaCost: "{2}{B}{B}", Power: 2, Toughness: 2},
			{Name: "Galian Beast", TypeLine: "Legendary Creature — Werewolf Beast", Power: 3, Toughness: 2},
		},
		Owner: owner, Controller: owner,
	})
}

func TestVincentMayTransformWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vincent := pushVincent(g, me.ID)

	declareAttack(t, g, opp.ID, vincent)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	c := findBattlefieldCardByID(g, vincent)
	if c == nil || c.ActiveFace != 1 {
		t.Fatalf("Vincent should have transformed (face=%v)", c != nil && c.ActiveFace == 1)
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if !game.HasKeyword(c, "trample") || !game.HasKeyword(c, "lifelink") {
		t.Error("Galian Beast has trample and lifelink")
	}
	if p, tough := sauronPT(t, g, vincent); p != 3 || tough != 2 {
		t.Errorf("Galian Beast is %d/%d, want 3/2", p, tough)
	}
}

func TestVincentStaysPutWhenYouDeclineTheTransform(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vincent := pushVincent(g, me.ID)

	declareAttack(t, g, opp.ID, vincent)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardByID(g, vincent); c == nil || c.ActiveFace != 0 {
		t.Error("declining the optional trigger leaves Vincent as he was")
	}
}

func TestGalianBeastReturnsTappedFrontFaceUpWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vincent := pushVincent(g, me.ID)
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if err := g.TransformPermanentForEffect(vincent); err != nil {
			t.Fatalf("TransformForEffect: %v", err)
		}
	})
	if c := findBattlefieldCardByID(g, vincent); c == nil || c.ActiveFace != 1 {
		t.Fatal("setup: Vincent should be the Galian Beast")
	}

	killCreature(t, g, me.ID, vincent)
	passPriorityAroundTable(t, g)

	c := findBattlefieldCardByID(g, vincent)
	if c == nil {
		t.Fatal("Galian Beast should return to the battlefield")
	}
	if !c.Tapped {
		t.Error("it returns tapped")
	}
	if c.ActiveFace != 0 || c.Name != "Vincent Valentine" {
		t.Errorf("it returns front face up, got face %d %q", c.ActiveFace, c.Name)
	}
}
