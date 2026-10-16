package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// entwine_cards_c_test.go — the third batch of entwine cards (CR 702.42,
// #3866): each card's entwined cast resolves every bullet, in the
// printed order (CR 702.42b), and a non-trivial bullet on its own.

const (
	ewcRudeAwakeningOracle      = "685a7f67-6809-4386-990f-c52d1846a584"
	ewcSavageBeatingOracle      = "89d1cd45-8148-4396-a9b8-e0f495e94b5b"
	ewcSecondSightOracle        = "1a4e948a-5a5e-4811-a6af-7ea93a051746"
	ewcSolarTideOracle          = "0f7a7348-3081-4fbb-9386-719a4ae7fa57"
	ewcStirThePrideOracle       = "7e46d2c7-59ae-4b0e-a2ec-d0ac1b68d594"
	ewcTemporalCascadeOracle    = "c65dc9cb-3aad-4d15-a8c0-c0c142058d45"
	ewcToothAndNailOracle       = "382b7873-e78d-473b-8346-d92d16e8f634"
	ewcUnboundedPotentialOracle = "d9ae08cf-eef0-407a-ad2c-08b4c89f99f0"
	ewcWailOfTheNimOracle       = "dcf503b5-c7ef-49e5-87ed-3b8758197446"
)

// ewcLand puts a basic land (no power or toughness of its own) onto the
// battlefield under `owner`.
func ewcLand(g *game.Game, owner uuid.UUID, tapped bool) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
		Tapped: tapped, Owner: owner, Controller: owner,
	})
}

func ewcOnBattlefield(g *game.Game, id uuid.UUID) bool {
	return findBattlefieldCardForTest(g, id) != nil
}

func ewcInGraveyard(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Graveyard.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func ewcCast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, oracle string, colors []string,
	pool string, modes []int, entwined bool, targets []game.TargetRef) {
	t.Helper()
	if _, err := castEntwineSpell(g, me, name, typeLine, cost, oracle, colors, pool, modes, entwined, targets); err != nil {
		t.Fatalf("%s %v (entwined %v): %v", name, modes, entwined, err)
	}
}

// --- Rude Awakening ----------------------------------------------------

func TestEwcRudeAwakeningEntwinedUntapsAndAnimatesYourLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	mine := ewcLand(g, me.ID, true)
	theirs := ewcLand(g, opp.ID, true)
	ewcCast(t, g, me, "Rude Awakening", "Sorcery", "{4}{G}", ewcRudeAwakeningOracle, []string{"G"},
		"{C}{C}{C}{C}{C}{C}{G}{G}", []int{0, 1}, true, nil)
	passPriorityAroundTable(t, g)
	if battlefieldTapped(g, mine) {
		t.Error("your land is untapped")
	}
	if !battlefieldTapped(g, theirs) {
		t.Error("an opponent's land is not yours to untap")
	}
	c := layeredCard(t, g, mine)
	if !c.IsCreature() || !c.IsLand() || c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("your land is a %d/%d (creature %v, land %v), want a 2/2 creature land", c.CurrentPower(), c.CurrentToughness(), c.IsCreature(), c.IsLand())
	}
	if layeredCard(t, g, theirs).IsCreature() {
		t.Error("an opponent's land became a creature")
	}
}

func TestEwcRudeAwakeningUnentwinedOnlyUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	mine := ewcLand(g, me.ID, true)
	ewcCast(t, g, me, "Rude Awakening", "Sorcery", "{4}{G}", ewcRudeAwakeningOracle, []string{"G"},
		"{C}{C}{C}{C}{G}", []int{0}, false, nil)
	passPriorityAroundTable(t, g)
	if battlefieldTapped(g, mine) {
		t.Error("your land is untapped")
	}
	if layeredCard(t, g, mine).IsCreature() {
		t.Error("the animate bullet was not chosen")
	}
}

// --- Savage Beating ------------------------------------------------------

func ewcCombatSteps(g *game.Game) int {
	n := 0
	for _, s := range g.TurnPlan {
		if s.Step == game.StepBeginCombat {
			n++
		}
	}
	return n
}

