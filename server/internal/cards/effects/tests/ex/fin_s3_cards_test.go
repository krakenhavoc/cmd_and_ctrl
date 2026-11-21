package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_s3_cards_test.go — Final Fantasy (FIN), the cards starting with S,
// batch 3 (#4462). Each test plays the real catalog entry and asserts
// what its text says. A Saga is either cast (chapter I) or put onto the
// battlefield with lore counters already on it, so the next precombat
// main adds the counter for the chapter under test (CR 714.3).

const (
	finS3AnimaOracle     = "0990e57f-279e-4348-a049-e06187cc38d4"
	finS3ChocoMogOracle  = "8b4bfdc5-8c0a-40c3-abbf-ada4a536dc34"
	finS3RamuhOracle     = "0d054301-8465-4a4c-91e7-9a28dda92e57"
	finS3FatChocoOracle  = "b5d2f8e7-89d3-4015-ab99-ba4d66918e37"
	finS3IfritOracle     = "96d1bfc1-b80f-466f-b07d-df1ab361e8e0"
	finS3GarudaOracle    = "aaf9e711-94e1-4869-81b0-baa6380ee873"
	finS3OdinOracle      = "032f59c3-9f47-4db3-a43c-065a99ed7efe"
	finS3ShivaOracle     = "5846718c-7d31-4d2c-80a2-b1534e2aebc1"
	finS3TitanOracle     = "66ce0c60-f406-4ece-915f-675da4e8b16d"
	finS3SandwormOracle  = "e92185f8-579c-4645-9cb8-ef5deb9eebc8"
	finS3SazhOracle      = "0e5b40dc-debf-48a1-8398-5ea00b5a5166"
	finS3ShamblingOracle = "d3f4aa87-a89e-43f8-abc3-712e6bf08889"
)

func TestFinS3EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		finS3AnimaOracle:     "Summon: Anima",
		finS3ChocoMogOracle:  "Summon: Choco/Mog",
		finS3RamuhOracle:     "Summon: Esper Ramuh",
		finS3FatChocoOracle:  "Summon: Fat Chocobo",
		finS3IfritOracle:     "Summon: G.F. Ifrit",
		finS3GarudaOracle:    "Summon: Primal Garuda",
		finS3OdinOracle:      "Summon: Primal Odin",
		finS3ShivaOracle:     "Summon: Shiva",
		finS3TitanOracle:     "Summon: Titan",
		finS3SandwormOracle:  "Sandworm",
		finS3SazhOracle:      "Sazh Katzroy",
		finS3ShamblingOracle: "Shambling Cie'th",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// finS3Opponents is every seat but the active one.
func finS3Opponents(g *game.Game) []*game.Player {
	var opps []*game.Player
	for i, p := range g.Seats {
		if i != g.Turn.ActiveSeat {
			opps = append(opps, p)
		}
	}
	return opps
}

// --- Summon: Anima -----------------------------------------------------------------

func TestFinS3AnimaChapterOneDrawsAndLosesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hand, life := me.Hand.Size(), me.Life
	effectstest.FinS3CastSaga(t, g, "Summon: Anima", "Enchantment Creature — Saga Horror", finS3AnimaOracle, "{4}{B}{B}", 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if got := life - me.Life; got != 1 {
		t.Errorf("lost %d life, want 1", got)
	}
}

func TestFinS3AnimaChapterFourEdictsAndDrainsEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opps := finS3Opponents(g)
	lone := effectstest.B12Creature(g, opps[0].ID, "Lone", "Creature — Bear", 2, 2)
	x := effectstest.B12Creature(g, opps[1].ID, "X", "Creature — Bear", 2, 2)
	y := effectstest.B12Creature(g, opps[1].ID, "Y", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.FinS3Saga(g, me.ID, "Summon: Anima", "Enchantment Creature — Saga Horror", finS3AnimaOracle, 4, 4, 3)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Error("the Saga's controller was asked to sacrifice")
	}
	effectstest.AnswerSacrifice(t, g, opps[0].ID, effectstest.SacrificeOptions(effectstest.SacrificeChoiceFor(g, opps[0].ID))...)
	effectstest.AnswerSacrifice(t, g, opps[1].ID, x)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(lone) || g.Battlefield.Contains(x) || !g.Battlefield.Contains(y) {
		t.Error("each opponent sacrifices exactly one creature of their choice")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("your own creature was sacrificed")
	}
	for _, opp := range opps {
		if got := lives[opp.ID] - opp.Life; got != 3 {
			t.Errorf("opponent %s lost %d life, want 3 (no creature still loses it)", opp.Name, got)
		}
	}
	if me.Life != lives[me.ID] {
		t.Errorf("the controller's life %d -> %d", lives[me.ID], me.Life)
	}
}

