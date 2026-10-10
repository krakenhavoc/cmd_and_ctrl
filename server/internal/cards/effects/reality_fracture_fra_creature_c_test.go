package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_fra_creature_c_test.go — tracker #2795, slice
// fra-creature-c: twenty Reality Fracture creatures.

const (
	rfcGuidingHydraOracle  = "24f1445c-16c9-45c0-be56-0266e6c78cdb"
	rfcHapatraFangOracle   = "6894345f-52a6-46e4-b278-0783087daee1"
	rfcHapatraFrostOracle  = "ac0f161a-e5ec-4a95-9f37-305bdf1ac900"
	rfcHeartstringOracle   = "2cdcf65c-03d2-416f-8a3f-0322f75e8595"
	rfcInvigoratorOracle   = "12d617b1-c95b-4b1c-a586-ccfdf5898b03"
	rfcPuppetbeastOracle   = "690ae865-87bd-46b2-8e65-c53bd80c1a18"
	rfcIngrisOracle        = "bda098cb-31ec-41a1-a9d7-122877cc69d2"
	rfcJhoiraOracle        = "cf9b0bb5-d545-4efd-9255-39b10c2be0ee"
	rfcJiangAloneOracle    = "fa8b3557-b35a-4a68-ab85-cf09683f3bc5"
	rfcJiangNeverOracle    = "4ba661d7-1cb8-4fec-bfaa-5f2797235fcf"
	rfcKarnDefenderOracle  = "aecc621f-67da-41a7-9d47-05ded302ae35"
	rfcKarnGuardianOracle  = "20b90a18-ff62-4995-b0d6-fbef44dca357"
	rfcKioraOracle         = "02b41d0e-82e7-4307-8cc4-8175ff78ea1f"
	rfcKothHomesteadOracle = "567aeb60-a441-41b1-98f4-54510d299317"
	rfcKothGeomancerOracle = "3b17da84-ae4d-4fc1-a327-85e54e6156c6"
	rfcKwiaOracle          = "07a28621-e617-46b0-af3e-2efb03b5056a"
	rfcLootOracle          = "a99ce9db-9b7d-45a0-bee7-0875fc62ca48"
	rfcLyraDawnOracle      = "0fca3328-484f-4050-925e-91840710b5d6"
	rfcLyraTolarianOracle  = "d9060fff-0b65-4b3d-930d-6e36315b802c"
	rfcMabelOracle         = "ab56f7cf-9538-45b5-a657-1e2ebef1e92b"

	rfcLegendaryCreature = "Legendary Creature — Test"
)

func rfcSeats(g *game.Game) (me *game.Player, opps []*game.Player) {
	me = g.Seats[g.Turn.ActiveSeat]
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opps = append(opps, p)
		}
	}
	return me, opps
}

func rfcCountersOf(t *testing.T, g *game.Game, id uuid.UUID, kind string) int {
	t.Helper()
	return b12Counter(t, g, id, kind)
}

func rfcGainLife(g *game.Game, who uuid.UUID, n int) {
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, who, n) })
}

// --- Guiding Hydra ------------------------------------------------

func TestGuidingHydraEntersWithXCounters(t *testing.T) {
	g := newCatalogGame(t)
	hydra := b12PlayFromHand(t, g, "Guiding Hydra", "Creature — Hydra Horror", rfcGuidingHydraOracle, game.CastSpellParams{XValue: 2})
	passPriorityAroundTable(t, g)
	if got := rfcCountersOf(t, g, hydra, game.CounterPlusOne); got != 2 {
		t.Fatalf("counters = %d, want 2", got)
	}
}

func TestGuidingHydraSpreadsACounterAtBeginningOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := rfcSeats(g)
	hydra := b12PlayFromHand(t, g, "Guiding Hydra", "Creature — Hydra Horror", rfcGuidingHydraOracle, game.CastSpellParams{XValue: 2})
	passPriorityAroundTable(t, g)
	bear := pushCreatureToBattlefieldForTest(g, me.ID, "Bear")
	other := pushCreatureToBattlefieldForTest(g, g.Seats[1].ID, "Their Bear")
	advanceTo(t, g, game.StepBeginCombat)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := rfcCountersOf(t, g, hydra, game.CounterPlusOne); got != 1 {
		t.Errorf("hydra counters = %d, want 1", got)
	}
	if got := rfcCountersOf(t, g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("my other creature counters = %d, want 1", got)
	}
	if got := rfcCountersOf(t, g, other, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's creature must get none, got %d", got)
	}
}

func TestGuidingHydraDeclinedChangesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := rfcSeats(g)
	hydra := b12PlayFromHand(t, g, "Guiding Hydra", "Creature — Hydra Horror", rfcGuidingHydraOracle, game.CastSpellParams{XValue: 2})
	passPriorityAroundTable(t, g)
	bear := pushCreatureToBattlefieldForTest(g, me.ID, "Bear")
	advanceTo(t, g, game.StepBeginCombat)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := rfcCountersOf(t, g, hydra, game.CounterPlusOne); got != 2 {
		t.Errorf("hydra counters = %d, want 2", got)
	}
	if got := rfcCountersOf(t, g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("bear counters = %d, want 0", got)
	}
}

// --- Hapatra, the Desert Fang -------------------------------------

