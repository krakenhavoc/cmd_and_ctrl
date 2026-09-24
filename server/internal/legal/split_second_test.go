package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// split_second_test.go — #1519, the bot's-eye view of a PRINTED split
// second. TestSplitSecondBlocksCastsNotPass sets the cache by hand;
// this one gets there the way a table does, by casting a Krosan Grip
// from its catalog entry, and then asks what the seat being answered
// may do. It must be offered pass and its mana abilities and nothing
// else, and every one of those must be a move the engine accepts —
// the property a bot seat relies on to never stall on a refusal.

const oracleKrosanGrip = "3e39224c-72ce-4ecc-aa17-12c071ea1f3e"

func TestAPrintedSplitSecondLeavesTheBotPassAndMana(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(opp)
	handCard(opp, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleLightningBolt})
	battlefieldCard(g, opp, basic("Mountain", "Mountain"))
	battlefieldCard(g, opp, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	bombardment := battlefieldCard(g, opp, game.Card{Name: "Goblin Bombardment", TypeLine: "Enchantment", OracleID: oracleGoblinBombardment})
	advanceTo(t, g, game.StepPrecombatMain)

	// Before the Grip: the opponent, given priority, has a Bolt and a
	// Bombardment to answer with. Establishes that the empty lists
	// below are split second's doing and not the board's.
	g.Turn.PriorityHolder = opp.Seat
	if moves := legal.EnumerateFor(g, opp.ID); countKind(moves, legal.KindCast) == 0 || countKind(moves, legal.KindActivate) == 0 {
		t.Fatalf("setup: the opponent has no cast or activation to lose: %v", labels(moves))
	}
	g.Turn.PriorityHolder = active.Seat

	grip := handCard(active, game.Card{Name: "Krosan Grip", TypeLine: "Instant", ManaCost: "{2}{G}", OracleID: oracleKrosanGrip})
	if err := g.CastSpell(active.ID, grip, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bombardment}},
	}); err != nil {
		t.Fatalf("cast Krosan Grip: %v", err)
	}
	if !g.SplitSecondActive {
		t.Fatal("Krosan Grip is on the stack without split second")
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("caster passes: %v", err)
	}
	if g.Turn.PriorityHolder != opp.Seat {
		t.Fatalf("priority went to seat %d, want the opponent's %d", g.Turn.PriorityHolder, opp.Seat)
	}

	moves := legal.EnumerateFor(g, opp.ID)
	if n := countKind(moves, legal.KindCast) + countKind(moves, legal.KindActivate); n != 0 {
		t.Errorf("CR 702.61a: %d cast/activate moves offered in response to Krosan Grip: %v", n, labels(moves))
	}
	if !hasLabel(moves, "Pass priority") {
		t.Errorf("pass is not offered under split second: %v", labels(moves))
	}
	if countKind(moves, legal.KindMana) == 0 {
		t.Errorf("CR 702.61b: no mana move offered under split second: %v", labels(moves))
	}
	dispatchAll(t, g, opp.ID, moves)
}
