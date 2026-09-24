package legal_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

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
			creature := func(c game.Card, name string) game.Card {
				c.Name, c.TypeLine, c.Toughness = name, "Creature — Test", 4
				if c.Power == 0 {
					c.Power = 2
				}
				return c
			}
			// setup seats the attacker and the given defending
			// creatures, and walks into declare_blockers with the
			// attack declared. The defending creatures exist BEFORE the
			// step begins: #1279 completes a defender's declaration as
			// the step begins when they have no legal block, so a
			// creature pushed afterwards is not asked about.
			setup := func(defending ...game.Card) (*game.Game, *game.Player, uuid.UUID, []uuid.UUID) {
				g := newTable(t)
				attackerSeat, defender := g.Seats[0], g.Seats[1]
				clearHand(attackerSeat)
				clearHand(defender)
				attacker := battlefieldCard(g, attackerSeat, creature(tc.attacker, "Attacker"))
				var ids []uuid.UUID
				for _, c := range defending {
					ids = append(ids, battlefieldCard(g, defender, c))
				}
				advanceTo(t, g, game.StepDeclareAttackers)
				if err := g.DeclareAttacker(attacker, defender.ID); err != nil {
					t.Fatal(err)
				}
				advanceTo(t, g, game.StepDeclareBlockers)
				return g, defender, attacker, ids
			}

			// Only the refused blocker: no legal block, nothing owed,
			// and the declaration — none — is already complete.
			g, defender, attacker, ids := setup(creature(tc.refused, "Refused blocker"))
			refused := ids[0]
			moves := legal.EnumerateFor(g, defender.ID)
			if n := blocksBy(moves, refused); n != 0 {
				t.Fatalf("offered %d illegal blocks: %v", n, labels(moves))
			}
			if g.SeatOwesBlockDecision(defender.ID) {
				t.Fatal("decision signal says a defender with no legal block owes a choice")
			}
			if got := g.BlockDeclarationStatusOf(defender.ID); got != game.BlockDeclarationDeclared {
				t.Fatalf("a defender with no legal block has declared none: status %q", got)
			}
			err := g.DeclareBlocker(refused, attacker)
			var refusal *game.BlockRefusedError
			if !errors.As(err, &refusal) || refusal.Reason != tc.reason {
				t.Fatalf("declaration refusal = %v, want %s", err, tc.reason)
			}

			// Both: the legal one is offered, the other is not.
			g, defender, attacker, ids = setup(creature(tc.refused, "Refused blocker"), creature(tc.allowed, "Allowed blocker"))
			refused, allowed := ids[0], ids[1]
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