func TestEwcSavageBeatingCastsOnlyDuringCombatOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	if _, err := castEntwineSpell(g, me, "Savage Beating", "Instant", "{3}{R}{R}", ewcSavageBeatingOracle, []string{"R"},
		"{C}{C}{C}{R}{R}", []int{0}, false, nil); err == nil {
		t.Fatal("cast in a main phase: refused")
	}
	ssToBeginCombat(t, g)
	if _, err := castEntwineSpell(g, me, "Savage Beating", "Instant", "{3}{R}{R}", ewcSavageBeatingOracle, []string{"R"},
		"{C}{C}{C}{R}{R}", []int{0}, false, nil); err != nil {
		t.Fatalf("cast during combat on your turn: %v", err)
	}
}

func TestEwcSavageBeatingEntwinedDoubleStrikeUntapAndExtraCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	ssToBeginCombat(t, g)
	findBattlefieldCardForTest(g, mine).Tapped = true
	before := ewcCombatSteps(g)
	ewcCast(t, g, me, "Savage Beating", "Instant", "{3}{R}{R}", ewcSavageBeatingOracle, []string{"R"},
		"{C}{C}{C}{C}{R}{R}{R}", []int{1, 0}, true, nil)
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, mine)
	if !game.HasKeyword(&c, "double strike") {
		t.Error("your creature has double strike")
	}
	o := layeredCard(t, g, theirs)
	if game.HasKeyword(&o, "double strike") {
		t.Error("an opponent's creature gained double strike")
	}
	if battlefieldTapped(g, mine) {
		t.Error("your creature is untapped")
	}
	if after := ewcCombatSteps(g); after != before+1 {
		t.Errorf("%d combat phases still to come, want %d", after, before+1)
	}
}

func TestEwcSavageBeatingUnentwinedAddsNoCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	ssToBeginCombat(t, g)
	before := ewcCombatSteps(g)
	ewcCast(t, g, me, "Savage Beating", "Instant", "{3}{R}{R}", ewcSavageBeatingOracle, []string{"R"},
		"{C}{C}{C}{R}{R}", []int{0}, false, nil)
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, mine)
	if !game.HasKeyword(&c, "double strike") {
		t.Error("your creature has double strike")
	}
	if after := ewcCombatSteps(g); after != before {
		t.Errorf("%d combat phases to come, want %d", after, before)
	}
}

// --- Second Sight ----------------------------------------------------------

func TestEwcSecondSightEntwinedOrdersTheOpponentsTopThenYours(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	theirs := seedLibrary(opp, "A", "B", "C", "D", "E", "F")
	mine := seedLibrary(me, "M1", "M2", "M3", "M4", "M5", "M6")
	ewcCast(t, g, me, "Second Sight", "Instant", "{2}{U}", ewcSecondSightOracle, []string{"U"},
		"{C}{C}{U}{U}", []int{1, 0}, true,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID, Mode: 1}})
	passPriorityAroundTable(t, g)

	first := putInLibraryChoiceFor(g, me.ID)
	if first == nil || !sameIDs(first.ScryCards, theirs[:5]) {
		t.Fatalf("the opponent's top five are ordered first (CR 702.42b): %+v", first)
	}
	wantTheirs := []uuid.UUID{theirs[4], theirs[3], theirs[2], theirs[1], theirs[0]}
	if err := g.ResolvePutInLibrary(first.ID, me.ID, nil, wantTheirs); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	second := putInLibraryChoiceFor(g, me.ID)
	if second == nil || !sameIDs(second.ScryCards, mine[:5]) {
		t.Fatalf("then your own top five: %+v", second)
	}
	wantMine := []uuid.UUID{mine[1], mine[0], mine[2], mine[3], mine[4]}
	if err := g.ResolvePutInLibrary(second.ID, me.ID, nil, wantMine); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	g.SettleResolution()
	if got := libraryTopIDs(opp, 6); !sameIDs(got, append(append([]uuid.UUID(nil), wantTheirs...), theirs[5])) {
		t.Errorf("opponent's top six %v", got)
	}
	if got := libraryTopIDs(me, 6); !sameIDs(got, append(append([]uuid.UUID(nil), wantMine...), mine[5])) {
		t.Errorf("your top six %v", got)
	}
}

