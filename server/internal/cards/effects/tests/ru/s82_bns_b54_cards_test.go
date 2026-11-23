package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82Bns54BondOracle      = "e9c9e730-3b73-48e3-845b-1b7f2d16050c"
	s82Bns54BountyOracle    = "e4bfe005-7c4e-4178-ae41-a28eb2352ad1"
	s82Bns54ConcentrateOr   = "a3d287ec-1a39-4f91-acf1-7c1d2c00ed89"
	s82Bns54ContractOracle  = "a7a0e473-77c1-4a85-b6d7-b0dfdadb696f"
	s82Bns54DragUnderOracle = "021d8596-782a-4cfd-8b70-ac5bb8a5aff1"
	s82Bns54ReprisalOracle  = "38e1ab26-95f9-43dd-9bda-13c7d8ba47ea"
	s82Bns54GiantOracle     = "395aa62e-0786-4bd8-a88f-998e4700742e"
	s82Bns54GargantuanOr    = "34f9c2a9-ca7c-4a02-aec5-e8f407d09580"
	s82Bns54GirdOracle      = "bc72d463-8818-4c19-bd5a-f47841d1adb0"
	s82Bns54GraveOracle     = "feb2f08e-79c7-4d51-bf55-e649f5ff5e07"
	s82Bns54GustOracle      = "7cfc14ce-9940-4c61-9daa-2c6dbf1a80ad"
)

func TestS82Bns54EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82Bns54BondOracle:      "Bond of Flourishing",
		s82Bns54BountyOracle:    "Bounty of Skemfar",
		s82Bns54ConcentrateOr:   "Concentrate",
		s82Bns54ContractOracle:  "Contract Killing",
		s82Bns54DragUnderOracle: "Drag Under",
		s82Bns54ReprisalOracle:  "Furious Reprisal",
		s82Bns54GiantOracle:     "Giant Opportunity",
		s82Bns54GargantuanOr:    "Gift of the Gargantuan",
		s82Bns54GirdOracle:      "Gird for Battle",
		s82Bns54GraveOracle:     "Grave Strength",
		s82Bns54GustOracle:      "Gust of Wind",
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

// s82Bns54Strict casts a sorcery from the active seat's hand with the strict
// legality checks and returns the announce error.
func s82Bns54Strict(t *testing.T, g *game.Game, name, oracle string, targets []game.TargetRef) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.HandCardForTest(active, name, "Sorcery", oracle)
	effectstest.ToMainForCost(t, g)
	return g.CastSpell(active.ID, id, game.CastSpellParams{Targets: targets, Strict: true})
}

func s82Bns54Push(p *game.Player, owner uuid.UUID, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: owner, Controller: owner})
	return id
}

// s82Bns54OrderTheRest answers the open "rest on the bottom in any order"
// prompt as the cards were offered.
func s82Bns54OrderTheRest(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	order := effectstest.PutInLibraryChoiceFor(g, p.ID)
	if order == nil {
		t.Fatal("no order prompt for the rest on the bottom")
	}
	if err := g.ResolvePutInLibrary(order.ID, p.ID, order.ScryCards, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
}

func TestS82Bns54ConcentrateDrawsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	before := effectstest.FdnB1HandSize(g, me)
	effectstest.CastCatalogSpell(t, g, "Concentrate", "Sorcery", s82Bns54ConcentrateOr, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// The spell itself left the hand when cast: net +3 cards.
	if got := effectstest.FdnB1HandSize(g, me); got != before+3 {
		t.Errorf("hand %d -> %d, want +3", before, got)
	}
}

func TestS82Bns54DragUnderBouncesAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B")
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	before := effectstest.FdnB1HandSize(g, me)
	effectstest.CastCatalogSpell(t, g, "Drag Under", "Sorcery", s82Bns54DragUnderOracle, effectstest.CardRefs(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) || !opp.Hand.Contains(ogre) {
		t.Error("the creature was not returned to its owner's hand")
	}
	if got := effectstest.FdnB1HandSize(g, me); got != before+1 {
		t.Errorf("hand %d -> %d, want +1 (the spell cast, one card drawn)", before, got)
	}
}

func TestS82Bns54DragUnderRefusesANoncreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	if err := s82Bns54Strict(t, g, "Drag Under", s82Bns54DragUnderOracle, effectstest.CardRefs(rock)); err == nil {
		t.Fatal("cast at an artifact that is not a creature")
	}
}

