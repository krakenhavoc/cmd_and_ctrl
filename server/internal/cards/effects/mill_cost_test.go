package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mill_cost_test.go — #3845: "Mill N cards" as a cost, on both owners of
// the component, and the two cards that were waiting on it (The Warring
// Triad, Millikin). The engine halves are against abilities registered
// for the test; the card halves are against the real catalog entries.

const (
	mcTestOracle = "test-mc-mill-two"
	mcShortOr    = "test-mc-short"

	warringTriadOracle = "df0b5995-117b-4ba8-964d-ea3c592621ef"
)

func mcMillEvents(g *game.Game) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventMill {
			n++
		}
	}
	return n
}

func mcFillGraveyard(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() {
		for i := 0; i < n; i++ {
			p.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: p.ID, Controller: p.ID})
		}
	})
}

// "Mill two cards" is paid at announce, through the mill door: the top
// two cards are in the graveyard before the ability resolves, one
// EventMill each, and the effect still runs afterwards.
func TestMillCostIsPaidAtAnnounceThroughTheMillDoor(t *testing.T) {
	var rec lcRecord
	registerForTest(t, lcSpec(mcTestOracle, "Mill Payer", MillCardsCost(2), &rec))
	g, me, _ := exileCostTable(t)
	lib := lcLibrary(g, me, 3)
	src := pushCatalogPermanent(g, me.ID, "Mill Payer", "Enchantment", mcTestOracle, false)
	events := mcMillEvents(g)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if rec.ran {
		t.Fatal("the effect ran at announce; only the cost is paid there")
	}
	if me.Library.Size() != 1 || !me.Library.Contains(lib[2]) {
		t.Fatalf("library holds %d cards, want only the bottom one", me.Library.Size())
	}
	if !me.Graveyard.Contains(lib[0]) || !me.Graveyard.Contains(lib[1]) {
		t.Fatal("the top two cards are not in the graveyard")
	}
	if got := mcMillEvents(g) - events; got != 2 {
		t.Errorf("EventMill fired %d times, want once per card milled", got)
	}
	if len(g.StackMeta) != 1 {
		t.Fatalf("stack holds %d items, want the paid ability", len(g.StackMeta))
	}
	passPriorityAroundTable(t, g)
	if !rec.ran {
		t.Error("the effect never ran")
	}
}

// CR 701.17b / CR 118.3: a cost that mills more cards than the library
// has can't be paid, and nothing else is paid either.
func TestMillCostOnAShortLibraryIsRefusedAndPaysNothing(t *testing.T) {
	var rec lcRecord
	registerForTest(t, lcSpec(mcShortOr, "Short Payer", Plus(ManaCost("{1}"), TapCost(), MillCardsCost(2)), &rec))
	g, me, _ := exileCostTable(t)
	lcLibrary(g, me, 1)
	src := pushCatalogPermanent(g, me.ID, "Short Payer", "Artifact", mcShortOr, false)
	floatMana(t, g, me, "{C}")

	err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{})
	if !errors.Is(err, game.ErrCantPayLibraryCost) {
		t.Fatalf("activate = %v, want ErrCantPayLibraryCost", err)
	}
	c, _ := g.LookupCardForEffect(src)
	if c.Tapped || me.Library.Size() != 1 || me.Graveyard.Size() != 0 || len(me.ManaPool) != 1 || len(g.StackMeta) != 0 {
		t.Fatal("a refused activation paid part of its cost")
	}
}

// The mill goes through the ordinary mill, so Rest in Peace applies:
// the cost is paid (the card left the library), the card is exiled
// instead of milled, and the ability still goes on the stack.
func TestMillCostIsReplacedByRestInPeace(t *testing.T) {
	var rec lcRecord
	registerForTest(t, lcSpec(mcTestOracle, "Mill Payer", MillCardsCost(1), &rec))
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castRestInPeace(t, g)
	lib := lcLibrary(g, me, 2)
	src := pushCatalogPermanent(g, me.ID, "Mill Payer", "Enchantment", mcTestOracle, false)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Library.Contains(lib[0]) {
		t.Fatal("the top card was not milled")
	}
	if me.Graveyard.Contains(lib[0]) {
		t.Error("the milled card reached the graveyard; Rest in Peace exiles it instead")
	}
	if !g.Exile.Contains(lib[0]) {
		t.Error("the milled card is not in exile")
	}
	if len(g.StackMeta) != 1 {
		t.Errorf("stack holds %d items, want the paid ability", len(g.StackMeta))
	}
}

