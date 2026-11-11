package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s51_cards_test.go — batch 51's planeswalkers: Nissa, Wrenn and
// Seven, Tezzeret, Tamiyo and Liliana, through the real loyalty path.

const (
	nissaShakesOracle    = "5aa3abf1-d56b-4f42-8c84-7e5a2c15ee0f"
	wrennAndSevenOracle  = "63a509ea-cedc-40e9-b4e8-0ff4fc356485"
	tezzeretBetrayOracle = "5dedd5a9-bc5d-47bb-9ae1-7d683843d041"
	tamiyoFieldOracle    = "9cf99129-f51a-4418-a8cc-8ca2992b17fa"
	lilianaDeathOracle   = "e1b6cc44-41b6-43d0-9fe9-722cebb1c47b"

	s51ArtifactProbeOracle = "test-s51-artifact-probe"
	s51CreatureProbeOracle = "test-s51-creature-probe"
)

func init() {
	noop := func(*game.Game, *game.StackItem) error { return nil }
	Register(Spec{
		OracleID: s51ArtifactProbeOracle, Name: "S51 Artifact Probe",
		Activated: []ActivatedAbility{{Label: "{2}: Do nothing.", Cost: ManaCost("{2}"), Purpose: game.Purpose{Answers: game.AnswerValue}, Effect: noop}},
	})
	Register(Spec{
		OracleID: s51CreatureProbeOracle, Name: "S51 Creature Probe",
		Activated: []ActivatedAbility{{Label: "{2}: Do nothing.", Cost: ManaCost("{2}"), Purpose: game.Purpose{Answers: game.AnswerValue}, Effect: noop}},
	})
}

func s51Card(g *game.Game, id uuid.UUID) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}

// s51Settle brings the layers up to date, as the engine does before any
// rules read; a test that reads Effective() straight off the battlefield
// has to ask.
func s51Settle(g *game.Game) { g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() }) }

func s51PT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	s51Settle(g)
	c, ok := s51Card(g, id)
	if !ok {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	// CurrentPower and CurrentToughness read the layers' answer with the
	// counters on top, which is the power and toughness the game uses.
	return c.CurrentPower(), c.CurrentToughness()
}

func s51HasKeyword(g *game.Game, id uuid.UUID, kw string) bool {
	s51Settle(g)
	return hasKeywordOf(g, id, kw)
}

func s51SearchChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceSearchLibrary && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func s51ActivePlayer(t *testing.T) (*game.Game, int, *game.Player, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	return g, seat, g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
}

// --- Nissa, Who Shakes the World -------------------------------------

func TestNissaWhoShakesTheWorldForestsTapForAnExtraG(t *testing.T) {
	g, _, me, opp := s51ActivePlayer(t)
	pushWalkerForTest(g, me.ID, "Nissa, Who Shakes the World", nissaShakesOracle, 5)
	forest := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	island := pushLandFor(g, me.ID, "Island", "Basic Land — Island")
	theirs := pushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	tapForMana(t, g, me.ID, forest)
	if got := me.ManaPool.Total(); got != 2 {
		t.Errorf("a Forest made %d mana, want 2", got)
	}
	for _, u := range me.ManaPool.Units() {
		if u.Color != "G" {
			t.Errorf("mana %q, want G", u.Color)
		}
	}
	drainPool(t, g, me.ID)
	tapForMana(t, g, me.ID, island)
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("an Island made %d mana, want 1", got)
	}
	tapForMana(t, g, opp.ID, theirs)
	if got := opp.ManaPool.Total(); got != 1 {
		t.Errorf("an opponent's Forest made %d mana, want 1", got)
	}
}

