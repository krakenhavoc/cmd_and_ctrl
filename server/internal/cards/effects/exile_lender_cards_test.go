package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_lender_cards_test.go — ADR 0155 PR 3 (#3201): the borrowers whose
// lenders are in exile. A resolving ability's exile now stamps
// Card.ExiledWith (ExileTarget.Linked), and the "exiled with" borrowers
// read that link. Each test asserts what a player gets: which permanent
// has which ability, and what using it does.

const (
	elImpostorOracle = "96e25887-e556-423a-8a69-0cd148216f6a"
	elCrawlerOracle  = "74cbe86b-cc8c-4a5f-93c3-f4eaced50393"
	elWelderOracle   = "ea2cbf5a-2def-450b-9db8-40a5b6a604d0"
	elCauldronOracle = "c259e16f-2a44-4552-8678-815f757a02e8"
	elStewardOracle  = "1bb7a942-c4b9-47a2-9782-3ec3efcdaff5"
	elForgeOracle    = "a14d89b6-23eb-4337-8dee-93c39b0efbf1"
	elRexOracle      = "7ad52f9c-48d8-486f-90b7-73ce5cb3b9db"
	elMairsilOracle  = "ed5fbae0-2dca-44a5-84dc-62674b3a92a6"
	elJewelOracle    = "c0d2bdd9-1e68-4d3a-83f8-ec2ba440a494"
)

// elRef is the source OBJECT a linked exile names.
func elRef(t *testing.T, g *game.Game, id uuid.UUID) game.PermissionCardRef {
	t.Helper()
	c := gaLayered(t, g, id)
	return game.PermissionCardRef{ID: id, Epoch: c.ObjectEpoch}
}

// elExiledCard is the card `id` as it sits in exile.
func elExiledCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			return c
		}
	}
	t.Fatalf("%s is not in exile", id)
	return game.Card{}
}

// elPushExiled puts a card straight into exile, linked to `by` (the zero
// ref for an unlinked card), and drops the layer cache.
func elPushExiled(g *game.Game, owner uuid.UUID, by game.PermissionCardRef, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: owner, Controller: owner, ExiledWith: by})
	g.BumpLayerVersionForTest()
	return id
}

// elActivate activates a permanent's own ability `index` at one card
// target, then resolves the stack.
func elActivate(t *testing.T, g *game.Game, p *game.Player, card uuid.UUID, index int, target uuid.UUID) {
	t.Helper()
	params := game.ActivateAbilityParams{}
	if target != uuid.Nil {
		params.Targets = []game.TargetRef{{Kind: game.TargetCard, ID: target}}
	}
	if err := g.ActivateCatalogAbility(p.ID, card, index, params); err != nil {
		t.Fatalf("activate ability %d: %v", index, err)
	}
	passPriorityAroundTable(t, g)
}

// elToHand moves a card out of exile to its owner's hand, the move a
// DependsOnExile static must notice.
func elToHand(t *testing.T, g *game.Game, owner, id uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneExile}, game.ZoneRef{Kind: game.ZoneHand, Owner: owner}, id); err != nil {
		t.Fatalf("move out of exile: %v", err)
	}
}

func TestExileLenderDarkImpostorTakesTheAbilitiesOfTheCreatureItExiled(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	impostor := gaPut(g, me.ID, "Dark Impostor", "Creature — Vampire Assassin", elImpostorOracle)
	elves := blElves(g, opp)
	// A creature card exiled by something else lends nothing.
	elPushExiled(g, opp.ID, game.PermissionCardRef{}, "Llanowar Elves", "Creature — Elf Druid", blLlanowarOracle)
	if idx, _ := blBorrowedMana(t, g, impostor); len(idx) != 0 {
		t.Fatalf("an unlinked exiled creature lent %d rows", len(idx))
	}
	b06AddMana(me, "B", "B", "C", "C", "C", "C")
	elActivate(t, g, me, impostor, 0, elves)
	if got := elExiledCard(t, g, elves).ExiledWith; got != elRef(t, g, impostor) {
		t.Fatalf("the exiled Elves are linked to %+v, want the Impostor object", got)
	}
	if n := gaLayered(t, g, impostor).Counters[game.CounterPlusOne]; n != 1 {
		t.Errorf("the Impostor has %d +1/+1 counters, want 1", n)
	}
	if idx, _ := blBorrowedMana(t, g, impostor); len(idx) != 1 {
		t.Fatalf("the Impostor has %d lent mana rows, want the Elves' 1", len(idx))
	}
	blTapLentMana(t, g, me, impostor)
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool = %v, want [G]", got)
	}
	// The Elves leave exile: the row goes with them.
	elToHand(t, g, opp.ID, elves)
	if idx, _ := blBorrowedMana(t, g, impostor); len(idx) != 0 {
		t.Errorf("the Elves left exile and %d rows are still lent", len(idx))
	}
}

