package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// top1000_sweep_a_test.go — the first slice of the missing top-1,000
// cards (#3654): the filter lands, the conditional "Activate only if"
// lands, Forgotten Cave, Hedge Maze, Unnatural Growth, Unearth,
// Overwhelming Stampede, Silence, Unbreakable Formation and Felidar
// Retreat. One test per behaviour.

const (
	sweepADesolateMireOracle        = "3edf9201-265f-4cd9-b27b-8073bf1a4cf2"
	sweepAShadowbloodRidgeOracle    = "15687ee3-3cdb-4a8f-a726-46b73bceb792"
	sweepAMossfireValleyOracle      = "23bbd091-f6ff-4514-97aa-42c08164b4eb"
	sweepATaintedWoodOracle         = "fa6d05a1-3df4-4751-b1a0-8d9693faec73"
	sweepATaintedFieldOracle        = "439de49b-1091-4688-9ffb-80a025df31c2"
	sweepAWastewoodVergeOracle      = "c6e0574c-3e2b-4c40-b17a-05bce3d49309"
	sweepAForgottenCaveOracle       = "394c6de5-7957-4a0b-a6b9-ee0c707cd022"
	sweepAHedgeMazeOracle           = "ca4b6689-04ee-4227-9bdc-cb5a9590c745"
	sweepAUnnaturalGrowthOracle     = "7324abaa-48da-439d-9339-b0ea5eea612e"
	sweepAUnearthOracle             = "4e0b647a-6e1a-421b-bca9-8d5e346bda80"
	sweepAOverwhelmingStampedeOracl = "e1e96802-fd0c-41f1-aa21-3287d75a0e88"
	sweepASilenceOracle             = "8aed54cb-d1bb-45ad-adbe-38e55d84ff31"
	sweepAUnbreakableFormationOracl = "97435e03-2fff-4fe0-8fa9-69ab0a046a33"
	sweepAFelidarRetreatOracle      = "16629f59-bae8-4c19-bf50-443eb0ed6856"
)

func TestSweepAFilterLandsTurnOneManaIntoTwoColours(t *testing.T) {
	for _, row := range []struct {
		name, oracle string
		want         []string
	}{
		{"Desolate Mire", sweepADesolateMireOracle, []string{"W", "B"}},
		{"Shadowblood Ridge", sweepAShadowbloodRidgeOracle, []string{"B", "R"}},
		{"Mossfire Valley", sweepAMossfireValleyOracle, []string{"R", "G"}},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		land := effectstest.SeedLand(g, me.ID, row.name, "Land", row.oracle)
		// Without the {1} the ability cannot be paid.
		if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err == nil {
			t.Errorf("%s: activated with an empty pool", row.name)
		}
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
		if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("%s: ActivateManaAbility: %v", row.name, err)
		}
		got := effectstest.PoolColors(me)
		if len(got) != 2 || got[0] != row.want[0] || got[1] != row.want[1] {
			t.Errorf("%s: pool %v, want %v", row.name, got, row.want)
		}
	}
}

func TestSweepATaintedLandsNeedASwampForColour(t *testing.T) {
	for _, row := range []struct {
		name, oracle string
		colours      []string
	}{
		{"Tainted Wood", sweepATaintedWoodOracle, []string{"B", "G"}},
		{"Tainted Field", sweepATaintedFieldOracle, []string{"W", "B"}},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		land := effectstest.SeedLand(g, me.ID, row.name, "Land", row.oracle)

		// {C} is always available.
		effectstest.B28TapForMana(t, g, me.ID, land, "")
		if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "C" {
			t.Fatalf("%s: tapped for %v, want {C}", row.name, got)
		}
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })

		// The coloured ability is refused with no Swamp, and tapping nothing.
		if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err == nil {
			t.Fatalf("%s: coloured mana without a Swamp", row.name)
		}
		if c, _ := effectstest.BattlefieldCard(g, land); c.Tapped {
			t.Errorf("%s: a refused activation tapped the land", row.name)
		}

		effectstest.SeedLand(g, me.ID, "Swamp", "Basic Land — Swamp", "")
		effectstest.B40TapForMana(t, g, me.ID, land, 1, row.colours[0])
		if got := effectstest.PoolColors(me); got[len(got)-1] != row.colours[0] {
			t.Errorf("%s: pool %v, want the chosen %s", row.name, got, row.colours[0])
		}
	}
}

