package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardslu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/lu"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestExileLenderDarkImpostorTakesTheAbilitiesOfTheCreatureItExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	impostor := effectstest.GaPut(g, me.ID, "Dark Impostor", "Creature — Vampire Assassin", effectstest.ElImpostorOracle)
	elves := effectstest.BlElves(g, opp)
	// A creature card exiled by something else lends nothing.
	effectstest.ElPushExiled(g, opp.ID, game.PermissionCardRef{}, "Llanowar Elves", "Creature — Elf Druid", effectstest.BlLlanowarOracle)
	if idx, _ := effectstest.BlBorrowedMana(t, g, impostor); len(idx) != 0 {
		t.Fatalf("an unlinked exiled creature lent %d rows", len(idx))
	}
	effectstest.B06AddMana(me, "B", "B", "C", "C", "C", "C")
	effectstest.ElActivate(t, g, me, impostor, 0, elves)
	if got := effectstest.ElExiledCard(t, g, elves).ExiledWith; got != effectstest.ElRef(t, g, impostor) {
		t.Fatalf("the exiled Elves are linked to %+v, want the Impostor object", got)
	}
	if n := effectstest.GaLayered(t, g, impostor).Counters[game.CounterPlusOne]; n != 1 {
		t.Errorf("the Impostor has %d +1/+1 counters, want 1", n)
	}
	if idx, _ := effectstest.BlBorrowedMana(t, g, impostor); len(idx) != 1 {
		t.Fatalf("the Impostor has %d lent mana rows, want the Elves' 1", len(idx))
	}
	effectstest.BlTapLentMana(t, g, me, impostor)
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool = %v, want [G]", got)
	}
	// The Elves leave exile: the row goes with them.
	effectstest.ElToHand(t, g, opp.ID, elves)
	if idx, _ := effectstest.BlBorrowedMana(t, g, impostor); len(idx) != 0 {
		t.Errorf("the Elves left exile and %d rows are still lent", len(idx))
	}
}

// A linked exile from a hand is a move nothing else re-reads the layers
// for; the borrower's DependsOnExile does (the arrival bump), with no
// test hook dropping the cache.
func TestExileLenderALinkedExileFromAHandLendsAtOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	impostor := effectstest.GaPut(g, me.ID, "Dark Impostor", "Creature — Vampire Assassin", effectstest.ElImpostorOracle)
	if idx, _ := effectstest.BlBorrowedMana(t, g, impostor); len(idx) != 0 {
		t.Fatalf("the Impostor alone has %d lent rows", len(idx))
	}
	elves := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: elves, Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid",
		OracleID: effectstest.BlLlanowarOracle, Owner: me.ID, Controller: me.ID})
	ref := effectstest.ElRef(t, g, impostor)
	var err error
	g.WithWriteLock(func() { err = g.ExileLinkedForEffect(game.ObjectRef(ref), elves) })
	if err != nil {
		t.Fatalf("ExileLinkedForEffect: %v", err)
	}
	if got := effectstest.ElExiledCard(t, g, elves).ExiledWith; got != ref {
		t.Fatalf("the card is linked to %+v, want %+v", got, ref)
	}
	if idx, _ := effectstest.BlBorrowedMana(t, g, impostor); len(idx) != 1 {
		t.Errorf("the Impostor has %d lent rows after a linked exile from a hand, want 1", len(idx))
	}
	// An unlinked exile stamps nothing.
	plain := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: plain, Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid",
		OracleID: effectstest.BlLlanowarOracle, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { err = g.ExileCardForEffect(plain) })
	if err != nil {
		t.Fatalf("ExileCardForEffect: %v", err)
	}
	if got := effectstest.ElExiledCard(t, g, plain).ExiledWith; got.ID != uuid.Nil {
		t.Errorf("a plain exile linked the card to %+v", got)
	}
}

