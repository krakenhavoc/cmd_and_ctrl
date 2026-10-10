package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// background_test.go — #2874, ADR 0144: a Background is the deck's
// second commander and not a creature. It waits in the command zone
// beside its creature, and a bot seat casts it there like any other
// commander rather than leaving it home all game.

func commandCastMove(t *testing.T, seat int, id, label string) legal.Move {
	return legal.Move{
		Type: legal.TypeCastSpell, Player: seatID(seat), Kind: legal.KindCast,
		Label: label, Source: uuid.MustParse(id),
		Params: mustJSON(t, map[string]any{"instance_id": id, "from_zone": "command"}),
	}
}

func background(id string, controller int, name, cost string) protocol.CardView {
	return protocol.CardView{
		InstanceID:  id,
		Name:        name,
		Owner:       seatID(controller).String(),
		Controller:  seatID(controller).String(),
		TypeLine:    "Legendary Enchantment — Background",
		ManaCost:    cost,
		IsCommander: true,
		KnownByYou:  true,
	}
}

func TestBackgroundCommanderIsCastFromTheCommandZone(t *testing.T) {
	karlach := creature(cardID(1), 0, "Karlach, Fury of Avernus", 5, 4, commander())
	seat := newSeat(0)
	bg := background(cardID(2), 0, "Agent of the Iron Throne", "{2}{B}")
	seat.Command = protocol.ZoneView{Kind: "command", Count: 1, Cards: []protocol.CardView{bg}}
	bf := []protocol.CardView{karlach}
	for i := 0; i < 4; i++ {
		bf = append(bf, land(cardID(30+i), 0))
	}
	v := newView([]protocol.PlayerView{seat, newSeat(1)}, withBattlefield(bf...))
	in := input(0, v,
		passMove(0),
		commandCastMove(t, 0, cardID(2), "Cast Agent of the Iron Throne"),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Cast Agent of the Iron Throne" {
		t.Fatalf("chose %q, want the Background cast from the command zone", got)
	}
}