// A linked exile from a hand is a move nothing else re-reads the layers
// for; the borrower's DependsOnExile does (the arrival bump), with no
// test hook dropping the cache.
func TestExileLenderALinkedExileFromAHandLendsAtOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	impostor := gaPut(g, me.ID, "Dark Impostor", "Creature — Vampire Assassin", elImpostorOracle)
	if idx, _ := blBorrowedMana(t, g, impostor); len(idx) != 0 {
		t.Fatalf("the Impostor alone has %d lent rows", len(idx))
	}
	elves := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: elves, Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid",
		OracleID: blLlanowarOracle, Owner: me.ID, Controller: me.ID})
	ref := elRef(t, g, impostor)
	var err error
	g.WithWriteLock(func() { err = g.ExileLinkedForEffect(game.ObjectRef(ref), elves) })
	if err != nil {
		t.Fatalf("ExileLinkedForEffect: %v", err)
	}
	if got := elExiledCard(t, g, elves).ExiledWith; got != ref {
		t.Fatalf("the card is linked to %+v, want %+v", got, ref)
	}
	if idx, _ := blBorrowedMana(t, g, impostor); len(idx) != 1 {
		t.Errorf("the Impostor has %d lent rows after a linked exile from a hand, want 1", len(idx))
	}
	// An unlinked exile stamps nothing.
	plain := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: plain, Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid",
		OracleID: blLlanowarOracle, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { err = g.ExileCardForEffect(plain) })
	if err != nil {
		t.Fatalf("ExileCardForEffect: %v", err)
	}
	if got := elExiledCard(t, g, plain).ExiledWith; got.ID != uuid.Nil {
		t.Errorf("a plain exile linked the card to %+v", got)
	}
}

func TestExileLenderPatchworkCrawlerExilesFromYourGraveyardOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crawler := gaPut(g, me.ID, "Patchwork Crawler", "Creature — Zombie Horror", elCrawlerOracle)
	theirs := blToGraveyard(g, opp, "Llanowar Elves", "Creature — Elf Druid", blLlanowarOracle)
	mine := blToGraveyard(g, me, "Llanowar Elves", "Creature — Elf Druid", blLlanowarOracle)
	b06AddMana(me, "U", "C", "C")
	err := g.ActivateCatalogAbility(me.ID, crawler, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	})
	if err == nil {
		t.Fatal("the Crawler targeted a creature card in an opponent's graveyard")
	}
	elActivate(t, g, me, crawler, 0, mine)
	if n := gaLayered(t, g, crawler).Counters[game.CounterPlusOne]; n != 1 {
		t.Errorf("the Crawler has %d +1/+1 counters, want 1", n)
	}
	if idx, _ := blBorrowedMana(t, g, crawler); len(idx) != 1 {
		t.Fatalf("the Crawler has %d lent rows, want 1", len(idx))
	}
}

func TestExileLenderMyrWelderLendsEveryCardItExiled(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	welder := gaPut(g, me.ID, "Myr Welder", "Artifact Creature — Myr", elWelderOracle)
	ring := blToGraveyard(g, opp, "Sol Ring", "Artifact", b15SolRingOracle)
	elActivate(t, g, me, welder, 0, ring)
	if !gaLayered(t, g, welder).Tapped {
		t.Fatal("the imprint {T} did not tap the Welder")
	}
	idx, origins := blBorrowedMana(t, g, welder)
	if len(idx) != 1 {
		t.Fatalf("the Welder has %d lent mana rows, want Sol Ring's 1", len(idx))
	}
	// A lent {T} needs the Welder untapped, as printed.
	if err := g.ActivateManaAbility(me.ID, welder, idx[0], game.ManaAbilityParams{Ref: origins[0].Ref}); err == nil {
		t.Error("the tapped Welder activated a lent {T} ability")
	}
}

