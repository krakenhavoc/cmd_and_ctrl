package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// borrowed_lender_cards_test.go — the first cards on ADR 0155's seam (its
// PR 2): the battlefield and graveyard lenders. Each test asserts what a
// player gets — who has which ability, who may activate it, what pays for
// it — rather than a label.

const (
	blMarvinOracle      = "22f0dbbb-9d8d-42bc-8903-fc0b65e00952"
	blKrajOracle        = "93cfa771-067f-44ed-9e29-6caf698ee9aa"
	blRobaranOracle     = "1d8c8c71-d569-450d-adc1-1a1b9aea0123"
	blDranaOracle       = "7abc2a1b-b758-480b-9415-f8f5369ab52f"
	blRefractorOracle   = "ae3517db-bb96-4c29-91dc-5bd68683dced"
	blSharkeyOracle     = "93a476c0-c77e-43b0-8695-ada40368a002"
	blOozeOracle        = "2fdeb920-2e48-4308-9517-743eec219c04"
	blSafehouseOracle   = "08826de5-a2d7-4732-a5f7-f580507a7844"
	blThranduilOracle   = "4a5f76e7-40be-4b06-9935-4a3b2672e1c2"
	blTrazynOracle      = "1ef3b48a-55ef-422e-ac7f-e1c1057eec2c"
	blLlanowarOracle    = "68954295-54e3-4303-a6bc-fc4547a4e3a3"
	blEvolvingWildsOrcl = "a75445d3-1303-4bb5-89ad-26ea93fecd48"
)

// blBorrowedMana is a permanent's lent mana rows, in order.
func blBorrowedMana(t *testing.T, g *game.Game, id uuid.UUID) (idx []int, origins game.AbilityOrigins) {
	t.Helper()
	_, all := game.ManaAbilitiesWithOrigins(gaLayered(t, g, id))
	for i, o := range all {
		if o.Borrowed {
			idx = append(idx, i)
			origins = append(origins, o)
		}
	}
	return idx, origins
}

// blBorrowedActivated is a permanent's lent activated rows, in order.
func blBorrowedActivated(t *testing.T, g *game.Game, id uuid.UUID) (idx []int, origins game.AbilityOrigins) {
	t.Helper()
	_, all := game.ActivatedAbilitiesWithOrigins(gaLayered(t, g, id))
	for i, o := range all {
		if o.Borrowed {
			idx = append(idx, i)
			origins = append(origins, o)
		}
	}
	return idx, origins
}

// blToGraveyard puts a card into a seat's graveyard and drops the layer
// cache, as the graveyard bump (#1117) does for a real move.
func blToGraveyard(g *game.Game, p *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	g.BumpLayerVersionForTest()
	return id
}

func blElves(g *game.Game, p *game.Player) uuid.UUID {
	return gaPut(g, p.ID, "Llanowar Elves", "Creature — Elf Druid", blLlanowarOracle)
}

func blTapLentMana(t *testing.T, g *game.Game, p *game.Player, card uuid.UUID) {
	t.Helper()
	idx, origins := blBorrowedMana(t, g, card)
	if len(idx) == 0 {
		t.Fatalf("%s has no lent mana ability", card)
	}
	if err := g.ActivateManaAbility(p.ID, card, idx[0], game.ManaAbilityParams{Ref: origins[0].Ref}); err != nil {
		t.Fatalf("activate the lent mana ability: %v", err)
	}
}

func TestLenderMarvinHasTheManaAbilitiesOfYourOtherNamedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	marvin := gaPut(g, me.ID, "Marvin, Murderous Mimic", "Artifact Creature — Toy", blMarvinOracle)
	if idx, _ := blBorrowedMana(t, g, marvin); len(idx) != 0 {
		t.Fatalf("Marvin alone has %d lent rows", len(idx))
	}
	elves := blElves(g, me)
	blElves(g, opp)
	if idx, _ := blBorrowedMana(t, g, marvin); len(idx) != 1 {
		t.Fatalf("Marvin has %d lent mana rows beside your Elves, want 1 (an opponent's creature lends nothing)", len(idx))
	}
	blTapLentMana(t, g, me, marvin)
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool = %v, want [G]", got)
	}
	if !gaLayered(t, g, marvin).Tapped || gaLayered(t, g, elves).Tapped {
		t.Error("the lent {T} taps Marvin, not the lender")
	}
	// A second creature with Marvin's own name lends nothing (the name
	// comparison), and a creature of another name does.
	gaPut(g, me.ID, "Marvin, Murderous Mimic", "Artifact Creature — Toy", blMarvinOracle)
	if idx, _ := blBorrowedMana(t, g, marvin); len(idx) != 1 {
		t.Errorf("a creature named Marvin lent rows: %d", len(idx))
	}
}

