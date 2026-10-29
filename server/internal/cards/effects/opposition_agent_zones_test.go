package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// opposition_agent_zones_test.go — #4519 under Opposition Agent (CR 723.2:
// "you control your opponents while they're searching their libraries").
// An and/or search asks which zones are searched before anything is shown.
// A searcher who has not chosen the library is not searching it: nothing
// controls them, and what they find is not exiled. Choosing the library
// begins the control, and the Agent's controller answers from there.

// oaZonesTable is oaTable with Tower Winder cast on the searcher's turn and
// a Command Tower in each of their zones; the zone question is open.
func oaZonesTable(t *testing.T) (g *game.Game, searcher, ctl *game.Player, libTower, gyTower uuid.UUID) {
	t.Helper()
	g, searcher, ctl, _, _ = oaTable(t)
	seedSearchLibrary(searcher, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	libTower = pushLibraryCardForTest(searcher, lgTower())
	gyTower = pushGraveyardCardWithTypeLineID(searcher, "Command Tower", "Land")
	lgCastTowerWinder(t, g)
	c := searchChoiceFor(g, searcher.ID)
	if c == nil || !c.SearchZonesPending() {
		t.Fatalf("Tower Winder must ask the zones first: %+v", c)
	}
	return g, searcher, ctl, libTower, gyTower
}

// oaChooseZones answers the zone question through the actions layer as the
// searcher.
func oaChooseZones(g *game.Game, c *game.PendingChoice, caller, player uuid.UUID, zones ...string) error {
	raw, _ := json.Marshal(map[string]any{"choice_id": c.ID.String(), "search_zones": zones})
	return actions.Dispatch(g, actions.Action{Type: actions.TypeResolveChoice, Player: player, Caller: caller, Params: raw})
}

// The question is the searcher's own: no control has begun, so the Agent's
// controller can neither answer it nor see it coming.
func TestOppositionAgentDoesNotControlTheZoneQuestion(t *testing.T) {
	g, searcher, ctl, _, _ := oaZonesTable(t)
	if got := g.ControllerOfPlayer(searcher.ID); got != searcher.ID {
		t.Fatalf("before the library is chosen the searcher is controlled by %v", got)
	}
	c := searchChoiceFor(g, searcher.ID)
	if err := oaChooseZones(g, c, ctl.ID, searcher.ID, game.SearchZoneLibrary); err == nil {
		t.Error("the Agent's controller answered the searcher's zone question")
	}
	if len(g.PlayerControls) != 0 {
		t.Errorf("a control record exists while the zones are unchosen: %+v", g.PlayerControls)
	}
}

// Graveyard only: not a search of the library, so no control, and the Tower
// found goes to the searcher's hand rather than into exile.
func TestOppositionAgentIgnoresAGraveyardOnlySearch(t *testing.T) {
	g, searcher, ctl, libTower, gyTower := oaZonesTable(t)
	c := searchChoiceFor(g, searcher.ID)
	if err := oaChooseZones(g, c, searcher.ID, searcher.ID, game.SearchZoneGraveyard); err != nil {
		t.Fatalf("the zone answer: %v", err)
	}
	if got := g.ControllerOfPlayer(searcher.ID); got != searcher.ID || len(g.PlayerControls) != 0 {
		t.Fatalf("a graveyard-only search is controlled by %v (%d records)", got, len(g.PlayerControls))
	}
	if err := oaAnswer(g, c, ctl.ID, searcher.ID, gyTower); err == nil {
		t.Error("the Agent's controller answered a search that is not of the library")
	}
	if err := oaAnswer(g, c, searcher.ID, searcher.ID, libTower); err == nil {
		t.Error("the library's Tower was a legal pick of a graveyard-only search")
	}
	if err := oaAnswer(g, c, searcher.ID, searcher.ID, gyTower); err != nil {
		t.Fatalf("the searcher's own answer: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !searcher.Hand.Contains(gyTower) || oaInExile(g, gyTower) {
		t.Error("the Tower found in the graveyard was exiled, though the library was not searched")
	}
}

// The library chosen: the control begins, only the Agent's controller may
// answer from here, and what is found is exiled.
func TestOppositionAgentControlsOnceTheLibraryIsChosen(t *testing.T) {
	g, searcher, ctl, libTower, _ := oaZonesTable(t)
	c := searchChoiceFor(g, searcher.ID)
	if err := oaChooseZones(g, c, searcher.ID, searcher.ID, game.SearchZoneLibrary, game.SearchZoneGraveyard); err != nil {
		t.Fatalf("the zone answer: %v", err)
	}
	c = oaOpenSearch(t, g, searcher, ctl)
	if err := oaAnswer(g, c, searcher.ID, searcher.ID, libTower); !errors.Is(err, game.ErrPlayerControlled) {
		t.Fatalf("the searcher answered their own library search: err = %v, want ErrPlayerControlled", err)
	}
	if err := oaAnswer(g, c, ctl.ID, ctl.ID, libTower); err != nil {
		t.Fatalf("the controller's answer: %v", err)
	}
	if !oaInExile(g, libTower) || searcher.Hand.Contains(libTower) {
		t.Error("the Tower found in the library was not exiled")
	}
	oaAssertControlOver(t, g, searcher)
}
