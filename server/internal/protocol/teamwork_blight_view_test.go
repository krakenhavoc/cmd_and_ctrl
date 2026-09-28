package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// teamwork_blight_view_test.go — the VIEW half of #1703. The client can
// only offer teamwork or a blight, and only pick what pays, if the hand
// card's optional_costs carries the number and the creatures that could
// pay it — and greys the toggle when nothing can.

const viewTeamworkBlightOracle = "test-view-teamwork-blight"

func stubTeamworkBlightCard(t *testing.T) {
	t.Helper()
	stubViewOptionalCosts(t, viewTeamworkBlightOracle, []game.AdditionalCost{
		{Optional: true, Key: game.TeamworkKey, Teamwork: 3, Label: "Teamwork 3"},
		{Optional: true, Key: game.BlightKey, Blight: 2, Label: "Blight 2"},
	})
}

func pushViewCreature(g *game.Game, owner uuid.UUID, power int, tapped bool) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: id, Name: "Bear", TypeLine: "Creature — Bear", Power: power, Toughness: 2,
			Owner: owner, Controller: owner, Tapped: tapped,
		})
	})
	return id
}

func TestTeamworkAndBlightOffersCarryTheirPayers(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stubTeamworkBlightCard(t)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Team Thing", TypeLine: "Sorcery", OracleID: viewTeamworkBlightOracle,
			ManaCost: "{1}", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})

	// Nothing on the board: both offers present, neither payable.
	c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	if len(c.OptionalCosts) != 2 {
		t.Fatalf("optional_costs = %+v, want teamwork and blight", c.OptionalCosts)
	}
	tw, bl := c.OptionalCosts[0], c.OptionalCosts[1]
	if tw.Teamwork != 3 || tw.TeamworkOptions == nil || len(tw.TeamworkOptions.Cards) != 0 {
		t.Errorf("teamwork with no creatures = %+v, want 3 with a present-and-empty option list", tw)
	}
	if bl.Blight != 2 || bl.BlightOptions == nil || len(bl.BlightOptions.Cards) != 0 {
		t.Errorf("blight with no creatures = %+v, want 2 with a present-and-empty option list", bl)
	}

	// A 2-power creature alone does not reach teamwork 3: still empty.
	a := pushViewCreature(g, me.ID, 2, false)
	c = handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	if got := c.OptionalCosts[0].TeamworkOptions.Cards; len(got) != 0 {
		t.Errorf("teamwork 3 with 2 power offered %v, want nothing", got)
	}
	if got := c.OptionalCosts[1].BlightOptions.Cards; len(got) != 1 || got[0] != a.String() {
		t.Errorf("blight options = %v, want the one creature", got)
	}

	// A second untapped creature makes 4: both offered. A tapped one and
	// an opponent's never are (the tapped one can still be blighted).
	b := pushViewCreature(g, me.ID, 2, false)
	tapped := pushViewCreature(g, me.ID, 5, true)
	pushViewCreature(g, opp.ID, 5, false)
	c = handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	got := map[string]bool{}
	for _, s := range c.OptionalCosts[0].TeamworkOptions.Cards {
		got[s] = true
	}
	if len(got) != 2 || !got[a.String()] || !got[b.String()] {
		t.Errorf("teamwork options = %v, want exactly the two untapped creatures", c.OptionalCosts[0].TeamworkOptions.Cards)
	}
	bl = c.OptionalCosts[1]
	if len(bl.BlightOptions.Cards) != 3 {
		t.Errorf("blight options = %v, want my three creatures (tapped %s included)", bl.BlightOptions.Cards, tapped)
	}
}
