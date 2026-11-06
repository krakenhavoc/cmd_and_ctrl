package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dft_cg_cards_test.go — Aetherdrift (DFT), cards from C to G (#4464).
// Each test plays the real catalog entry and asserts what its text says.
// The fixtures are dft_ab_cards_test.go's, dft_st_cards_test.go's and
// dft_mp_cards_test.go's.

const (
	dftCGCaelornaOracle      = "4bbc5df9-60d7-46c1-8856-91c9c74abf18"
	dftCGCameraOracle        = "b46bac37-5b45-4be5-93d0-76e48dbd9f74"
	dftCGIroncladOracle      = "a08580da-e5aa-424b-86bf-d4535f3480d1"
	dftCGCrashOracle         = "38405a37-921a-49ab-a95b-28eb8c26474a"
	dftCGMechanicOracle      = "9b1d79b7-a812-4f48-9be0-789294da3b56"
	dftCGBeetleOracle        = "2b15d571-2f90-44f7-9153-20ed51253045"
	dftCGDiversionOracle     = "4d739edc-4e02-4b35-a2a0-e9685c6e2889"
	dftCGDraconauticsOracle  = "380d8d64-9aec-4656-b5b8-d336ebf803c8"
	dftCGEngineRatOracle     = "7e29aae9-815a-4999-ace6-7bc860acde32"
	dftCGFangGuardianOracle  = "1e0b9767-8e51-4021-a8fb-22259e2f5347"
	dftCGFuelOracle          = "e7d4e952-d7b9-489d-a9a7-c61d08bf17cc"
	dftCGGallantStrikeOracle = "04dcda3b-3401-4523-8b97-b700a14feb36"
	dftCGSerpentOracle       = "fdbd8a95-1dc8-4df2-bab0-a93d1941a405"
	dftCGBaubleOracle        = "431d6a2b-e22c-4921-a5f1-dd672f764f78"
)

func TestDftCGEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dftCGCaelornaOracle:      "Caelorna, Coral Tyrant",
		dftCGCameraOracle:        "Camera Launcher",
		dftCGIroncladOracle:      "Clamorous Ironclad",
		dftCGCrashOracle:         "Crash and Burn",
		dftCGMechanicOracle:      "Daring Mechanic",
		dftCGBeetleOracle:        "Debris Beetle",
		dftCGDiversionOracle:     "Diversion Unit",
		dftCGDraconauticsOracle:  "Draconautics Engineer",
		dftCGEngineRatOracle:     "Engine Rat",
		dftCGFangGuardianOracle:  "Fang Guardian",
		dftCGFuelOracle:          "Fuel the Flames",
		dftCGGallantStrikeOracle: "Gallant Strike",
		dftCGSerpentOracle:       "Gearseeker Serpent",
		dftCGBaubleOracle:        "Grim Bauble",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- fixtures ------------------------------------------------------------------

// dftCGCast casts a catalog card from the active player's hand with no
// payment, announcing the given modes and targets, and lets it resolve.
func dftCGCast(t *testing.T, g *game.Game, me *game.Player, c game.Card, modes []int, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = me.ID, me.ID
	me.Hand.PushTop(c)
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Modes: modes, Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	passPriorityAroundTable(t, g)
	return c.InstanceID
}

// dftCGCastErr is dftCGCast for a cast that is supposed to be refused:
// the announce error comes back instead of failing the test.
func dftCGCastErr(g *game.Game, me *game.Player, c game.Card, modes []int, targets []game.TargetRef) error {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = me.ID, me.ID
	me.Hand.PushTop(c)
	return g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Modes: modes, Targets: targets})
}

// dftCGPickRefused asserts that the open target pick refuses each of
// the given cards.
func dftCGPickRefused(t *testing.T, g *game.Game, chooser uuid.UUID, bad ...uuid.UUID) {
	t.Helper()
	pick := latestPickTarget(g, chooser)
	if pick == nil {
		t.Fatal("no target pick")
	}
	for _, id := range bad {
		if err := g.ResolvePickTarget(pick.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: id}); err == nil {
			t.Fatalf("%s was a legal target", id)
		}
	}
}