// --- Solar Tide ------------------------------------------------------------

func ewcCastSolarTide(g *game.Game, me *game.Player, modes []int, entwined bool, sacrifice []uuid.UUID) error {
	id := handCardFull(me, "Solar Tide", "Sorcery", "{4}{W}{W}", ewcSolarTideOracle, []string{"W"})
	me.ManaPool.EmptyPool()
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}{W}{W}"); err != nil {
		return err
	}
	var optional []int
	if entwined {
		optional = []int{0}
	}
	err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict: true, Modes: modes, OptionalCosts: optional, SacrificeIDs: sacrifice,
	})
	if err != nil {
		_, _ = me.Hand.Remove(id)
		me.ManaPool.EmptyPool()
	}
	return err
}

func TestEwcSolarTideEntwinedSacrificesTwoLandsAndSweepsBoth(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	small := pushVanillaCreature(g, opp.ID, "Small", 2, 2)
	big := pushVanillaCreature(g, opp.ID, "Big", 4, 4)
	a, b := ewcLand(g, me.ID, false), ewcLand(g, me.ID, false)
	if err := ewcCastSolarTide(g, me, []int{0, 1}, true, []uuid.UUID{a}); err == nil {
		t.Fatal("entwined naming one land: the cost is two")
	}
	if err := ewcCastSolarTide(g, me, []int{0, 1}, true, nil); err == nil {
		t.Fatal("entwined naming no land: refused")
	}
	if err := ewcCastSolarTide(g, me, []int{0, 1}, true, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("entwined, two lands: %v", err)
	}
	if ewcOnBattlefield(g, a) || ewcOnBattlefield(g, b) {
		t.Fatal("the two lands are sacrificed as the spell is cast")
	}
	if !ewcInGraveyard(me, a) || !ewcInGraveyard(me, b) {
		t.Error("the sacrificed lands are in your graveyard")
	}
	passPriorityAroundTable(t, g)
	if ewcOnBattlefield(g, small) || ewcOnBattlefield(g, big) {
		t.Error("entwined: both sweeps destroy")
	}
}

func TestEwcSolarTideEntwineRefusedWithOneLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	pushVanillaCreature(g, opp.ID, "Small", 2, 2)
	only := ewcLand(g, me.ID, false)
	if err := ewcCastSolarTide(g, me, []int{0, 1}, true, []uuid.UUID{only, only}); err == nil {
		t.Fatal("one land named twice cannot pay two")
	}
	if err := ewcCastSolarTide(g, me, []int{0}, false, nil); err != nil {
		t.Fatalf("unentwined, one bullet: %v", err)
	}
}

func TestEwcSolarTideUnentwinedDestroysOnlyTheBigOnes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	small := pushVanillaCreature(g, opp.ID, "Small", 2, 2)
	big := pushVanillaCreature(g, opp.ID, "Big", 3, 3)
	if err := ewcCastSolarTide(g, me, []int{1}, false, nil); err != nil {
		t.Fatalf("unentwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !ewcOnBattlefield(g, small) {
		t.Error("power 2 survives the power-3-or-greater sweep")
	}
	if ewcOnBattlefield(g, big) {
		t.Error("power 3 is destroyed")
	}
}

// --- Stir the Pride ----------------------------------------------------------

func TestEwcStirThePrideEntwinedPumpsAndGrantsTheLifeTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Wall", 0, 10)
	ewcCast(t, g, me, "Stir the Pride", "Instant", "{4}{W}", ewcStirThePrideOracle, []string{"W"},
		"{C}{C}{C}{C}{C}{W}{W}", []int{0, 1}, true, nil)
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, mine)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("your bear is %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	if o := layeredCard(t, g, theirs); o.CurrentPower() != 0 {
		t.Error("an opponent's creature was pumped")
	}
	life := me.Life
	pr6Damage(t, g, mine, theirs, 3)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d — the granted trigger gains the damage dealt", me.Life, life+3)
	}
	// The opponent's creature did not gain it.
	oppLife := opp.Life
	pr6Damage(t, g, theirs, mine, 1)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife {
		t.Error("an opponent's creature has the trigger")
	}
}

func TestEwcStirThePrideUnentwinedGrantOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Wall", 0, 10)
	ewcCast(t, g, me, "Stir the Pride", "Instant", "{4}{W}", ewcStirThePrideOracle, []string{"W"},
		"{C}{C}{C}{C}{W}", []int{1}, false, nil)
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, mine); c.CurrentPower() != 2 {
		t.Error("the pump bullet was not chosen")
	}
	life := me.Life
	pr6Damage(t, g, mine, theirs, 2)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

// --- Temporal Cascade ----------------------------------------------------------

func TestEwcTemporalCascadeEntwinedShufflesThenDrawsSeven(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	dead := b17GraveyardCard(opp, "Dead Bear", "Creature — Bear", "{1}{G}")
	ewcCast(t, g, me, "Temporal Cascade", "Sorcery", "{5}{U}{U}", ewcTemporalCascadeOracle, []string{"U"},
		"{C}{C}{C}{C}{C}{C}{C}{U}{U}", []int{1, 0}, true, nil)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.Hand.Size() != 7 {
			t.Errorf("seat %s holds %d cards, want 7", p.Name, p.Hand.Size())
		}
		if p.Graveyard.Size() != 0 && p.ID != me.ID {
			t.Errorf("seat %s's graveyard has %d cards, want 0", p.Name, p.Graveyard.Size())
		}
	}
	if ewcInGraveyard(opp, dead) {
		t.Error("the opponent's graveyard was shuffled away")
	}
	// The spell itself goes to your graveyard after resolving.
	if n := me.Graveyard.Size(); n != 1 {
		t.Errorf("your graveyard holds %d, want only Temporal Cascade", n)
	}
}

func TestEwcTemporalCascadeUnentwinedShufflesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	b17GraveyardCard(opp, "Dead Bear", "Creature — Bear", "{1}{G}")
	oppLib := opp.Library.Size() + opp.Hand.Size() + opp.Graveyard.Size()
	ewcCast(t, g, me, "Temporal Cascade", "Sorcery", "{5}{U}{U}", ewcTemporalCascadeOracle, []string{"U"},
		"{C}{C}{C}{C}{C}{U}{U}", []int{0}, false, nil)
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != 0 || opp.Graveyard.Size() != 0 {
		t.Errorf("opponent: hand %d, graveyard %d — both shuffled in, nobody draws", opp.Hand.Size(), opp.Graveyard.Size())
	}
	if opp.Library.Size() != oppLib {
		t.Errorf("opponent's library %d, want %d", opp.Library.Size(), oppLib)
	}
}

// --- Tooth and Nail ----------------------------------------------------------

func TestEwcToothAndNailEntwinedSearchesThenPutsThemOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	a := pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Big One", TypeLine: "Creature — Beast", Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID})
	b := pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Big Two", TypeLine: "Creature — Beast", Power: 7, Toughness: 7, Owner: me.ID, Controller: me.ID})
	ewcCast(t, g, me, "Tooth and Nail", "Sorcery", "{5}{G}{G}", ewcToothAndNailOracle, []string{"G"},
		"{C}{C}{C}{C}{C}{C}{C}{G}{G}", []int{1, 0}, true, nil)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("the put waits for the search (CR 702.42b)")
	}
	answerSearchByID(t, g, me.ID, a, b)
	answerChooseCards(t, g, me.ID, a, b)
	g.SettleResolution()
	if !ewcOnBattlefield(g, a) || !ewcOnBattlefield(g, b) {
		t.Error("both found creatures are put onto the battlefield")
	}
}