func TestS82Bns54GustOfWindCostsTwoLessWithAFlyerAndBouncesOnlyTheirPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	if got := effectstest.SelfPricedMV(t, g, me, s82Bns54GustOracle, "Sorcery", "{3}{U}"); got != 4 {
		t.Errorf("with no flyer: %d, want 4", got)
	}
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Walker", TypeLine: "Creature — Bird", Power: 1, Toughness: 1})
	if got := effectstest.SelfPricedMV(t, g, me, s82Bns54GustOracle, "Sorcery", "{3}{U}"); got != 4 {
		t.Errorf("with a creature without flying: %d, want 4", got)
	}
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	if got := effectstest.SelfPricedMV(t, g, me, s82Bns54GustOracle, "Sorcery", "{3}{U}"); got != 2 {
		t.Errorf("with a flyer: %d, want 2", got)
	}

	mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	if err := s82Bns54Strict(t, g, "Gust of Wind", s82Bns54GustOracle, effectstest.CardRefs(mine)); err == nil {
		t.Fatal("cast at a permanent I control")
	}
	land := effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	if err := s82Bns54Strict(t, g, "Gust of Wind", s82Bns54GustOracle, effectstest.CardRefs(land)); err == nil {
		t.Fatal("cast at a land")
	}
}

func TestS82Bns54GustOfWindBouncesAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B")
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	before := effectstest.FdnB1HandSize(g, me)
	effectstest.CastCatalogSpell(t, g, "Gust of Wind", "Sorcery", s82Bns54GustOracle, effectstest.CardRefs(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Hand.Contains(rock) {
		t.Error("the artifact was not returned to its owner's hand")
	}
	if got := effectstest.FdnB1HandSize(g, me); got != before+1 {
		t.Errorf("hand %d -> %d, want +1 (the spell cast, one card drawn)", before, got)
	}
}

func TestS82Bns54ContractKillingDestroysAndMakesTwoTreasures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Contract Killing", "Sorcery", s82Bns54ContractOracle, effectstest.CardRefs(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("the creature survived")
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Treasure"); got != 2 {
		t.Errorf("the caster has %d Treasures, want 2", got)
	}
}

func TestS82Bns54ContractKillingStillMakesTreasuresWhenTheCreatureSurvives(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Their Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4, Keywords: []string{"indestructible"}})
	effectstest.CastCatalogSpell(t, g, "Contract Killing", "Sorcery", s82Bns54ContractOracle, effectstest.CardRefs(wall))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wall) {
		t.Error("the indestructible creature was destroyed")
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Treasure"); got != 2 {
		t.Errorf("the caster has %d Treasures, want 2", got)
	}
}

func TestS82Bns54GirdForBattlePutsACounterOnEachOfUpToTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Gird for Battle", "Sorcery", s82Bns54GirdOracle, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if n := effectstest.PlusOneCounters(g, id); n != 1 {
			t.Errorf("%s has %d +1/+1 counters, want 1", id, n)
		}
	}
	if n := effectstest.PlusOneCounters(g, c); n != 0 {
		t.Errorf("the untargeted creature has %d counters, want 0", n)
	}
	d := effectstest.B12Creature(g, me.ID, "Bear D", "Creature — Bear", 2, 2)
	if err := s82Bns54Strict(t, g, "Gird for Battle", s82Bns54GirdOracle, effectstest.CardRefs(a, b, c)); err == nil {
		t.Error("cast at three creatures")
	}
	_ = d
}

func TestS82Bns54FuriousReprisalDealsTwoToEachOfTwoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := opp.Life
	refs := []game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetPlayer, ID: opp.ID}}
	effectstest.CastCatalogSpell(t, g, "Furious Reprisal", "Sorcery", s82Bns54ReprisalOracle, refs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the 2/2 survived 2 damage")
	}
	if opp.Life != life-2 {
		t.Errorf("life %d -> %d, want -2", life, opp.Life)
	}
}