// --- Caelorna, Coral Tyrant -----------------------------------------------------

func TestDftCGCaelornaResolvesAsAPlainCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := dftABCast(t, g, me, game.Card{Name: "Caelorna, Coral Tyrant", TypeLine: "Legendary Creature — Octopus",
		OracleID: dftCGCaelornaOracle, Power: 0, Toughness: 8})
	if !g.Battlefield.Contains(id) {
		t.Fatal("Caelorna did not enter")
	}
	if p, tough := effectivePTOf(t, g, id); p != 0 || tough != 8 {
		t.Errorf("Caelorna is %d/%d, want 0/8", p, tough)
	}
	if len(g.PendingChoices) != 0 || g.Stack.Size() != 0 {
		t.Errorf("Caelorna left something behind: choices %v, stack %d", g.PendingChoices, g.Stack.Size())
	}
}

// --- Camera Launcher --------------------------------------------------------------

func TestDftCGCameraLauncherExhaustsForACounterAndAThopterOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	launcher := dftABPush(g, me.ID, "Camera Launcher", "Artifact Creature — Construct", dftCGCameraOracle, 2, 2)
	advanceToMain(t, g)
	before := len(g.Battlefield.Cards)
	fillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, launcher, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, launcher, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if got := len(g.Battlefield.Cards) - before; got != 1 {
		t.Fatalf("%d permanents made, want one Thopter", got)
	}
	thopter := g.Battlefield.Cards[len(g.Battlefield.Cards)-1]
	if thopter.Name != "Thopter" || !thopter.IsArtifact() || !thopter.IsCreature() {
		t.Errorf("the token is %q (%q), want a Thopter artifact creature", thopter.Name, thopter.TypeLine)
	}
	if !hasEffectiveKeyword(t, g, thopter.InstanceID, "flying") {
		t.Error("the Thopter has no flying")
	}
	fillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, launcher, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the exhaust ability was activated twice")
	}
}

// --- Clamorous Ironclad -----------------------------------------------------------

func TestDftCGClamorousIroncladHasMenaceCrewsThreeAndCycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ironclad := dftABPush(g, me.ID, "Clamorous Ironclad", dftABVehicle, dftCGIroncladOracle, 6, 3)
	if !hasEffectiveKeyword(t, g, ironclad, "menace") {
		t.Error("Clamorous Ironclad has no menace")
	}
	two := b12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	one := b12Creature(g, me.ID, "One", "Creature — Rat", 1, 1)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, ironclad, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{two}}); err == nil {
		t.Fatal("crewed with total power 2; crew 3 wants 3")
	}
	dftABCrew(t, g, me, ironclad, 0, two, one)
	if p, tough := effectivePTOf(t, g, ironclad); p != 6 || tough != 3 {
		t.Errorf("the crewed Ironclad is %d/%d, want 6/3", p, tough)
	}

	// Cycling {R}: colourless mana can't pay it.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	advanceToMain(t, g2)
	seedLibrary(me2, "Drawn")
	id := pushCatalogHandCard(me2, "Clamorous Ironclad", dftABVehicle, dftCGIroncladOracle)
	fillPool(me2, 1)
	if err := g2.ActivateCatalogAbility(me2.ID, id, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("cycled for colourless mana; cycling costs {R}")
	}
	fillPoolColored(me2, "R", 1)
	hand := me2.Hand.Size()
	if err := g2.ActivateCatalogAbility(me2.ID, id, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if !me2.Graveyard.Contains(id) || me2.Hand.Size() != hand {
		t.Errorf("cycling did not discard the Ironclad for a card: graveyard %v, hand %d want %d",
			me2.Graveyard.Contains(id), me2.Hand.Size(), hand)
	}
}

// --- Crash and Burn ---------------------------------------------------------------