// Register refuses a mill beside an exile of the top, which no printed
// cost does, and a negative count.
func TestRegisterRefusesAMalformedMillCost(t *testing.T) {
	for name, spec := range map[string]Spec{
		"negative": {OracleID: "test-mc-neg", Name: "Negative Mill", Activated: []ActivatedAbility{{
			Label: "x", Cost: MillCardsCost(-1), Effect: func(*game.Game, *game.StackItem) error { return nil }}}},
		"with an exile": {OracleID: "test-mc-both", Name: "Mill And Exile", Activated: []ActivatedAbility{{
			Label: "x", Cost: Plus(MillCardsCost(1), ExileTopOfLibrary(1)), Effect: func(*game.Game, *game.StackItem) error { return nil }}}},
		"negative on a mana ability": {OracleID: "test-mc-negmana", Name: "Negative Mill Mana", ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{Tap: true, Mill: -1}, Produced: "{C}", Label: "x"}}},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: Register accepted it", name)
					delete(registry, spec.OracleID)
				}
			}()
			Register(spec)
		}()
	}
}

// --- The Warring Triad ------------------------------------------------

const warringTriadType = "Legendary Artifact Creature — God"

func triadTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, opp := exileCostTable(t)
	lcLibrary(g, me, 5)
	// pushCatalogPermanent stamps the permanent as having been there
	// all along, so a Triad that is a creature is not summoning sick.
	id := pushCatalogPermanent(g, me.ID, "The Warring Triad", warringTriadType, warringTriadOracle, false)
	return g, me, opp, id
}

func TestWarringTriadIsRegisteredAsAFullCardWithAStackedAbility(t *testing.T) {
	spec, ok := Lookup(warringTriadOracle)
	if !ok {
		t.Fatal("The Warring Triad is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness = %v caveats = %v, want Full with none", spec.Completeness, spec.Caveats)
	}
	if len(spec.ManaAbilities) != 0 || len(spec.Activated) != 1 {
		t.Fatalf("mana abilities = %d, activated = %d: a targeted ability is not a mana ability (CR 605.1a)",
			len(spec.ManaAbilities), len(spec.Activated))
	}
	ab := spec.Activated[0]
	if !ab.Cost.Tap || ab.Cost.Mill != 1 || ab.Targets == nil || ab.SorcerySpeed {
		t.Errorf("cost = %+v targets = %v sorcery = %v, want {T}, mill one, a target and instant speed", ab.Cost, ab.Targets, ab.SorcerySpeed)
	}
}

// Tap and mill are paid at announce, the ability is on the stack (not a
// mana ability, so nothing is in the pool yet), and on resolution the
// target player picks the colour and the mana lands in THEIR pool.
func TestWarringTriadMillsAndTheTargetPlayerAddsTheMana(t *testing.T) {
	g, me, opp, triad := triadTable(t)
	top, _ := me.Library.Top()
	yard := me.Graveyard.Size()

	err := g.ActivateCatalogAbility(me.ID, triad, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	c, _ := g.LookupCardForEffect(triad)
	if !c.Tapped {
		t.Error("the Triad is not tapped")
	}
	if me.Graveyard.Size() != yard+1 || !me.Graveyard.Contains(top.InstanceID) {
		t.Error("the top card was not milled")
	}
	if len(me.ManaPool) != 0 || len(opp.ManaPool) != 0 {
		t.Fatal("mana arrived before the ability resolved — it is not a mana ability")
	}

	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceMana, opp.ID)
	if pick == nil {
		t.Fatal("the target player was not asked for a colour")
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceMana, me.ID) != nil {
		t.Error("the activator was asked for the colour; the target player adds the mana")
	}
	if len(pick.ColorOptions) != 5 {
		t.Errorf("colour options = %v, want all five colours", pick.ColorOptions)
	}
	if err := g.ResolveManaChoice(pick.ID, opp.ID, pick.ColorOptions[0]); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if len(opp.ManaPool) != 1 || len(me.ManaPool) != 0 {
		t.Errorf("pools: target %d, activator %d, want 1 and 0", len(opp.ManaPool), len(me.ManaPool))
	}
}

