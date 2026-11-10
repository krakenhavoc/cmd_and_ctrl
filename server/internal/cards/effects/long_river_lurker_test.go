package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const longRiverLurkerOracle = "c65c4d01-29da-47e6-993c-a265cf1f198c"

func TestLongRiverLurkerIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(longRiverLurkerOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Long River Lurker: %+v %v", spec, ok)
	}
}

// lrlCastLurker casts the Lurker from hand, with `target` as its entry
// trigger's target, and settles the stack.
func lrlCastLurker(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Long River Lurker", TypeLine: "Creature — Frog Scout", OracleID: longRiverLurkerOracle,
			ManaCost: "{2}{U}", Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID})
	})
	floatForTest(g, me, "UUU")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, target)
	passPriorityAroundTable(t, g)
	return id
}

func TestLongRiverLurkerMakesTheTargetUnblockableAndMayFlickerItOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	runner := b12Creature(g, me.ID, "Runner", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterByKeyedForEffect(uuid.Nil, runner, game.CounterPlusOne, 1, game.KeyedTail{}) })
	blocker := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	advanceToMain(t, g)
	lrlCastLurker(t, g, me, runner)
	r, _ := battlefieldCard(g, runner)
	oldEpoch := r.ObjectEpoch
	assertRestrictions(t, g, runner, game.CantBeBlocked)
	advanceTo(t, g, game.StepDeclareAttackers)
	declareAttack(t, g, opp.ID, runner)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, runner); err == nil {
		t.Error("the unblockable creature was blocked")
	}
	life := opp.Life
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if life-opp.Life != 3 {
		t.Errorf("the defender lost %d, want 3 (2/2 with a counter)", life-opp.Life)
	}
	if _, still := battlefieldCard(g, runner); still {
		t.Error("the creature is still the same object after being exiled and returned")
	}
	back := battlefieldIDsNamed(g, "Runner")
	if len(back) != 1 {
		t.Fatalf("the battlefield has %d Runners, want the creature back as a new object", len(back))
	}
	r2, _ := battlefieldCard(g, back[0])
	if r2.ObjectEpoch == oldEpoch && back[0] == runner {
		t.Error("the same object")
	}
	if counterOn(g, back[0], game.CounterPlusOne) > 0 {
		t.Error("the counter survived the blink")
	}
}

func TestLongRiverLurkerGivesOtherFrogsWardOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dskB1Creature(g, me.ID, "Long River Lurker", "Creature — Frog Scout", longRiverLurkerOracle, 2, 3)
	frog := b12Creature(g, me.ID, "Other Frog", "Creature — Frog", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	// A Frog is warded; a Bear is not.
	castAtWardedCreature(t, g, opp, bear)
	for i := 0; i < 6; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if hasPayUnlessFor(g, opp.ID) {
		t.Error("a non-Frog has ward")
	}
	passPriorityAroundTable(t, g)
	castAtWardedCreature(t, g, opp, frog)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt for a spell aimed at the other Frog")
	}
}
