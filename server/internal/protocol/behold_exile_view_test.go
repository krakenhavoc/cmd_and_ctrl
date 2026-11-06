package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Champion of the Clachan
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// behold_exile_view_test.go — #4687: the Champions' mandatory "behold a
// Kithkin and exile it" ships as the card's own additional_cost with
// `reveal`, `behold`, `behold_exile` and the cards that could pay it, off
// the engine's own walk (#544).

const viewClachanOracle = "4a220e0c-e463-4264-9a96-658cf842b82e"

func TestBeholdExileCostListsWhatCouldPayIt(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	kithkin := uuid.New()
	onTable := uuid.New()
	champ := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: kithkin, Name: "Kithkin Card", TypeLine: "Creature — Kithkin", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Creature — Goblin", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
		g.Battlefield.PushTop(game.Card{InstanceID: onTable, Name: "Kithkin Guard", TypeLine: "Creature — Kithkin", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
		me.Hand.PushTop(game.Card{
			InstanceID: champ, Name: "Champion of the Clachan", TypeLine: "Creature — Kithkin Knight", OracleID: viewClachanOracle,
			ManaCost: "{3}{W}", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})
	c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), champ.String())
	ac := c.AdditionalCost
	if ac == nil || !ac.Reveal || !ac.Behold || !ac.BeholdExile || ac.RevealOptions == nil || len(ac.Branches) != 0 {
		t.Fatalf("additional_cost = %+v, want reveal, behold, behold_exile and options, no branches", ac)
	}
	if ac.Label != "Behold a Kithkin and exile it" {
		t.Errorf("label = %q", ac.Label)
	}
	got := ac.RevealOptions.Cards
	if len(got) != 2 || !slices.Contains(got, kithkin.String()) || !slices.Contains(got, onTable.String()) {
		t.Errorf("reveal_options = %v, want the Kithkin card and the Kithkin permanent (not the Goblin or the spell)", got)
	}
}