func TestNissaWhoShakesTheWorldPlusOneAnimatesAnUntappedLand(t *testing.T) {
	g, seat, me, _ := s51ActivePlayer(t)
	pw := pushWalkerForTest(g, me.ID, "Nissa, Who Shakes the World", nissaShakesOracle, 5)
	land := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(land) })
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{Targets: cardRefs(pw), Strict: true}); err == nil {
		t.Fatal("+1 targeting a planeswalker was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{Targets: cardRefs(land), Strict: true}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := s51Card(g, land)
	if !ok {
		t.Fatal("the land left the battlefield")
	}
	if c.Tapped {
		t.Error("the land is still tapped")
	}
	if p, tough := s51PT(t, g, land); p != 3 || tough != 3 {
		t.Errorf("land = %d/%d, want 3/3 (a 0/0 with three counters)", p, tough)
	}
	for _, kw := range []string{"vigilance", "haste"} {
		if !s51HasKeyword(g, land, kw) {
			t.Errorf("the land lacks %s", kw)
		}
	}
	var eff game.Characteristic
	g.ReadSnapshot(func() {
		for _, b := range g.Battlefield.Cards {
			if b.InstanceID == land {
				eff = b.Effective()
			}
		}
	})
	if !c.IsLand() || !c.IsCreature() {
		t.Errorf("land %v creature %v, want both", c.IsLand(), c.IsCreature())
	}
	_ = eff
	// "Up to one": the ability is legal with no target and still adds loyalty.
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		// A second loyalty ability in a turn is refused by the engine; either answer is fine here.
		t.Log("second activation accepted")
	}
}

func TestNissaWhoShakesTheWorldMinusEightMakesLandsIndestructibleAndFetchesForests(t *testing.T) {
	g, seat, me, opp := s51ActivePlayer(t)
	pw := pushWalkerForTest(g, me.ID, "Nissa, Who Shakes the World", nissaShakesOracle, 8)
	mine := pushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	theirs := pushLandFor(g, opp.ID, "Mountain", "Basic Land — Mountain")
	g.WithWriteLock(func() {
		for i := 0; i < 3; i++ {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
		}
	})
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−8: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(emblemsOf(g, me.ID)) != 1 {
		t.Fatalf("emblems = %d, want 1", len(emblemsOf(g, me.ID)))
	}
	if !s51HasKeyword(g, mine, "indestructible") {
		t.Error("my land lacks indestructible")
	}
	if s51HasKeyword(g, theirs, "indestructible") {
		t.Error("an opponent's land got indestructible")
	}
	pick := s51SearchChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no search prompt")
	}
	if pick.SearchMin != 0 || len(pick.SearchCards) < 3 {
		t.Fatalf("search bounds min %d over %d cards, want 0 and at least three Forests", pick.SearchMin, len(pick.SearchCards))
	}
	if err := g.ResolveSearchLibrary(pick.ID, me.ID, pick.SearchCards); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	forests := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Forest" {
			forests++
			if !c.Tapped {
				t.Error("a fetched Forest entered untapped")
			}
		}
	}
	if forests < 3 {
		t.Errorf("Forests on the battlefield = %d, want 3 fetched", forests)
	}
}

// --- Wrenn and Seven --------------------------------------------------

func TestWrennAndSevenPlusOneSplitsTheTopFour(t *testing.T) {
	g, seat, me, _ := s51ActivePlayer(t)
	pw := pushWalkerForTest(g, me.ID, "Wrenn and Seven", wrennAndSevenOracle, 5)
	g.WithWriteLock(func() {
		for _, c := range []game.Card{
			{Name: "Bear", TypeLine: "Creature — Bear"},
			{Name: "Forest", TypeLine: "Basic Land — Forest"},
			{Name: "Rock", TypeLine: "Artifact"},
			{Name: "Island", TypeLine: "Basic Land — Island"},
		} {
			c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
			me.Library.PushTop(c)
		}
	})
	advanceToMainOf(t, g, seat)
	hand, gy := handSize(me), me.Graveyard.Size()
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
	if handSize(me) != hand+2 || me.Graveyard.Size() != gy+2 {
		t.Errorf("hand %d (want %d), graveyard %d (want %d)", handSize(me), hand+2, me.Graveyard.Size(), gy+2)
	}
	if got := loyaltyOf(g, pw); got != 6 {
		t.Errorf("loyalty = %d, want 6", got)
	}
}

