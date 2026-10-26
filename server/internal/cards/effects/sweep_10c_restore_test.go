package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_10c_restore_test.go — #4253 slice 10c: the card-name choice, the
// choose-a-source prompt, the land-mana walk and the draw-instead `done`
// are data, so a table waiting on one of them is a restore point. Each test
// takes a STRICT restore (RestoreStrict, the boot path) while the prompt
// is open, checks the snapshot wrote the prompt with its key, and answers
// it on the restored table.

// restoreAtKeyedPrompt strict-restores g after checking that an open prompt
// in its snapshot names the key (PickThen), which is what makes a binary
// without the key refuse the file.
func restoreAtKeyedPrompt(t *testing.T, g *game.Game, key string) *game.Game {
	t.Helper()
	snap := g.CaptureSnapshot()
	found := false
	for _, c := range snap.PendingChoices {
		if c.PickThen == key {
			found = true
		}
	}
	if !found {
		t.Fatalf("no prompt in the snapshot names %q: %+v", key, snap.Continuations)
	}
	if snap.Continuations.ChoiceResumeFrames != 0 {
		t.Fatalf("the open prompt still counts a closure frame: %+v", snap.Continuations)
	}
	return restoreRoundTrip(t, g, true)
}

// --- the card-name choice ---------------------------------------------

// Demonic Consultation's name prompt is a restore point: the restored
// answer exiles the top six and reveals for the named card.
func TestDemonicConsultationNamePromptIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	after := sweepFTop(me, "After", "Sorcery")
	wanted := sweepFTop(me, "Wanted", "Sorcery")
	other := sweepFTop(me, "Other", "Sorcery")
	var six []uuid.UUID
	for i := 0; i < 6; i++ {
		six = append(six, sweepFTop(me, "Junk", "Sorcery"))
	}

	castCatalogSpell(t, g, "Demonic Consultation", "Instant", sweepFConsultationOracle, nil)
	passPriorityAroundTable(t, g)

	g = restoreAtKeyedPrompt(t, g, "then/demonic-consultation-named")
	me = g.Seats[g.Turn.ActiveSeat]
	for _, id := range six {
		if inExile(g, id) {
			t.Fatal("the restored table exiled the top six before the name")
		}
	}
	actNameFor(t, g, me.ID, "Wanted")
	passPriorityAroundTable(t, g)

	for _, id := range six {
		if !inExile(g, id) {
			t.Errorf("a top-six card %s was not exiled after the restored answer", id)
		}
	}
	if !inExile(g, other) {
		t.Error("a revealed card that was not named was not exiled")
	}
	if !inHandOf(me, wanted) {
		t.Error("the named card is not in hand after the restored answer")
	}
	if !me.Library.Contains(after) {
		t.Error("a card below the named one left the library")
	}
}

// Tamiyo's +1 asks for a name; the restored table reveals four and takes
// the matches, the rest going to the graveyard.
func TestTamiyoNamePromptIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tamiyo := pushCatalogWalker(g, me.ID, "Tamiyo, Collector of Tales", tamiyoCollectorOracle, 5)
	for _, c := range []game.Card{
		{Name: "Sol Ring", TypeLine: "Artifact"},
		{Name: "Grizzly Bears", TypeLine: "Creature — Bear"},
		{Name: "Sol Ring", TypeLine: "Artifact"},
		{Name: "Forest", TypeLine: "Basic Land — Forest"},
	} {
		c.Owner, c.Controller, c.InstanceID = me.ID, me.ID, uuid.New()
		me.Library.PushTop(c)
	}
	advanceToMainOf(t, g, g.Turn.ActiveSeat)
	if err := g.ActivateCatalogAbility(me.ID, tamiyo, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)

	g = restoreAtKeyedPrompt(t, g, "then/tamiyo-collector-named")
	me = g.Seats[0]
	hand, grave := len(me.Hand.Cards), len(me.Graveyard.Cards)
	actNameFor(t, g, me.ID, "sol ring")
	passPriorityAroundTable(t, g)

	if got := len(me.Hand.Cards) - hand; got != 2 {
		t.Errorf("two Sol Rings go to hand after the restored answer, got %d", got)
	}
	if got := len(me.Graveyard.Cards) - grave; got != 2 {
		t.Errorf("the other two go to the graveyard after the restored answer, got %d", got)
	}
}