func TestExileLenderPatchworkCrawlerExilesFromYourGraveyardOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crawler := effectstest.GaPut(g, me.ID, "Patchwork Crawler", "Creature — Zombie Horror", effectstest.ElCrawlerOracle)
	theirs := effectstest.BlToGraveyard(g, opp, "Llanowar Elves", "Creature — Elf Druid", effectstest.BlLlanowarOracle)
	mine := effectstest.BlToGraveyard(g, me, "Llanowar Elves", "Creature — Elf Druid", effectstest.BlLlanowarOracle)
	effectstest.B06AddMana(me, "U", "C", "C")
	err := g.ActivateCatalogAbility(me.ID, crawler, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	})
	if err == nil {
		t.Fatal("the Crawler targeted a creature card in an opponent's graveyard")
	}
	effectstest.ElActivate(t, g, me, crawler, 0, mine)
	if n := effectstest.GaLayered(t, g, crawler).Counters[game.CounterPlusOne]; n != 1 {
		t.Errorf("the Crawler has %d +1/+1 counters, want 1", n)
	}
	if idx, _ := effectstest.BlBorrowedMana(t, g, crawler); len(idx) != 1 {
		t.Fatalf("the Crawler has %d lent rows, want 1", len(idx))
	}
}

func TestExileLenderMyrWelderLendsEveryCardItExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	welder := effectstest.GaPut(g, me.ID, "Myr Welder", "Artifact Creature — Myr", effectstest.ElWelderOracle)
	ring := effectstest.BlToGraveyard(g, opp, "Sol Ring", "Artifact", effectstest.B15SolRingOracle)
	effectstest.ElActivate(t, g, me, welder, 0, ring)
	if !effectstest.GaLayered(t, g, welder).Tapped {
		t.Fatal("the imprint {T} did not tap the Welder")
	}
	idx, origins := effectstest.BlBorrowedMana(t, g, welder)
	if len(idx) != 1 {
		t.Fatalf("the Welder has %d lent mana rows, want Sol Ring's 1", len(idx))
	}
	// A lent {T} needs the Welder untapped, as printed.
	if err := g.ActivateManaAbility(me.ID, welder, idx[0], game.ManaAbilityParams{Ref: origins[0].Ref}); err == nil {
		t.Error("the tapped Welder activated a lent {T} ability")
	}
}

func TestExileLenderHazelsBrewmasterLendsToYourFoods(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brewmaster := effectstest.GaPut(g, me.ID, "Hazel's Brewmaster", "Creature — Squirrel Warlock", effectstest.B21HazelsBrewmasterOracle)
	food := effects.FoodToken()
	food.InstanceID, food.Owner, food.Controller = uuid.New(), me.ID, me.ID
	foodID := effectstest.PushBattlefieldCardWithTimestamp(g, food)
	theirFood := effects.FoodToken()
	theirFood.InstanceID, theirFood.Owner, theirFood.Controller = uuid.New(), opp.ID, opp.ID
	theirFoodID := effectstest.PushBattlefieldCardWithTimestamp(g, theirFood)

	effectstest.ElPushExiled(g, opp.ID, effectstest.ElRef(t, g, brewmaster), "Llanowar Elves", "Creature — Elf Druid", effectstest.BlLlanowarOracle)
	effectstest.ElPushExiled(g, opp.ID, effectstest.ElRef(t, g, brewmaster), "Sol Ring", "Artifact", effectstest.B15SolRingOracle)
	if idx, _ := effectstest.BlBorrowedMana(t, g, foodID); len(idx) != 1 {
		t.Fatalf("your Food has %d lent mana rows, want the Elves' 1 (Sol Ring is not a creature card)", len(idx))
	}
	if idx, _ := effectstest.BlBorrowedMana(t, g, theirFoodID); len(idx) != 0 {
		t.Errorf("an opponent's Food borrowed %d rows", len(idx))
	}
	if idx, _ := effectstest.BlBorrowedMana(t, g, brewmaster); len(idx) != 0 {
		t.Errorf("the Brewmaster itself borrowed %d rows", len(idx))
	}
	// The Food keeps its own ability beside the lent one.
	if own := game.ActivatedAbilitiesForCard(effectstest.GaLayered(t, g, foodID)); len(own) != 1 {
		t.Errorf("the Food has %d activated abilities, want its own 1", len(own))
	}
	effectstest.BlTapLentMana(t, g, me, foodID)
	if !effectstest.GaLayered(t, g, foodID).Tapped {
		t.Error("the lent {T} did not tap the Food")
	}

	// An uncatalogued creature card exiled with it: the Food names it.
	bear := effectstest.ElPushExiled(g, opp.ID, effectstest.ElRef(t, g, brewmaster), "Mystery Bear", "Creature — Bear", "")
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elves := effectstest.BlToGraveyard(g, opp, "Llanowar Elves", "Creature — Elf Druid", effectstest.BlLlanowarOracle)
	brewmaster := effectstest.CastAndResolveCreature(t, g, "Hazel's Brewmaster", "Creature — Squirrel Warlock", effectstest.B21HazelsBrewmasterOracle)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, elves)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ElExiledCard(t, g, elves).ExiledWith; got != effectstest.ElRef(t, g, brewmaster) {
		t.Fatalf("the trigger's exile is linked to %+v, want the Brewmaster object", got)
	}
	var food uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Food" {
			food = c.InstanceID
		}
	}
	if idx, _ := effectstest.BlBorrowedMana(t, g, food); len(idx) != 1 {
		t.Errorf("the new Food has %d lent mana rows, want the Elves' 1", len(idx))
	}
}

