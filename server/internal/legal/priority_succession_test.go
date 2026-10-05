package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// priority_succession_test.go — #2275, CR 117.4, from the bot's side.
//
// The enumerator offers moves to whoever holds priority, so it is only
// as good as the engine's answer to "who holds priority now". Before
// #2275 a non-active player's spell resolved on its caster's pass, so
// the active seat was never asked about it: a bot holding Counterspell
// in its own turn watched an opponent's Bolt resolve. Here the round
// restarts from the caster and comes back to the active seat with the
// Bolt still on the stack — and Counterspell is on offer, aimed at it.
func TestActiveSeatIsOfferedACounterToAnOpponentsSpell(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clearHand(active)
	clearHand(opp)
	cs := handCard(active, game.Card{Name: "Counterspell", TypeLine: "Instant", ManaCost: "{U}{U}", OracleID: oracleCounterspell})
	bolt := handCard(opp, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleLightningBolt})
	battlefieldCard(g, active, basic("Island", "Island"))
	battlefieldCard(g, active, basic("Island", "Island"))
	battlefieldCard(g, opp, basic("Mountain", "Mountain"))
	advanceTo(t, g, game.StepPrecombatMain)

	// The active player passes first; the opponent bolts them and
	// passes; the other two seats pass.
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: active.ID}},
		Strict:  true, AutoTap: true,
	}); err != nil {
		t.Fatal(err)
	}
	for range 3 {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}

	if !holdsPriority(g, active.ID) {
		t.Fatalf("seat %d holds priority, want the active seat with the Bolt on the stack", g.Turn.PriorityHolder)
	}
	if _, ok := g.StackMeta[bolt]; !ok {
		t.Fatal("the Bolt resolved before the active player held priority over it (CR 117.4)")
	}
	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	var counter *legal.Move
	for i := range moves {
		if moves[i].Source == cs && moves[i].Kind == legal.KindCast {
			counter = &moves[i]
		}
	}
	if counter == nil {
		t.Fatalf("Counterspell not offered against the opponent's Bolt: %v", labels(moves))
	}
	if !counter.TargetsStack {
		t.Errorf("the Counterspell move should target the stack: %+v", counter)
	}
}
