package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mishras_command_test.go — #1112. Mishra's Command is a "choose
// two" with per-mode targets (#764, Kolaghan's Command's shape),
// with an {X} in the cast cost besides: every test here also pins
// that ctx.X() reads back the announced value inside a modal bullet,
// and that a bullet's target is checked against ITS OWN clause (CR
// 608.2b), never the clause of the bullet chosen beside it.

const mishrasCommandOracle = "25434ea3-bcfa-4dae-a16e-10ab87ce32af"

// --- damage bullets: creature and planeswalker, two clauses -------

func TestMishrasCommandDealsXDamageToCreatureAndPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	opp, other := g.Seats[1], g.Seats[2]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: other.ID, Controller: other.ID, Counters: map[string]int{game.CounterLoyalty: 5},
	})

	b19CastXModal(t, g, "Mishra's Command", "Sorcery", mishrasCommandOracle, "{X}{R}", 3,
		[]int{1, 2},
		[]game.TargetRef{
			modeRef(game.TargetCard, bear, 0, 0),   // occurrence 0 = "deals X damage to target creature"
			modeRef(game.TargetCard, walker, 1, 0), // occurrence 1 = "deals X damage to target planeswalker"
		})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) {
		t.Error("X=3 damage should have killed the 2/2")
	}
	if got := counterCount(g, walker, game.CounterLoyalty); got != 2 {
		t.Errorf("X=3 damage to a planeswalker removes 3 loyalty: %d, want 2", got)
	}
}

// A creature named in the planeswalker slot is illegal — each
// bullet's clause is checked on its own, never the union of both.
func TestMishrasCommandRefusesATargetForTheWrongMode(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	active := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)

	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Mishra's Command", TypeLine: "Sorcery", ManaCost: "{X}{R}",
		OracleID: mishrasCommandOracle, Owner: active.ID, Controller: active.ID,
	})
	err := g.CastSpell(active.ID, id, game.CastSpellParams{
		XValue: 3,
		Modes:  []int{1, 2},
		Targets: []game.TargetRef{
			modeRef(game.TargetCard, bear, 0, 0),
			modeRef(game.TargetCard, bear, 1, 0),
		},
	})
	if err != game.ErrIllegalTarget {
		t.Errorf("a creature in the 'target planeswalker' slot: err %v, want ErrIllegalTarget", err)
	}
}

// --- pump bullet, targeting a DIFFERENT creature than the burn ----

func TestMishrasCommandPumpsOneCreatureWhileBurningAnother(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Elk", 3, 3)

	b19CastXModal(t, g, "Mishra's Command", "Sorcery", mishrasCommandOracle, "{X}{R}", 3,
		[]int{1, 3},
		[]game.TargetRef{
			modeRef(game.TargetCard, theirs, 0, 0), // occurrence 0 = the damage bullet
			modeRef(game.TargetCard, mine, 1, 0),   // occurrence 1 = the pump bullet
		})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(theirs) {
		t.Error("X=3 damage should have killed the 3/3")
	}
	if got := effectivePower(t, g, mine); got != 2+3 {
		t.Errorf("power after +X/+0: %d, want %d", got, 2+3)
	}
	if got := effectiveToughness(t, g, mine); got != 2 {
		t.Errorf("toughness is unaffected by +X/+0: %d, want 2", got)
	}
	abilities := effectiveAbilities(t, g, mine)
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

// The pump bullet resolves even though the discard bullet's own
// question is still open: the mode-effect loop runs every chosen
// bullet in one pass (CR 700.2c); only the discard's OWN
// continuation waits on the target player's answer. And the draw
// counts what was REALLY discarded, not X (#1027,
// PlayerDiscardsThenForEffect).
func TestMishrasCommandDiscardsUpToXThenDrawsExactlyWhatWasDiscarded(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	oppHandBefore := opp.Hand.Size()

	b19CastXModal(t, g, "Mishra's Command", "Sorcery", mishrasCommandOracle, "{X}{R}", 3,
		[]int{0, 3},
		[]game.TargetRef{
			modeRef(game.TargetPlayer, opp.ID, 0, 0), // occurrence 0 = the discard-then-draw bullet
			modeRef(game.TargetCard, mine, 1, 0),     // occurrence 1 = the pump bullet
		})
	passPriorityAroundTable(t, g)

	if got := effectivePower(t, g, mine); got != 2+3 {
		t.Fatalf("the pump bullet should have resolved already: power %d, want %d", got, 2+3)
	}
	c := discardChoiceFor(g, opp.ID)
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
	answerDiscard(t, g, opp.ID, ids...)

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