func TestExileLenderAgathasSoulCauldronLendsToCreaturesWithCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cauldron := effectstest.GaPut(g, me.ID, "Agatha's Soul Cauldron", "Legendary Artifact", effectstest.ElCauldronOracle)
	bear := effectstest.GaPut(g, me.ID, "Bear", "Creature — Bear", "")
	otherBear := effectstest.GaPut(g, me.ID, "Other Bear", "Creature — Bear", "")
	theirBear := effectstest.GaPut(g, opp.ID, "Their Bear", "Creature — Bear", "")
	elves := effectstest.BlToGraveyard(g, opp, "Llanowar Elves", "Creature — Elf Druid", effectstest.BlLlanowarOracle)

	if err := g.ActivateCatalogAbility(me.ID, cauldron, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: elves}},
	}); err != nil {
		t.Fatalf("activate the Cauldron: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The reflexive trigger asks for a creature you control.
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(p.PickTargetCards, theirBear) {
		t.Error("the counter can target an opponent's creature")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.GaLayered(t, g, bear).Counters[game.CounterPlusOne]; n != 1 {
		t.Fatalf("the Bear has %d +1/+1 counters, want 1", n)
	}
	if idx, _ := effectstest.BlBorrowedMana(t, g, bear); len(idx) != 1 {
		t.Errorf("the Bear with a counter has %d lent rows, want the Elves' 1", len(idx))
	}
	if idx, _ := effectstest.BlBorrowedMana(t, g, otherBear); len(idx) != 0 {
		t.Errorf("a creature with no counter borrowed %d rows", len(idx))
	}
	// The spend rider: any ability of a creature you control.
	card := effectstest.GaLayered(t, g, bear)
	theirs := effectstest.GaLayered(t, g, theirBear)
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
	if idx, _ := effectstest.BlBorrowedMana(t, g, bear); len(idx) != 0 {
		t.Errorf("the Bear lost its counter and still has %d lent rows", len(idx))
	}
}

func TestExileLenderAgathasSoulCauldronNoncreatureMakesNoTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cauldron := effectstest.GaPut(g, me.ID, "Agatha's Soul Cauldron", "Legendary Artifact", effectstest.ElCauldronOracle)
	effectstest.GaPut(g, me.ID, "Bear", "Creature — Bear", "")
	ring := effectstest.BlToGraveyard(g, opp, "Sol Ring", "Artifact", effectstest.B15SolRingOracle)
	effectstest.ElActivate(t, g, me, cauldron, 0, ring)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a noncreature card exiled this way made the counter trigger")
	}
	if !effectstest.InExile(g, ring) {
		t.Error("the card is not exiled")
	}
}

func TestExileLenderStewardOfTheHarvestGivesYourCreaturesTheLandsAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := effectstest.BlToGraveyard(g, me, "Forest", "Basic Land — Forest", "")
	island := effectstest.BlToGraveyard(g, me, "Island", "Basic Land — Island", "")
	theirs := effectstest.BlToGraveyard(g, opp, "Mountain", "Basic Land — Mountain", "")
	bear := effectstest.GaPut(g, me.ID, "Bear", "Creature — Bear", "")
	theirBear := effectstest.GaPut(g, opp.ID, "Their Bear", "Creature — Bear", "")
	steward := effectstest.CastAndResolveCreature(t, g, "Steward of the Harvest", "Creature — Human Druid", effectstest.ElStewardOracle)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(p.PickTargetCards, theirs) {
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
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, steward} {
		if idx, _ := effectstest.BlBorrowedMana(t, g, id); len(idx) != 2 {
			t.Errorf("%s has %d lent mana rows, want the Forest's and the Island's", effectstest.GaLayered(t, g, id).Name, len(idx))
		}
	}
	if idx, _ := effectstest.BlBorrowedMana(t, g, theirBear); len(idx) != 0 {
		t.Errorf("an opponent's creature borrowed %d rows", len(idx))
	}
	effectstest.BlTapLentMana(t, g, me, bear)
	if len(effectstest.Batch01PoolColors(me)) != 1 {
		t.Errorf("the lent land ability made %v", effectstest.Batch01PoolColors(me))
	}
}