func TestSweepAWastewoodVergeNeedsASwampOrAForestForBlack(t *testing.T) {
	for _, typeLine := range []string{"Basic Land — Swamp", "Basic Land — Forest", "Land — Swamp Forest"} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		land := effectstest.SeedLand(g, me.ID, "Wastewood Verge", "Land", sweepAWastewoodVergeOracle)
		effectstest.SeedLand(g, me.ID, "Plains", "Basic Land — Plains", "")
		if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err == nil {
			t.Fatalf("%s: {B} with only a Plains", typeLine)
		}
		effectstest.B28TapForMana(t, g, me.ID, land, "")
		if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "G" {
			t.Fatalf("%s: the green ability gave %v, want {G}", typeLine, got)
		}
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
		me.ManaPool = nil

		effectstest.SeedLand(g, me.ID, "Other", typeLine, "")
		effectstest.B40TapForMana(t, g, me.ID, land, 1, "")
		if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "B" {
			t.Errorf("%s: pool %v, want {B}", typeLine, got)
		}
	}
}

func TestSweepAForgottenCaveEntersTappedTapsForRedAndCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.B12PlayFromHand(t, g, "Forgotten Cave", "Land", sweepAForgottenCaveOracle, game.CastSpellParams{})
	if !effectstest.B16Tapped(t, g, land) {
		t.Fatal("Forgotten Cave enters untapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	effectstest.B28TapForMana(t, g, me.ID, land, "")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Fatalf("pool %v, want {R}", got)
	}
	me.ManaPool = nil

	id, _ := effectstest.CycleFromHand(t, g, "Forgotten Cave", "Land", sweepAForgottenCaveOracle, "{R}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled land is not in the graveyard")
	}
	before := len(me.Hand.Cards)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != before+1 {
		t.Errorf("hand %d -> %d, want the cycling draw", before, got)
	}
}

func TestSweepAHedgeMazeEntersTappedAndSurveils(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PlayLandFromHand(t, g, "Hedge Maze", sweepAHedgeMazeOracle)
	if !effectstest.B16Tapped(t, g, land) {
		t.Fatal("Hedge Maze enters untapped")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID) == nil {
		t.Error("the enters trigger did not ask to surveil 1")
	}
}

func TestSweepAUnnaturalGrowthDoublesYourCreaturesAtEachCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Unnatural Growth", "Enchantment", sweepAUnnaturalGrowthOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)

	effectstest.AdvanceToStepOf(t, g, 0, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.S58p6View(t, g, bear); c.CurrentPower() != 4 || c.CurrentToughness() != 6 {
		t.Errorf("bear %d/%d, want 4/6", c.CurrentPower(), c.CurrentToughness())
	}
	if c := effectstest.S58p6View(t, g, theirs); c.CurrentPower() != 2 {
		t.Errorf("an opponent's creature was doubled to %d", c.CurrentPower())
	}

	effectstest.AdvanceToStepOf(t, g, 1, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.S58p6View(t, g, bear); c.CurrentPower() != 4 {
		t.Errorf("on the opponent's turn the bear is %d, want 4 (each combat)", c.CurrentPower())
	}
}

func TestSweepAUnearthReturnsACheapCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cheap := effectstest.SeedGraveyardCreature(me, "Cheap", "{2}{B}")
	dear := effectstest.SeedGraveyardCreature(me, "Dear", "{3}{B}")

	if err := effectstest.CastCatalogSpellErr(t, g, "Unearth", "Sorcery", sweepAUnearthOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: dear}}); err == nil {
		t.Fatal("Unearth targeted a mana value 4 creature")
	}
	effectstest.CastCatalogSpell(t, g, "Unearth", "Sorcery", sweepAUnearthOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: cheap}})
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, cheap); !ok || c.Controller != me.ID {
		t.Error("the mana value 3 creature did not come back under your control")
	}
	if effectstest.BattlefieldHas(g, dear) {
		t.Error("the mana value 4 creature came back")
	}
}

func TestSweepAUnearthCyclesFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id, _ := effectstest.CycleFromHand(t, g, "Unearth", "Sorcery", sweepAUnearthOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
}

