package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Haunted Screen
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// activate_only_once_view_test.go — the wire half of #5221. `once_used`
// is on a row that says "Activate only once" (CR 602.5b) exactly while
// this object has already activated it, the row stays listed, and the
// row is not stamped as a spent exhaust ability.

const hauntedScreenOracle = "49f92c0b-16ac-4cf0-8ec3-e3638990ee75"

func TestOnceUsedIsStampedOnceTheAbilityIsActivated(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0].ID
	screen := game.NewCard("Haunted Screen", me)
	screen.TypeLine = "Artifact"
	screen.OracleID = hauntedScreenOracle
	screen.ManaCost = "{3}"
	screen.Controller = me
	g.Battlefield.PushTop(screen)
	for range 7 {
		plains := game.NewCard("Plains", me)
		plains.TypeLine = "Basic Land — Plains"
		plains.Controller = me
		g.Battlefield.PushTop(plains)
	}

	c := conditionCardView(t, g, screen.InstanceID)
	if len(c.ActivatedAbilities) != 1 {
		t.Fatalf("got %d activated abilities, want 1", len(c.ActivatedAbilities))
	}
	if c.ActivatedAbilities[0].OnceUsed {
		t.Error("once_used before anything was activated")
	}
	raw, _ := json.Marshal(c)
	if strings.Contains(string(raw), "once_used") {
		t.Errorf("an unused ability must leave the flag off the wire: %s", raw)
	}

	if err := g.ActivateCatalogAbility(me, screen.InstanceID, 0, game.ActivateAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("activate the {7} ability: %v", err)
	}
	c = conditionCardView(t, g, screen.InstanceID)
	if len(c.ActivatedAbilities) != 1 || !c.ActivatedAbilities[0].OnceUsed {
		t.Fatalf("after the activation: want the row listed and stamped once_used, got %+v", c.ActivatedAbilities)
	}
	if c.ActivatedAbilities[0].Exhausted || c.ActivatedAbilities[0].PowerUpUsed {
		t.Error("an only-once row was stamped exhausted or power_up_used — it is neither keyword")
	}
	raw, _ = json.Marshal(c)
	if !strings.Contains(string(raw), `"once_used":true`) {
		t.Errorf("once_used missing from the wire: %s", raw)
	}
}
