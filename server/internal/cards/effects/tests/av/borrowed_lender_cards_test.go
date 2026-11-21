package testsav

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blBorrowedActivated is a permanent's lent activated rows, in order.
func blBorrowedActivated(t *testing.T, g *game.Game, id uuid.UUID) (idx []int, origins game.AbilityOrigins) {
	t.Helper()
	_, all := game.ActivatedAbilitiesWithOrigins(effectstest.GaLayered(t, g, id))
	for i, o := range all {
		if o.Borrowed {
			idx = append(idx, i)
			origins = append(origins, o)
		}
	}
	return idx, origins
}

func TestLenderMarvinHasTheManaAbilitiesOfYourOtherNamedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	marvin := effectstest.GaPut(g, me.ID, "Marvin, Murderous Mimic", "Artifact Creature — Toy", effectstest.BlMarvinOracle)
	if idx, _ := effectstest.BlBorrowedMana(t, g, marvin); len(idx) != 0 {
		t.Fatalf("Marvin alone has %d lent rows", len(idx))
	}
	elves := effectstest.BlElves(g, me)
	effectstest.BlElves(g, opp)
	if idx, _ := effectstest.BlBorrowedMana(t, g, marvin); len(idx) != 1 {
		t.Fatalf("Marvin has %d lent mana rows beside your Elves, want 1 (an opponent's creature lends nothing)", len(idx))
	}
	effectstest.BlTapLentMana(t, g, me, marvin)
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool = %v, want [G]", got)
	}
	if !effectstest.GaLayered(t, g, marvin).Tapped || effectstest.GaLayered(t, g, elves).Tapped {
		t.Error("the lent {T} taps Marvin, not the lender")
	}
	// A second creature with Marvin's own name lends nothing (the name
	// comparison), and a creature of another name does.
	effectstest.GaPut(g, me.ID, "Marvin, Murderous Mimic", "Artifact Creature — Toy", effectstest.BlMarvinOracle)
	if idx, _ := effectstest.BlBorrowedMana(t, g, marvin); len(idx) != 1 {
		t.Errorf("a creature named Marvin lent rows: %d", len(idx))
	}
}

