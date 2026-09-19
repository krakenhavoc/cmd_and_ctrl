package effects

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_cost_cards_test.go — #625, the five catalog cards the
// counter-removal cost component made whole (or whole but for an
// unrelated caveat): Heart of Kiran, Dragon's Hoard, Mikaeus, the
// Lunarch, Benevolent Hydra and Fain, the Broker. The component's own
// rules are pinned in game/counter_cost_test.go; these tests are about
// each card doing what its oracle text says.

const (
	heartOfKiranOracle    = "e2ee410f-2467-4f1f-84a0-8a79faedc0b3"
	ccDragonsHoardOracle  = "8cf77dc4-763b-41b7-a5da-0ef2734f08e6"
	ccMikaeusOracle       = "82f3faa8-39fa-450b-843f-d60a4c36d8f7"
	ccBenevolentHydra     = "01dbf1bc-ca62-4fb6-959c-ef7c0dc03bb0"
	ccFainTheBrokerOracle = "31b990e3-9bad-4b0a-a9d5-f5b9ed2ad0b0"
)

// pushWalkerForCounterCost seats a planeswalker with `loyalty` loyalty
// counters under `owner`'s control.
func pushWalkerForCounterCost(g *game.Game, owner uuid.UUID, loyalty int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: owner, Controller: owner,
		Counters: map[string]int{game.CounterLoyalty: loyalty},
	})
}

// renameBattlefieldForCounterCost gives a battlefield card a distinct
// name, so two pushWalkerForCounterCost walkers under one controller
// are two different legendary permanents rather than a legend-rule
// prompt.
func renameBattlefieldForCounterCost(g *game.Game, id uuid.UUID, name string) {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Name = name
		}
	}
}

func setCounters(g *game.Game, id uuid.UUID, counters map[string]int) {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Counters = counters
		}
	}
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
}

// --- Heart of Kiran ------------------------------------------------

func TestHeartOfKiranOffersBothCrewAbilities(t *testing.T) {
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: heartOfKiranOracle})
	if len(abilities) != 2 {
		t.Fatalf("Heart of Kiran lists %d abilities, want Crew 3 and the counter alternative", len(abilities))
	}
	if abilities[0].Cost.Crew != 3 || abilities[0].Cost.RemoveCounters != nil {
		t.Errorf("ability 0 should be Crew 3 alone: %+v", abilities[0].Cost)
	}
	alt := abilities[1].Cost
	if alt.Crew != 0 || alt.Loyalty != nil || alt.Tap {
		t.Errorf("the alternative pays no crew, no loyalty and no tap: %+v", alt)
	}
	if rc := alt.RemoveCounters; rc == nil || rc.Counter != game.CounterLoyalty || rc.N != 1 || rc.From == nil {
		t.Errorf("the alternative removes one loyalty counter from another permanent: %+v", rc)
	}
	if spec, _ := Lookup(heartOfKiranOracle); spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Heart of Kiran has no remaining gap: %s %v", spec.Completeness, spec.Caveats)
	}
}

// The alternative crews the Heart, paid with a loyalty counter from a
// planeswalker — at instant speed, on an opponent's turn, without
// spending the walker's loyalty activation.
func TestHeartOfKiranCrewsByRemovingALoyaltyCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	heart := pushVehicleForTest(g, me.ID, "Heart of Kiran", heartOfKiranOracle, 4, 4)
	walker := pushWalkerForCounterCost(g, me.ID, 3)
	advanceToMainOf(t, g, 1)

	if err := g.ActivateCatalogAbility(me.ID, heart, 1, game.ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{walker}}); err != nil {
		t.Fatalf("crew by loyalty on the opponent's turn: %v", err)
	}
	if got := counterCount(g, walker, game.CounterLoyalty); got != 2 {
		t.Errorf("walker loyalty %d, want 2", got)
	}
	if g.LoyaltyActivatedThisTurn[walker] {
		t.Error("crewing spent the walker's loyalty activation")
	}
	passPriorityAroundTable(t, g)
	v, _ := battlefieldCardByID(g, heart)
	if !v.IsCreature() {
		t.Fatal("the Heart is not a creature after the counter crew resolved")
	}
	if got := v.CurrentPower(); got != 4 {
		t.Errorf("crewed power %d, want the printed 4", got)
	}
}