// Aimed at themselves, the activator gets the mana.
func TestWarringTriadCanAddTheManaToItsOwnController(t *testing.T) {
	g, me, _, triad := triadTable(t)
	if err := g.ActivateCatalogAbility(me.ID, triad, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceMana, me.ID)
	if pick == nil {
		t.Fatal("no colour pick")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, pick.ColorOptions[0]); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("pool = %d, want 1", len(me.ManaPool))
	}
}

// CR 701.17b: with no card to mill the cost can't be paid, so the
// Triad stays untapped and the ability is not on the stack.
func TestWarringTriadWithAnEmptyLibraryCannotActivate(t *testing.T) {
	g, me, _, triad := triadTable(t)
	g.WithWriteLock(func() { me.Library.Cards = nil })

	err := g.ActivateCatalogAbility(me.ID, triad, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	})
	if !errors.Is(err, game.ErrCantPayLibraryCost) {
		t.Fatalf("activate = %v, want ErrCantPayLibraryCost", err)
	}
	if c, _ := g.LookupCardForEffect(triad); c.Tapped || len(g.StackMeta) != 0 {
		t.Error("a refused activation tapped the Triad or stacked the ability")
	}
}

// The layer-cache note on the card: the Triad is not a creature below
// eight cards in the graveyard, and the mill its own ability pays wakes
// it at once — the eighth card reaching the graveyard bumps the layers.
func TestWarringTriadWakesOnTheEighthCard(t *testing.T) {
	g, me, _, triad := triadTable(t)
	mcFillGraveyard(g, me, 7)
	g.BumpLayerVersionForTest()
	if containsString(effectiveTypes(t, g, triad), "Creature") {
		t.Fatal("the Triad is a creature with seven cards in the graveyard")
	}

	if err := g.ActivateCatalogAbility(me.ID, triad, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Graveyard.Size() != 8 {
		t.Fatalf("graveyard = %d, want 8", me.Graveyard.Size())
	}
	if !containsString(effectiveTypes(t, g, triad), "Creature") {
		t.Error("the Triad is still not a creature with eight cards in the graveyard")
	}
}

// While it is not a creature, a Triad that came in this turn is not
// summoning sick (CR 302.6 is about creatures), so it can pay the {T}.
func TestWarringTriadThatJustEnteredCanTapWhileNotACreature(t *testing.T) {
	g, me, _, _ := triadTable(t)
	fresh := pushCatalogPermanent(g, me.ID, "The Warring Triad", warringTriadType, warringTriadOracle, true)
	if err := g.ActivateCatalogAbility(me.ID, fresh, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err != nil {
		t.Fatalf("a non-creature Triad that just entered could not activate: %v", err)
	}
}

// --- Millikin ----------------------------------------------------------

// Millikin's mill is the real cost: the library test, the pay and the
// replacement all go through the shared component, so Rest in Peace
// exiles the milled card while the mana still lands.
func TestMillikinMillIsReplacedByRestInPeace(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castRestInPeace(t, g)
	lib := lcLibrary(g, me, 2)
	millikin := pushCatalogPermanent(g, me.ID, "Millikin", "Artifact Creature — Construct", b11MillikinOracle, false)

	if _, err := activateMillMana(t, g, me, millikin); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("pool = %d, want the {C}", len(me.ManaPool))
	}
	if me.Graveyard.Contains(lib[0]) || !g.Exile.Contains(lib[0]) {
		t.Error("the milled card should have been exiled by Rest in Peace")
	}
}

// Millikin is an activated ability with a cost the player chooses, so the
// auto-tapper never takes it: a spell is not paid for by milling a
// library behind the player's back.
func TestMillikinIsNeverAutoTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := me.Library.Size()
	pushCatalogPermanent(g, me.ID, "Millikin", "Artifact Creature — Construct", b11MillikinOracle, false)

	cost, err := game.ParseCost("{C}")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := g.AutoTapForCost(me.ID, cost, 0); ok {
		t.Error("the auto-tapper planned Millikin's mill-paid mana")
	}
	if me.Library.Size() != lib {
		t.Error("a probe milled the library")
	}
}
