package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// may_pay_x_cards_test.go — #2727: "you may pay {X}{R}" as a trigger
// resolves (ADR 0129's amendment of 2026-10-09).

const (
	TilonallisSummonerOracle = "830d3450-18e7-4bac-b330-3906ef5e4514"
	FlameblastDragonOracle   = "3180ee4f-66d7-4b73-bc10-9ee0c790ec18"
)

// summonerAttacks puts Tilonalli's Summoner on the active seat's
// battlefield, fills its pool with `generic` colourless and `red` red
// mana, attacks the next seat with it and resolves the trigger.
func SummonerAttacks(t *testing.T, g *game.Game, generic, red int) (me, opp *game.Player) {
	t.Helper()
	me = g.Seats[g.Turn.ActiveSeat]
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	summoner := PushDiesCreatureForTest(g, me.ID, "Tilonalli's Summoner", TilonallisSummonerOracle, "Creature — Human Shaman", 1, 1)
	AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(summoner, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	LockInAttacks(t, g)
	FillPool(me, generic)
	FillPoolColored(me, "R", red)
	PassPriorityAroundTable(t, g)
	return me, opp
}

// answerNumber answers the open number prompt for `p` with n.
func AnswerNumber(t *testing.T, g *game.Game, p uuid.UUID, n int) *game.PayAmountPrompt {
	t.Helper()
	c := PayAmountPrompt(g, p)
	if c == nil {
		t.Fatalf("no number prompt for %s", p)
	}
	pa := *c.PayAmount
	if err := g.ResolvePayAmount(c.ID, p, n); err != nil {
		t.Fatalf("answer %d: %v", n, err)
	}
	return &pa
}

func Elementals(g *game.Game, controller uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elemental" && c.Controller == controller {
			out = append(out, c)
		}
	}
	return out
}