func TestLenderExperimentKrajHasTheAbilitiesOfCreaturesWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kraj := effectstest.GaPut(g, me.ID, "Experiment Kraj", "Legendary Creature — Ooze Mutant", effectstest.BlKrajOracle)
	elves := effectstest.BlElves(g, me)
	if idx, _ := effectstest.BlBorrowedMana(t, g, kraj); len(idx) != 0 {
		t.Fatalf("an Elf with no counter lent %d rows", len(idx))
	}
	if err := g.AddCounter(elves, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if idx, _ := effectstest.BlBorrowedMana(t, g, kraj); len(idx) != 1 {
		t.Fatalf("the Elf with a counter lent %d rows, want 1", len(idx))
	}
	if err := g.AddCounter(elves, game.CounterPlusOne, -1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if idx, _ := effectstest.BlBorrowedMana(t, g, kraj); len(idx) != 0 {
		t.Errorf("the counter is gone and %d rows are still lent", len(idx))
	}
	// Kraj's own ability still reads as its own and puts a counter on.
	if own := game.ActivatedAbilitiesForCard(effectstest.GaLayered(t, g, kraj)); len(own) != 1 {
		t.Fatalf("Kraj has %d activated abilities, want its own 1", len(own))
	}
}

func TestLenderExperimentKrajsCounterAbilityIsLentBackByAnotherKraj(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	first := effectstest.GaPut(g, me.ID, "Experiment Kraj", "Legendary Creature — Ooze Mutant", effectstest.BlKrajOracle)
	second := effectstest.GaPut(g, me.ID, "Other Kraj", "Legendary Creature — Ooze Mutant", effectstest.BlKrajOracle)
	if err := g.AddCounter(second, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	idx, origins := blBorrowedActivated(t, g, first)
	if len(idx) != 1 {
		t.Fatalf("the first Kraj has %d lent activated rows, want the other Kraj's 1", len(idx))
	}
	// The lent row is the other Kraj's catalog row, and it works: it
	// puts a counter on a target creature and taps the BORROWER.
	target := effectstest.GaPut(g, me.ID, "Bear", "Creature — Bear", "")
	err := g.ActivateCatalogAbility(me.ID, first, idx[0], game.ActivateAbilityParams{
		Ref:     origins[0].Ref,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	})
	if err != nil {
		t.Fatalf("activate the lent counter ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var counters int
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(target)
		counters = c.Counters[game.CounterPlusOne]
	})
	if counters != 1 {
		t.Errorf("the target has %d counters, want 1", counters)
	}
	if !effectstest.GaLayered(t, g, first).Tapped {
		t.Error("the borrower was not tapped")
	}
}

func TestLenderRobaranMercenariesBorrowsFromLegendaryCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	robaran := effectstest.GaPut(g, me.ID, "Robaran Mercenaries", "Creature — Human Mercenary", effectstest.BlRobaranOracle)
	effectstest.BlElves(g, me)
	if idx, _ := effectstest.BlBorrowedMana(t, g, robaran); len(idx) != 0 {
		t.Fatalf("a non-legendary Elf lent %d rows", len(idx))
	}
	effectstest.GaPut(g, opp.ID, "Their Legend", "Legendary Creature — Elf Druid", effectstest.BlLlanowarOracle)
	if idx, _ := effectstest.BlBorrowedMana(t, g, robaran); len(idx) != 0 {
		t.Fatalf("an opponent's legend lent %d rows", len(idx))
	}
	effectstest.GaPut(g, me.ID, "My Legend", "Legendary Creature — Elf Druid", effectstest.BlLlanowarOracle)
	if idx, _ := effectstest.BlBorrowedMana(t, g, robaran); len(idx) != 1 {
		t.Errorf("your legend lent %d rows, want 1", len(idx))
	}
}

func TestLenderDranaAndLinvalaStopsTheLendersAndNotTheCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drana := effectstest.GaPut(g, me.ID, "Drana and Linvala", "Legendary Creature — Vampire Angel", effectstest.BlDranaOracle)
	mine := effectstest.BlElves(g, me)
	theirs := effectstest.BlElves(g, opp)
	if idx, _ := effectstest.BlBorrowedMana(t, g, drana); len(idx) != 1 {
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
	effectstest.BlTapLentMana(t, g, me, drana)
}

func TestLenderDranaAndLinvalasAnyColourRiderCoversOnlyTheLentRows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drana := effectstest.GaPut(g, me.ID, "Drana and Linvala", "Legendary Creature — Vampire Angel", effectstest.BlDranaOracle)
	effectstest.BlElves(g, opp)
	_, origins := effectstest.BlBorrowedMana(t, g, drana)
	if len(origins) != 1 {
		t.Fatalf("Drana has %d lent rows", len(origins))
	}
	card := effectstest.GaLayered(t, g, drana)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	refractor := effectstest.GaPut(g, me.ID, "Manascape Refractor", "Artifact", effectstest.BlRefractorOracle)
	effectstest.GaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	effectstest.GaPut(g, opp.ID, "Island", "Basic Land — Island", "")
	idx, origins := effectstest.BlBorrowedMana(t, g, refractor)
	if len(idx) != 2 {
		t.Fatalf("the Refractor has %d lent mana rows, want one per land (2)", len(idx))
	}
	card := effectstest.GaLayered(t, g, refractor)
	var covered bool
	g.ReadSnapshot(func() {
		covered = g.SpendsManaAsAnyColorForEffect(me.ID, game.ManaSpendForAbilityRow(card, origins[0]))
	})
	if !covered {
		t.Error("the Refractor's rider does not cover its lent rows")
	}
	effectstest.BlTapLentMana(t, g, me, refractor)
	if len(effectstest.Batch01PoolColors(me)) != 1 {
		t.Errorf("the lent land ability made %v", effectstest.Batch01PoolColors(me))
	}
}