func TestHeartOfKiranAlternativeNeedsAPlaneswalkerWithLoyalty(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	heart := pushVehicleForTest(g, me.ID, "Heart of Kiran", heartOfKiranOracle, 4, 4)
	advanceToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, heart, 1, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("no planeswalker named: err = %v, want ErrInvalidParam", err)
	}
	theirs := pushWalkerForCounterCost(g, opp.ID, 4)
	if err := g.ActivateCatalogAbility(me.ID, heart, 1, game.ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{theirs}}); !errors.Is(err, game.ErrCardCallerMismatch) {
		t.Errorf("an opponent's walker: err = %v, want ErrCardCallerMismatch", err)
	}
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	setCounters(g, bear, map[string]int{game.CounterLoyalty: 2})
	if err := g.ActivateCatalogAbility(me.ID, heart, 1, game.ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{bear}}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("a creature holding loyalty counters is not a planeswalker: err = %v, want ErrIllegalTarget", err)
	}
	if len(g.StackMeta) != 0 {
		t.Error("an unpayable crew reached the stack")
	}
	if v, _ := battlefieldCardByID(g, heart); v.IsCreature() {
		t.Error("the Heart became a creature without paying")
	}
}

// Picks among several planeswalkers, and a walker at 1 loyalty dies to
// the SBA while the crew still resolves.
func TestHeartOfKiranPicksTheWalkerAndCanKillIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	heart := pushVehicleForTest(g, me.ID, "Heart of Kiran", heartOfKiranOracle, 4, 4)
	big := pushWalkerForCounterCost(g, me.ID, 5)
	small := pushWalkerForCounterCost(g, me.ID, 1)
	// Two walkers of the same legendary name are a legend-rule prompt,
	// and since #730 an unanswered prompt gates the table. Two real
	// walkers have two names.
	renameBattlefieldForCounterCost(g, big, "Other Test Walker")
	advanceToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, heart, 1, game.ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{small}}); err != nil {
		t.Fatalf("crew from the 1-loyalty walker: %v", err)
	}
	if g.Battlefield.Contains(small) {
		t.Error("the walker paid down to 0 survived CR 704.5i")
	}
	if got := counterCount(g, big, game.CounterLoyalty); got != 5 {
		t.Errorf("the other walker paid: %d", got)
	}
	passPriorityAroundTable(t, g)
	if v, _ := battlefieldCardByID(g, heart); !v.IsCreature() {
		t.Error("the crew did not resolve after its cost killed the walker")
	}
}

// --- Dragon's Hoard ----------------------------------------------

func TestDragonsHoardSpendsAGoldCounterToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hoard := pushCatalogPermanent(g, me.ID, "Dragon's Hoard", "Artifact", ccDragonsHoardOracle, false)
	advanceToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, hoard, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrInsufficientCounters) {
		t.Fatalf("no gold counter: err = %v, want ErrInsufficientCounters", err)
	}
	if b31Tapped(t, g, hoard) {
		t.Fatal("an unpayable draw tapped the Hoard")
	}

	setCounters(g, hoard, map[string]int{"gold": 2})
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, hoard, 0, game.ActivateAbilityParams{})
	if got := counterCount(g, hoard, "gold"); got != 1 {
		t.Errorf("gold counters %d, want 1", got)
	}
	if !b31Tapped(t, g, hoard) {
		t.Error("the draw taps the Hoard")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
	if spec, _ := Lookup(ccDragonsHoardOracle); spec.Completeness != CompletenessFull {
		t.Error("Dragon's Hoard has no remaining gap")
	}
}

// --- Mikaeus, the Lunarch ------------------------------------------

func TestMikaeusSpendsACounterToPumpTheTeam(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mik := pushCatalogPermanent(g, me.ID, "Mikaeus, the Lunarch", "Legendary Creature — Human Cleric", ccMikaeusOracle, false)
	setCounters(g, mik, map[string]int{game.CounterPlusOne: 2})
	bear := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	elf := pushVanillaCreature(g, me.ID, "My Elf", 1, 1)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	advanceToMain(t, g)

	b16Activate(t, g, me.ID, mik, 1, game.ActivateAbilityParams{})
	if got := counterCount(g, mik, game.CounterPlusOne); got != 1 {
		t.Errorf("Mikaeus has %d counters, want 1 (paid one, gets none back)", got)
	}
	for _, id := range []uuid.UUID{bear, elf} {
		if got := counterCount(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("another creature of mine has %d counters, want 1", got)
		}
	}
	if got := counterCount(g, theirs, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's creature got %d counters", got)
	}
	if !b31Tapped(t, g, mik) {
		t.Error("the pump taps Mikaeus")
	}
}