func TestDftCGCrashAndBurnDestroysAnUncrewedVehicle(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vehicle := dftABPush(g, opp.ID, "Their Vehicle", dftABVehicle, "", 3, 3)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if err := dftCGCastErr(g, me, game.Card{Name: "Crash and Burn", TypeLine: "Instant", OracleID: dftCGCrashOracle},
		[]int{0}, dftABCardRef(bear)); err == nil {
		t.Fatal("the destroy bullet took a creature that is not a Vehicle")
	}
	dftCGCast(t, g, me, game.Card{Name: "Crash and Burn", TypeLine: "Instant", OracleID: dftCGCrashOracle},
		[]int{0}, dftABCardRef(vehicle))
	if g.Battlefield.Contains(vehicle) {
		t.Error("the Vehicle was not destroyed")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bear was destroyed")
	}
}

func TestDftCGCrashAndBurnDealsSixToACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	big := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 7)
	dftCGCast(t, g, me, game.Card{Name: "Crash and Burn", TypeLine: "Instant", OracleID: dftCGCrashOracle},
		[]int{1}, dftABCardRef(big))
	if got := damageOn(t, g, big); got != 6 {
		t.Errorf("%d damage on the Giant, want 6", got)
	}
}

// --- Daring Mechanic --------------------------------------------------------------

func TestDftCGDaringMechanicCountersAMountOrVehicleOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mechanic := dftABPush(g, me.ID, "Daring Mechanic", "Creature — Human Artificer", dftCGMechanicOracle, 3, 3)
	vehicle := dftABPush(g, me.ID, "My Vehicle", dftABVehicle, "", 3, 3)
	mount := b12Creature(g, opp.ID, "Their Mount", "Creature — Horse Mount", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	pay := func() {
		fillPool(me, 3)
		fillPoolColored(me, "W", 1)
	}
	pay()
	if err := g.ActivateCatalogAbility(me.ID, mechanic, 0, game.ActivateAbilityParams{Strict: true, Targets: dftABCardRef(bear)}); err == nil {
		t.Fatal("a creature that is neither a Mount nor a Vehicle was a legal target")
	}
	for _, id := range []uuid.UUID{vehicle, mount} {
		pay()
		if err := g.ActivateCatalogAbility(me.ID, mechanic, 0, game.ActivateAbilityParams{Strict: true, Targets: dftABCardRef(id)}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		passPriorityAroundTable(t, g)
		if got := countersOn(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("%d +1/+1 counters on %s, want 1", got, id)
		}
	}
}

// --- Debris Beetle ----------------------------------------------------------------

func TestDftCGDebrisBeetleDrainsEachOpponentForThreeAndCrewsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	lives, life := lifeOfOpponents(g), me.Life
	beetle := dftABCast(t, g, me, game.Card{Name: "Debris Beetle", TypeLine: dftABVehicle,
		OracleID: dftCGBeetleOracle, Power: 6, Toughness: 6})
	for i, now := range lifeOfOpponents(g) {
		if now != lives[i]-3 {
			t.Errorf("opponent %d: life %d, want %d", i, now, lives[i]-3)
		}
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
	if !hasEffectiveKeyword(t, g, beetle, "trample") {
		t.Error("Debris Beetle has no trample")
	}
	if c, _ := battlefieldCard(g, beetle); c.IsCreature() {
		t.Error("the uncrewed Beetle is a creature")
	}
	one := b12Creature(g, me.ID, "One", "Creature — Rat", 1, 1)
	other := b12Creature(g, me.ID, "Other", "Creature — Rat", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, beetle, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{one}}); err == nil {
		t.Fatal("crewed with total power 1; crew 2 wants 2")
	}
	dftABCrew(t, g, me, beetle, 0, one, other)
}

// --- Diversion Unit ---------------------------------------------------------------

