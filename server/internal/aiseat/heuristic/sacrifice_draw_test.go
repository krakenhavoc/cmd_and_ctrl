package heuristic_test

import (
	"context"
	"math"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// sacrifice_draw_test.go is #4991: Sai, Master Thopterist's "{1}{U},
// Sacrifice two artifacts: Draw a card." declares its draw, so the row
// is priced at a card in hand (Weights.Hand) instead of the flat
// ActivateBase, less the Thopters it gives up (sacrifice.go).

// saiWithThopters is Sai, as the view shows her (one triggered and one
// activated row, the activated one declaring `draws`), and two 1/1
// flying Thopter tokens.
func saiWithThopters(draws int) []protocol.CardView {
	sai := creature(cardID(1), 0, "Sai, Master Thopterist", 1, 4, rows("triggered", "activated"))
	sai.ManaCost = "{2}{U}"
	sai.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, ManaCost: "{1}{U}", SacrificeLabel: "two artifacts",
		Purpose: &protocol.PurposeView{Draws: draws, Answers: &[]string{"value"}},
	}}
	out := []protocol.CardView{sai}
	for i := 2; i <= 3; i++ {
		th := creature(cardID(i), 0, "Thopter", 1, 1, keywords("flying"))
		th.TypeLine = "Artifact Creature — Thopter"
		th.ManaCost = ""
		th.IsToken = true
		out = append(out, th)
	}
	return out
}

// An opponent's Wrath of God is on the stack, so both Thopters are dying
// anyway and cost nothing to sacrifice: the bot draws in response, and
// the card is priced as one (1.20, less the 0.30 every untapped
// creature's activation pays). Backed out (PricePurposes off) it is the
// flat ActivateBase, 0.20 net, which still clears the dying-anyway bar.
func TestSaiDrawsWithThoptersASweepWillTake(t *testing.T) {
	wrath := spell(cardID(9), 1, "Wrath of God", "{2}{W}{W}")
	wrath.TypeLine = "Sorcery"
	wrath.Purpose = &protocol.PurposeView{Sweep: &protocol.SweepView{Matches: "creatures", How: "destroy"}}
	item := protocol.StackItemView{ID: "w", Kind: "spell", Controller: seatID(1).String(), SourceCardID: cardID(9)}
	v := newView(
		[]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(saiWithThopters(1)...),
		withStack(wrath), withStackItems(item),
		withTurn(5, 1, "precombat_main"),
	)
	v.Turn.PriorityHolder = 0
	in := input(0, v, passMove(0),
		activate(t, 0, cardID(1), "Sai: draw", map[string]any{"sacrifice_ids": []string{cardID(2), cardID(3)}}))

	for _, c := range []struct {
		name   string
		prices bool
		want   float64
		reason string
	}{
		{"declared draw", true, 0.90, "declared purpose"},
		{"back-out", false, 0.20, "activate"},
	} {
		t.Run(c.name, func(t *testing.T) {
			cfg := heuristic.DefaultConfig()
			cfg.PricePurposes = c.prices
			pol := heuristic.NewWithConfig(cfg)
			if got := chose(t, in, decide(t, pol, in)); got != "Sai: draw" {
				t.Fatalf("the heuristic chose %q, want Sai's draw", got)
			}
			var found bool
			for _, r := range pol.Rank(context.Background(), in) {
				if r.Index != 1 {
					continue
				}
				found = true
				if math.Abs(r.Value-c.want) > 1e-9 || !strings.Contains(r.Reason, c.reason) {
					t.Errorf("Sai's draw priced %.3f (%s), want %.2f (%s)", r.Value, r.Reason, c.want, c.reason)
				}
			}
			if !found {
				t.Fatal("Sai's draw is not ranked")
			}
		})
	}
}