func TestExileLenderTerritoryForgeTriggersOnlyWhenCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ring := effectstest.GaPut(g, opp.ID, "Sol Ring", "Artifact", effectstest.B15SolRingOracle)
	forge := effectstest.CastAndResolveCreature(t, g, "Territory Forge", "Artifact", effectstest.ElForgeOracle)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, ring)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ElExiledCard(t, g, ring).ExiledWith; got != effectstest.ElRef(t, g, forge) {
		t.Fatalf("the exiled Sol Ring is linked to %+v, want the Forge object", got)
	}
	if idx, _ := effectstest.BlBorrowedMana(t, g, forge); len(idx) != 1 {
		t.Fatalf("the Forge has %d lent mana rows, want Sol Ring's 1", len(idx))
	}
	effectstest.BlTapLentMana(t, g, me, forge)
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 {
		t.Errorf("the lent Sol Ring made %v, want two colourless", got)
	}
	// Put onto the battlefield without being cast: no trigger.
	before := g.Exile.Size()
	other := effectstest.GaPut(g, me.ID, "Territory Forge", "Artifact", effectstest.ElForgeOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || g.Exile.Size() != before {
		t.Error("a Forge that was not cast triggered")
	}
	if idx, _ := effectstest.BlBorrowedMana(t, g, other); len(idx) != 0 {
		t.Errorf("the second Forge borrowed the first one's card: %d rows", len(idx))
	}
}

func TestExileLenderRexBorrowsFromEveryCardWithABrainCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rex := effectstest.GaPut(g, me.ID, "Rex, Cyber-Hound", "Legendary Artifact Creature — Robot Dog", effectstest.ElRexOracle)
	elves := effectstest.BlToGraveyard(g, opp, "Llanowar Elves", "Creature — Elf Druid", effectstest.BlLlanowarOracle)
	g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(me.ID, game.CounterEnergy, 2) })
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.ElActivate(t, g, me, rex, 0, elves)
	if n := me.Counters[game.CounterEnergy]; n != 0 {
		t.Errorf("energy left %d, want 0", n)
	}
	if n := effectstest.ElExiledCard(t, g, elves).Counters[effects.RexBrainCounter]; n != 1 {
		t.Fatalf("the exiled Elves have %d brain counters, want 1", n)
	}
	if idx, _ := effectstest.BlBorrowedMana(t, g, rex); len(idx) != 1 {
		t.Fatalf("Rex has %d lent rows, want the Elves' 1", len(idx))
	}
	// Any card in exile with a brain counter lends, whoever exiled it;
	// one without lends nothing.
	other := effectstest.ElPushExiled(g, opp.ID, game.PermissionCardRef{}, "Llanowar Elves", "Creature — Elf Druid", effectstest.BlLlanowarOracle)
	if idx, _ := effectstest.BlBorrowedMana(t, g, rex); len(idx) != 1 {
		t.Fatalf("a card with no brain counter lent: %d rows", len(idx))
	}
	if err := g.AddCounter(other, effects.RexBrainCounter, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if idx, _ := effectstest.BlBorrowedMana(t, g, rex); len(idx) != 2 {
		t.Errorf("a brain counter put on an exiled card: Rex has %d rows, want 2", len(idx))
	}
}