func TestDftCGDiversionUnitTaxesAnInstantOrSorcerySpell(t *testing.T) {
	for _, pay := range []bool{false, true} {
		name := "declined"
		if pay {
			name = "paid"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			unit := dftABPush(g, me.ID, "Diversion Unit", "Artifact Creature — Robot", dftCGDiversionOracle, 2, 1)
			if !hasEffectiveKeyword(t, g, unit, "flying") {
				t.Error("Diversion Unit has no flying")
			}
			creature := castCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
			fillPoolColored(me, "U", 1)
			if err := g.ActivateCatalogAbility(me.ID, unit, 0, game.ActivateAbilityParams{Strict: true, Targets: dftABCardRef(creature)}); err == nil {
				t.Fatal("a creature spell was a legal target")
			}
			// An instant, since the creature spell is still on the stack.
			spell := castCatalogSpell(t, g, "An Instant", "Instant", "", nil)
			if err := g.ActivateCatalogAbility(me.ID, unit, 0, game.ActivateAbilityParams{Strict: true, Targets: dftABCardRef(spell)}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			if g.Battlefield.Contains(unit) {
				t.Error("the Unit was not sacrificed as part of the cost")
			}
			passUntilTaxed(t, g, me.ID)
			if pay {
				fillPool(me, 3)
			}
			answerPayUnless(t, g, me.ID, pay)
			if g.Stack.Contains(spell) == !pay {
				t.Errorf("paid=%v: the instant is still on the stack = %v", pay, g.Stack.Contains(spell))
			}
		})
	}
}

// --- Draconautics Engineer --------------------------------------------------------

func TestDftCGDraconauticsEngineerHastesTheOthersAndGrowsOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	engineer := dftABPush(g, me.ID, "Draconautics Engineer", "Creature — Goblin Artificer", dftCGDraconauticsOracle, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	fillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, engineer, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("my other creature did not gain haste")
	}
	if hasEffectiveKeyword(t, g, engineer, "haste") {
		t.Error("the Engineer gave itself haste; the text says other creatures")
	}
	if hasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature gained haste")
	}
	if got := countersOn(g, engineer, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	fillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, engineer, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the first exhaust ability was activated twice")
	}
	dftMPPassTurn(t, g)
	if hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the haste outlasted the turn")
	}
}

func TestDftCGDraconauticsEngineerMakesADinosaurDragonOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	engineer := dftABPush(g, me.ID, "Draconautics Engineer", "Creature — Goblin Artificer", dftCGDraconauticsOracle, 2, 2)
	advanceToMain(t, g)
	before := len(g.Battlefield.Cards)
	pay := func() {
		fillPool(me, 3)
		fillPoolColored(me, "R", 1)
	}
	pay()
	if err := g.ActivateCatalogAbility(me.ID, engineer, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := len(g.Battlefield.Cards) - before; got != 1 {
		t.Fatalf("%d permanents made, want one Dragon", got)
	}
	dragon := g.Battlefield.Cards[len(g.Battlefield.Cards)-1]
	if dragon.Name != "Dinosaur Dragon" || !dragon.HasSubtype("Dinosaur") || !dragon.HasSubtype("Dragon") {
		t.Errorf("the token is %q (%q), want a Dinosaur Dragon", dragon.Name, dragon.TypeLine)
	}
	if p, tough := effectivePTOf(t, g, dragon.InstanceID); p != 4 || tough != 4 {
		t.Errorf("the Dragon is %d/%d, want 4/4", p, tough)
	}
	if !hasEffectiveKeyword(t, g, dragon.InstanceID, "flying") {
		t.Error("the Dragon has no flying")
	}
	if len(dragon.Colors) != 1 || dragon.Colors[0] != "R" {
		t.Errorf("the Dragon's colours are %v, want red", dragon.Colors)
	}
	if got := countersOn(g, engineer, game.CounterPlusOne); got != 0 {
		t.Errorf("the second exhaust put %d counters on the Engineer", got)
	}
	pay()
	if err := g.ActivateCatalogAbility(me.ID, engineer, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the second exhaust ability was activated twice")
	}
	// Spending the second exhaust left the first one unspent.
	fillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, engineer, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Errorf("the first exhaust was spent by the second: %v", err)
	}
}

