package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// free_commander_cast_test.go — #6242: the enumerator answers "you may
// cast your commander from the command zone without paying its mana
// cost" with the free cast of the commander out of the command zone, and
// the actions layer takes it with no mana.

const oracleFreeCommanderOffer = "test-6242-free-commander"

func init() {
	effects.Register(effects.Spec{
		OracleID: oracleFreeCommanderOffer,
		Name:     "Test Free Commander",
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: You may cast your commander from the command zone without paying its mana cost.",
			Cost:  effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				effects.OfferFreeCommanderCast(effects.NewContext(g, item), "Test Free Commander")
				return nil
			},
		}},
	})
}

func TestEnumeratorOffersTheFreeCommanderCast(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	src := battlefieldCard(g, seat, game.Card{Name: "Test Free Commander", TypeLine: "Artifact", OracleID: oracleFreeCommanderOffer})
	cmd := game.NewCard("Test Commander", seat.ID)
	cmd.TypeLine, cmd.ManaCost, cmd.IsCommander = "Legendary Creature — Bear", "{4}{G}{G}", true
	cmd.Power, cmd.Toughness = 5, 5
	seat.Command.PushTop(cmd)

	c := resolveCopier(t, g, seat, src)
	var cast *legal.Move
	for _, m := range legal.EnumerateFor(g, seat.ID) {
		if m.Kind == legal.KindCast && m.Source == cmd.InstanceID && m.ResolutionCast == c.ID.String() {
			cast = &m
			break
		}
	}
	if cast == nil {
		t.Fatalf("no free cast of the commander is offered: %v", labels(legal.EnumerateFor(g, seat.ID)))
	}
	var p struct {
		FromZone string `json:"from_zone"`
	}
	_ = json.Unmarshal(cast.Params, &p)
	if p.FromZone != string(game.ZoneCommand) {
		t.Errorf("the cast is from %q, want the command zone", p.FromZone)
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(cast.Type), Player: seat.ID, Caller: seat.ID, Params: cast.Params}); err != nil {
		t.Fatalf("dispatch the free cast: %v", err)
	}
	if !g.Stack.Contains(cmd.InstanceID) {
		t.Fatal("the commander is not on the stack")
	}
	if n := seat.CommanderCasts[cmd.InstanceID]; n != 1 {
		t.Errorf("CommanderCasts = %d, want 1", n)
	}
}