func TestExileLenderMairsilCagesFromYourHandOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elves := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: elves, Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid",
		OracleID: effectstest.BlLlanowarOracle, Owner: me.ID, Controller: me.ID})
	land := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	mairsil := effectstest.CastAndResolveCreature(t, g, "Mairsil, the Pretender", "Legendary Creature — Human Wizard", effectstest.ElMairsilOracle)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no cage prompt")
	}
	if effectstest.HasID(pick.ChooseCards, land) {
		t.Error("a land is offered for the cage")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Errorf("cage bounds %d..%d, want at most one, optional", pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{elves}); err != nil {
		t.Fatalf("cage: %v", err)
	}
	if n := effectstest.ElExiledCard(t, g, elves).Counters[cardslu.MairsilCageCounter]; n != 1 {
		t.Fatalf("the caged Elves have %d cage counters, want 1", n)
	}
	// Arriving from a hand re-read the layers.
	idx, origins := effectstest.BlBorrowedMana(t, g, mairsil)
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
	theirs := effectstest.ElPushExiled(g, opp.ID, game.PermissionCardRef{}, "Llanowar Elves", "Creature — Elf Druid", effectstest.BlLlanowarOracle)
	if err := g.AddCounter(theirs, cardslu.MairsilCageCounter, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if idx, _ := effectstest.BlBorrowedMana(t, g, mairsil); len(idx) != 1 {
		t.Errorf("an opponent's caged card lent: %d rows", len(idx))
	}
}

func TestExileLenderLocusOfEnlightenmentBorrowsItsCraftMaterials(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	locus := effectstest.GaPut(g, me.ID, "Locus of Enlightenment", "Legendary Artifact", effectstest.ElJewelOracle+"#1")
	elves := effectstest.ElPushExiled(g, me.ID, game.PermissionCardRef{}, "Llanowar Elves", "Creature — Elf Druid", effectstest.BlLlanowarOracle)
	ring := effectstest.ElPushExiled(g, me.ID, game.PermissionCardRef{}, "Sol Ring", "Artifact", effectstest.B15SolRingOracle)
	effectstest.ElPushExiled(g, me.ID, game.PermissionCardRef{}, "Not A Material", "Creature — Elf Druid", effectstest.BlLlanowarOracle)
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
	idx, origins := effectstest.BlBorrowedMana(t, g, locus)
	if len(idx) != 2 {
		t.Fatalf("the Locus has %d lent mana rows, want its two materials'", len(idx))
	}
	if err := g.ActivateManaAbility(me.ID, locus, idx[1], game.ManaAbilityParams{Ref: origins[1].Ref}); err != nil {
		t.Fatalf("activate a material's row: %v", err)
	}
	if !effectstest.GaLayered(t, g, locus).Tapped {
		t.Error("the lent {T} did not tap the Locus")
	}
	// A material that leaves exile stops lending.
	effectstest.ElToHand(t, g, me.ID, elves)
	if idx, _ := effectstest.BlBorrowedMana(t, g, locus); len(idx) != 1 {
		t.Errorf("a material left exile and the Locus has %d rows, want 1", len(idx))
	}
}

func TestExileLenderLocusCopiesEachNonManaAbilityYouActivate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Locus of Enlightenment", "Legendary Artifact", effectstest.ElJewelOracle+"#1", false)
	monolith := effectstest.PushCatalogPermanent(g, me.ID, "Basalt Monolith", "Artifact", effectstest.AcBasaltMonolithOracle, false)
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, monolith, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if effectstest.AcItemOnStack(g, "Locus of Enlightenment") == nil {
		t.Fatal("activating an ability did not trigger the Locus")
	}
	// The trigger resolves first and copies the untap; no payment, no
	// "may".
	for i := 0; i < 16 && effectstest.AcItemOnStack(g, "Locus of Enlightenment") != nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if n := effectstest.AcCopiesOnStack(g); n != 1 {
		t.Fatalf("the Locus made %d copies, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.AcCountEvents(g, game.EventActivateAbility); n != 1 {
		t.Errorf("%d activations announced, want 1: a copy is created, not activated (CR 707.10a)", n)
	}
	// A mana ability is never copied (CR 605.3b).
	elf := effectstest.BlElves(g, me)
	if err := g.ActivateManaAbility(me.ID, elf, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Elves: %v", err)
	}
	if effectstest.AcItemOnStack(g, "Locus of Enlightenment") != nil {
		t.Error("a mana ability triggered the Locus")
	}
}

func TestExileLenderEnigmaJewelCraftsWithNonlandsWithActivatedAbilities(t *testing.T) {
	cost := effects.CraftWithNonlandsWithActivatedAbilities(4).ExilePermanents
	elves := game.Card{Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid", OracleID: effectstest.BlLlanowarOracle}
	vanilla := game.Card{Name: "Bear", TypeLine: "Creature — Bear"}
	land := game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	jewel := effectstest.GaPut(g, me.ID, "The Enigma Jewel", "Legendary Artifact", effectstest.ElJewelOracle)
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
	extra := game.Card{InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact", OracleID: effectstest.B15SolRingOracle, Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(extra)
	g.ReadSnapshot(func() { payable = g.ExilePermanentsPayable(me.ID, jewel, cost) })
	if !payable {
		t.Error("four nonlands with activated abilities did not pay the craft")
	}
}
