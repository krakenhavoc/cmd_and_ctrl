package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #3172 cav-h2: Venser's caveat said a token copy skips the copied card's
// as-enters clause. It does not (the token runs the ordinary entry
// pipeline, #762); this is the card-level proof.
func TestVenserTokenCopyAsksForTheCopiedCardsEntryChoice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	target := pushTokenReplacementCard(g, automatonOracle, "Adaptive Automaton", "Artifact Creature — Construct", opp.ID)
	rfCrFCast(t, g, "Venser, Fervent Forger", "Legendary Creature — Human Sorcerer", oracleVenserFervent, 5, 3)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	pickCard(t, g, me.ID, target)
	passPriorityAroundTable(t, g)

	if got := b28TokensNamed(g, me.ID, "Adaptive Automaton"); got != 2 {
		t.Fatalf("token copies of the Automaton: %d, want 2", got)
	}
	asked := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceCreatureType && c.Chooser == me.ID {
			asked++
		}
	}
	if asked == 0 {
		t.Error("the token copies did not ask for their creature type")
	}
}