func TestLenderExperimentKrajHasTheAbilitiesOfCreaturesWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kraj := gaPut(g, me.ID, "Experiment Kraj", "Legendary Creature — Ooze Mutant", blKrajOracle)
	elves := blElves(g, me)
	if idx, _ := blBorrowedMana(t, g, kraj); len(idx) != 0 {
		t.Fatalf("an Elf with no counter lent %d rows", len(idx))
	}
	if err := g.AddCounter(elves, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if idx, _ := blBorrowedMana(t, g, kraj); len(idx) != 1 {
		t.Fatalf("the Elf with a counter lent %d rows, want 1", len(idx))
	}
	if err := g.AddCounter(elves, game.CounterPlusOne, -1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if idx, _ := blBorrowedMana(t, g, kraj); len(idx) != 0 {
		t.Errorf("the counter is gone and %d rows are still lent", len(idx))
	}
	// Kraj's own ability still reads as its own and puts a counter on.
	if own := game.ActivatedAbilitiesForCard(gaLayered(t, g, kraj)); len(own) != 1 {
		t.Fatalf("Kraj has %d activated abilities, want its own 1", len(own))
	}
}

func TestLenderExperimentKrajsCounterAbilityIsLentBackByAnotherKraj(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	first := gaPut(g, me.ID, "Experiment Kraj", "Legendary Creature — Ooze Mutant", blKrajOracle)
	second := gaPut(g, me.ID, "Other Kraj", "Legendary Creature — Ooze Mutant", blKrajOracle)
	if err := g.AddCounter(second, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	idx, origins := blBorrowedActivated(t, g, first)
	if len(idx) != 1 {
		t.Fatalf("the first Kraj has %d lent activated rows, want the other Kraj's 1", len(idx))
	}
	// The lent row is the other Kraj's catalog row, and it works: it
	// puts a counter on a target creature and taps the BORROWER.
	target := gaPut(g, me.ID, "Bear", "Creature — Bear", "")
	err := g.ActivateCatalogAbility(me.ID, first, idx[0], game.ActivateAbilityParams{
		Ref:     origins[0].Ref,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	})
	if err != nil {
		t.Fatalf("activate the lent counter ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	var counters int
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(target)
		counters = c.Counters[game.CounterPlusOne]
	})
	if counters != 1 {
		t.Errorf("the target has %d counters, want 1", counters)
	}
	if !gaLayered(t, g, first).Tapped {
		t.Error("the borrower was not tapped")
	}
}

func TestLenderRobaranMercenariesBorrowsFromLegendaryCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	robaran := gaPut(g, me.ID, "Robaran Mercenaries", "Creature — Human Mercenary", blRobaranOracle)
	blElves(g, me)
	if idx, _ := blBorrowedMana(t, g, robaran); len(idx) != 0 {
		t.Fatalf("a non-legendary Elf lent %d rows", len(idx))
	}
	gaPut(g, opp.ID, "Their Legend", "Legendary Creature — Elf Druid", blLlanowarOracle)
	if idx, _ := blBorrowedMana(t, g, robaran); len(idx) != 0 {
		t.Fatalf("an opponent's legend lent %d rows", len(idx))
	}
	gaPut(g, me.ID, "My Legend", "Legendary Creature — Elf Druid", blLlanowarOracle)
	if idx, _ := blBorrowedMana(t, g, robaran); len(idx) != 1 {
		t.Errorf("your legend lent %d rows, want 1", len(idx))
	}
}

func TestLenderDranaAndLinvalaStopsTheLendersAndNotTheCopy(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drana := gaPut(g, me.ID, "Drana and Linvala", "Legendary Creature — Vampire Angel", blDranaOracle)
	mine := blElves(g, me)
	theirs := blElves(g, opp)
	if idx, _ := blBorrowedMana(t, g, drana); len(idx) != 1 {
		t.Fatalf("Drana has %d lent rows, want the opponent's Elves only", len(idx))
	}
	// The gate: the opponent's creature can't be activated, your own can.
	g.Turn.PriorityHolder = 1
	g.Turn.ActiveSeat = 1
	if err := g.ActivateManaAbility(opp.ID, theirs, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrCantActivate) {
		t.Errorf("the opponent's Elves activated under Drana and Linvala: %v", err)
	}
	g.Turn.PriorityHolder = 0
	g.Turn.ActiveSeat = 0
	if err := g.ActivateManaAbility(me.ID, mine, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf("your own Elves were stopped: %v", err)
	}
	// Drana's lent copy is not stopped.
	blTapLentMana(t, g, me, drana)
}

func TestLenderDranaAndLinvalasAnyColourRiderCoversOnlyTheLentRows(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drana := gaPut(g, me.ID, "Drana and Linvala", "Legendary Creature — Vampire Angel", blDranaOracle)
	blElves(g, opp)
	_, origins := blBorrowedMana(t, g, drana)
	if len(origins) != 1 {
		t.Fatalf("Drana has %d lent rows", len(origins))
	}
	card := gaLayered(t, g, drana)
	var lent, own bool
	g.ReadSnapshot(func() {
		lent = g.SpendsManaAsAnyColorForEffect(me.ID, game.ManaSpendForAbilityRow(card, origins[0]))
		own = g.SpendsManaAsAnyColorForEffect(me.ID, game.ManaSpendForAbility(card))
	})
	if !lent {
		t.Error("mana is not spendable as any colour on a lent row")
	}
	if own {
		t.Error("the rider reached a payment that is not a lent row")
	}
}

func TestLenderManascapeRefractorBorrowsEveryLandOnTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	refractor := gaPut(g, me.ID, "Manascape Refractor", "Artifact", blRefractorOracle)
	gaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	gaPut(g, opp.ID, "Island", "Basic Land — Island", "")
	idx, origins := blBorrowedMana(t, g, refractor)
	if len(idx) != 2 {
		t.Fatalf("the Refractor has %d lent mana rows, want one per land (2)", len(idx))
	}
	card := gaLayered(t, g, refractor)
	var covered bool
	g.ReadSnapshot(func() {
		covered = g.SpendsManaAsAnyColorForEffect(me.ID, game.ManaSpendForAbilityRow(card, origins[0]))
	})
	if !covered {
		t.Error("the Refractor's rider does not cover its lent rows")
	}
	blTapLentMana(t, g, me, refractor)
	if len(batch01PoolColors(me)) != 1 {
		t.Errorf("the lent land ability made %v", batch01PoolColors(me))
	}
}