// --- Engine Rat -------------------------------------------------------------------

func TestDftCGEngineRatDrainsEachOpponentForTwoPerActivation(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rat := dftABPush(g, me.ID, "Engine Rat", "Creature — Zombie Rat", dftCGEngineRatOracle, 1, 1)
	if !hasEffectiveKeyword(t, g, rat, "deathtouch") {
		t.Error("Engine Rat has no deathtouch")
	}
	advanceToMain(t, g)
	lives, life := lifeOfOpponents(g), me.Life
	for i := 0; i < 2; i++ {
		fillPool(me, 5)
		fillPoolColored(me, "B", 1)
		if err := g.ActivateCatalogAbility(me.ID, rat, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activation %d: %v", i+1, err)
		}
		passPriorityAroundTable(t, g)
	}
	for i, now := range lifeOfOpponents(g) {
		if now != lives[i]-4 {
			t.Errorf("opponent %d: life %d after two activations, want %d", i, now, lives[i]-4)
		}
	}
	if me.Life != life {
		t.Errorf("my life changed: %d -> %d", life, me.Life)
	}
}

// --- Fang Guardian ----------------------------------------------------------------

func TestDftCGFangGuardianPumpsAnotherCreatureOrVehicleYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	// A real catalog Vehicle, so it has a crew ability to animate it with.
	vehicle := dftABPush(g, me.ID, "Debris Beetle", dftABVehicle, dftCGBeetleOracle, 6, 6)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	guardian := dftABCast(t, g, me, game.Card{Name: "Fang Guardian", TypeLine: "Creature — Ape Druid",
		OracleID: dftCGFangGuardianOracle, Power: 4, Toughness: 2})
	if !hasEffectiveKeyword(t, g, guardian, "flash") {
		t.Error("Fang Guardian has no flash")
	}
	dftCGPickRefused(t, g, me.ID, guardian, theirs)
	pickTriggerTarget(t, g, me.ID, vehicle)
	passPriorityAroundTable(t, g)
	dftABCrew(t, g, me, vehicle, 0, bear)
	if p, tough := effectivePTOf(t, g, vehicle); p != 8 || tough != 8 {
		t.Errorf("the crewed Vehicle is %d/%d, want 8/8 with the boost", p, tough)
	}
	dftMPPassTurn(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want an untouched 2/2", p, tough)
	}
}

func TestDftCGFangGuardianPumpsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	dftABCast(t, g, me, game.Card{Name: "Fang Guardian", TypeLine: "Creature — Ape Druid",
		OracleID: dftCGFangGuardianOracle, Power: 4, Toughness: 2})
	pickTriggerTarget(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tough)
	}
	dftMPPassTurn(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the Bear is %d/%d after the turn ended, want 2/2", p, tough)
	}
}

// --- Fuel the Flames --------------------------------------------------------------

func TestDftCGFuelTheFlamesDealsTwoToEachCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "My Giant", "Creature — Giant", 3, 5)
	theirs := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 3, 5)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	vehicle := dftABPush(g, opp.ID, "Their Vehicle", dftABVehicle, "", 3, 3)
	dftCGCast(t, g, me, game.Card{Name: "Fuel the Flames", TypeLine: "Instant", OracleID: dftCGFuelOracle}, nil, nil)
	for _, id := range []uuid.UUID{mine, theirs} {
		if got := damageOn(t, g, id); got != 2 {
			t.Errorf("%d damage on %s, want 2", got, id)
		}
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the 2/2 survived 2 damage")
	}
	if got := damageOn(t, g, vehicle); got != 0 {
		t.Errorf("the uncrewed Vehicle took %d damage; it is not a creature", got)
	}
	dftSTCycles(t, "Fuel the Flames", "Instant", dftCGFuelOracle, 0)
}

// --- Gallant Strike ---------------------------------------------------------------

