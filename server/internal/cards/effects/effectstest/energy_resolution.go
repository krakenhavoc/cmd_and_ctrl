package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// energy_resolution.go — energy-resolution test helpers (ADR 0129) shared across test packages.

// ErPlay is how a test answers whatever the table asks `me` while it
// settles: the target to pick, whether to pay a fixed energy amount,
// the amount for a pay_amount and the option for an option pick.
type ErPlay struct {
	Target uuid.UUID
	Pay    bool
	Amount int
	Option int
}

// SettleEnergy passes priority until the table is quiet, answering each
// prompt owed by `me` from `play`. It returns how many energy prompts
// (either kind) it answered.
func SettleEnergy(t *testing.T, g *game.Game, me uuid.UUID, play ErPlay) int {
	t.Helper()
	answered := 0
	for i := 0; i < 20; i++ {
		PassPriorityAroundTable(t, g)
		switch {
		case LatestPickTarget(g, me) != nil:
			AnswerTriggerTargets(t, g, me, play.Target)
		case EnergyPrompt(g, me) != nil:
			if err := g.ResolvePayUnless(EnergyPrompt(g, me).ID, me, play.Pay); err != nil {
				t.Fatalf("answer energy: %v", err)
			}
			answered++
		case PayAmountPrompt(g, me) != nil:
			if err := g.ResolvePayAmount(PayAmountPrompt(g, me).ID, me, play.Amount); err != nil {
				t.Fatalf("answer amount: %v", err)
			}
			answered++
		case LatestOptionPickFor(g, me) != nil:
			AnswerOptionPick(t, g, me, play.Option)
		default:
			return answered
		}
	}
	t.Fatal("the table did not settle")
	return answered
}

// AttackAndAnswerEnergy attacks with `attacker`, resolves its attack
// trigger, and answers the energy prompt with `pay`.
func AttackAndAnswerEnergy(t *testing.T, g *game.Game, attacker uuid.UUID, pay bool) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	DeclareAttack(t, g, opp.ID, attacker)
	PassPriorityAroundTable(t, g)
	c := EnergyPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no energy prompt after the attack trigger resolved")
	}
	if err := g.ResolvePayUnless(c.ID, me.ID, pay); err != nil {
		t.Fatalf("answer: %v", err)
	}
	PassPriorityAroundTable(t, g)
}
