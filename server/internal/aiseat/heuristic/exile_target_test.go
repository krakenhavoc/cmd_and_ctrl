package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// exile_target_test.go -- #3259. A card in exile is a target the policy
// can price: The Darkness Crystal's "put target creature card exiled
// with this onto the battlefield under your control" is worth the card
// it takes. Before the change an exile target priced 0, so the bot
// could not tell the 6/4 from the 1/1.

func withExile(cards ...protocol.CardView) viewOpt {
	return func(v *protocol.GameView) {
		v.Exile = protocol.ZoneView{Kind: "exile", Count: len(cards), Cards: cards}
	}
}

func darknessCrystalActivation(t *testing.T, src, target string) legal.Move {
	return legal.Move{
		Type: legal.TypeActivateAbility, Player: seatID(0), Kind: legal.KindActivate,
		Label: "The Darkness Crystal: return " + target, Source: uuid.MustParse(src),
		Params: mustJSON(t, map[string]any{
			"source_card_id": src, "ability_index": 0,
			"targets": []map[string]string{{"kind": "card", "id": target}},
		}),
	}
}

func TestTheBotTakesTheBestCreatureOutOfExile(t *testing.T) {
	crystal := equipment(cardID(5), 0, "The Darkness Crystal")
	crystal.TypeLine = "Legendary Artifact"
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(crystal, land(cardID(10), 0)),
		withExile(
			graveyardCardView(cardID(1), 1, "Goblin Token Maker", "Creature -- Goblin", "{R}"),
			graveyardCardView(cardID(2), 1, "Craw Wurm", "Creature -- Wurm", "{4}{G}{G}"),
		))
	// Give the exiled cards the stats the evaluator reads.
	for i := range v.Exile.Cards {
		switch v.Exile.Cards[i].InstanceID {
		case cardID(1):
			v.Exile.Cards[i].Power, v.Exile.Cards[i].Toughness = 1, 1
		case cardID(2):
			v.Exile.Cards[i].Power, v.Exile.Cards[i].Toughness = 6, 4
		}
	}
	// The weaker card is offered first so a tie would pick it.
	weak := darknessCrystalActivation(t, cardID(5), cardID(1))
	strong := darknessCrystalActivation(t, cardID(5), cardID(2))
	in := input(0, v, passMove(0), weak, strong)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != strong.Label {
		t.Errorf("the bot chose %q, want the 6/4 over the 1/1", got)
	}
}
