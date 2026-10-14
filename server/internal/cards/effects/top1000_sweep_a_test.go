package effects

import (
	"testing"

	"github.com/google/uuid"

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
		g := newCatalogGame(t)
		me := g.Seats[0]
		land := seedLand(g, me.ID, row.name, "Land", row.oracle)
		// Without the {1} the ability cannot be paid.
		if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err == nil {
			t.Errorf("%s: activated with an empty pool", row.name)
		}
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
		if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("%s: ActivateManaAbility: %v", row.name, err)
		}
		got := poolColors(me)
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
		g := newCatalogGame(t)
		me := g.Seats[0]
		land := seedLand(g, me.ID, row.name, "Land", row.oracle)

		// {C} is always available.
		b28TapForMana(t, g, me.ID, land, "")
		if got := poolColors(me); len(got) != 1 || got[0] != "C" {
			t.Fatalf("%s: tapped for %v, want {C}", row.name, got)
		}
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })

		// The coloured ability is refused with no Swamp, and tapping nothing.
		if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err == nil {
			t.Fatalf("%s: coloured mana without a Swamp", row.name)
		}
		if c, _ := battlefieldCard(g, land); c.Tapped {
			t.Errorf("%s: a refused activation tapped the land", row.name)
		}

		seedLand(g, me.ID, "Swamp", "Basic Land — Swamp", "")
		b40TapForMana(t, g, me.ID, land, 1, row.colours[0])
		if got := poolColors(me); got[len(got)-1] != row.colours[0] {
			t.Errorf("%s: pool %v, want the chosen %s", row.name, got, row.colours[0])
		}
	}
}

func TestSweepAWastewoodVergeNeedsASwampOrAForestForBlack(t *testing.T) {
	for _, typeLine := range []string{"Basic Land — Swamp", "Basic Land — Forest", "Land — Swamp Forest"} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		land := seedLand(g, me.ID, "Wastewood Verge", "Land", sweepAWastewoodVergeOracle)
		seedLand(g, me.ID, "Plains", "Basic Land — Plains", "")
		if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err == nil {
			t.Fatalf("%s: {B} with only a Plains", typeLine)
		}
		b28TapForMana(t, g, me.ID, land, "")
		if got := poolColors(me); len(got) != 1 || got[0] != "G" {
			t.Fatalf("%s: the green ability gave %v, want {G}", typeLine, got)
		}
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
		me.ManaPool = nil

		seedLand(g, me.ID, "Other", typeLine, "")
		b40TapForMana(t, g, me.ID, land, 1, "")
		if got := poolColors(me); len(got) != 1 || got[0] != "B" {
			t.Errorf("%s: pool %v, want {B}", typeLine, got)
		}
	}
}

func TestSweepAForgottenCaveEntersTappedTapsForRedAndCycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := b12PlayFromHand(t, g, "Forgotten Cave", "Land", sweepAForgottenCaveOracle, game.CastSpellParams{})
	if !b16Tapped(t, g, land) {
		t.Fatal("Forgotten Cave enters untapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	b28TapForMana(t, g, me.ID, land, "")
	if got := poolColors(me); len(got) != 1 || got[0] != "R" {
		t.Fatalf("pool %v, want {R}", got)
	}
	me.ManaPool = nil

	id, _ := cycleFromHand(t, g, "Forgotten Cave", "Land", sweepAForgottenCaveOracle, "{R}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled land is not in the graveyard")
	}
	before := len(me.Hand.Cards)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != before+1 {
		t.Errorf("hand %d -> %d, want the cycling draw", before, got)
	}
}

func TestSweepAHedgeMazeEntersTappedAndSurveils(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := playLandFromHand(t, g, "Hedge Maze", sweepAHedgeMazeOracle)
	if !b16Tapped(t, g, land) {
		t.Fatal("Hedge Maze enters untapped")
	}
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID) == nil {
		t.Error("the enters trigger did not ask to surveil 1")
	}
}

func TestSweepAUnnaturalGrowthDoublesYourCreaturesAtEachCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Unnatural Growth", "Enchantment", sweepAUnnaturalGrowthOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 3)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)

	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if c := s58p6View(t, g, bear); c.CurrentPower() != 4 || c.CurrentToughness() != 6 {
		t.Errorf("bear %d/%d, want 4/6", c.CurrentPower(), c.CurrentToughness())
	}
	if c := s58p6View(t, g, theirs); c.CurrentPower() != 2 {
		t.Errorf("an opponent's creature was doubled to %d", c.CurrentPower())
	}

	advanceToStepOf(t, g, 1, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if c := s58p6View(t, g, bear); c.CurrentPower() != 4 {
		t.Errorf("on the opponent's turn the bear is %d, want 4 (each combat)", c.CurrentPower())
	}
}

func TestSweepAUnearthReturnsACheapCreatureOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cheap := seedGraveyardCreature(me, "Cheap", "{2}{B}")
	dear := seedGraveyardCreature(me, "Dear", "{3}{B}")

	if err := castCatalogSpellErr(t, g, "Unearth", "Sorcery", sweepAUnearthOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: dear}}); err == nil {
		t.Fatal("Unearth targeted a mana value 4 creature")
	}
	castCatalogSpell(t, g, "Unearth", "Sorcery", sweepAUnearthOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: cheap}})
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, cheap); !ok || c.Controller != me.ID {
		t.Error("the mana value 3 creature did not come back under your control")
	}
	if battlefieldHas(g, dear) {
		t.Error("the mana value 4 creature came back")
	}
}

func TestSweepAUnearthCyclesFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id, _ := cycleFromHand(t, g, "Unearth", "Sorcery", sweepAUnearthOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
}

func TestSweepAOverwhelmingStampedeAddsTheGreatestPowerAndTrample(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := b12Creature(g, me.ID, "Big", "Creature — Beast", 5, 5)
	small := b12Creature(g, me.ID, "Small", "Creature — Bear", 1, 1)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 9, 9)

	castCatalogSpell(t, g, "Overwhelming Stampede", "Sorcery", sweepAOverwhelmingStampedeOracl, nil)
	passPriorityAroundTable(t, g)

	if c := s58p6View(t, g, big); c.CurrentPower() != 10 || c.CurrentToughness() != 10 || !game.HasKeyword(&c, "trample") {
		t.Errorf("big %d/%d trample=%v, want 10/10 with trample", c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "trample"))
	}
	if c := s58p6View(t, g, small); c.CurrentPower() != 6 || c.CurrentToughness() != 6 {
		t.Errorf("small %d/%d, want 6/6 (X is read once, from the 5-power creature)", c.CurrentPower(), c.CurrentToughness())
	}
	if c := s58p6View(t, g, theirs); c.CurrentPower() != 9 || game.HasKeyword(&c, "trample") {
		t.Errorf("an opponent's creature was pumped: %d/%d", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestSweepASilenceStopsEveryOpponentCastingThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	castInPlace(t, g, me.ID, "Silence", sweepASilenceOracle)
	passPriorityAroundTable(t, g)

	spell := instantCardFor(t, g)
	for _, opp := range g.Seats[1:] {
		if err := castGateFor(g, opp.ID, spell, game.ZoneHand); err == nil {
			t.Errorf("seat %s cast a spell under Silence", opp.ID)
		}
	}
	if err := castGateFor(g, me.ID, spell, game.ZoneHand); err != nil {
		t.Errorf("Silence stopped its own caster: %v", err)
	}
	advanceOneTurnForTest(t, g)
	if err := castGateFor(g, g.Seats[1].ID, spell, game.ZoneHand); err != nil {
		t.Errorf("the ban outlasted the turn: %v", err)
	}
}

func TestSweepAUnbreakableFormationGrantsIndestructibleAndAddendumInMain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Unbreakable Formation", "Instant", sweepAUnbreakableFormationOracl, nil)
	passPriorityAroundTable(t, g)

	c := s58p6View(t, g, mine)
	if !game.HasKeyword(&c, "indestructible") || !game.HasKeyword(&c, "vigilance") {
		t.Error("cast in the main phase, the creature should be indestructible with vigilance")
	}
	if c.Counters["+1/+1"] != 1 {
		t.Errorf("+1/+1 counters %d, want 1 (addendum)", c.Counters["+1/+1"])
	}
	o := s58p6View(t, g, theirs)
	if game.HasKeyword(&o, "indestructible") || o.Counters["+1/+1"] != 0 {
		t.Error("an opponent's creature was protected")
	}
}

func TestSweepAUnbreakableFormationHasNoAddendumOutsideYourMainPhase(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)

	castInPlace(t, g, me.ID, "Unbreakable Formation", sweepAUnbreakableFormationOracl)
	passPriorityAroundTable(t, g)

	c := s58p6View(t, g, mine)
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
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Felidar Retreat", TypeLine: "Enchantment", OracleID: sweepAFelidarRetreatOracle,
		Owner: me.ID, Controller: me.ID,
	})
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatal("landfall did not ask which mode")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestSweepAFelidarRetreatLandfallMakesACatBeast(t *testing.T) {
	g := newCatalogGame(t)
	felidarRetreatLandfall(t, g, 0)
	if findBattlefieldByName(g, "Cat Beast") == uuid.Nil {
		t.Fatal("no Cat Beast token")
	}
	c, _ := battlefieldCard(g, findBattlefieldByName(g, "Cat Beast"))
	if c.Power != 2 || c.Toughness != 2 {
		t.Errorf("token is %d/%d, want 2/2", c.Power, c.Toughness)
	}
}

func TestSweepAFelidarRetreatLandfallPutsACounterOnEachCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	felidarRetreatLandfall(t, g, 1)
	for _, id := range []uuid.UUID{a, b} {
		c := s58p6View(t, g, id)
		if c.Counters["+1/+1"] != 1 || !game.HasKeyword(&c, "vigilance") {
			t.Errorf("%s: counters %v, vigilance %v", c.Name, c.Counters, game.HasKeyword(&c, "vigilance"))
		}
	}
	if c := s58p6View(t, g, theirs); c.Counters["+1/+1"] != 0 {
		t.Error("an opponent's creature got a counter")
	}
	if findBattlefieldByName(g, "Cat Beast") != uuid.Nil {
		t.Error("the counter mode also made a token")
	}
}