// Cast for X=1, Mikaeus is a printed 0/0 with one counter. Paying the
// pump with it leaves a 0/0 with no counters, which dies (CR 704.5f) —
// before #625's review it stayed on the battlefield, skipped as a
// placeholder, and could tap to grow and pump again forever.
func TestMikaeusDiesPayingThePumpWithHisLastCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mik := b12PlayFromHand(t, g, "Mikaeus, the Lunarch", "Legendary Creature — Human Cleric", ccMikaeusOracle, game.CastSpellParams{XValue: 1})
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, mik); !ok || c.Toughness != 0 || counterCount(g, mik, game.CounterPlusOne) != 1 {
		t.Fatalf("want a printed 0/0 Mikaeus with one +1/+1 counter, got on battlefield %v %+v", ok, c.Counters)
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == mik {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	}
	bear := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)

	if err := g.ActivateCatalogAbility(me.ID, mik, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("pump with the last counter: %v", err)
	}
	if g.Battlefield.Contains(mik) {
		t.Fatal("Mikaeus survived as a 0/0 after spending his last counter")
	}
	if !me.Graveyard.Contains(mik) {
		t.Error("Mikaeus did not reach the graveyard")
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the pump still resolves: the bear has %d counters, want 1", got)
	}
	// Neither the pump nor an X=0 cast is stronger than printed any
	// more — #691 made a printed 0/0 with a printing behind it die at
	// once, and the X=0 caveat went with the gap
	// (TestMikaeusCastForXZeroDiesAtOnce). What is left is the
	// X-counter timing, and nothing here may declare an X=0 gap again.
	spec, _ := Lookup(ccMikaeusOracle)
	if len(spec.Caveats) != 1 {
		t.Errorf("want the X-counter timing caveat alone, got %v", spec.Caveats)
	}
	for _, c := range spec.Caveats {
		if strings.Contains(c, "X=0") {
			t.Errorf("an X=0 caveat is back and the engine no longer has that gap: %q", c)
		}
	}
}

// --- Benevolent Hydra ----------------------------------------------

func TestBenevolentHydraMovesACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hydra := pushCatalogPermanent(g, me.ID, "Benevolent Hydra", "Creature — Hydra", ccBenevolentHydra, false)
	setCounters(g, hydra, map[string]int{game.CounterPlusOne: 3})
	bear := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	advanceToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, hydra, 0, game.ActivateAbilityParams{Targets: b16TargetCard(hydra)}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("targeting itself: err = %v, want ErrIllegalTarget (\"another\")", err)
	}
	if got := counterCount(g, hydra, game.CounterPlusOne); got != 3 {
		t.Fatalf("a rejected activation removed a counter: %d", got)
	}

	b16Activate(t, g, me.ID, hydra, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)})
	if got := counterCount(g, hydra, game.CounterPlusOne); got != 2 {
		t.Errorf("the Hydra has %d counters, want 2", got)
	}
	// The placement is an ordinary one, so the Hydra's own "that many
	// plus one" applies to it.
	if got := counterCount(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("the bear has %d counters, want 2 (one, plus the Hydra's one)", got)
	}
}

// --- Fain, the Broker ----------------------------------------------

func TestFainRemovesACounterOfAnyKindForATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fain := b31Push(g, me.ID, "Fain, the Broker", "Legendary Creature — Human Warlock", ccFainTheBrokerOracle, "{2}{B}", 3, 3, "B")
	bear := pushVanillaCreature(g, me.ID, "Stunned Bear", 2, 2)
	setCounters(g, bear, map[string]int{game.CounterStun: 1, game.CounterPlusOne: 1})
	advanceToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, fain, 1, game.ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{bear}}); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("no kind named for \"a counter\": err = %v, want ErrInvalidParam", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, fain, 1, game.ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{fain}, CounterKind: game.CounterPlusOne}); !errors.Is(err, game.ErrInsufficientCounters) {
		t.Errorf("Fain has no counters: err = %v, want ErrInsufficientCounters", err)
	}
	if b31Tapped(t, g, fain) {
		t.Fatal("a rejected activation tapped Fain")
	}

	b16Activate(t, g, me.ID, fain, 1, game.ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{bear}, CounterKind: game.CounterStun})
	if got := counterCount(g, bear, game.CounterStun); got != 0 {
		t.Errorf("stun counters %d, want 0", got)
	}
	if got := counterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the +1/+1 counter was touched: %d", got)
	}
	if findBattlefieldByName(g, "Treasure") == uuid.Nil {
		t.Error("no Treasure token")
	}
	if !b31Tapped(t, g, fain) {
		t.Error("the ability taps Fain")
	}
}

