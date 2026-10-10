package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// any_player_view_test.go — the view and digest half of ADR 0106 §1
// (#1793): an "Any player may activate this ability" row reaches every
// seat marked `any_player`, stamped with THAT seat as the activator; the
// non-controller's digest lists it under the permanent's ID; and the
// digest and the activator-only option lists stay own-seat only.

const oracleXantchaView = "0f0f3712-8d13-41a5-b332-2ab34e48d79d"

const anyPlayerViewOracle = "test-protocol-any-player-saproling-cluster"

// stubAnyPlayerView wires one test permanent whose only row is an
// any-player "Discard a card" ability that may be activated only
// during the activator's own turn — both halves read the ACTIVATOR.
func stubAnyPlayerView(t *testing.T) {
	t.Helper()
	prev := game.CatalogActivatedAbilities
	game.CatalogActivatedAbilities = func(key string) []game.ActivatedAbilityShape {
		if key == anyPlayerViewOracle {
			return []game.ActivatedAbilityShape{{
				Label:     "Discard a card: Nothing. Any player may activate this ability but only during their turn.",
				Cost:      game.AbilityCost{DiscardCards: &game.DiscardCost{N: 1, Label: "a card"}},
				AnyPlayer: true,
				Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
					return g.Seats[g.Turn.ActiveSeat].ID == controller
				},
				Effect: func(*game.Game, *game.StackItem) error { return nil },
			}}
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogActivatedAbilities = prev })
}

func rowOf(t *testing.T, v GameView, id uuid.UUID) ActivatedAbilityView {
	t.Helper()
	c := cardInFrame(v, id.String())
	if c == nil || len(c.ActivatedAbilities) != 1 {
		t.Fatalf("frame for %s: card %v has no single ability row", v.ID, c)
	}
	return c.ActivatedAbilities[0]
}

func TestAnyPlayerRowIsInTheNonControllersDigest(t *testing.T) {
	g := busyTable(t, 1)
	active := g.Seats[g.Turn.ActiveSeat]
	owner := g.Seats[(g.Turn.ActiveSeat+1)%4]
	controller := g.Seats[(g.Turn.ActiveSeat+2)%4]
	xantcha := put(g.Battlefield, owner, game.Card{
		Name: "Xantcha, Sleeper Agent", TypeLine: "Legendary Creature — Phyrexian Minion",
		Power: 5, Toughness: 5, OracleID: oracleXantchaView,
	})
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == xantcha {
			g.Battlefield.Cards[i].Controller = controller.ID
		}
	}
	knowTheTable(g)

	av := ViewOfGameFor(g, active.ID.String())
	if av.LegalActions == nil {
		t.Fatal("the active seat got no digest")
	}
	e := av.LegalActions.Sources[xantcha.String()]
	if e == nil || !slices.Equal(e.Abilities, []string{"own:0"}) {
		t.Fatalf("active non-controller's digest entry for Xantcha = %+v, want abilities [own:0]", e)
	}
	row := rowOf(t, av, xantcha)
	if !row.AnyPlayer {
		t.Error("Xantcha's row is not marked any_player")
	}
	if row.Purpose == nil || !row.Purpose.Priced() || withoutAnswers(*row.Purpose) != (ActivationPurposeView{Draws: 1, ControllerLosesLife: 2}) {
		t.Errorf("purpose = %+v", row.Purpose)
	}
	if row.Ref != e.Abilities[0] {
		t.Errorf("row ref %q does not join the digest's %q", row.Ref, e.Abilities[0])
	}
	// Nobody else holds priority, so nobody else has a digest — and the
	// controller's frame still carries the row, marked.
	for _, p := range g.Seats {
		if p.ID == active.ID {
			continue
		}
		v := ViewOfGameFor(g, p.ID.String())
		if v.LegalActions != nil {
			t.Errorf("seat %s holds no priority but got a digest", p.Name)
		}
		if !rowOf(t, v, xantcha).AnyPlayer {
			t.Errorf("seat %s's copy of the row is not marked any_player", p.Name)
		}
	}
	if raw := ViewOfGame(g); raw.LegalActions != nil {
		t.Error("the raw view carries a digest")
	}
}

