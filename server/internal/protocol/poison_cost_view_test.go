package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

var poisonCostViewThen = game.RegisterConfirmThen("test/4613-poison-cost-view", func(*game.Game, game.ConfirmAnswer) error { return nil })

// #4613: a confirm whose accept branch gives poison counters ("Ward—Get five
// poison counters") ships the count, so a client or bot holding only the wire
// payload does not price it like "shuffle your library".
func TestConfirmViewCarriesThePoisonCost(t *testing.T) {
	g := buildActiveGame(t)
	payer := g.Seats[0]
	g.WithWriteLock(func() {
		if _, err := g.QueueConfirmKeyedForEffect(game.ConfirmPrompt{
			Chooser:      payer.ID,
			Question:     "Ward — get 5 poison counters, or it is countered",
			AcceptLabel:  "Get 5 poison counters",
			DeclineLabel: "Let it be countered",
			PoisonCost:   5,
		}, poisonCostViewThen, nil); err != nil {
			t.Fatal(err)
		}
	})
	v := ViewOfGame(g)
	if len(v.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want 1", len(v.PendingChoices))
	}
	c := v.PendingChoices[0]
	if c.Kind != "confirm" || c.PoisonCost != 5 || c.LifeCost != 0 || c.LifeCostUnpayable {
		t.Fatalf("wrong confirm view: kind %q poison %d life %d unpayable %v", c.Kind, c.PoisonCost, c.LifeCost, c.LifeCostUnpayable)
	}
	raw, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"poison_cost":5`) {
		t.Errorf("poison_cost missing from the wire: %s", raw)
	}
}
