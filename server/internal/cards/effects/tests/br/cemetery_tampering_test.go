package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const cemeteryTamperingOracle = "9babcd0a-531f-417d-905e-b916a87112bb"

func TestCemeteryTamperingIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(cemeteryTamperingOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Cemetery Tampering: %+v %v", spec, ok)
	}
}

// cemeteryTamperingTable enters the Tampering through the real cast, hides
// the third of five cards, and puts n cards in the graveyard.
func cemeteryTamperingTable(t *testing.T, n int) (g *game.Game, me *game.Player, hidden uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	top := effectstest.StackTopLibrary(me, "A", "B", "C", "D", "E")
	card := game.NewCard("Cemetery Tampering", me.ID)
	card.OracleID = cemeteryTamperingOracle
	card.TypeLine = "Enchantment"
	card.ManaCost = "{2}{B}"
	effectstest.EnterWithHideaway(t, g, card, top[2])
	g.WithWriteLock(func() {
		for i := 0; i < n; i++ {
			me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
		}
	})
	return g, me, top[2]
}

func TestCemeteryTamperingMillsThenPlaysTheExiledCardAtTwentyCards(t *testing.T) {
	g, me, hidden := cemeteryTamperingTable(t, 17)
	effectstest.BeginUpkeep(t, g, me.ID)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Graveyard.Size(); got < 20 {
		t.Fatalf("graveyard = %d, want at least 20 after milling three", got)
	}
	if effectstest.PlayHiddenGrant(g, hidden) == nil {
		t.Error("twenty cards in the graveyard but the exiled card has no free play")
	}
}

func TestCemeteryTamperingBelowTwentyGivesNothing(t *testing.T) {
	g, me, hidden := cemeteryTamperingTable(t, 10)
	effectstest.BeginUpkeep(t, g, me.ID)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlayHiddenGrant(g, hidden) != nil {
		t.Error("the exiled card is playable with fewer than twenty cards in the graveyard")
	}
}

func TestCemeteryTamperingDecliningTheMillStillChecksTheGraveyard(t *testing.T) {
	g, me, hidden := cemeteryTamperingTable(t, 20)
	effectstest.BeginUpkeep(t, g, me.ID)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlayHiddenGrant(g, hidden) == nil {
		t.Error("twenty cards without milling should still offer the exiled card (\"Then if\")")
	}
	if me.Graveyard.Size() != 20 {
		t.Errorf("graveyard = %d, want 20 (declined to mill)", me.Graveyard.Size())
	}
}

func TestCemeteryTamperingOnlyOnYourUpkeep(t *testing.T) {
	g, me, hidden := cemeteryTamperingTable(t, 25)
	effectstest.BeginUpkeep(t, g, g.Seats[1].ID)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID) != nil || effectstest.PlayHiddenGrant(g, hidden) != nil {
		t.Error("an opponent's upkeep triggered the Tampering")
	}
}