// TestAnyPlayerRowIsStampedForEachActivator: the row's verdicts and its
// hand-read option list are the VIEWER's when the viewer would be the
// activator, and a hand-read list reaches its own seat alone (#1369).
func TestAnyPlayerRowIsStampedForEachActivator(t *testing.T) {
	stubAnyPlayerView(t)
	g := busyTable(t, 1)
	active := g.Seats[g.Turn.ActiveSeat]
	controller := g.Seats[(g.Turn.ActiveSeat+1)%4]
	bystander := g.Seats[(g.Turn.ActiveSeat+2)%4]
	cluster := put(g.Battlefield, controller, game.Card{Name: "Test Cluster", TypeLine: "Enchantment", OracleID: anyPlayerViewOracle})
	knowTheTable(g)

	handIDs := func(p *game.Player) []string {
		var out []string
		for _, c := range p.Hand.Cards {
			out = append(out, c.InstanceID.String())
		}
		return out
	}
	for _, tc := range []struct {
		who       *game.Player
		condition bool // condition_unmet
	}{
		{active, false},    // the activator's turn
		{controller, true}, // the controller's condition reads the controller
		{bystander, true},  // and a bystander's reads the bystander
	} {
		v := ViewOfGameFor(g, tc.who.ID.String())
		row := rowOf(t, v, cluster)
		if row.ConditionUnmet != tc.condition {
			t.Errorf("%s: condition_unmet = %v, want %v", tc.who.Name, row.ConditionUnmet, tc.condition)
		}
		if want := handIDs(tc.who); !slices.Equal(row.DiscardCostOptions, want) {
			t.Errorf("%s: discard options %v, want their own hand %v", tc.who.Name, row.DiscardCostOptions, want)
		}
	}
	// A spectator sees the public row: no hand of anybody's.
	if row := rowOf(t, ViewOfGameFor(g, ""), cluster); len(row.DiscardCostOptions) != 0 {
		t.Errorf("spectator sees discard options %v", row.DiscardCostOptions)
	}
	// And the active seat's digest names the row.
	av := ViewOfGameFor(g, active.ID.String())
	if e := av.LegalActions.Sources[cluster.String()]; e == nil || !slices.Equal(e.Abilities, []string{"own:0"}) {
		t.Errorf("active non-controller's digest entry = %+v", e)
	}
}

// TestAnyPlayerActivationIsNarrated: the public log names both players
// when one reaches across to another's permanent (ADR 0106 §1 decision
// 6), and says nothing for the controller's own activation.
func TestAnyPlayerActivationIsNarrated(t *testing.T) {
	g := busyTable(t, 1)
	active := g.Seats[g.Turn.ActiveSeat]
	controller := g.Seats[(g.Turn.ActiveSeat+1)%4]
	xantcha := put(g.Battlefield, controller, game.Card{
		Name: "Xantcha, Sleeper Agent", TypeLine: "Legendary Creature — Phyrexian Minion",
		Power: 5, Toughness: 5, OracleID: oracleXantchaView,
	})
	knowTheTable(g)
	if err := g.ActivateCatalogAbility(active.ID, xantcha, 0, game.ActivateAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	entry := findLog(t, ViewOfGame(g).Log, LogActivateAcross)
	if want := active.Name + " activated " + controller.Name + "'s Xantcha, Sleeper Agent"; entry.Text != want {
		t.Errorf("text %q, want %q", entry.Text, want)
	}

	g2 := busyTable(t, 1)
	own := g2.Seats[g2.Turn.ActiveSeat]
	x2 := put(g2.Battlefield, own, game.Card{
		Name: "Xantcha, Sleeper Agent", TypeLine: "Legendary Creature — Phyrexian Minion",
		Power: 5, Toughness: 5, OracleID: oracleXantchaView,
	})
	knowTheTable(g2)
	if err := g2.ActivateCatalogAbility(own.ID, x2, 0, game.ActivateAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("own activation: %v", err)
	}
	for _, e := range ViewOfGame(g2).Log {
		if e.Kind == LogActivateAcross {
			t.Errorf("the controller's own activation was narrated as reaching across: %q", e.Text)
		}
	}
}

// withoutAnswers drops a declared Answers, which this test does not pin.
func withoutAnswers(p ActivationPurposeView) ActivationPurposeView {
	p.Answers = nil
	return p
}