func TestLenderSharkeyBorrowsOpponentsLandAbilitiesExceptManaAndGatesThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sharkey := effectstest.GaPut(g, me.ID, "Sharkey, Tyrant of the Shire", "Legendary Creature — Avatar Rogue", effectstest.BlSharkeyOracle)
	wilds := effectstest.GaPut(g, opp.ID, "Evolving Wilds", "Land", effectstest.BlEvolvingWildsOrcl)
	forest := effectstest.GaPut(g, opp.ID, "Forest", "Basic Land — Forest", "")
	effectstest.GaPut(g, me.ID, "My Wilds", "Land", effectstest.BlEvolvingWildsOrcl)
	if idx, _ := effectstest.BlBorrowedMana(t, g, sharkey); len(idx) != 0 {
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

// #3863: "Mana of any TYPE can be spent to activate Sharkey's
// abilities" reaches a colourless {C} too. Three red mana pay a lent Sea
// Gate Wreckage's {2}{C}; the any-colour spend would leave the {C}
// unpaid.
func TestLenderSharkeyPaysALentColourlessCostWithColouredMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sharkey := effectstest.GaPut(g, me.ID, "Sharkey, Tyrant of the Shire", "Legendary Creature — Avatar Rogue", effectstest.BlSharkeyOracle)
	effectstest.GaPut(g, opp.ID, "Sea Gate Wreckage", "Land", "91f34686-cb96-49c0-b4a7-49dd1fd076e2")
	idx, origins := blBorrowedActivated(t, g, sharkey)
	if len(idx) != 1 {
		t.Fatalf("Sharkey has %d lent rows, want the Wreckage's draw", len(idx))
	}
	me.Hand.Cards = nil
	me.ManaPool = nil
	me.ManaPool.AddMana(game.ManaToken{Color: "R"}, game.ManaToken{Color: "R"}, game.ManaToken{Color: "R"})
	g.Turn.PriorityHolder = 0
	g.Turn.ActiveSeat = 0
	if err := g.ActivateCatalogAbility(me.ID, sharkey, idx[0], game.ActivateAbilityParams{Ref: origins[0].Ref}); err != nil {
		t.Fatalf("three red mana did not pay the lent {2}{C}: %v", err)
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("pool has %d mana left, want all three spent", n)
	}
	if n := len(g.StackMeta); n != 1 {
		t.Fatalf("%d items on the stack, want the lent draw", n)
	}
	for _, it := range g.StackMeta {
		if it.Paid.OnPaper || it.Paid.PaperMana != 0 {
			t.Errorf("the {C} was left on paper (%d unpaid), want it paid with red", it.Paid.PaperMana)
		}
	}
}

func TestLenderNecroticOozeHasTheAbilitiesOfCreatureCardsInAnyGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ooze := effectstest.GaPut(g, me.ID, "Necrotic Ooze", "Creature — Ooze", effectstest.BlOozeOracle)
	if idx, _ := effectstest.BlBorrowedMana(t, g, ooze); len(idx) != 0 {
		t.Fatalf("an empty graveyard lent %d rows", len(idx))
	}
	effectstest.BlToGraveyard(g, me, "Llanowar Elves", "Creature — Elf Druid", effectstest.BlLlanowarOracle)
	elfID := effectstest.BlToGraveyard(g, opp, "Llanowar Elves", "Creature — Elf Druid", effectstest.BlLlanowarOracle)
	effectstest.BlToGraveyard(g, opp, "Not A Creature", "Artifact", effectstest.BlLlanowarOracle)
	if idx, _ := effectstest.BlBorrowedMana(t, g, ooze); len(idx) != 2 {
		t.Fatalf("two Elves in two graveyards lent %d rows, want 2 (CR 113.2c)", len(idx))
	}
	effectstest.BlTapLentMana(t, g, me, ooze)
	// Exiling one takes its instance away.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneExile}, elfID); err != nil {
		t.Fatalf("exile: %v", err)
	}
	if idx, _ := effectstest.BlBorrowedMana(t, g, ooze); len(idx) != 1 {
		t.Errorf("after one Elf was exiled the Ooze has %d rows, want 1", len(idx))
	}
}

