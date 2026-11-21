package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- the card-name choice ---------------------------------------------

// Demonic Consultation's name prompt is a restore point: the restored
// answer exiles the top six and reveals for the named card.
func TestDemonicConsultationNamePromptIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	after := effectstest.SweepFTop(me, "After", "Sorcery")
	wanted := effectstest.SweepFTop(me, "Wanted", "Sorcery")
	other := effectstest.SweepFTop(me, "Other", "Sorcery")
	var six []uuid.UUID
	for i := 0; i < 6; i++ {
		six = append(six, effectstest.SweepFTop(me, "Junk", "Sorcery"))
	}

	effectstest.CastCatalogSpell(t, g, "Demonic Consultation", "Instant", effectstest.SweepFConsultationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	g = effectstest.RestoreAtKeyedPrompt(t, g, "then/demonic-consultation-named")
	me = g.Seats[g.Turn.ActiveSeat]
	for _, id := range six {
		if effectstest.InExile(g, id) {
			t.Fatal("the restored table exiled the top six before the name")
		}
	}
	effectstest.ActNameFor(t, g, me.ID, "Wanted")
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range six {
		if !effectstest.InExile(g, id) {
			t.Errorf("a top-six card %s was not exiled after the restored answer", id)
		}
	}
	if !effectstest.InExile(g, other) {
		t.Error("a revealed card that was not named was not exiled")
	}
	if !effectstest.InHandOf(me, wanted) {
		t.Error("the named card is not in hand after the restored answer")
	}
	if !me.Library.Contains(after) {
		t.Error("a card below the named one left the library")
	}
}

// Tamiyo's +1 asks for a name; the restored table reveals four and takes
// the matches, the rest going to the graveyard.
func TestTamiyoNamePromptIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tamiyo := effectstest.PushCatalogWalker(g, me.ID, "Tamiyo, Collector of Tales", effectstest.TamiyoCollectorOracle, 5)
	for _, c := range []game.Card{
		{Name: "Sol Ring", TypeLine: "Artifact"},
		{Name: "Grizzly Bears", TypeLine: "Creature — Bear"},
		{Name: "Sol Ring", TypeLine: "Artifact"},
		{Name: "Forest", TypeLine: "Basic Land — Forest"},
	} {
		c.Owner, c.Controller, c.InstanceID = me.ID, me.ID, uuid.New()
		me.Library.PushTop(c)
	}
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	if err := g.ActivateCatalogAbility(me.ID, tamiyo, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	g = effectstest.RestoreAtKeyedPrompt(t, g, "then/tamiyo-collector-named")
	me = g.Seats[0]
	hand, grave := len(me.Hand.Cards), len(me.Graveyard.Cards)
	effectstest.ActNameFor(t, g, me.ID, "sol ring")
	effectstest.PassPriorityAroundTable(t, g)

	if got := len(me.Hand.Cards) - hand; got != 2 {
		t.Errorf("two Sol Rings go to hand after the restored answer, got %d", got)
	}
	if got := len(me.Graveyard.Cards) - grave; got != 2 {
		t.Errorf("the other two go to the graveyard after the restored answer, got %d", got)
	}
}