func TestS82Bns54FuriousReprisalNeedsTwoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	if err := s82Bns54Strict(t, g, "Furious Reprisal", s82Bns54ReprisalOracle, effectstest.PlayerRefs(opp.ID)); err == nil {
		t.Fatal("cast at a single target")
	}
}

func TestS82Bns54GraveStrengthMillsThenCountsCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Old Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	})
	s82Bns54Push(me, me.ID, "Milled Land", "Basic Land — Forest")
	s82Bns54Push(me, me.ID, "Milled Bear", "Creature — Bear")
	s82Bns54Push(me, me.ID, "Milled Elf", "Creature — Elf")
	effectstest.CastCatalogSpell(t, g, "Grave Strength", "Sorcery", s82Bns54GraveOracle, effectstest.CardRefs(target))
	effectstest.PassPriorityAroundTable(t, g)
	// Three creature cards: the one already there and two milled.
	if n := effectstest.PlusOneCounters(g, target); n != 3 {
		t.Errorf("target has %d +1/+1 counters, want 3", n)
	}
}

func TestS82Bns54GraveStrengthWithNoTargetDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 1, 1)
	s82Bns54Push(me, me.ID, "Milled Bear", "Creature — Bear")
	s82Bns54Push(me, me.ID, "Milled Land", "Basic Land — Forest")
	s82Bns54Push(me, me.ID, "Milled Elf", "Creature — Elf")
	library := me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Grave Strength", "Sorcery", s82Bns54GraveOracle, effectstest.CardRefs(target))
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, target); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != library {
		t.Errorf("library %d -> %d: the spell milled with its only target gone", library, me.Library.Size())
	}
}

func TestS82Bns54BondOfFlourishingTakesAPermanentAndGainsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	inst := s82Bns54Push(me, me.ID, "Shock", "Instant")
	rock := s82Bns54Push(me, me.ID, "Rock", "Artifact")
	bear := s82Bns54Push(me, me.ID, "Bear", "Creature — Bear")
	_ = inst
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Bond of Flourishing", "Sorcery", s82Bns54BondOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no pick opened")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	// Library top-down: Bear, Rock, Shock. Only the permanent cards are on offer.
	if len(offered) != 2 || !offered[bear] || !offered[rock] || pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("offered %v (min %d max %d), want the two permanent cards, up to one", pick.ChooseCards, pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{rock}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(rock) {
		t.Error("the permanent card is not in hand")
	}
	s82Bns54OrderTheRest(t, g, me)
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3", life, me.Life)
	}
}

func TestS82Bns54BondOfFlourishingGainsLifeEvenTakingNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Bond of Flourishing", "Sorcery", s82Bns54BondOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	s82Bns54OrderTheRest(t, g, me)
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3", life, me.Life)
	}
}

func TestS82Bns54GiftOfTheGargantuanTakesACreatureAndALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	s82Bns54Push(me, me.ID, "Deep Card", "Instant")
	land := s82Bns54Push(me, me.ID, "Forest", "Basic Land — Forest")
	spell := s82Bns54Push(me, me.ID, "Shock", "Instant")
	bear := s82Bns54Push(me, me.ID, "Bear", "Creature — Bear")
	elf := s82Bns54Push(me, me.ID, "Elf", "Creature — Elf")
	_ = spell
	effectstest.CastCatalogSpell(t, g, "Gift of the Gargantuan", "Sorcery", s82Bns54GargantuanOr, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no pick opened")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	// Top four are Elf, Bear, Shock, Forest.
	if len(offered) != 3 || !offered[elf] || !offered[bear] || !offered[land] {
		t.Fatalf("offered %v, want the two creature cards and the land", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bear, elf}); err == nil {
		t.Fatal("two creature cards were accepted for one creature slot")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bear, land}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(bear) || !me.Hand.Contains(land) {
		t.Error("the creature and land cards are not in hand")
	}
	if effectstest.PutInLibraryChoiceFor(g, me.ID) == nil {
		t.Error("no order prompt for the rest on the bottom")
	}
}

