package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// extort_log_test.go — the public log tells the client an extort drain
// resolved (#4544). The client's Scales of Souls effect
// (client/src/lib/extort.ts) reads three things off the wire and nothing
// else: a `resolve` entry whose `label` ends in " — extort", and the
// `life` entries that share its `batch` (each opponent's loss, the
// controller's gain). This pins all three, so a rename of the label or a
// batch change fails here and not silently on the table.
func TestExtortDrainIsReadableFromThePublicLog(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sorin := sorinCard(me.ID)
	// A permanent face up on the battlefield is public: every seat knows it,
	// so the log names its abilities (a face-down one is redacted).
	sorin.KnownBy = map[uuid.UUID]bool{}
	for _, s := range g.Seats {
		sorin.KnownBy[s.ID] = true
	}
	pushBattlefieldCardWithTimestamp(g, sorin)

	e2CastSpellOfColor(t, g, "Some Spell", "Sorcery", "U")
	for i := 0; i < 8 && !hasPayUnlessFor(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("casting a spell offers the extort payment")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	answerPayUnless(t, g, me.ID, true)

	log := protocol.FilterViewFor(protocol.ViewOfGame(g), me.ID.String()).Log
	var batch uint64
	found := false
	for _, e := range log {
		if e.Kind == protocol.LogResolve && strings.HasSuffix(e.Label, " — extort") {
			batch, found = e.Batch, true
		}
	}
	if !found {
		for _, e := range log {
			t.Logf("%s seat=%d card=%s label=%q batch=%d amount=%d text=%q stack=%s", e.Kind, e.Seat, e.CardID, e.Label, e.Batch, e.Amount, e.Text, e.StackItemID)
		}
		t.Fatal(`no resolve entry labelled "… — extort"`)
	}
	if batch == 0 {
		t.Fatal("the extort resolve entry carries no batch")
	}
	lost, gained := 0, 0
	for _, e := range log {
		if e.Kind != protocol.LogLife || e.Batch != batch {
			continue
		}
		if e.Amount < 0 {
			lost++
		} else {
			gained++
			if e.Seat != 0 {
				t.Errorf("the gain is on seat %d, want the extort player's seat 0", e.Seat)
			}
		}
	}
	if want := len(g.Seats) - 1; lost != want || gained != 1 {
		t.Errorf("life entries in the extort batch: %d losses, %d gains; want %d and 1", lost, gained, want)
	}
}
