package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const tardisOracle = "9167a585-b742-427d-b840-e024f6878a5c"

func TestTardisIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(tardisOracle)
	if !ok || spec.Name != "TARDIS" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("TARDIS spec = %+v, registered %v", spec, ok)
	}
}

// tardisAttack puts a crewed TARDIS (a Vehicle that is a creature) onto the
// battlefield, with a Time Lord beside it when timeLord, and attacks with it.
func tardisAttack(t *testing.T, g *game.Game, timeLord bool) uuid.UUID {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	if timeLord {
		effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "The Doctor", TypeLine: "Legendary Creature — Time Lord Doctor", Power: 2, Toughness: 2})
	}
	tardis := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "TARDIS", OracleID: tardisOracle,
		TypeLine: "Artifact Creature — Vehicle", Power: 2, Toughness: 4, Keywords: []string{"flying"}})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(tardis, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	return tardis
}

func TestTardisAttackWithATimeLordGivesTheNextSpellCascadeOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedCheapLibrary(g, me, 6)
	tardisAttack(t, g, true)

	effectstest.CastWithCost(t, g, "First Sorcery", "Sorcery", "{3}{G}", "")
	if got, _ := tardisSettleCast(t, g, me.ID, false); got != 1 {
		t.Fatalf("the next spell got %d cascade offers, want 1", got)
	}

	effectstest.CastWithCost(t, g, "Second Sorcery", "Sorcery", "{3}{G}", "")
	if got, asked := tardisSettleCast(t, g, me.ID, false); got != 0 || asked {
		t.Errorf("the second spell got %d cascade offers (planeswalk asked %v), want none", got, asked)
	}
}

func TestTardisAttackWithoutATimeLordDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedCheapLibrary(g, me, 6)
	tardisAttack(t, g, false)

	effectstest.CastWithCost(t, g, "First Sorcery", "Sorcery", "{3}{G}", "")
	if got, asked := tardisSettleCast(t, g, me.ID, false); got != 0 || asked {
		t.Errorf("no Time Lord, yet the next spell got %d cascade offers (planeswalk asked %v)", got, asked)
	}
}

func TestTardisPlaneswalkIsOptionalAndHappensOnTheNextCast(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g, up := effectstest.NewPlaneGame(t, effectstest.PlaneFixture{
			Oracle: "22222222-3333-4444-8555-666666666666", Name: "Tardis Test Plane", TypeLine: "Plane — Testworld"})
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		effectstest.SeedCheapLibrary(g, me, 6)
		tardisAttack(t, g, true)
		if stillUp(g, up.InstanceID) == false {
			t.Fatal("the plane changed before any spell was cast")
		}
		effectstest.CastWithCost(t, g, "First Sorcery", "Sorcery", "{3}{G}", "")
		if _, asked := tardisSettleCast(t, g, me.ID, yes); !asked {
			t.Fatal("the planeswalk question was never asked")
		}
		if got := !stillUp(g, up.InstanceID); got != yes {
			t.Errorf("answered %v: planeswalked = %v", yes, got)
		}
	}
}

// tardisSettleCast settles the stack after a spell was cast: the two delayed
// triggers are ordered, the planeswalk question (when asked) is answered
// planeswalk, and every cascade offer is declined. It returns how many cascade
// offers the cast produced and whether the planeswalk question was asked.
func tardisSettleCast(t *testing.T, g *game.Game, me uuid.UUID, planeswalk bool) (offers int, asked bool) {
	t.Helper()
	for i := 0; i < 16; i++ {
		effectstest.DrcSettle(t, g, me)
		if effectstest.LatestTriggerPromptFor(g, me) != nil {
			asked = true
			effectstest.AnswerLatestTriggerPrompt(t, g, me, planeswalk)
			continue
		}
		if n := effectstest.CountMayCastPrompts(g); n > 0 {
			offers += n
			effectstest.AnswerAllMayCast(t, g, me, false)
			continue
		}
		if effectstest.StackFullyEmpty(g) {
			return offers, asked
		}
	}
	t.Fatal("the stack never settled")
	return 0, false
}

func stillUp(g *game.Game, id uuid.UUID) bool {
	for _, c := range g.PlanarFaceUpCards() {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}