// --- Summon: Choco/Mog -------------------------------------------------------------

func TestFinS3ChocoMogPumpsOnlyOtherCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	saga := effectstest.FinS3CastSaga(t, g, "Summon: Choco/Mog", "Enchantment Creature — Saga Bird Moogle", finS3ChocoMogOracle, "{2}{W}", 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, bear); p != 3 || th != 2 {
		t.Errorf("your Bear is %d/%d, want 3/2", p, th)
	}
	if p := effectstest.EffectivePower(t, g, saga); p != 3 {
		t.Errorf("the Saga is power %d, want 3: it pumps only the others", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
		t.Errorf("the opponent's creature is power %d, want 2", p)
	}
}

// --- Summon: Esper Ramuh -----------------------------------------------------------

func TestFinS3RamuhChapterOneCountsNoncreatureNonlandCards(t *testing.T) {
	for _, tc := range []struct {
		name      string
		toughness int
		dies      bool
	}{{"two damage kills a 2/2", 2, true}, {"two damage leaves a 2/3", 3, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
			effectstest.PushGraveyardCardWithTypeLineID(me, "Bolt", "Instant")
			effectstest.PushGraveyardCardWithTypeLineID(me, "Ponder", "Sorcery")
			effectstest.PushGraveyardCardWithTypeLineID(me, "Bear Card", "Creature — Bear")
			effectstest.PushGraveyardCardWithTypeLineID(me, "Forest", "Basic Land — Forest")
			victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, tc.toughness)
			effectstest.FinS3CastSaga(t, g, "Summon: Esper Ramuh", "Enchantment Creature — Saga Wizard", finS3RamuhOracle, "{2}{R}{R}", 3, 3)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.B04WaitForPick(t, g, me.ID)
			effectstest.AnswerPickTarget(t, g, victim)
			effectstest.PassPriorityAroundTable(t, g)
			if dead := !g.Battlefield.Contains(victim); dead != tc.dies {
				t.Errorf("victim died %v, want %v", dead, tc.dies)
			}
		})
	}
}

func TestFinS3RamuhChapterTwoPumpsWizards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	mage := effectstest.B12Creature(g, me.ID, "Mage", "Creature — Human Wizard", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	saga := effectstest.FinS3Saga(g, me.ID, "Summon: Esper Ramuh", "Enchantment Creature — Saga Wizard", finS3RamuhOracle, 3, 3, 1)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, mage); p != 2 {
		t.Errorf("the Wizard is power %d, want 2", p)
	}
	if p := effectstest.EffectivePower(t, g, saga); p != 4 {
		t.Errorf("the Saga, a Wizard, is power %d, want 4", p)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("the Bear is power %d, want 2", p)
	}
}

// --- Summon: Fat Chocobo -----------------------------------------------------------

func TestFinS3FatChocoboMakesABirdThenGrantsTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.FinS3CastSaga(t, g, "Summon: Fat Chocobo", "Enchantment Creature — Saga Bird", finS3FatChocoOracle, "{4}{G}", 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Bird", me.ID); got != 1 {
		t.Fatalf("%d Bird tokens after chapter I, want 1", got)
	}
	bird := effectstest.FindBattlefieldByName(g, "Bird")
	if p, th := effectstest.EffectivePTOf(t, g, bird); p != 2 || th != 2 {
		t.Errorf("the Bird is %d/%d, want 2/2", p, th)
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat) // chapter II
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bird, "trample") {
		t.Error("chapter II did not give the Bird trample")
	}
}

// --- Summon: G.F. Ifrit ------------------------------------------------------------

func TestFinS3IfritLootsThenAddsRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.FinS3CastSaga(t, g, "Summon: G.F. Ifrit", "Enchantment Creature — Saga Demon", finS3IfritOracle, "{2}{R}", 3, 2)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("chapter I offered no discard")
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want a discard and a draw", hand, me.Hand.Size())
	}
}

func TestFinS3IfritChapterThreeAddsRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.FinS3Saga(g, me.ID, "Summon: G.F. Ifrit", "Enchantment Creature — Saga Demon", finS3IfritOracle, 3, 2, 2)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.ManaPool.CountColor("R"); got != 1 {
		t.Errorf("%d red mana in the pool, want 1", got)
	}
}

// --- Summon: Primal Garuda ---------------------------------------------------------

func TestFinS3GarudaBlastsATappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 4, 4)
	effectstest.FinS3Tap(g, victim)
	effectstest.FinS3CastSaga(t, g, "Summon: Primal Garuda", "Enchantment Creature — Saga Harpy", finS3GarudaOracle, "{3}{W}", 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("4 damage did not kill the tapped 4/4")
	}
}

func TestFinS3GarudaSlipstreamPumpsAndGrantsFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FinS3Saga(g, me.ID, "Summon: Primal Garuda", "Enchantment Creature — Saga Harpy", finS3GarudaOracle, 3, 3, 1)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, bear); p != 3 || th != 2 {
		t.Errorf("the Bear is %d/%d, want 3/2", p, th)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the Bear did not gain flying")
	}
}

// --- Summon: Primal Odin -----------------------------------------------------------

func TestFinS3OdinChapterOneDestroys(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 5, 5)
	effectstest.FinS3CastSaga(t, g, "Summon: Primal Odin", "Enchantment Creature — Saga Knight", finS3OdinOracle, "{4}{B}{B}", 5, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("chapter I did not destroy the target")
	}
}

func TestFinS3OdinZantetsukenMakesTheDamagedPlayerLose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], effectstest.FinABOpp(g)
	odin := effectstest.FinS3Saga(g, me.ID, "Summon: Primal Odin", "Enchantment Creature — Saga Knight", finS3OdinOracle, 5, 3, 1)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FinABAttack(t, g, opp.ID, odin)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Eliminated {
		t.Error("the player Odin hit did not lose the game")
	}
	if me.Eliminated {
		t.Error("Odin's controller lost the game")
	}
}

func TestFinS3OdinHallOfSorrowDrawsTwoAndEveryoneLosesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.FinS3Saga(g, me.ID, "Summon: Primal Odin", "Enchantment Creature — Saga Knight", finS3OdinOracle, 5, 3, 2)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	for _, p := range g.Seats {
		if got := lives[p.ID] - p.Life; got != 2 {
			t.Errorf("%s lost %d life, want 2", p.Name, got)
		}
	}
}

// --- Summon: Shiva -----------------------------------------------------------------

func TestFinS3ShivaTapsAndStuns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.FinS3CastSaga(t, g, "Summon: Shiva", "Enchantment Creature — Saga Elemental", finS3ShivaOracle, "{3}{U}{U}", 4, 5)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedNow(g, victim) || effectstest.CountersOn(g, victim, game.CounterStun) != 1 {
		t.Errorf("the target: tapped %v, %d stun counters; want tapped with one", effectstest.TappedNow(g, victim), effectstest.CountersOn(g, victim, game.CounterStun))
	}
}

func TestFinS3ShivaDiamondDustDrawsPerTappedOpposingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opps := finS3Opponents(g)
	a := effectstest.B12Creature(g, opps[0].ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opps[0].ID, "B", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opps[0].ID, "Untapped", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opps[1].ID, "C", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.FinS3Saga(g, me.ID, "Summon: Shiva", "Enchantment Creature — Saga Elemental", finS3ShivaOracle, 4, 5, 2)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	// Tapped now, after every untap step the advance walked through; the
	// count is taken as the chapter resolves.
	for _, id := range []uuid.UUID{a, b, c, mine} {
		effectstest.FinS3Tap(g, id)
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3 (your own tapped creature does not count)", got)
	}
}

// --- Summon: Titan -----------------------------------------------------------------

func TestFinS3TitanMillsFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	grave := me.Graveyard.Size()
	effectstest.FinS3CastSaga(t, g, "Summon: Titan", "Enchantment Creature — Saga Giant", finS3TitanOracle, "{3}{G}{G}", 7, 7)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Graveyard.Size() - grave; got != 5 {
		t.Errorf("milled %d, want 5", got)
	}
}