func TestExileLenderHazelsBrewmasterLendsToYourFoods(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brewmaster := gaPut(g, me.ID, "Hazel's Brewmaster", "Creature — Squirrel Warlock", b21HazelsBrewmasterOracle)
	food := FoodToken()
	food.InstanceID, food.Owner, food.Controller = uuid.New(), me.ID, me.ID
	foodID := pushBattlefieldCardWithTimestamp(g, food)
	theirFood := FoodToken()
	theirFood.InstanceID, theirFood.Owner, theirFood.Controller = uuid.New(), opp.ID, opp.ID
	theirFoodID := pushBattlefieldCardWithTimestamp(g, theirFood)

	elPushExiled(g, opp.ID, elRef(t, g, brewmaster), "Llanowar Elves", "Creature — Elf Druid", blLlanowarOracle)
	elPushExiled(g, opp.ID, elRef(t, g, brewmaster), "Sol Ring", "Artifact", b15SolRingOracle)
	if idx, _ := blBorrowedMana(t, g, foodID); len(idx) != 1 {
		t.Fatalf("your Food has %d lent mana rows, want the Elves' 1 (Sol Ring is not a creature card)", len(idx))
	}
	if idx, _ := blBorrowedMana(t, g, theirFoodID); len(idx) != 0 {
		t.Errorf("an opponent's Food borrowed %d rows", len(idx))
	}
	if idx, _ := blBorrowedMana(t, g, brewmaster); len(idx) != 0 {
		t.Errorf("the Brewmaster itself borrowed %d rows", len(idx))
	}
	// The Food keeps its own ability beside the lent one.
	if own := game.ActivatedAbilitiesForCard(gaLayered(t, g, foodID)); len(own) != 1 {
		t.Errorf("the Food has %d activated abilities, want its own 1", len(own))
	}
	blTapLentMana(t, g, me, foodID)
	if !gaLayered(t, g, foodID).Tapped {
		t.Error("the lent {T} did not tap the Food")
	}

	// An uncatalogued creature card exiled with it: the Food names it.
	bear := elPushExiled(g, opp.ID, elRef(t, g, brewmaster), "Mystery Bear", "Creature — Bear", "")
	var lenders []*game.Card
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(foodID)
		lenders = g.UnautomatedLendersOf(&c)
	})
	if len(lenders) != 1 || lenders[0].InstanceID != bear {
		t.Errorf("the Food's not-automated lenders = %d, want the Mystery Bear", len(lenders))
	}
}

func TestExileLenderHazelsBrewmasterTriggerLinksItsExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elves := blToGraveyard(g, opp, "Llanowar Elves", "Creature — Elf Druid", blLlanowarOracle)
	brewmaster := castAndResolveCreature(t, g, "Hazel's Brewmaster", "Creature — Squirrel Warlock", b21HazelsBrewmasterOracle)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, elves)
	passPriorityAroundTable(t, g)
	if got := elExiledCard(t, g, elves).ExiledWith; got != elRef(t, g, brewmaster) {
		t.Fatalf("the trigger's exile is linked to %+v, want the Brewmaster object", got)
	}
	var food uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Food" {
			food = c.InstanceID
		}
	}
	if idx, _ := blBorrowedMana(t, g, food); len(idx) != 1 {
		t.Errorf("the new Food has %d lent mana rows, want the Elves' 1", len(idx))
	}
}

