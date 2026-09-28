package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// optional_cost_clause_view_test.go — the VIEW half of #1716. When a
// teamwork (or kicker) cost widens the target clause, the offer carries
// the widened clause's legal set, so the client's picker — which takes
// its clause from the first ticked offer that carries one
// (castTargetOverride) — offers exactly what the announce gate accepts.
// An offer that rewrites nothing stamps nothing, and the card's own
// legal set stays the printed clause's.

const viewWidenOracle = "test-view-optional-widen"

func TestOptionalCostCarriesItsWidenedClause(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	smallOnly := func(_ *game.Game, caster uuid.UUID, c game.Card, _ game.ZoneKind) bool {
		return c.Owner == caster && c.Name == "Small"
	}
	anyOfMine := func(_ *game.Game, caster uuid.UUID, c game.Card, _ game.ZoneKind) bool {
		return c.Owner == caster
	}
	withTargetSpec(t, viewWidenOracle, func() *game.TargetSpec {
		return &game.TargetSpec{Mode: "card", Label: "target small card in your graveyard",
			Zones: []game.ZoneKind{game.ZoneGraveyard}, CardOK: smallOnly, Min: 1, Max: 1}
	})
	stubViewOptionalCosts(t, viewWidenOracle, []game.AdditionalCost{
		{Optional: true, Key: game.KickerKey, ManaCost: "{1}", Label: "Kicker {1}"},
		{Optional: true, Key: game.TeamworkKey, Teamwork: 2, Label: "Teamwork 2",
			Targets: &game.TargetSpec{Mode: "card", Label: "target card in your graveyard",
				Zones: []game.ZoneKind{game.ZoneGraveyard}, CardOK: anyOfMine, Min: 1, Max: 1}},
	})
	id := uuid.New()
	small, big := uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Widen Thing", TypeLine: "Sorcery", OracleID: viewWidenOracle,
			ManaCost: "{1}", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
		me.Graveyard.PushTop(game.Card{InstanceID: small, Name: "Small", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: big, Name: "Big", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	})

	c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	if c.LegalTargets == nil || len(c.LegalTargets.Cards) != 1 || c.LegalTargets.Cards[0] != small.String() {
		t.Errorf("the card's own clause offers only the small card, got %+v", c.LegalTargets)
	}
	if len(c.OptionalCosts) != 2 {
		t.Fatalf("optional_costs = %+v, want kicker and teamwork", c.OptionalCosts)
	}
	if kick := c.OptionalCosts[0]; kick.TargetMode != "" || kick.LegalTargets != nil || len(kick.Clauses) != 0 {
		t.Errorf("an offer that rewrites nothing must stamp no clause: %+v", kick)
	}
	tw := c.OptionalCosts[1]
	got := map[string]bool{}
	if tw.LegalTargets != nil {
		for _, s := range tw.LegalTargets.Cards {
			got[s] = true
		}
	}
	if tw.TargetMode != "card" || !got[small.String()] || !got[big.String()] {
		t.Errorf("the teamwork offer must carry the widened clause with both cards: mode %q, legal %+v", tw.TargetMode, tw.LegalTargets)
	}
}