func TestEwcToothAndNailUnentwinedPutsUpToTwoFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	a := handCardForTest(me, "Hand One", "Creature — Beast", "")
	b := handCardForTest(me, "Hand Two", "Creature — Beast", "")
	c := handCardForTest(me, "Hand Three", "Creature — Beast", "")
	ewcCast(t, g, me, "Tooth and Nail", "Sorcery", "{5}{G}{G}", ewcToothAndNailOracle, []string{"G"},
		"{C}{C}{C}{C}{C}{G}{G}", []int{1}, false, nil)
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) != nil {
		t.Fatal("the search bullet was not chosen")
	}
	answerChooseCards(t, g, me.ID, a, c)
	g.SettleResolution()
	if !ewcOnBattlefield(g, a) || !ewcOnBattlefield(g, c) || ewcOnBattlefield(g, b) {
		t.Error("exactly the two chosen creatures enter")
	}
}

// --- Unbounded Potential ----------------------------------------------------------

func TestEwcUnboundedPotentialEntwinedCountersThenProliferate(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	a := pushVanillaCreature(g, me.ID, "A", 1, 1)
	b := pushVanillaCreature(g, me.ID, "B", 1, 1)
	ewcCast(t, g, me, "Unbounded Potential", "Instant", "{1}{W}", ewcUnboundedPotentialOracle, []string{"W"},
		"{C}{C}{C}{C}{W}{W}", []int{0, 1}, true,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a, Mode: 0}, {Kind: game.TargetCard, ID: b, Mode: 0}})
	passPriorityAroundTable(t, g)
	if plusOneCounters(g, a) != 1 || plusOneCounters(g, b) != 1 {
		t.Fatalf("the counters go on before the proliferate: %d, %d", plusOneCounters(g, a), plusOneCounters(g, b))
	}
	answerProliferate(t, g, a)
	g.SettleResolution()
	if plusOneCounters(g, a) != 2 || plusOneCounters(g, b) != 1 {
		t.Errorf("counters %d, %d, want 2, 1", plusOneCounters(g, a), plusOneCounters(g, b))
	}
}

func TestEwcUnboundedPotentialUnentwinedUpToTwoTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	a := pushVanillaCreature(g, me.ID, "A", 1, 1)
	ewcCast(t, g, me, "Unbounded Potential", "Instant", "{1}{W}", ewcUnboundedPotentialOracle, []string{"W"},
		"{C}{W}", []int{0}, false,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a, Mode: 0}})
	passPriorityAroundTable(t, g)
	if plusOneCounters(g, a) != 1 {
		t.Errorf("one target, one counter: %d", plusOneCounters(g, a))
	}
	if proliferatePrompt(g) != nil {
		t.Error("the proliferate bullet was not chosen")
	}
}

// --- Wail of the Nim ----------------------------------------------------------

func TestEwcWailOfTheNimEntwinedRegeneratesYoursBeforeTheDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	mine := pushVanillaCreature(g, me.ID, "My Sprite", 1, 1)
	theirs := pushVanillaCreature(g, opp.ID, "Their Sprite", 1, 1)
	myLife, oppLife := me.Life, opp.Life
	ewcCast(t, g, me, "Wail of the Nim", "Instant", "{2}{B}", ewcWailOfTheNimOracle, []string{"B"},
		"{C}{C}{B}{B}", []int{1, 0}, true, nil)
	passPriorityAroundTable(t, g)
	if !ewcOnBattlefield(g, mine) {
		t.Error("your creature regenerated (the shield went up first, CR 702.42b)")
	} else if !battlefieldTapped(g, mine) {
		t.Error("a regenerated creature is tapped")
	}
	if ewcOnBattlefield(g, theirs) {
		t.Error("an opponent's creature had no shield")
	}
	if me.Life != myLife-1 || opp.Life != oppLife-1 {
		t.Errorf("life %d / %d, want %d / %d", me.Life, opp.Life, myLife-1, oppLife-1)
	}
}

func TestEwcWailOfTheNimUnentwinedDamageKillsYoursToo(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	mine := pushVanillaCreature(g, me.ID, "My Sprite", 1, 1)
	ewcCast(t, g, me, "Wail of the Nim", "Instant", "{2}{B}", ewcWailOfTheNimOracle, []string{"B"},
		"{C}{C}{B}", []int{1}, false, nil)
	passPriorityAroundTable(t, g)
	if ewcOnBattlefield(g, mine) {
		t.Error("no shield without the regenerate bullet")
	}
}
