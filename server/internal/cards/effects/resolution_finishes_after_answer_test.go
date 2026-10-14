package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// resolution_finishes_after_answer_test.go — #3613, CR 608.2c and
// 608.2n: a spell or ability that asks something as it resolves has not
// resolved until the answer is in. The public log's "… resolved" line
// and, for a spell, the move to the graveyard follow the answer, and
// the item stays on the stack while the prompt is open.

// resolvedLineAt is the index in the public log of the "resolved" line
// whose card or stack item is id, or -1.
func resolvedLineAt(log []protocol.LogEvent, id uuid.UUID) int {
	for i, e := range log {
		if e.Kind == protocol.LogResolve && (e.CardID == id.String() || e.StackItemID == id.String()) {
			return i
		}
	}
	return -1
}

// lastLineAbout is the index of the last log line naming card id, or -1.
func lastLineAbout(log []protocol.LogEvent, kind protocol.LogKind, id uuid.UUID) int {
	at := -1
	for i, e := range log {
		if e.Kind == kind && e.CardID == id.String() {
			at = i
		}
	}
	return at
}

// stackShows reports whether the view's stack lane lists item id.
func stackShows(v protocol.GameView, id uuid.UUID) bool {
	for _, it := range v.StackItems {
		if it.ID == id.String() {
			return true
		}
	}
	return false
}

func TestCouncilsJudgmentResolvesAfterTheVote(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	a := ctrlPushCreature(g, order[1].ID, "Voted Bear")
	ctrlPushCreature(g, order[2].ID, "Other Bear")
	spell := castCatalogSpell(t, g, "Council's Judgment", "Sorcery", councilsJudgmentOracle, nil)
	passPriorityAroundTable(t, g)
	if ballotFor(g, me.ID) == nil {
		t.Fatal("Council's Judgment opens a vote as it resolves")
	}

	// The vote is part of the resolution, so the spell is still on the
	// stack and has not resolved.
	if !g.Stack.Contains(spell) {
		t.Error("CR 608.2n: the spell is still on the stack while the vote is open")
	}
	if me.Graveyard.Contains(spell) {
		t.Error("the spell is in the graveyard before the vote is over")
	}
	v := protocol.ViewOfGame(g)
	if at := resolvedLineAt(v.Log, spell); at >= 0 {
		t.Errorf("the log says %q before the vote", v.Log[at].Text)
	}
	if !stackShows(v, spell) {
		t.Error("the stack lane drops the spell while its vote is open")
	}

	for _, p := range order {
		voteFor(t, g, p.ID, "Voted Bear")
	}
	if g.Battlefield.Contains(a) {
		t.Fatal("the voted permanent is exiled")
	}
	if g.Stack.Contains(spell) || !me.Graveyard.Contains(spell) {
		t.Error("once the vote is over the spell goes to its owner's graveyard")
	}
	v = protocol.ViewOfGame(g)
	resolved := resolvedLineAt(v.Log, spell)
	if resolved < 0 {
		t.Fatal("no \"resolved\" line once the vote is over")
	}
	if exiled := lastLineAbout(v.Log, protocol.LogZone, a); exiled < 0 || exiled > resolved {
		t.Errorf("the exile (line %d) comes before \"resolved\" (line %d)", exiled, resolved)
	}
	if !strings.Contains(v.Log[resolved].Text, "Council's Judgment resolved") {
		t.Errorf("resolved line = %q", v.Log[resolved].Text)
	}
	if stackShows(v, spell) {
		t.Error("the spell is still in the stack lane after it resolved")
	}
	if n := countResolvedLines(v.Log, spell); n != 1 {
		t.Errorf("%d \"resolved\" lines for the spell, want 1", n)
	}
}

// countResolvedLines counts the "resolved" lines naming id.
func countResolvedLines(log []protocol.LogEvent, id uuid.UUID) int {
	n := 0
	for _, e := range log {
		if e.Kind == protocol.LogResolve && (e.CardID == id.String() || e.StackItemID == id.String()) {
			n++
		}
	}
	return n
}

// fabricateResolveLine is the index of the fabricate trigger's
// "resolved" line, or -1.
func fabricateResolveLine(log []protocol.LogEvent) int {
	for i, e := range log {
		if e.Kind == protocol.LogResolve && strings.Contains(e.Text, "fabricate") {
			return i
		}
	}
	return -1
}

func TestFabricateResolvesAfterItsChoice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	angel := b36CastCreature(t, g, "Angel of Invention", "Creature — Angel", b40AngelOfInventionOracle, 2, 1)
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil {
		t.Fatal("fabricate 2 asks which mode")
	}
	v := protocol.ViewOfGame(g)
	if at := fabricateResolveLine(v.Log); at >= 0 {
		t.Errorf("the log says %q before the fabricate choice", v.Log[at].Text)
	}
	if len(v.StackItems) != 1 || !strings.Contains(v.StackItems[0].Label, "fabricate") {
		t.Errorf("the stack lane shows %+v while the fabricate choice is open, want the trigger", v.StackItems)
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if counterOn(g, angel, game.CounterPlusOne) != 2 {
		t.Fatal("the counters branch puts two +1/+1 counters on the Angel")
	}
	v = protocol.ViewOfGame(g)
	resolved := fabricateResolveLine(v.Log)
	if resolved < 0 {
		t.Fatal("no \"resolved\" line once the choice is made")
	}
	if counters := lastLineAbout(v.Log, protocol.LogCounters, angel); counters < 0 || counters > resolved {
		t.Errorf("the counters (line %d) come before \"resolved\" (line %d)", counters, resolved)
	}
	if len(v.StackItems) != 0 {
		t.Errorf("the stack lane still shows %+v after the trigger resolved", v.StackItems)
	}
}