func TestLenderSharkeyBorrowsOpponentsLandAbilitiesExceptManaAndGatesThem(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sharkey := gaPut(g, me.ID, "Sharkey, Tyrant of the Shire", "Legendary Creature — Avatar Rogue", blSharkeyOracle)
	wilds := gaPut(g, opp.ID, "Evolving Wilds", "Land", blEvolvingWildsOrcl)
	forest := gaPut(g, opp.ID, "Forest", "Basic Land — Forest", "")
	gaPut(g, me.ID, "My Wilds", "Land", blEvolvingWildsOrcl)
	if idx, _ := blBorrowedMana(t, g, sharkey); len(idx) != 0 {
		t.Errorf("Sharkey borrowed %d mana abilities, which it leaves out", len(idx))
	}
	idx, origins := blBorrowedActivated(t, g, sharkey)
	if len(idx) != 1 {
		t.Fatalf("Sharkey has %d lent rows, want the opponent's Evolving Wilds only", len(idx))
	}
	// The gate stops the land and spares its mana ability (the Forest).
	g.Turn.PriorityHolder = 1
	g.Turn.ActiveSeat = 1
	if err := g.ActivateCatalogAbility(opp.ID, wilds, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrCantActivate) {
		t.Errorf("the opponent's Evolving Wilds activated under Sharkey: %v", err)
	}
	if err := g.ActivateManaAbility(opp.ID, forest, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf("the gate stopped a mana ability: %v", err)
	}
	g.Turn.PriorityHolder = 0
	g.Turn.ActiveSeat = 0
	// Sharkey's lent copy is not stopped.
	if err := g.ActivateCatalogAbility(me.ID, sharkey, idx[0], game.ActivateAbilityParams{Ref: origins[0].Ref}); err != nil {
		t.Errorf("Sharkey's lent Evolving Wilds was stopped: %v", err)
	}
}

