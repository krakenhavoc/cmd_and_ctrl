package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- damage bullets: creature and planeswalker, two clauses -------

func TestMishrasCommandDealsXDamageToCreatureAndPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp, other := g.Seats[1], g.Seats[2]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: other.ID, Controller: other.ID, Counters: map[string]int{game.CounterLoyalty: 5},
	})

	effectstest.B19CastXModal(t, g, "Mishra's Command", "Sorcery", effectstest.MishrasCommandOracle, "{X}{R}", 3,
		[]int{1, 2},
		[]game.TargetRef{
			effectstest.ModeRef(game.TargetCard, bear, 0, 0),   // occurrence 0 = "deals X damage to target creature"
			effectstest.ModeRef(game.TargetCard, walker, 1, 0), // occurrence 1 = "deals X damage to target planeswalker"
		})
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) {
		t.Error("X=3 damage should have killed the 2/2")
	}
	if got := effectstest.CounterCount(g, walker, game.CounterLoyalty); got != 2 {
		t.Errorf("X=3 damage to a planeswalker removes 3 loyalty: %d, want 2", got)
	}
}

// A creature named in the planeswalker slot is illegal — each
// bullet's clause is checked on its own, never the union of both.
func TestMishrasCommandRefusesATargetForTheWrongMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	active := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)

	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Mishra's Command", TypeLine: "Sorcery", ManaCost: "{X}{R}",
		OracleID: effectstest.MishrasCommandOracle, Owner: active.ID, Controller: active.ID,
	})
	err := g.CastSpell(active.ID, id, game.CastSpellParams{
		XValue: 3,
		Modes:  []int{1, 2},
		Targets: []game.TargetRef{
			effectstest.ModeRef(game.TargetCard, bear, 0, 0),
			effectstest.ModeRef(game.TargetCard, bear, 1, 0),
		},
	})
	if err != game.ErrIllegalTarget {
		t.Errorf("a creature in the 'target planeswalker' slot: err %v, want ErrIllegalTarget", err)
	}
}

// --- pump bullet, targeting a DIFFERENT creature than the burn ----

func TestMishrasCommandPumpsOneCreatureWhileBurningAnother(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Elk", 3, 3)

	effectstest.B19CastXModal(t, g, "Mishra's Command", "Sorcery", effectstest.MishrasCommandOracle, "{X}{R}", 3,
		[]int{1, 3},
		[]game.TargetRef{
			effectstest.ModeRef(game.TargetCard, theirs, 0, 0), // occurrence 0 = the damage bullet
			effectstest.ModeRef(game.TargetCard, mine, 1, 0),   // occurrence 1 = the pump bullet
		})
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(theirs) {
		t.Error("X=3 damage should have killed the 3/3")
	}
	if got := effectstest.EffectivePower(t, g, mine); got != 2+3 {
		t.Errorf("power after +X/+0: %d, want %d", got, 2+3)
	}
	if got := effectstest.EffectiveToughness(t, g, mine); got != 2 {
		t.Errorf("toughness is unaffected by +X/+0: %d, want 2", got)
	}
	abilities := effectstest.EffectiveAbilities(t, g, mine)
	found := false
	for _, a := range abilities {
		if a == "haste" {
			found = true
		}
	}
	if !found {
		t.Errorf("the pumped creature should have gained haste: %v", abilities)
	}
}

// --- discard-up-to-X, then draw exactly what was discarded --------

// The pump bullet waits for the discard bullet's question: CR
// 608.2c follows the instructions in the order written, so the
// bullet printed later runs only once the target player has answered
// (#2789). And the draw counts what was REALLY discarded, not X
// (#1027, PlayerDiscardsThenKeyedForEffect).
func TestMishrasCommandDiscardsUpToXThenDrawsExactlyWhatWasDiscarded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	oppHandBefore := opp.Hand.Size()

	effectstest.B19CastXModal(t, g, "Mishra's Command", "Sorcery", effectstest.MishrasCommandOracle, "{X}{R}", 3,
		[]int{0, 3},
		[]game.TargetRef{
			effectstest.ModeRef(game.TargetPlayer, opp.ID, 0, 0), // occurrence 0 = the discard-then-draw bullet
			effectstest.ModeRef(game.TargetCard, mine, 1, 0),     // occurrence 1 = the pump bullet
		})
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.EffectivePower(t, g, mine); got != 2 {
		t.Fatalf("the pump bullet ran before the discard was answered: power %d, want 2", got)
	}
	c := effectstest.DiscardChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the discard bullet queued no prompt for the target player")
	}
	if c.ChooseMin != 0 || c.ChooseMax != 3 {
		t.Errorf("up to X=3: bounds [%d,%d], want [0,3]", c.ChooseMin, c.ChooseMax)
	}

	// Discard FEWER than X — the draw must follow the actual count.
	ids := make([]uuid.UUID, 2)
	for i := range ids {
		ids[i] = opp.Hand.Cards[i].InstanceID
	}
	effectstest.AnswerDiscard(t, g, opp.ID, ids...)
	g.SettleResolution()

	if got := effectstest.EffectivePower(t, g, mine); got != 2+3 {
		t.Errorf("the pump bullet runs once the discard is answered: power %d, want %d", got, 2+3)
	}

	if got := opp.Hand.Size(); got != oppHandBefore {
		t.Errorf("discarded 2, drew 2 back: hand %d, want %d (net unchanged)", got, oppHandBefore)
	}
	if got := opp.Graveyard.Size(); got != 2 {
		t.Errorf("graveyard should hold the 2 discarded cards, has %d", got)
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("the discard is answered, so priority passes: %v", err)
	}
}