// Fain can spend any creature's counter, so he reaches every printed
// 0/0 counter creature in the catalog. Taking the last +1/+1 counter off
// one kills it (CR 704.5f); it does not survive as a placeholder.
func TestFainSpendingAZeroZerosLastCounterKillsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fain := b31Push(g, me.ID, "Fain, the Broker", "Legendary Creature — Human Warlock", ccFainTheBrokerOracle, "{2}{B}", 3, 3, "B")
	walker := pushVanillaCreature(g, me.ID, "Zero Zero Construct", 0, 0)
	setCounters(g, walker, map[string]int{game.CounterPlusOne: 1})
	advanceToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, fain, 1, game.ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{walker}, CounterKind: game.CounterPlusOne}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(walker) {
		t.Fatal("the 0/0 survived losing its last +1/+1 counter to Fain")
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Treasure") == uuid.Nil {
		t.Error("no Treasure token")
	}
	if spec, _ := Lookup(ccFainTheBrokerOracle); spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Fain is complete: %s %v", spec.Completeness, spec.Caveats)
	}
}

// --- the printed text ----------------------------------------------

// TestCounterCostCardsMatchTheirOracleText checks each card's counter
// component against the printed cost it models, read from the Scryfall
// dump: the phrase is really on the card, and the component removes
// the kind and number the phrase says, from the permanent it says.
//
// Dump-gated like the other real-dump tests:
//
//	CMDCTRL_SCRYFALL_DUMP=data/scryfall/default-cards.json go test ./internal/cards/effects/ -run OracleText
func TestCounterCostCardsMatchTheirOracleText(t *testing.T) {
	path := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if path == "" {
		t.Skip("set CMDCTRL_SCRYFALL_DUMP to check the counter costs against the real oracle text")
	}
	idx := cards.NewIndex()
	if _, err := idx.Load(path); err != nil {
		t.Fatalf("load %s: %v", path, err)
	}
	cases := []struct {
		oracle   string
		name     string
		ability  int
		phrase   string
		kind     string
		fromThis bool
	}{
		{heartOfKiranOracle, "Heart of Kiran", 1, "remove a loyalty counter from a planeswalker you control rather than pay", game.CounterLoyalty, false},
		{ccDragonsHoardOracle, "Dragon's Hoard", 0, "{T}, Remove a gold counter from this artifact: Draw a card.", "gold", true},
		{ccMikaeusOracle, "Mikaeus, the Lunarch", 1, "{T}, Remove a +1/+1 counter from Mikaeus: Put a +1/+1 counter on each other creature you control.", game.CounterPlusOne, true},
		{ccBenevolentHydra, "Benevolent Hydra", 0, "{T}, Remove a +1/+1 counter from this creature: Put a +1/+1 counter on another target creature you control.", game.CounterPlusOne, true},
		{ccFainTheBrokerOracle, "Fain, the Broker", 1, "{T}, Remove a counter from a creature you control: Create a Treasure token.", "", false},
	}
	for _, tc := range cases {
		card, ok := idx.FindByOracleID(uuid.MustParse(tc.oracle))
		if !ok {
			t.Errorf("%s: oracle %s not in the dump", tc.name, tc.oracle)
			continue
		}
		if card.Name != tc.name {
			t.Errorf("oracle %s is %q in the dump, want %q", tc.oracle, card.Name, tc.name)
		}
		if !strings.Contains(card.OracleText, tc.phrase) {
			t.Errorf("%s: oracle text does not contain %q:\n%s", tc.name, tc.phrase, card.OracleText)
		}
		abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: tc.oracle})
		if tc.ability >= len(abilities) {
			t.Errorf("%s: no ability %d", tc.name, tc.ability)
			continue
		}
		ab := abilities[tc.ability]
		rc := ab.Cost.RemoveCounters
		if rc == nil {
			t.Errorf("%s ability %d (%q) has no counter component", tc.name, tc.ability, ab.Label)
			continue
		}
		if rc.Counter != tc.kind || rc.N != 1 || (rc.From == nil) != tc.fromThis {
			t.Errorf("%s: counter component %+v, want kind %q N 1 fromThis %v", tc.name, *rc, tc.kind, tc.fromThis)
		}
		// Every one of these but Heart's alternative prints {T}.
		if wantTap := strings.HasPrefix(tc.phrase, "{T}"); ab.Cost.Tap != wantTap {
			t.Errorf("%s: Tap = %v, want %v", tc.name, ab.Cost.Tap, wantTap)
		}
	}
}