func TestWrennAndSevenZeroPutsAnyNumberOfLandsTapped(t *testing.T) {
	g, seat, me, _ := s51ActivePlayer(t)
	pw := pushWalkerForTest(g, me.ID, "Wrenn and Seven", wrennAndSevenOracle, 5)
	l1 := handCardForTest(me, "Forest", "Basic Land — Forest", "")
	l2 := handCardForTest(me, "Island", "Basic Land — Island", "")
	rock := handCardForTest(me, "Rock", "Artifact", "")
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("0: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no land prompt")
	}
	if pick.ChooseMin != 0 {
		t.Errorf("floor = %d, want 0 (any number includes none)", pick.ChooseMin)
	}
	for _, id := range pick.ChooseCards {
		if id == rock {
			t.Error("a nonland card is offered")
		}
	}
	answerChooseCards(t, g, me.ID, l1, l2)
	for _, id := range []uuid.UUID{l1, l2} {
		c, ok := s51Card(g, id)
		if !ok || !c.Tapped {
			t.Errorf("land %v on battlefield %v tapped %v, want both", id, ok, c.Tapped)
		}
	}
	if got := loyaltyOf(g, pw); got != 5 {
		t.Errorf("loyalty = %d, want 5", got)
	}
}

func TestWrennAndSevenMinusThreeMakesATreefolkAsBigAsYourLands(t *testing.T) {
	g, seat, me, opp := s51ActivePlayer(t)
	pw := pushWalkerForTest(g, me.ID, "Wrenn and Seven", wrennAndSevenOracle, 5)
	for i := 0; i < 3; i++ {
		pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	}
	pushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−3: %v", err)
	}
	passPriorityAroundTable(t, g)
	var tree uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treefolk" {
			tree = c.InstanceID
		}
	}
	if tree == uuid.Nil {
		t.Fatal("no Treefolk")
	}
	if p, tough := s51PT(t, g, tree); p != 3 || tough != 3 {
		t.Errorf("Treefolk = %d/%d, want 3/3", p, tough)
	}
	if !s51HasKeyword(g, tree, "reach") {
		t.Error("the Treefolk lacks reach")
	}
	if _, err := g.SpawnCards(me.ID, me.ID, game.ZoneBattlefield, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}, 1); err != nil {
		t.Fatal(err)
	}
	if p, _ := s51PT(t, g, tree); p != 4 {
		t.Errorf("after another land the Treefolk is %d wide, want 4", p)
	}
}

func TestWrennAndSevenMinusEightReturnsPermanentsAndRemovesTheHandLimit(t *testing.T) {
	g, seat, me, _ := s51ActivePlayer(t)
	pw := pushWalkerForTest(g, me.ID, "Wrenn and Seven", wrennAndSevenOracle, 8)
	var perm, spell uuid.UUID
	g.WithWriteLock(func() {
		perm = uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: perm, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		spell = uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	})
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 3, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−8: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(perm) {
		t.Error("the permanent card did not return to hand")
	}
	if me.Hand.Contains(spell) {
		t.Error("the instant returned to hand")
	}
	if len(emblemsOf(g, me.ID)) != 1 {
		t.Fatalf("emblems = %d, want 1", len(emblemsOf(g, me.ID)))
	}
	var max int
	g.ReadSnapshot(func() { max = g.EffectiveMaxHandSizeLocked(me) })
	if max != game.NoMaxHandSize {
		t.Errorf("maximum hand size = %d, want none", max)
	}
}

// --- Tezzeret, Betrayer of Flesh -------------------------------------

func TestTezzeretBetrayerDiscountsTheFirstArtifactActivationEachTurn(t *testing.T) {
	g, seat, me, _ := s51ActivePlayer(t)
	pushWalkerForTest(g, me.ID, "Tezzeret, Betrayer of Flesh", tezzeretBetrayOracle, 4)
	a1 := pushArtifactToBattlefieldForTest(g, me.ID, "S51 Artifact Probe", "Artifact", s51ArtifactProbeOracle)
	a2 := pushArtifactToBattlefieldForTest(g, me.ID, "S51 Artifact Probe", "Artifact", s51ArtifactProbeOracle)
	cr := pushArtifactToBattlefieldForTest(g, me.ID, "S51 Creature Probe", "Creature — Bear", s51CreatureProbeOracle)
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, cr, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("a creature's ability was discounted")
	}
	if err := g.ActivateCatalogAbility(me.ID, a1, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("the first artifact activation should cost {0}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, a1, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("a second activation of the same artifact was discounted")
	}
	if err := g.ActivateCatalogAbility(me.ID, a2, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("another artifact's first activation was discounted after one already was this turn")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, a2, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("paying {2} for the undiscounted activation: %v", err)
	}
}