func TestLenderNecroticOozeHasTheAbilitiesOfCreatureCardsInAnyGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ooze := gaPut(g, me.ID, "Necrotic Ooze", "Creature — Ooze", blOozeOracle)
	if idx, _ := blBorrowedMana(t, g, ooze); len(idx) != 0 {
		t.Fatalf("an empty graveyard lent %d rows", len(idx))
	}
	blToGraveyard(g, me, "Llanowar Elves", "Creature — Elf Druid", blLlanowarOracle)
	elfID := blToGraveyard(g, opp, "Llanowar Elves", "Creature — Elf Druid", blLlanowarOracle)
	blToGraveyard(g, opp, "Not A Creature", "Artifact", blLlanowarOracle)
	if idx, _ := blBorrowedMana(t, g, ooze); len(idx) != 2 {
		t.Fatalf("two Elves in two graveyards lent %d rows, want 2 (CR 113.2c)", len(idx))
	}
	blTapLentMana(t, g, me, ooze)
	// Exiling one takes its instance away.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneExile}, elfID); err != nil {
		t.Fatalf("exile: %v", err)
	}
	if idx, _ := blBorrowedMana(t, g, ooze); len(idx) != 1 {
		t.Errorf("after one Elf was exiled the Ooze has %d rows, want 1", len(idx))
	}
}

func TestLenderMirranSafehouseHasTheManaAbilitiesOfLandCardsInGraveyards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	safehouse := gaPut(g, me.ID, "Mirran Safehouse", "Artifact", blSafehouseOracle)
	blToGraveyard(g, opp, "Forest", "Basic Land — Forest", "")
	blToGraveyard(g, me, "Llanowar Elves", "Creature — Elf Druid", blLlanowarOracle)
	idx, origins := blBorrowedMana(t, g, safehouse)
	if len(idx) != 1 {
		t.Fatalf("the Safehouse has %d lent mana rows, want the Forest's 1", len(idx))
	}
	blTapLentMana(t, g, me, safehouse)
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool = %v, want [G] (%s)", got, origins[0].Ref)
	}
}

func TestLenderThranduilBorrowsFromElvesInYourGraveyardOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thranduil := gaPut(g, me.ID, "Thranduil, the Elvenking", "Legendary Creature — Elf Noble", blThranduilOracle)
	blToGraveyard(g, opp, "Llanowar Elves", "Creature — Elf Druid", blLlanowarOracle)
	blToGraveyard(g, me, "Not An Elf", "Creature — Human", blLlanowarOracle)
	if idx, _ := blBorrowedMana(t, g, thranduil); len(idx) != 0 {
		t.Fatalf("a foreign Elf and your Human lent %d rows", len(idx))
	}
	blToGraveyard(g, me, "Llanowar Elves", "Creature — Elf Druid", blLlanowarOracle)
	if idx, _ := blBorrowedMana(t, g, thranduil); len(idx) != 1 {
		t.Errorf("your Elf lent %d rows, want 1", len(idx))
	}
}

func TestLenderThranduilLootsForAnotherLegendaryElf(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gaPut(g, me.ID, "Thranduil, the Elvenking", "Legendary Creature — Elf Noble", blThranduilOracle)
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Plain Elf", "Creature — Elf", "", nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before {
		t.Fatalf("a non-legendary Elf drew cards: hand %d → %d", before, me.Hand.Size())
	}
	castCatalogSpell(t, g, "Legendary Elf", "Legendary Creature — Elf Warrior", "", nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before+2 {
		t.Errorf("hand %d → %d, want two drawn", before, me.Hand.Size())
	}
	asked := false
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == me.ID && c.FromPlayer == me.ID {
			asked = true
		}
	}
	if !asked {
		t.Error("no discard prompt after the draw")
	}
}

func TestLenderTrazynHasTheAbilitiesOfArtifactCardsInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trazyn := gaPut(g, me.ID, "Trazyn the Infinite", "Legendary Artifact Creature — Necron", blTrazynOracle)
	blToGraveyard(g, opp, "Artifact Elf", "Artifact Creature — Elf Druid", blLlanowarOracle)
	blToGraveyard(g, me, "Not An Artifact", "Creature — Elf Druid", blLlanowarOracle)
	if idx, _ := blBorrowedMana(t, g, trazyn); len(idx) != 0 {
		t.Fatalf("a foreign artifact and your creature lent %d rows", len(idx))
	}
	blToGraveyard(g, me, "Artifact Elf", "Artifact Creature — Elf Druid", blLlanowarOracle)
	if idx, _ := blBorrowedMana(t, g, trazyn); len(idx) != 1 {
		t.Errorf("your artifact lent %d rows, want 1", len(idx))
	}
}