func TestDftCGGallantStrikeDestroysOnlyToughnessFourOrGreater(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	small := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 3)
	big := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
	advanceToMain(t, g)
	if err := castCatalogSpellErr(t, g, "Gallant Strike", "Instant", dftCGGallantStrikeOracle, dftABCardRef(small)); err == nil {
		t.Fatal("a creature with toughness 3 was a legal target")
	}
	dftCGCast(t, g, me, game.Card{Name: "Gallant Strike", TypeLine: "Instant", OracleID: dftCGGallantStrikeOracle},
		nil, dftABCardRef(big))
	if g.Battlefield.Contains(big) {
		t.Error("the toughness 4 creature was not destroyed")
	}
	if !g.Battlefield.Contains(small) {
		t.Error("the other creature was destroyed")
	}
	dftSTCycles(t, "Gallant Strike", "Instant", dftCGGallantStrikeOracle, 0)
}

// --- Gearseeker Serpent -----------------------------------------------------------

func TestDftCGGearseekerSerpentHasAffinityAndCanBeMadeUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 4; i++ {
		pushPermanent(g, me.ID, game.Card{Name: "Signet", TypeLine: "Artifact"})
	}
	pushPermanent(g, opp.ID, game.Card{Name: "Their Signet", TypeLine: "Artifact"})
	const typeLine = "Creature — Serpent"
	if got := selfPricedMV(t, g, me, dftCGSerpentOracle, typeLine, "{5}{U}{U}"); got != 3 {
		t.Errorf("with four artifacts it costs mana value %d, want 3 ({1}{U}{U})", got)
	}
	for i := 0; i < 3; i++ {
		pushPermanent(g, me.ID, game.Card{Name: "Signet", TypeLine: "Artifact"})
	}
	if got := selfPricedMV(t, g, me, dftCGSerpentOracle, typeLine, "{5}{U}{U}"); got != 2 {
		t.Errorf("with seven artifacts it costs mana value %d, want 2 ({U}{U} is never reduced)", got)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	serpent := dftABPush(g2, me2.ID, "Gearseeker Serpent", typeLine, dftCGSerpentOracle, 5, 6)
	advanceToMain(t, g2)
	fillPool(me2, 5)
	fillPoolColored(me2, "U", 1)
	if err := g2.ActivateCatalogAbility(me2.ID, serpent, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if restrictionsOf(t, g2, serpent)&game.CantBeBlocked == 0 {
		t.Error("the Serpent can be blocked after its ability resolved")
	}
	dftMPPassTurn(t, g2)
	if restrictionsOf(t, g2, serpent)&game.CantBeBlocked != 0 {
		t.Error("the Serpent's evasion outlasted the turn")
	}
}

// --- Grim Bauble ------------------------------------------------------------------

func TestDftCGGrimBaubleShrinksAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 3, 3)
	advanceToMain(t, g)
	dftABCast(t, g, me, game.Card{Name: "Grim Bauble", TypeLine: "Artifact", OracleID: dftCGBaubleOracle})
	dftCGPickRefused(t, g, me.ID, mine)
	pickTriggerTarget(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("their creature is %d/%d, want 1/1", p, tough)
	}
	dftMPPassTurn(t, g)
	if p, tough := effectivePTOf(t, g, theirs); p != 3 || tough != 3 {
		t.Errorf("their creature is %d/%d after the turn ended, want 3/3", p, tough)
	}
}

func TestDftCGGrimBaubleSacrificesToSurveilTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bauble := dftABPush(g, me.ID, "Grim Bauble", "Artifact", dftCGBaubleOracle, 0, 0)
	advanceToMain(t, g)
	seedLibrary(me, "Top A", "Top B")
	fillPool(me, 2)
	fillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, bauble, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(bauble) {
		t.Error("the Bauble was not sacrificed as part of the cost")
	}
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("no surveil 2: %+v", c)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	for _, id := range c.ScryCards {
		if !me.Graveyard.Contains(id) {
			t.Errorf("surveilled card %s is not in the graveyard", id)
		}
	}
}