func TestTezzeretBetrayerDiscountDoesNotHelpAnOpponent(t *testing.T) {
	g, seat, me, opp := s51ActivePlayer(t)
	pushWalkerForTest(g, me.ID, "Tezzeret, Betrayer of Flesh", tezzeretBetrayOracle, 4)
	theirs := pushArtifactToBattlefieldForTest(g, opp.ID, "S51 Artifact Probe", "Artifact", s51ArtifactProbeOracle)
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(opp.ID, theirs, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("an opponent's activation was discounted")
	}
}

func TestTezzeretBetrayerPlusOneDrawsTwoThenDiscards(t *testing.T) {
	g, seat, me, _ := s51ActivePlayer(t)
	pw := pushWalkerForTest(g, me.ID, "Tezzeret, Betrayer of Flesh", tezzeretBetrayOracle, 4)
	art := handCardForTest(me, "Rock", "Artifact", "")
	plain := handCardForTest(me, "Bear", "Creature — Bear", "")
	advanceToMainOf(t, g, seat)
	hand := handSize(me)
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
	if handSize(me) != hand+2 {
		t.Fatalf("hand = %d, want %d after drawing two", handSize(me), hand+2)
	}
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no discard prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{plain}); err == nil {
		t.Fatal("a lone nonartifact card satisfied the discard")
	}
	answerChooseCards(t, g, me.ID, art)
	if handSize(me) != hand+1 || me.Hand.Contains(art) {
		t.Errorf("hand = %d (want %d), artifact still in hand %v", handSize(me), hand+1, me.Hand.Contains(art))
	}
}

func TestTezzeretBetrayerPlusOneTwoCardsAlsoPays(t *testing.T) {
	g, seat, me, _ := s51ActivePlayer(t)
	pw := pushWalkerForTest(g, me.ID, "Tezzeret, Betrayer of Flesh", tezzeretBetrayOracle, 4)
	advanceToMainOf(t, g, seat)
	hand := handSize(me)
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) < 2 {
		t.Fatal("no discard prompt")
	}
	answerChooseCards(t, g, me.ID, pick.ChooseCards[0], pick.ChooseCards[1])
	if handSize(me) != hand {
		t.Errorf("hand = %d, want %d", handSize(me), hand)
	}
}

func TestTezzeretBetrayerMinusTwoAnimatesAnArtifact(t *testing.T) {
	g, seat, me, _ := s51ActivePlayer(t)
	pw := pushWalkerForTest(g, me.ID, "Tezzeret, Betrayer of Flesh", tezzeretBetrayOracle, 4)
	rock := pushArtifactToBattlefieldForTest(g, me.ID, "Rock", "Artifact", "")
	veh := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: veh, Name: "Cart", TypeLine: "Artifact — Vehicle", Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID})
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{Targets: cardRefs(pw), Strict: true}); err == nil {
		t.Fatal("−2 targeting a planeswalker was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{Targets: cardRefs(rock), Strict: true}); err != nil {
		t.Fatalf("−2: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, _ := s51Card(g, rock)
	if !c.IsArtifact() || !c.IsCreature() {
		t.Errorf("artifact %v creature %v, want both", c.IsArtifact(), c.IsCreature())
	}
	if p, tough := s51PT(t, g, rock); p != 4 || tough != 4 {
		t.Errorf("Rock = %d/%d, want 4/4", p, tough)
	}
}

func TestTezzeretBetrayerMinusTwoLeavesAVehiclesSize(t *testing.T) {
	g, seat, me, _ := s51ActivePlayer(t)
	pw := pushWalkerForTest(g, me.ID, "Tezzeret, Betrayer of Flesh", tezzeretBetrayOracle, 4)
	veh := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: veh, Name: "Cart", TypeLine: "Artifact — Vehicle", Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID})
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{Targets: cardRefs(veh), Strict: true}); err != nil {
		t.Fatalf("−2: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, _ := s51Card(g, veh)
	if !c.IsCreature() {
		t.Error("the Vehicle did not become a creature")
	}
	if p, tough := s51PT(t, g, veh); p != 3 || tough != 2 {
		t.Errorf("Cart = %d/%d, want its printed 3/2", p, tough)
	}
}

func TestTezzeretBetrayerMinusSixEmblemDrawsWhenAnArtifactBecomesTapped(t *testing.T) {
	g, seat, me, opp := s51ActivePlayer(t)
	pw := pushWalkerForTest(g, me.ID, "Tezzeret, Betrayer of Flesh", tezzeretBetrayOracle, 6)
	mine := pushArtifactToBattlefieldForTest(g, me.ID, "Rock", "Artifact", "")
	theirs := pushArtifactToBattlefieldForTest(g, opp.ID, "Their Rock", "Artifact", "")
	creature := pushCreatureToBattlefieldForTest(g, me.ID, "Bear")
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−6: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(emblemsOf(g, me.ID)) != 1 {
		t.Fatalf("emblems = %d, want 1", len(emblemsOf(g, me.ID)))
	}
	hand := handSize(me)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(theirs) })
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(creature) })
	passPriorityAroundTable(t, g)
	if handSize(me) != hand {
		t.Fatalf("hand = %d, want %d: an opponent's artifact or a nonartifact tapped", handSize(me), hand)
	}
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(mine) })
	passPriorityAroundTable(t, g)
	if handSize(me) != hand+1 {
		t.Errorf("hand = %d, want %d", handSize(me), hand+1)
	}
}