func TestExileLenderAgathasSoulCauldronLendsToCreaturesWithCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cauldron := gaPut(g, me.ID, "Agatha's Soul Cauldron", "Legendary Artifact", elCauldronOracle)
	bear := gaPut(g, me.ID, "Bear", "Creature — Bear", "")
	otherBear := gaPut(g, me.ID, "Other Bear", "Creature — Bear", "")
	theirBear := gaPut(g, opp.ID, "Their Bear", "Creature — Bear", "")
	elves := blToGraveyard(g, opp, "Llanowar Elves", "Creature — Elf Druid", blLlanowarOracle)

	if err := g.ActivateCatalogAbility(me.ID, cauldron, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: elves}},
	}); err != nil {
		t.Fatalf("activate the Cauldron: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The reflexive trigger asks for a creature you control.
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if hasID(p.PickTargetCards, theirBear) {
		t.Error("the counter can target an opponent's creature")
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if n := gaLayered(t, g, bear).Counters[game.CounterPlusOne]; n != 1 {
		t.Fatalf("the Bear has %d +1/+1 counters, want 1", n)
	}
	if idx, _ := blBorrowedMana(t, g, bear); len(idx) != 1 {
		t.Errorf("the Bear with a counter has %d lent rows, want the Elves' 1", len(idx))
	}
	if idx, _ := blBorrowedMana(t, g, otherBear); len(idx) != 0 {
		t.Errorf("a creature with no counter borrowed %d rows", len(idx))
	}
	// The spend rider: any ability of a creature you control.
	card := gaLayered(t, g, bear)
	theirs := gaLayered(t, g, theirBear)
	var mineCovered, theirsCovered bool
	g.ReadSnapshot(func() {
		mineCovered = g.SpendsManaAsAnyColorForEffect(me.ID, game.ManaSpendForAbility(card))
		theirsCovered = g.SpendsManaAsAnyColorForEffect(me.ID, game.ManaSpendForAbility(theirs))
	})
	if !mineCovered {
		t.Error("the rider does not cover an ability of your creature")
	}
	if theirsCovered {
		t.Error("the rider covers an ability of an opponent's creature")
	}
	// Losing the counter loses the rows.
	if err := g.AddCounter(bear, game.CounterPlusOne, -1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if idx, _ := blBorrowedMana(t, g, bear); len(idx) != 0 {
		t.Errorf("the Bear lost its counter and still has %d lent rows", len(idx))
	}
}

func TestExileLenderAgathasSoulCauldronNoncreatureMakesNoTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cauldron := gaPut(g, me.ID, "Agatha's Soul Cauldron", "Legendary Artifact", elCauldronOracle)
	gaPut(g, me.ID, "Bear", "Creature — Bear", "")
	ring := blToGraveyard(g, opp, "Sol Ring", "Artifact", b15SolRingOracle)
	elActivate(t, g, me, cauldron, 0, ring)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a noncreature card exiled this way made the counter trigger")
	}
	if !inExile(g, ring) {
		t.Error("the card is not exiled")
	}
}

func TestExileLenderStewardOfTheHarvestGivesYourCreaturesTheLandsAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := blToGraveyard(g, me, "Forest", "Basic Land — Forest", "")
	island := blToGraveyard(g, me, "Island", "Basic Land — Island", "")
	theirs := blToGraveyard(g, opp, "Mountain", "Basic Land — Mountain", "")
	bear := gaPut(g, me.ID, "Bear", "Creature — Bear", "")
	theirBear := gaPut(g, opp.ID, "Their Bear", "Creature — Bear", "")
	steward := castAndResolveCreature(t, g, "Steward of the Harvest", "Creature — Human Druid", elStewardOracle)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if hasID(p.PickTargetCards, theirs) {
		t.Error("the Steward can exile a land from an opponent's graveyard")
	}
	if p.PickTargetMax != 3 || p.PickTargetMin != 0 {
		t.Errorf("pick bounds %d..%d, want up to three", p.PickTargetMin, p.PickTargetMax)
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: forest}, {Kind: game.TargetCard, ID: island},
	}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, steward} {
		if idx, _ := blBorrowedMana(t, g, id); len(idx) != 2 {
			t.Errorf("%s has %d lent mana rows, want the Forest's and the Island's", gaLayered(t, g, id).Name, len(idx))
		}
	}
	if idx, _ := blBorrowedMana(t, g, theirBear); len(idx) != 0 {
		t.Errorf("an opponent's creature borrowed %d rows", len(idx))
	}
	blTapLentMana(t, g, me, bear)
	if len(batch01PoolColors(me)) != 1 {
		t.Errorf("the lent land ability made %v", batch01PoolColors(me))
	}
}

func TestExileLenderTerritoryForgeTriggersOnlyWhenCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ring := gaPut(g, opp.ID, "Sol Ring", "Artifact", b15SolRingOracle)
	forge := castAndResolveCreature(t, g, "Territory Forge", "Artifact", elForgeOracle)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, ring)
	passPriorityAroundTable(t, g)
	if got := elExiledCard(t, g, ring).ExiledWith; got != elRef(t, g, forge) {
		t.Fatalf("the exiled Sol Ring is linked to %+v, want the Forge object", got)
	}
	if idx, _ := blBorrowedMana(t, g, forge); len(idx) != 1 {
		t.Fatalf("the Forge has %d lent mana rows, want Sol Ring's 1", len(idx))
	}
	blTapLentMana(t, g, me, forge)
	if got := batch01PoolColors(me); len(got) != 2 {
		t.Errorf("the lent Sol Ring made %v, want two colourless", got)
	}
	// Put onto the battlefield without being cast: no trigger.
	before := g.Exile.Size()
	other := gaPut(g, me.ID, "Territory Forge", "Artifact", elForgeOracle)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || g.Exile.Size() != before {
		t.Error("a Forge that was not cast triggered")
	}
	if idx, _ := blBorrowedMana(t, g, other); len(idx) != 0 {
		t.Errorf("the second Forge borrowed the first one's card: %d rows", len(idx))
	}
}

