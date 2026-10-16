package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// any_player_mana_view_test.go — the view, digest and log half of #1864
// (ADR 0106 §1 amendment 2026-10-16): Mana Cache's "Any player may
// activate this ability" MANA row reaches every seat marked
// `any_player`, stamped with that seat as the activator ("their turn",
// the counter it would remove); the active non-controller's digest lists
// it under `mana_abilities`; and an activation across the table is told
// in the log while the controller's own stays silent (CR 605.3b).

const oracleManaCacheView = "1c09f2d1-413b-4a92-aa39-99c1a2106814"

func manaRowOf(t *testing.T, v GameView, id uuid.UUID) ManaAbilityView {
	t.Helper()
	c := cardInFrame(v, id.String())
	if c == nil || len(c.ManaAbilities) != 1 {
		t.Fatalf("frame for %s: card %v has no single mana row", v.ID, c)
	}
	return c.ManaAbilities[0]
}

func putManaCache(g *game.Game, controller *game.Player, charge int) uuid.UUID {
	id := put(g.Battlefield, controller, game.Card{Name: "Mana Cache", TypeLine: "Enchantment", OracleID: oracleManaCacheView})
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Counters = map[string]int{game.CounterCharge: charge}
		}
	}
	return id
}

func TestAnyPlayerManaRowIsStampedForEachActivator(t *testing.T) {
	g := busyTable(t, 1)
	active := g.Seats[g.Turn.ActiveSeat]
	controller := g.Seats[(g.Turn.ActiveSeat+1)%4]
	bystander := g.Seats[(g.Turn.ActiveSeat+2)%4]
	cache := putManaCache(g, controller, 2)
	knowTheTable(g)

	// The active player's copy: their turn, so the condition holds.
	av := ViewOfGameFor(g, active.ID.String())
	row := manaRowOf(t, av, cache)
	if !row.AnyPlayer {
		t.Error("Mana Cache's mana row is not marked any_player")
	}
	if row.ConditionUnmet {
		t.Error("the active player's copy says the condition is unmet; \"their turn\" is the activator's")
	}
	if av.LegalActions == nil {
		t.Fatal("the active seat got no digest")
	}
	e := av.LegalActions.Sources[cache.String()]
	if e == nil || !slices.Equal(e.ManaAbilities, []string{row.Ref}) {
		t.Fatalf("active non-controller's digest entry = %+v, want mana_abilities [%s]", e, row.Ref)
	}

	// The controller's and a bystander's copies: not their turn.
	for _, p := range []*game.Player{controller, bystander} {
		r := manaRowOf(t, ViewOfGameFor(g, p.ID.String()), cache)
		if !r.AnyPlayer {
			t.Errorf("%s's copy is not marked any_player", p.Name)
		}
		if !r.ConditionUnmet {
			t.Errorf("%s's copy says the condition holds on another player's turn", p.Name)
		}
	}
}

// TestAnyPlayerManaActivationIsNarrated: the activator gets the mana and
// the log names both players; the controller's own activation is not a
// log line, as no mana activation is (CR 605.3b).
func TestAnyPlayerManaActivationIsNarrated(t *testing.T) {
	g := busyTable(t, 1)
	active := g.Seats[g.Turn.ActiveSeat]
	controller := g.Seats[(g.Turn.ActiveSeat+1)%4]
	cache := putManaCache(g, controller, 1)
	knowTheTable(g)
	if err := g.ActivateManaAbility(active.ID, cache, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate across: %v", err)
	}
	if active.ManaPool.CountColor("C") != 1 {
		t.Errorf("activator's pool = %v, want one {C}", active.ManaPool)
	}
	entry := findLog(t, ViewOfGame(g).Log, LogActivateAcross)
	if want := active.Name + " activated " + controller.Name + "'s Mana Cache"; entry.Text != want {
		t.Errorf("text %q, want %q", entry.Text, want)
	}

	g2 := busyTable(t, 1)
	own := g2.Seats[g2.Turn.ActiveSeat]
	c2 := putManaCache(g2, own, 1)
	knowTheTable(g2)
	if err := g2.ActivateManaAbility(own.ID, c2, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("own activation: %v", err)
	}
	for _, e := range ViewOfGame(g2).Log {
		if e.Kind == LogActivateAcross {
			t.Errorf("the controller's own mana activation was narrated: %q", e.Text)
		}
	}
}