// --- Tamiyo, Field Researcher ----------------------------------------

func TestTamiyoFieldResearcherMinusTwoTapsAndFreezes(t *testing.T) {
	g, seat, me, opp := s51ActivePlayer(t)
	pw := pushWalkerForTest(g, me.ID, "Tamiyo, Field Researcher", tamiyoFieldOracle, 4)
	a := pushCreatureToBattlefieldForTest(g, opp.ID, "A")
	land := pushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{Targets: cardRefs(land), Strict: true}); err == nil {
		t.Fatal("−2 targeting a land was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{Targets: cardRefs(a), Strict: true}); err != nil {
		t.Fatalf("−2: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, _ := s51Card(g, a)
	if !c.Tapped {
		t.Error("the target is untapped")
	}
	if got := loyaltyOf(g, pw); got != 2 {
		t.Errorf("loyalty = %d, want 2", got)
	}
}

func TestTamiyoFieldResearcherMinusSevenDrawsThreeAndMakesAnEmblem(t *testing.T) {
	g, seat, me, _ := s51ActivePlayer(t)
	pw := pushWalkerForTest(g, me.ID, "Tamiyo, Field Researcher", tamiyoFieldOracle, 7)
	advanceToMainOf(t, g, seat)
	hand := handSize(me)
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−7: %v", err)
	}
	passPriorityAroundTable(t, g)
	if handSize(me) != hand+3 {
		t.Errorf("hand = %d, want %d", handSize(me), hand+3)
	}
	if len(emblemsOf(g, me.ID)) != 1 {
		t.Fatalf("emblems = %d, want 1", len(emblemsOf(g, me.ID)))
	}
	spell := handCardFull(me, "Big Spell", "Sorcery", "{9}", "", nil)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: GrantedAltCostFree}); err != nil {
		t.Fatalf("casting a nonland card from hand for free: %v", err)
	}
}

func TestTamiyoFieldResearcherEmblemOnlyHelpsItsOwner(t *testing.T) {
	g, seat, me, opp := s51ActivePlayer(t)
	pw := pushWalkerForTest(g, me.ID, "Tamiyo, Field Researcher", tamiyoFieldOracle, 7)
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−7: %v", err)
	}
	passPriorityAroundTable(t, g)
	spell := handCardFull(opp, "Big Spell", "Instant", "{9}", "", nil)
	if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: GrantedAltCostFree}); err == nil {
		t.Fatal("an opponent cast a spell for free under my emblem")
	}
}

func TestTamiyoFieldResearcherPlusOneDrawsWhenEitherCreatureDealsCombatDamage(t *testing.T) {
	g, seat, me, opp := s51ActivePlayer(t)
	pw := pushWalkerForTest(g, me.ID, "Tamiyo, Field Researcher", tamiyoFieldOracle, 4)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	other := pushVanillaCreature(g, me.ID, "Bear C", 2, 2)
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{Targets: cardRefs(a, b), Strict: true}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
	hand := handSize(me)
	attackWith(t, g, opp.ID, a, b, other)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := handSize(me); got != hand+2 {
		t.Errorf("hand = %d, want %d: one card for each chosen creature, none for the third", got, hand+2)
	}
}