func TestHapatraFangPutsXMinusCountersOnOnePerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := rfcSeats(g)
	pushGraveyardPermanent(me, "Big Thing", "Creature — Test", "{4}{B}")
	pushGraveyardPermanent(me, "Small Thing", "Creature — Test", "{1}")
	a := pushDiesCreatureForTest(g, opps[0].ID, "A's Wall", "", "Creature — Test", 9, 9)
	a2 := pushDiesCreatureForTest(g, opps[0].ID, "A's Other Wall", "", "Creature — Test", 9, 9)
	b := pushDiesCreatureForTest(g, opps[1].ID, "B's Wall", "", "Creature — Test", 9, 9)
	castHoldCreature(t, g, "Hapatra, the Desert Fang", rfcLegendaryCreature, rfcHapatraFangOracle, 3, 3)
	answerPickTarget(t, g, a)
	answerPickTarget(t, g, b)
	passPriorityAroundTable(t, g)
	if got := rfcCountersOf(t, g, a, game.CounterMinusOne); got != 5 {
		t.Errorf("first opponent's pick has %d -1/-1 counters, want 5 (greatest mana value in the graveyard)", got)
	}
	if got := rfcCountersOf(t, g, b, game.CounterMinusOne); got != 5 {
		t.Errorf("second opponent's pick has %d -1/-1 counters, want 5", got)
	}
	if got := rfcCountersOf(t, g, a2, game.CounterMinusOne); got != 0 {
		t.Errorf("only one creature per opponent, other has %d", got)
	}
}

func TestHapatraFangWithAnEmptyGraveyardDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	_, opps := rfcSeats(g)
	a := pushCreatureToBattlefieldForTest(g, opps[0].ID, "A's Bear")
	castHoldCreature(t, g, "Hapatra, the Desert Fang", rfcLegendaryCreature, rfcHapatraFangOracle, 3, 3)
	answerPickTarget(t, g, a)
	passPriorityAroundTable(t, g)
	if got := rfcCountersOf(t, g, a, game.CounterMinusOne); got != 0 {
		t.Errorf("X is zero, got %d counters", got)
	}
}

// --- Hapatra, the Desert Frost ------------------------------------

func TestHapatraFrostTapsAndStunsOnePerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	_, opps := rfcSeats(g)
	a := pushCreatureToBattlefieldForTest(g, opps[0].ID, "A's Bear")
	b := pushCreatureToBattlefieldForTest(g, opps[1].ID, "B's Bear")
	castHoldCreature(t, g, "Hapatra, the Desert Frost", rfcLegendaryCreature, rfcHapatraFrostOracle, 4, 3)
	answerPickTarget(t, g, a)
	answerPickTarget(t, g, b)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		c, ok := battlefieldCard(g, id)
		if !ok || !c.Tapped {
			t.Errorf("%s should be tapped", id)
		}
		if got := rfcCountersOf(t, g, id, game.CounterStun); got != 1 {
			t.Errorf("%s has %d stun counters, want 1", id, got)
		}
	}
}

func TestHapatraFrostUntapAbility(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := rfcSeats(g)
	toMain(t, g)
	hap := pushCatalogPermanent(g, me.ID, "Hapatra, the Desert Frost", rfcLegendaryCreature, rfcHapatraFrostOracle, false)
	bear := pushCreatureToBattlefieldForTest(g, me.ID, "Bear")
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear) })
	if err := g.ActivateCatalogAbility(me.ID, hap, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, bear); c.Tapped {
		t.Error("the creature should be untapped")
	}
}

// --- Heartstring Puller -------------------------------------------

func TestHeartstringPullerMakesACadet(t *testing.T) {
	g := newCatalogGame(t)
	castHoldCreature(t, g, "Heartstring Puller", "Creature — Elf Sorcerer", rfcHeartstringOracle, 3, 1)
	cadets := battlefieldIDsNamed(g, "Cadet")
	if len(cadets) != 1 {
		t.Fatalf("Cadets = %d, want 1", len(cadets))
	}
	if effectivePower(t, g, cadets[0]) != 2 {
		t.Error("Cadet should be a 2/2")
	}
}

// --- Hexhaven Invigorator -----------------------------------------

func TestHexhavenInvigoratorFetchesThatManyLandsTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := rfcSeats(g)
	toMain(t, g)
	inv := pushDiesCreatureForTest(g, me.ID, "Hexhaven Invigorator", rfcInvigoratorOracle, "Creature — Chimera Horror", 6, 6)
	var lib []uuid.UUID
	for i := 0; i < 4; i++ {
		lib = append(lib, pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}))
	}
	before := len(battlefieldIDsNamed(g, "Forest"))
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.Nil, inv, 3) })
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, lib[0], lib[1], lib[2])
	passPriorityAroundTable(t, g)
	forests := battlefieldIDsNamed(g, "Forest")
	if len(forests)-before != 3 {
		t.Fatalf("lands fetched = %d, want 3", len(forests)-before)
	}
	for _, id := range forests {
		if c, _ := battlefieldCard(g, id); !c.Tapped {
			t.Error("fetched lands enter tapped")
		}
	}
}

// rfcPlayLand is playLandFromHand with a real type line, so a land
// carries the basic land type a landfall rider reads.
func rfcPlayLand(t *testing.T, g *game.Game, name, typeLine string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Owner: active.ID, Controller: active.ID,
	})
	advanceToMain(t, g)
	active.LandDropsPerTurn++
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play %s: %v", name, err)
	}
	return id
}
