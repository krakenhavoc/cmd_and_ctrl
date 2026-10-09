package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tomik_orzhov_lawmage_test.go — #2821: an attack limit granted to other
// permanents. Tomik gives every planeswalker its controller controls "No
// more than one creature can attack this planeswalker each combat."

const tomikOrzhovLawmageOracle = "7a501f7e-eec8-45f7-9ac3-483fd8f5ca5e"

// Each of the opponent's planeswalkers takes one attacker, counted on its
// own; the player is not limited; a planeswalker of another player is not
// limited; and the enumerator offers exactly what the verb accepts.
func TestTomikLimitsEachOfYourPlaneswalkersToOneAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, opp.ID, "Tomik, Orzhov Lawmage", "Legendary Creature — Human Advisor", tomikOrzhovLawmageOracle, 2, 1)
	first := pushWalkerForTest(g, opp.ID, "First Walker", "", 4)
	second := pushWalkerForTest(g, opp.ID, "Second Walker", "", 4)
	var bears []uuid.UUID
	for i := 0; i < 5; i++ {
		bears = append(bears, b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	}
	advanceTo(t, g, game.StepDeclareAttackers)

	if err := g.DeclareAttacker(bears[0], first); err != nil {
		t.Fatalf("one attacker at the first walker: %v", err)
	}
	at := clTargets(clAgreeOnAttacksAnywhere(t, g, me.ID))
	if at[first] != 0 || at[second] != 4 || at[opp.ID] != 4 {
		t.Fatalf("offers after one attacker at the first walker: %v", at)
	}
	le := clAttackRefusal(t, g.DeclareAttacker(bears[1], first))
	if le.Source != first || le.Max != 1 {
		t.Errorf("refusal = %+v", le)
	}
	if got, want := le.Sentence(me.ID), "No more than one creature can attack First Walker each combat."; got != want {
		t.Errorf("sentence %q, want %q", got, want)
	}
	if err := g.DeclareAttacker(bears[1], second); err != nil {
		t.Fatalf("the second walker has a limit of its own: %v", err)
	}
	clAttackRefusal(t, g.DeclareAttacker(bears[2], second))
	if err := g.DeclareAttacker(bears[2], opp.ID); err != nil {
		t.Fatalf("Tomik's controller: %v", err)
	}
	if err := g.DeclareAttacker(bears[3], opp.ID); err != nil {
		t.Fatalf("Tomik's controller, twice: %v", err)
	}
}

// The grant is read live: a planeswalker of a player without Tomik is not
// limited, and once Tomik leaves the battlefield neither is any other.
func TestTomikLimitEndsWithTomik(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tomik := b12Push(g, opp.ID, "Tomik, Orzhov Lawmage", "Legendary Creature — Human Advisor", tomikOrzhovLawmageOracle, 2, 1)
	walker := pushWalkerForTest(g, opp.ID, "Walker", "", 4)
	other := pushWalkerForTest(g, g.Seats[2].ID, "Unprotected Walker", "", 4)
	var bears []uuid.UUID
	for i := 0; i < 4; i++ {
		bears = append(bears, b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	}
	advanceTo(t, g, game.StepDeclareAttackers)

	for i := 0; i < 2; i++ {
		if err := g.DeclareAttacker(bears[i], other); err != nil {
			t.Fatalf("attacker %d at another player's walker: %v", i, err)
		}
	}
	if err := g.DeclareAttacker(bears[2], walker); err != nil {
		t.Fatalf("one attacker at Tomik's walker: %v", err)
	}
	clAttackRefusal(t, g.DeclareAttacker(bears[3], walker))

	b25Destroy(g, tomik)
	if err := g.DeclareAttacker(bears[3], walker); err != nil {
		t.Errorf("with Tomik gone the walker is unlimited: %v", err)
	}
}