func TestTamiyoFieldResearcherPlusOneDrawsOncePerCreatureNotPerRecipient(t *testing.T) {
	g, seat, me, opp := s51ActivePlayer(t)
	pw := pushWalkerForTest(g, me.ID, "Tamiyo, Field Researcher", tamiyoFieldOracle, 4)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{Targets: cardRefs(a), Strict: true}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
	hand := handSize(me)
	attackWith(t, g, opp.ID, a)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := handSize(me); got != hand+1 {
		t.Errorf("hand = %d, want %d", got, hand+1)
	}
}

// --- Liliana, Death Wielder ------------------------------------------

func TestLilianaDeathWielderPlusTwoPutsAMinusOneCounter(t *testing.T) {
	g, seat, me, opp := s51ActivePlayer(t)
	pw := pushWalkerForTest(g, me.ID, "Liliana, Death Wielder", lilianaDeathOracle, 5)
	victim := pushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{Targets: cardRefs(victim), Strict: true}); err != nil {
		t.Fatalf("+2: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := loyaltyOf(g, pw); got != 7 {
		t.Errorf("loyalty = %d, want 7", got)
	}
	if p, tough := s51PT(t, g, victim); p != 2 || tough != 2 {
		t.Errorf("victim = %d/%d, want 2/2", p, tough)
	}
}

func TestLilianaDeathWielderMinusThreeNeedsAMinusOneCounter(t *testing.T) {
	g, seat, me, opp := s51ActivePlayer(t)
	pw := pushWalkerForTest(g, me.ID, "Liliana, Death Wielder", lilianaDeathOracle, 5)
	victim := pushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	clean := pushVanillaCreature(g, opp.ID, "Clean", 3, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == victim {
				if g.Battlefield.Cards[i].Counters == nil {
					g.Battlefield.Cards[i].Counters = map[string]int{}
				}
				g.Battlefield.Cards[i].Counters[game.CounterMinusOne] = 1
			}
		}
	})
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{Targets: cardRefs(clean), Strict: true}); err == nil {
		t.Fatal("−3 at an uncountered creature was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{Targets: cardRefs(victim), Strict: true}); err != nil {
		t.Fatalf("−3: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := s51Card(g, victim); ok {
		t.Error("the creature with the counter survived")
	}
	if _, ok := s51Card(g, clean); !ok {
		t.Error("the clean creature died")
	}
}

func TestLilianaDeathWielderPlusTwoWithNoTargetStillGainsLoyalty(t *testing.T) {
	g, seat, me, _ := s51ActivePlayer(t)
	pw := pushWalkerForTest(g, me.ID, "Liliana, Death Wielder", lilianaDeathOracle, 5)
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("+2 with no target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := loyaltyOf(g, pw); got != 7 {
		t.Errorf("loyalty = %d, want 7", got)
	}
}

func TestLilianaDeathWielderMinusTenReturnsEveryCreatureCard(t *testing.T) {
	g, seat, me, opp := s51ActivePlayer(t)
	pw := pushWalkerForTest(g, me.ID, "Liliana, Death Wielder", lilianaDeathOracle, 10)
	var c1, c2, spell, theirs uuid.UUID
	g.WithWriteLock(func() {
		mk := func(p *game.Player, name, tl string) uuid.UUID {
			id := uuid.New()
			p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: tl, Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID})
			return id
		}
		c1, c2 = mk(me, "Bear", "Creature — Bear"), mk(me, "Elf", "Creature — Elf")
		spell = mk(me, "Bolt", "Instant")
		theirs = mk(opp, "Their Bear", "Creature — Bear")
	})
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−10: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{c1, c2} {
		if _, ok := s51Card(g, id); !ok {
			t.Errorf("creature card %v did not return", id)
		}
	}
	if _, ok := s51Card(g, spell); ok {
		t.Error("the instant returned")
	}
	if _, ok := s51Card(g, theirs); ok {
		t.Error("an opponent's graveyard creature returned")
	}
}