func TestFinS3TitanReturnsLandsTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	f1 := effectstest.PushGraveyardCardWithTypeLineID(me, "Forest", "Basic Land — Forest")
	f2 := effectstest.PushGraveyardCardWithTypeLineID(me, "Wastes", "Basic Land")
	bolt := effectstest.PushGraveyardCardWithTypeLineID(me, "Bolt", "Instant")
	effectstest.FinS3Saga(g, me.ID, "Summon: Titan", "Enchantment Creature — Saga Giant", finS3TitanOracle, 7, 7, 1)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{f1, f2} {
		c, ok := effectstest.BattlefieldCard(g, id)
		if !ok || !c.Tapped || c.Controller != me.ID {
			t.Errorf("land %s: on the battlefield %v, tapped %v; want both under you", id, ok, c.Tapped)
		}
	}
	if !me.Graveyard.Contains(bolt) {
		t.Error("the instant left the graveyard")
	}
}

func TestFinS3TitanChapterThreePumpsByLandsAndGrantsTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	for i := 0; i < 3; i++ {
		effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FinS3Saga(g, me.ID, "Summon: Titan", "Enchantment Creature — Saga Giant", finS3TitanOracle, 7, 7, 2)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	lands := effects.CountControlled(g, me.ID, func(c game.Card) bool { return c.IsLand() })
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, bear); p != 2+lands || th != 2+lands {
		t.Errorf("the Bear is %d/%d, want %d/%d", p, th, 2+lands, 2+lands)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the Bear did not gain trample")
	}
}

// --- Sandworm ----------------------------------------------------------------------

func TestFinS3SandwormDestroysALandAndItsControllerFetchesTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	forest := effectstest.StapleLibraryCard(opp, "Forest", "Basic Land — Forest")
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Sandworm", TypeLine: "Creature — Worm", OracleID: finS3SandwormOracle,
		ManaCost: "{4}{R}", Power: 5, Toughness: 4}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.AnswerPickTarget(t, g, land)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land survived")
	}
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("the caster was offered the search; it belongs to the land's controller")
	}
	effectstest.AnswerSearchByID(t, g, opp.ID, forest)
	got, ok := effectstest.BattlefieldCard(g, forest)
	if !ok || !got.Tapped || got.Controller != opp.ID {
		t.Errorf("the fetched basic must be on the battlefield tapped under its owner, got %+v (found %v)", got, ok)
	}
}

// --- Sazh Katzroy ------------------------------------------------------------------

func TestFinS3SazhFetchesABird(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bird := effectstest.StapleLibraryCard(me, "Chocobo", "Creature — Bird")
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Sazh Katzroy", TypeLine: "Legendary Creature — Human Pilot", OracleID: finS3SazhOracle,
		ManaCost: "{3}{G}", Power: 3, Toughness: 3}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, bird)
	if !me.Hand.Contains(bird) {
		t.Error("the Bird card did not go to hand")
	}
}

func TestFinS3SazhAttackPutsACounterThenDoubles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	sazh := effectstest.B12Push(g, me.ID, "Sazh Katzroy", "Legendary Creature — Human Pilot", finS3SazhOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 1}
			}
		}
	})
	effectstest.FinABAttack(t, g, opp.ID, sazh)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 4 {
		t.Errorf("%d +1/+1 counters, want 4 (one plus one, then doubled)", got)
	}
}

// --- Shambling Cie'th --------------------------------------------------------------

func TestFinS3ShamblingCiethReturnsOnANoncreatureSpellForB(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost, pool string
		offered                    bool
	}{
		{"sorcery", "Sorcery", "{1}{U}", "UU", true},
		{"creature", "Creature — Bear", "{1}{G}", "GG", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			cieth := effectstest.PushCatalogGraveyardCard(me, "Shambling Cie'th", "Creature — Mutant Horror", finS3ShamblingOracle, 3, 3)
			finSZ2CastPaid(t, g, tc.typeLine, tc.cost, tc.pool)
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.HasPayUnlessFor(g, me.ID) != tc.offered {
				t.Fatalf("offered %v, want %v", !tc.offered, tc.offered)
			}
			if !tc.offered {
				return
			}
			effectstest.FloatForTest(g, me, "B")
			prompt := effectstest.AnswerPayUnless(t, g, me.ID, true)
			if prompt.PayCost != "{B}" {
				t.Errorf("payment = %q, want {B}", prompt.PayCost)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !me.Hand.Contains(cieth) || me.Graveyard.Contains(cieth) {
				t.Errorf("paying did not return it to hand (hand=%t graveyard=%t)", me.Hand.Contains(cieth), me.Graveyard.Contains(cieth))
			}
		})
	}
}