func TestSweepAOverwhelmingStampedeAddsTheGreatestPowerAndTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Beast", 5, 5)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 9, 9)

	effectstest.CastCatalogSpell(t, g, "Overwhelming Stampede", "Sorcery", sweepAOverwhelmingStampedeOracl, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if c := effectstest.S58p6View(t, g, big); c.CurrentPower() != 10 || c.CurrentToughness() != 10 || !game.HasKeyword(&c, "trample") {
		t.Errorf("big %d/%d trample=%v, want 10/10 with trample", c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "trample"))
	}
	if c := effectstest.S58p6View(t, g, small); c.CurrentPower() != 6 || c.CurrentToughness() != 6 {
		t.Errorf("small %d/%d, want 6/6 (X is read once, from the 5-power creature)", c.CurrentPower(), c.CurrentToughness())
	}
	if c := effectstest.S58p6View(t, g, theirs); c.CurrentPower() != 9 || game.HasKeyword(&c, "trample") {
		t.Errorf("an opponent's creature was pumped: %d/%d", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestSweepASilenceStopsEveryOpponentCastingThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.CastInPlace(t, g, me.ID, "Silence", sweepASilenceOracle)
	effectstest.PassPriorityAroundTable(t, g)

	spell := effectstest.InstantCardFor(t, g)
	for _, opp := range g.Seats[1:] {
		if err := effectstest.CastGateFor(g, opp.ID, spell, game.ZoneHand); err == nil {
			t.Errorf("seat %s cast a spell under Silence", opp.ID)
		}
	}
	if err := effectstest.CastGateFor(g, me.ID, spell, game.ZoneHand); err != nil {
		t.Errorf("Silence stopped its own caster: %v", err)
	}
	effectstest.AdvanceOneTurnForTest(t, g)
	if err := effectstest.CastGateFor(g, g.Seats[1].ID, spell, game.ZoneHand); err != nil {
		t.Errorf("the ban outlasted the turn: %v", err)
	}
}

func TestSweepAUnbreakableFormationGrantsIndestructibleAndAddendumInMain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Unbreakable Formation", "Instant", sweepAUnbreakableFormationOracl, nil)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.S58p6View(t, g, mine)
	if !game.HasKeyword(&c, "indestructible") || !game.HasKeyword(&c, "vigilance") {
		t.Error("cast in the main phase, the creature should be indestructible with vigilance")
	}
	if c.Counters["+1/+1"] != 1 {
		t.Errorf("+1/+1 counters %d, want 1 (addendum)", c.Counters["+1/+1"])
	}
	o := effectstest.S58p6View(t, g, theirs)
	if game.HasKeyword(&o, "indestructible") || o.Counters["+1/+1"] != 0 {
		t.Error("an opponent's creature was protected")
	}
}

func TestSweepAUnbreakableFormationHasNoAddendumOutsideYourMainPhase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)

	effectstest.CastInPlace(t, g, me.ID, "Unbreakable Formation", sweepAUnbreakableFormationOracl)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.S58p6View(t, g, mine)
	if !game.HasKeyword(&c, "indestructible") {
		t.Error("the creature should still be indestructible")
	}
	if game.HasKeyword(&c, "vigilance") || c.Counters["+1/+1"] != 0 {
		t.Error("the addendum applied outside the main phase")
	}
}

func felidarRetreatLandfall(t *testing.T, g *game.Game, mode int) {
	t.Helper()
	me := g.Seats[0]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Felidar Retreat", TypeLine: "Enchantment", OracleID: sweepAFelidarRetreatOracle,
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatal("landfall did not ask which mode")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestSweepAFelidarRetreatLandfallMakesACatBeast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	felidarRetreatLandfall(t, g, 0)
	if effectstest.FindBattlefieldByName(g, "Cat Beast") == uuid.Nil {
		t.Fatal("no Cat Beast token")
	}
	c, _ := effectstest.BattlefieldCard(g, effectstest.FindBattlefieldByName(g, "Cat Beast"))
	if c.Power != 2 || c.Toughness != 2 {
		t.Errorf("token is %d/%d, want 2/2", c.Power, c.Toughness)
	}
}

func TestSweepAFelidarRetreatLandfallPutsACounterOnEachCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	felidarRetreatLandfall(t, g, 1)
	for _, id := range []uuid.UUID{a, b} {
		c := effectstest.S58p6View(t, g, id)
		if c.Counters["+1/+1"] != 1 || !game.HasKeyword(&c, "vigilance") {
			t.Errorf("%s: counters %v, vigilance %v", c.Name, c.Counters, game.HasKeyword(&c, "vigilance"))
		}
	}
	if c := effectstest.S58p6View(t, g, theirs); c.Counters["+1/+1"] != 0 {
		t.Error("an opponent's creature got a counter")
	}
	if effectstest.FindBattlefieldByName(g, "Cat Beast") != uuid.Nil {
		t.Error("the counter mode also made a token")
	}
}