func TestExileLenderRexBorrowsFromEveryCardWithABrainCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rex := gaPut(g, me.ID, "Rex, Cyber-Hound", "Legendary Artifact Creature — Robot Dog", elRexOracle)
	elves := blToGraveyard(g, opp, "Llanowar Elves", "Creature — Elf Druid", blLlanowarOracle)
	g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(me.ID, game.CounterEnergy, 2) })
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	elActivate(t, g, me, rex, 0, elves)
	if n := me.Counters[game.CounterEnergy]; n != 0 {
		t.Errorf("energy left %d, want 0", n)
	}
	if n := elExiledCard(t, g, elves).Counters[rexBrainCounter]; n != 1 {
		t.Fatalf("the exiled Elves have %d brain counters, want 1", n)
	}
	if idx, _ := blBorrowedMana(t, g, rex); len(idx) != 1 {
		t.Fatalf("Rex has %d lent rows, want the Elves' 1", len(idx))
	}
	// Any card in exile with a brain counter lends, whoever exiled it;
	// one without lends nothing.
	other := elPushExiled(g, opp.ID, game.PermissionCardRef{}, "Llanowar Elves", "Creature — Elf Druid", blLlanowarOracle)
	if idx, _ := blBorrowedMana(t, g, rex); len(idx) != 1 {
		t.Fatalf("a card with no brain counter lent: %d rows", len(idx))
	}
	if err := g.AddCounter(other, rexBrainCounter, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if idx, _ := blBorrowedMana(t, g, rex); len(idx) != 2 {
		t.Errorf("a brain counter put on an exiled card: Rex has %d rows, want 2", len(idx))
	}
}

func TestExileLenderMairsilCagesFromYourHandOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elves := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: elves, Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid",
		OracleID: blLlanowarOracle, Owner: me.ID, Controller: me.ID})
	land := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	mairsil := castAndResolveCreature(t, g, "Mairsil, the Pretender", "Legendary Creature — Human Wizard", elMairsilOracle)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no cage prompt")
	}
	if hasID(pick.ChooseCards, land) {
		t.Error("a land is offered for the cage")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Errorf("cage bounds %d..%d, want at most one, optional", pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{elves}); err != nil {
		t.Fatalf("cage: %v", err)
	}
	if n := elExiledCard(t, g, elves).Counters[mairsilCageCounter]; n != 1 {
		t.Fatalf("the caged Elves have %d cage counters, want 1", n)
	}
	// Arriving from a hand re-read the layers.
	idx, origins := blBorrowedMana(t, g, mairsil)
	if len(idx) != 1 {
		t.Fatalf("Mairsil has %d lent rows, want the Elves' 1", len(idx))
	}
	// Mairsil entered this turn: give it haste-free tapping by clearing
	// sickness, then use the row once; a second time is refused.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == mairsil {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	if err := g.ActivateManaAbility(me.ID, mairsil, idx[0], game.ManaAbilityParams{Ref: origins[0].Ref}); err != nil {
		t.Fatalf("activate the caged row: %v", err)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == mairsil {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	if err := g.ActivateManaAbility(me.ID, mairsil, idx[0], game.ManaAbilityParams{Ref: origins[0].Ref}); err == nil {
		t.Error("the caged row was activated twice in one turn")
	}
	// A card an opponent owns with a cage counter lends nothing.
	opp := g.Seats[1]
	theirs := elPushExiled(g, opp.ID, game.PermissionCardRef{}, "Llanowar Elves", "Creature — Elf Druid", blLlanowarOracle)
	if err := g.AddCounter(theirs, mairsilCageCounter, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if idx, _ := blBorrowedMana(t, g, mairsil); len(idx) != 1 {
		t.Errorf("an opponent's caged card lent: %d rows", len(idx))
	}
}

func TestExileLenderLocusOfEnlightenmentBorrowsItsCraftMaterials(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	locus := gaPut(g, me.ID, "Locus of Enlightenment", "Legendary Artifact", elJewelOracle+"#1")
	elves := elPushExiled(g, me.ID, game.PermissionCardRef{}, "Llanowar Elves", "Creature — Elf Druid", blLlanowarOracle)
	ring := elPushExiled(g, me.ID, game.PermissionCardRef{}, "Sol Ring", "Artifact", b15SolRingOracle)
	elPushExiled(g, me.ID, game.PermissionCardRef{}, "Not A Material", "Creature — Elf Druid", blLlanowarOracle)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == locus {
				var refs []game.ObjectRef
				for _, c := range g.Exile.Cards {
					if c.InstanceID == elves || c.InstanceID == ring {
						refs = append(refs, game.ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch})
					}
				}
				g.Battlefield.Cards[i].CraftedWith = refs
			}
		}
	})
	g.BumpLayerVersionForTest()
	idx, origins := blBorrowedMana(t, g, locus)
	if len(idx) != 2 {
		t.Fatalf("the Locus has %d lent mana rows, want its two materials'", len(idx))
	}
	if err := g.ActivateManaAbility(me.ID, locus, idx[1], game.ManaAbilityParams{Ref: origins[1].Ref}); err != nil {
		t.Fatalf("activate a material's row: %v", err)
	}
	if !gaLayered(t, g, locus).Tapped {
		t.Error("the lent {T} did not tap the Locus")
	}
	// A material that leaves exile stops lending.
	elToHand(t, g, me.ID, elves)
	if idx, _ := blBorrowedMana(t, g, locus); len(idx) != 1 {
		t.Errorf("a material left exile and the Locus has %d rows, want 1", len(idx))
	}
}

