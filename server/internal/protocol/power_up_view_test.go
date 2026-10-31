package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Brave Brawler
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// power_up_view_test.go — the wire half of #4527. `power_up_used` is on
// a power-up row exactly while this object has already activated it
// (CR 702.193a), the row stays listed, and the charged cost is the
// keyword's reduction the turn the creature entered.

const braveBrawlerOracle = "a1a791be-71e6-4c1f-8717-08ca62d40de2"

func TestPowerUpUsedIsStampedOnceTheAbilityIsActivated(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0].ID
	brawler := game.NewCard("Brave Brawler", me)
	brawler.TypeLine = "Creature — Human Warrior Hero"
	brawler.OracleID = braveBrawlerOracle
	brawler.ManaCost = "{1}{W}"
	brawler.Controller = me
	brawler.Power, brawler.Toughness = 2, 1
	g.Battlefield.PushTop(brawler)
	for range 5 {
		plains := game.NewCard("Plains", me)
		plains.TypeLine = "Basic Land — Plains"
		plains.Controller = me
		g.Battlefield.PushTop(plains)
	}

	c := conditionCardView(t, g, brawler.InstanceID)
	if len(c.ActivatedAbilities) != 1 {
		t.Fatalf("got %d activated abilities, want 1", len(c.ActivatedAbilities))
	}
	if c.ActivatedAbilities[0].PowerUpUsed {
		t.Error("power_up_used before anything was activated")
	}
	if got := c.ActivatedAbilities[0].ChargedManaCost; got == nil || *got != "{4}{W}" {
		t.Errorf("charged cost of a Brawler that did not enter this turn = %v, want {4}{W}", got)
	}
	raw, _ := json.Marshal(c)
	if strings.Contains(string(raw), "power_up_used") {
		t.Errorf("an unused power-up must leave the flag off the wire: %s", raw)
	}

	if err := g.ActivateCatalogAbility(me, brawler.InstanceID, 0, game.ActivateAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("activate the power-up ability: %v", err)
	}
	c = conditionCardView(t, g, brawler.InstanceID)
	if len(c.ActivatedAbilities) != 1 || !c.ActivatedAbilities[0].PowerUpUsed {
		t.Fatalf("after the activation: want the row listed and stamped power_up_used, got %+v", c.ActivatedAbilities)
	}
	if c.ActivatedAbilities[0].Exhausted {
		t.Error("a power-up row was stamped exhausted — the keywords are distinct")
	}
	raw, _ = json.Marshal(c)
	if !strings.Contains(string(raw), `"power_up_used":true`) {
		t.Errorf("power_up_used missing from the wire: %s", raw)
	}
}

func pushPowerUpViewCreature(g *game.Game, me uuid.UUID, name, typeLine, oracle, cost string, p, tough int) game.Card {
	c := game.NewCard(name, me)
	c.TypeLine = typeLine
	c.OracleID = oracle
	c.ManaCost = cost
	c.Controller = me
	c.Power, c.Toughness = p, tough
	g.Battlefield.PushTop(c)
	return c
}

// #4571: Wonder Man's "an additional time" keeps `power_up_used` off until
// the second activation, and a power-up ban (Kang the Conqueror's extra
// turn) stamps `cant_activate` with the printed clause.
func TestPowerUpAdditionalTimeAndBanAreProjected(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0].ID
	brawler := pushPowerUpViewCreature(g, me, "Brave Brawler", "Creature — Human Warrior Hero", braveBrawlerOracle, "{1}{W}", 2, 1)
	pushPowerUpViewCreature(g, me, "Wonder Man, Hollywood Hero", "Legendary Creature — Human Performer Hero",
		"e2c49754-711e-43b3-82f1-5376a4d99701", "{3}{R}{R}", 4, 4)
	for range 10 {
		plains := game.NewCard("Plains", me)
		plains.TypeLine = "Basic Land — Plains"
		plains.Controller = me
		g.Battlefield.PushTop(plains)
	}

	if err := g.ActivateCatalogAbility(me, brawler.InstanceID, 0, game.ActivateAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	if c := conditionCardView(t, g, brawler.InstanceID); c.ActivatedAbilities[0].PowerUpUsed {
		t.Error("power_up_used after one activation under Wonder Man — a second is allowed")
	}
	if err := g.ActivateCatalogAbility(me, brawler.InstanceID, 0, game.ActivateAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("second activation: %v", err)
	}
	if c := conditionCardView(t, g, brawler.InstanceID); !c.ActivatedAbilities[0].PowerUpUsed {
		t.Error("power_up_used missing after two activations under Wonder Man")
	}

	other := pushPowerUpViewCreature(g, me, "Brave Brawler", "Creature — Human Warrior Hero", braveBrawlerOracle, "{1}{W}", 2, 1)
	if c := conditionCardView(t, g, other.InstanceID); c.ActivatedAbilities[0].CantActivate != "" {
		t.Fatalf("cant_activate = %q before any ban", c.ActivatedAbilities[0].CantActivate)
	}
	g.WithWriteLock(func() {
		g.RegisterScopedRuleEffectForEffect(uuid.Nil, game.ScopeGame, uuid.Nil,
			[]game.Mod{{Kind: game.ModCantActivatePowerUp}}, g.UntilEndOfTurnDuration(), "Kang the Conqueror")
	})
	c := conditionCardView(t, g, other.InstanceID)
	if !strings.Contains(c.ActivatedAbilities[0].CantActivate, "Power-up abilities can't be activated") {
		t.Errorf("cant_activate = %q under the ban, want the printed clause", c.ActivatedAbilities[0].CantActivate)
	}
}
