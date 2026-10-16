package legal_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// #1943: Breaking of the Fellowship's second target is "another target
// creature that player controls". The enumerator offers only pairs
// whose second creature shares the first's controller, never the same
// creature twice, and never a first pick whose controller has no
// second creature. dispatchAll proves the cast path accepts every one.
func TestBreakingOfTheFellowshipEnumeratesSameControllerPairs(t *testing.T) {
	g := newTable(t)
	me, opp, lone := g.Seats[0], g.Seats[1], g.Seats[2]
	clearHand(me)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, me, 2, "Mountain")
	battlefieldCard(g, me, creature("Mine", "{1}{G}", 2, 2))
	giant := battlefieldCard(g, opp, creature("Giant", "{2}{R}", 3, 3))
	bear := battlefieldCard(g, opp, creature("Bear", "{1}{G}", 2, 2))
	alone := battlefieldCard(g, lone, creature("Alone", "{1}{G}", 2, 2))
	spell := handCard(me, game.Card{
		Name: "Breaking of the Fellowship", TypeLine: "Sorcery", ManaCost: "{1}{R}", Colors: []string{"R"},
		OracleID: "96e081d2-ec72-41fc-ad81-2e31ab823ae0",
	})

	controller := map[string]uuid.UUID{giant.String(): opp.ID, bear.String(): opp.ID, alone.String(): lone.ID}
	moves := legal.EnumerateFor(g, me.ID)
	seen := map[[2]string]bool{}
	for _, c := range castPayloadsOf(t, moves, spell) {
		if len(c.Targets) != 2 {
			t.Fatalf("a cast names %d targets: %+v", len(c.Targets), c)
		}
		first, second := c.Targets[0].ID, c.Targets[1].ID
		if first == alone.String() || second == alone.String() {
			t.Fatalf("offered a creature with no partner under its controller: %+v", c)
		}
		if first == second || controller[first] != controller[second] {
			t.Fatalf("offered an unrelated pair: %+v", c)
		}
		seen[[2]string{first, second}] = true
	}
	if !seen[[2]string{giant.String(), bear.String()}] || !seen[[2]string{bear.String(), giant.String()}] || len(seen) != 2 {
		t.Fatalf("offered pairs: %v (%v)", seen, labels(moves))
	}
	dispatchAll(t, g, me.ID, moves)
}
