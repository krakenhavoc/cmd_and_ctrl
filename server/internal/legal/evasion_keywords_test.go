package legal_test

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

func TestEvasionBlocksAgreeWithDeclarationsAndDecisionSignal(t *testing.T) {
	for _, tc := range []struct {
		name                       string
		attacker, refused, allowed game.Card
		reason                     game.BlockReason
	}{
		{"fear", game.Card{Keywords: []string{"fear"}}, game.Card{}, game.Card{Colors: []string{"B"}}, "fear"},
		{"intimidate", game.Card{Keywords: []string{"intimidate"}, Colors: []string{"R"}}, game.Card{Colors: []string{"G"}}, game.Card{Colors: []string{"R"}}, "intimidate"},
		{"shadow attacker", game.Card{Keywords: []string{"shadow"}}, game.Card{}, game.Card{Keywords: []string{"shadow"}}, "shadow"},
		{"shadow blocker", game.Card{}, game.Card{Keywords: []string{"shadow"}}, game.Card{}, "shadow"},
		{"horsemanship", game.Card{Keywords: []string{"horsemanship"}}, game.Card{Keywords: []string{"flying", "reach"}}, game.Card{Keywords: []string{"horsemanship"}}, "horsemanship"},
		{"skulk", game.Card{Keywords: []string{"skulk"}, Power: 2}, game.Card{Power: 3}, game.Card{Power: 2}, "skulk"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newTable(t)
			attackerSeat, defender := g.Seats[0], g.Seats[1]
			clearHand(attackerSeat)
			clearHand(defender)
			creature := func(c game.Card, name string) game.Card {
				c.Name, c.TypeLine, c.Toughness = name, "Creature — Test", 4
				if c.Power == 0 {
					c.Power = 2
				}
				return c
			}
			attacker := battlefieldCard(g, attackerSeat, creature(tc.attacker, "Attacker"))
			refused := battlefieldCard(g, defender, creature(tc.refused, "Refused blocker"))
			advanceTo(t, g, game.StepDeclareAttackers)
			if err := g.DeclareAttacker(attacker, defender.ID); err != nil {
				t.Fatal(err)
			}
			advanceTo(t, g, game.StepDeclareBlockers)

			moves := legal.EnumerateFor(g, defender.ID)
			if n := blocksBy(moves, refused); n != 0 {
				t.Fatalf("offered %d illegal blocks: %v", n, labels(moves))
			}
			if g.SeatOwesBlockDecision(defender.ID) {
				t.Fatal("decision signal says a defender with no legal block owes a choice")
			}
			err := g.DeclareBlocker(refused, attacker)
			var refusal *game.BlockRefusedError
			if !errors.As(err, &refusal) || refusal.Reason != tc.reason {
				t.Fatalf("declaration refusal = %v, want %s", err, tc.reason)
			}

			allowed := battlefieldCard(g, defender, creature(tc.allowed, "Allowed blocker"))
			moves = legal.EnumerateFor(g, defender.ID)
			if blocksBy(moves, allowed) != 1 || blocksBy(moves, refused) != 0 {
				t.Fatalf("wrong block choices: %v", labels(moves))
			}
			if !g.SeatOwesBlockDecision(defender.ID) {
				t.Fatal("decision signal skipped a legal block")
			}
			dispatchAll(t, g, defender.ID, moves)
			if err := g.DeclareBlocker(allowed, attacker); err != nil {
				t.Fatalf("offered block refused: %v", err)
			}
		})
	}
}