func TestExileLenderLocusCopiesEachNonManaAbilityYouActivate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Locus of Enlightenment", "Legendary Artifact", elJewelOracle+"#1", false)
	monolith := pushCatalogPermanent(g, me.ID, "Basalt Monolith", "Artifact", acBasaltMonolithOracle, false)
	floatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, monolith, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if acItemOnStack(g, "Locus of Enlightenment") == nil {
		t.Fatal("activating an ability did not trigger the Locus")
	}
	// The trigger resolves first and copies the untap; no payment, no
	// "may".
	for i := 0; i < 16 && acItemOnStack(g, "Locus of Enlightenment") != nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if n := acCopiesOnStack(g); n != 1 {
		t.Fatalf("the Locus made %d copies, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if n := acCountEvents(g, game.EventActivateAbility); n != 1 {
		t.Errorf("%d activations announced, want 1: a copy is created, not activated (CR 707.10a)", n)
	}
	// A mana ability is never copied (CR 605.3b).
	elf := blElves(g, me)
	if err := g.ActivateManaAbility(me.ID, elf, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Elves: %v", err)
	}
	if acItemOnStack(g, "Locus of Enlightenment") != nil {
		t.Error("a mana ability triggered the Locus")
	}
}

func TestExileLenderEnigmaJewelCraftsWithNonlandsWithActivatedAbilities(t *testing.T) {
	cost := CraftWithNonlandsWithActivatedAbilities(4).ExilePermanents
	elves := game.Card{Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid", OracleID: blLlanowarOracle}
	vanilla := game.Card{Name: "Bear", TypeLine: "Creature — Bear"}
	land := game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}
	g := newCatalogGame(t)
	me := g.Seats[0]
	jewel := gaPut(g, me.ID, "The Enigma Jewel", "Legendary Artifact", elJewelOracle)
	for _, c := range []game.Card{elves, elves, elves, vanilla, land} {
		c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
		me.Graveyard.PushTop(c)
	}
	var opts []uuid.UUID
	g.ReadSnapshot(func() { opts = g.ExilePermanentsOptionsForEffect(me.ID, jewel, cost) })
	if len(opts) != 3 {
		t.Fatalf("%d materials offered, want the three Elves (not the vanilla Bear or the land)", len(opts))
	}
	var payable bool
	g.ReadSnapshot(func() { payable = g.ExilePermanentsPayable(me.ID, jewel, cost) })
	if payable {
		t.Error("three materials paid a four-or-more craft")
	}
	extra := game.Card{InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact", OracleID: b15SolRingOracle, Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(extra)
	g.ReadSnapshot(func() { payable = g.ExilePermanentsPayable(me.ID, jewel, cost) })
	if !payable {
		t.Error("four nonlands with activated abilities did not pay the craft")
	}
}
