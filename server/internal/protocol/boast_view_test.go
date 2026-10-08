package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Fearless Liberator, Birgi
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// boast_view_test.go — the wire half of #2697 (CR 702.142). A boast
// ability's row says WHICH half of its instruction fails —
// `boast_blocked: "not_attacked"` or `"used"` — so the client can say
// why the row is grey, and says nothing when nothing objects.

const (
	boastViewLiberatorOracle = "b0b93253-6432-401e-9a01-94c41fb72c10"
	boastViewBirgiOracle     = "fb81e4d3-1d8c-4779-be62-87cf49277e51"
)

func pushBoastCard(g *game.Game, controller uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	c := game.NewCard(name, controller)
	c.TypeLine = typeLine
	c.OracleID = oracle
	c.Controller = controller
	c.Power, c.Toughness = 2, 2
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func boastBlocked(t *testing.T, g *game.Game, id uuid.UUID) string {
	t.Helper()
	c := conditionCardView(t, g, id)
	if len(c.ActivatedAbilities) != 1 {
		t.Fatalf("got %d activated abilities, want 1", len(c.ActivatedAbilities))
	}
	return c.ActivatedAbilities[0].BoastBlocked
}

func advanceToCombatDamage(t *testing.T, g *game.Game, attacker uuid.UUID) {
	t.Helper()
	for i := 0; i < 20 && g.Turn.Step != game.StepDeclareAttackers; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	defender := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if err := g.DeclareAttacker(attacker, defender.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	for i := 0; i < 20 && g.Turn.Step != game.StepCombatDamage; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

func TestBoastBlockedNamesTheFailingHalf(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	dwarf := pushBoastCard(g, me, "Fearless Liberator", "Creature — Dwarf Berserker", boastViewLiberatorOracle)
	for range 4 {
		mountain := game.NewCard("Mountain", me)
		mountain.TypeLine = "Basic Land — Mountain"
		mountain.Controller = me
		g.Battlefield.PushTop(mountain)
	}

	if got := boastBlocked(t, g, dwarf); got != "not_attacked" {
		t.Fatalf("before attacking: boast_blocked = %q, want not_attacked", got)
	}

	advanceToCombatDamage(t, g, dwarf)
	c := conditionCardView(t, g, dwarf)
	if got := c.ActivatedAbilities[0].BoastBlocked; got != "" {
		t.Fatalf("after attacking: boast_blocked = %q, want none", got)
	}
	raw, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(raw), "boast_blocked") {
		t.Errorf("a boast with nothing objecting must leave the flag off the wire: %s", raw)
	}

	if err := g.ActivateCatalogAbility(me, dwarf, 0, game.ActivateAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := boastBlocked(t, g, dwarf); got != "used" {
		t.Errorf("after the boast: boast_blocked = %q, want used", got)
	}
}

func TestBoastBlockedHonoursBirgisSecondActivation(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	dwarf := pushBoastCard(g, me, "Fearless Liberator", "Creature — Dwarf Berserker", boastViewLiberatorOracle)
	pushBoastCard(g, me, "Birgi, God of Storytelling", "Legendary Creature — God", boastViewBirgiOracle)
	for range 8 {
		mountain := game.NewCard("Mountain", me)
		mountain.TypeLine = "Basic Land — Mountain"
		mountain.Controller = me
		g.Battlefield.PushTop(mountain)
	}
	advanceToCombatDamage(t, g, dwarf)

	if err := g.ActivateCatalogAbility(me, dwarf, 0, game.ActivateAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("first boast: %v", err)
	}
	if got := boastBlocked(t, g, dwarf); got != "" {
		t.Errorf("one boast under Birgi: boast_blocked = %q, want none — a second is allowed", got)
	}
	if err := g.ActivateCatalogAbility(me, dwarf, 0, game.ActivateAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("second boast: %v", err)
	}
	if got := boastBlocked(t, g, dwarf); got != "used" {
		t.Errorf("two boasts under Birgi: boast_blocked = %q, want used", got)
	}
}

func TestANonBoastAbilityNeverCarriesBoastBlocked(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	elf := pushBoastCard(g, me, "Greenbelt Guardian", "Creature — Elf Ranger", greenbeltGuardianOracle)
	for _, a := range conditionCardView(t, g, elf).ActivatedAbilities {
		if a.BoastBlocked != "" {
			t.Errorf("%q carries boast_blocked %q", a.Label, a.BoastBlocked)
		}
	}
}