func TestS82Bns54BountyOfSkemfarPutsALandAndTakesAnElf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	s82Bns54Push(me, me.ID, "Seventh Card", "Instant")
	s82Bns54Push(me, me.ID, "Plain Bear", "Creature — Bear")
	elf := s82Bns54Push(me, me.ID, "Elf Scout", "Creature — Elf Scout")
	s82Bns54Push(me, me.ID, "Shock", "Instant")
	island := s82Bns54Push(me, me.ID, "Island", "Basic Land — Island")
	forest := s82Bns54Push(me, me.ID, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Bounty of Skemfar", "Sorcery", s82Bns54BountyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	land := effectstest.LatestChooseCards(g, me.ID)
	if land == nil {
		t.Fatal("no land pick opened")
	}
	if len(land.ChooseCards) != 2 || land.ChooseMin != 0 || land.ChooseMax != 1 {
		t.Fatalf("land pick offered %v (min %d max %d), want the two lands, up to one", land.ChooseCards, land.ChooseMin, land.ChooseMax)
	}
	if err := g.ResolveChooseCards(land.ID, me.ID, []uuid.UUID{forest}); err != nil {
		t.Fatalf("land pick: %v", err)
	}
	if !g.Battlefield.Contains(forest) {
		t.Fatal("the land did not enter the battlefield")
	}
	if c, _ := g.LookupCardForEffect(forest); !c.Tapped {
		t.Error("the land did not enter tapped")
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no Elf pick opened")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != elf || pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("Elf pick offered %v, want only the Elf card, up to one", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{elf}); err != nil {
		t.Fatalf("Elf pick: %v", err)
	}
	if !me.Hand.Contains(elf) {
		t.Error("the Elf card is not in hand")
	}
	if me.Library.Contains(island) == false {
		t.Error("the declined land left the library")
	}
}

func s82Bns54Foods(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(p.ID, effects.FoodToken(), n) })
}

func TestS82Bns54GiantOpportunitySacrificingTwoFoodsMakesAGiant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	s82Bns54Foods(g, me, 3)
	effectstest.CastCatalogSpell(t, g, "Giant Opportunity", "Sorcery", s82Bns54GiantOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("no sacrifice prompt opened")
	}
	if len(pick.ChooseCards) != 3 || pick.ChooseMin != 0 || pick.ChooseMax != 2 {
		t.Fatalf("offered %v (min %d max %d), want the three Foods, none or two", pick.ChooseCards, pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, pick.ChooseCards[:1]); err == nil {
		t.Fatal("a single Food was accepted")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, pick.ChooseCards[:2]); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Foods left, want 1", got)
	}
	if got := effectstest.TokensNamedControlledBy(g, "Giant", me.ID); got != 1 {
		t.Fatalf("%d Giants, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Giant" {
			eff := c.Effective()
			if eff.Power != 7 || eff.Toughness != 7 || !c.IsCreature() {
				t.Errorf("the Giant is %d/%d, creature %v", eff.Power, eff.Toughness, c.IsCreature())
			}
		}
	}
}

func TestS82Bns54GiantOpportunityDecliningMakesThreeFoods(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	s82Bns54Foods(g, me, 2)
	effectstest.CastCatalogSpell(t, g, "Giant Opportunity", "Sorcery", s82Bns54GiantOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("no sacrifice prompt opened")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Food"); got != 5 {
		t.Errorf("%d Foods, want 5", got)
	}
	if got := effectstest.TokensNamedControlledBy(g, "Giant", me.ID); got != 0 {
		t.Errorf("%d Giants, want 0", got)
	}
}

func TestS82Bns54GiantOpportunityWithoutTwoFoodsMakesThreeWithoutAsking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	s82Bns54Foods(g, me, 1)
	effectstest.CastCatalogSpell(t, g, "Giant Opportunity", "Sorcery", s82Bns54GiantOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil {
		t.Error("asked to sacrifice with only one Food")
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Food"); got != 4 {
		t.Errorf("%d Foods, want 4", got)
	}
}