func TestLenderMirranSafehouseHasTheManaAbilitiesOfLandCardsInGraveyards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	safehouse := effectstest.GaPut(g, me.ID, "Mirran Safehouse", "Artifact", effectstest.BlSafehouseOracle)
	effectstest.BlToGraveyard(g, opp, "Forest", "Basic Land — Forest", "")
	effectstest.BlToGraveyard(g, me, "Llanowar Elves", "Creature — Elf Druid", effectstest.BlLlanowarOracle)
	idx, origins := effectstest.BlBorrowedMana(t, g, safehouse)
	if len(idx) != 1 {
		t.Fatalf("the Safehouse has %d lent mana rows, want the Forest's 1", len(idx))
	}
	effectstest.BlTapLentMana(t, g, me, safehouse)
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool = %v, want [G] (%s)", got, origins[0].Ref)
	}
}

func TestLenderThranduilBorrowsFromElvesInYourGraveyardOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thranduil := effectstest.GaPut(g, me.ID, "Thranduil, the Elvenking", "Legendary Creature — Elf Noble", effectstest.BlThranduilOracle)
	effectstest.BlToGraveyard(g, opp, "Llanowar Elves", "Creature — Elf Druid", effectstest.BlLlanowarOracle)
	effectstest.BlToGraveyard(g, me, "Not An Elf", "Creature — Human", effectstest.BlLlanowarOracle)
	if idx, _ := effectstest.BlBorrowedMana(t, g, thranduil); len(idx) != 0 {
		t.Fatalf("a foreign Elf and your Human lent %d rows", len(idx))
	}
	effectstest.BlToGraveyard(g, me, "Llanowar Elves", "Creature — Elf Druid", effectstest.BlLlanowarOracle)
	if idx, _ := effectstest.BlBorrowedMana(t, g, thranduil); len(idx) != 1 {
		t.Errorf("your Elf lent %d rows, want 1", len(idx))
	}
}

func TestLenderThranduilLootsForAnotherLegendaryElf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.GaPut(g, me.ID, "Thranduil, the Elvenking", "Legendary Creature — Elf Noble", effectstest.BlThranduilOracle)
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Plain Elf", "Creature — Elf", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before {
		t.Fatalf("a non-legendary Elf drew cards: hand %d → %d", before, me.Hand.Size())
	}
	effectstest.CastCatalogSpell(t, g, "Legendary Elf", "Legendary Creature — Elf Warrior", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trazyn := effectstest.GaPut(g, me.ID, "Trazyn the Infinite", "Legendary Artifact Creature — Necron", effectstest.BlTrazynOracle)
	effectstest.BlToGraveyard(g, opp, "Artifact Elf", "Artifact Creature — Elf Druid", effectstest.BlLlanowarOracle)
	effectstest.BlToGraveyard(g, me, "Not An Artifact", "Creature — Elf Druid", effectstest.BlLlanowarOracle)
	if idx, _ := effectstest.BlBorrowedMana(t, g, trazyn); len(idx) != 0 {
		t.Fatalf("a foreign artifact and your creature lent %d rows", len(idx))
	}
	effectstest.BlToGraveyard(g, me, "Artifact Elf", "Artifact Creature — Elf Druid", effectstest.BlLlanowarOracle)
	if idx, _ := effectstest.BlBorrowedMana(t, g, trazyn); len(idx) != 1 {
		t.Errorf("your artifact lent %d rows, want 1", len(idx))
	}
}
