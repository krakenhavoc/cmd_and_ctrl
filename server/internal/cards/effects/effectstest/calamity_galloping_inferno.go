package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// calamity_galloping_inferno_test.go — #3548, ADR 0186 slice 6.

const (
	calamityOracle      = "eb37496e-374f-41b9-bd52-f52f37bdac31"
	calamityProbeOracle = "test-calamity-attack-probe"
)

// calamityProbeAttacks counts "whenever this creature attacks" triggers of
// the probe creature below, which Calamity copies.
var CalamityProbeAttacks int

func init() {
	// A creature with "Whenever this creature attacks, ..." so a token copy
	// that was (wrongly) declared as an attacker would show it.
	effects.Register(effects.Spec{
		OracleID: calamityProbeOracle,
		Name:     "Calamity Probe Saddler",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			Key:     "Calamity Probe Saddler — count an attack",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclared(ev, source)
			},
			Effect: func(_ *game.Game, _ *game.StackItem) error {
				CalamityProbeAttacks++
				return nil
			},
		}},
	})
}

func PushCalamity(g *game.Game, owner uuid.UUID) uuid.UUID {
	return PushDiesCreatureForTest(g, owner, "Calamity, Galloping Inferno", calamityOracle, "Legendary Creature — Horse Mount", 4, 6)
}

func PushProbeSaddler(g *game.Game, owner uuid.UUID) uuid.UUID {
	return PushDiesCreatureForTest(g, owner, "Calamity Probe Saddler", calamityProbeOracle, "Creature — Test", 2, 2)
}

// calamityAttackPrompt saddles Calamity, attacks `defender` with it and
// settles the stack up to the first choice, which it returns.
func CalamityAttackPrompt(t *testing.T, g *game.Game, me *game.Player, defender uuid.UUID, calamity uuid.UUID, saddlers ...uuid.UUID) *game.PendingChoice {
	t.Helper()
	ToMain(t, g)
	if err := SaddleAbility(t, g, me.ID, calamity, saddlers...); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	PassPriorityAroundTable(t, g)
	DeclareAttack(t, g, defender, calamity)
	PassPriorityAroundTable(t, g)
	return LatestChooseCards(g, me.ID)
}

// answerAttackSplit answers the "how many of the tokens attack X" prompts
// of ONE token: no for the first `skip` players offered, yes for the next.
// The last player offered takes the token without a prompt.
func AnswerAttackSplit(t *testing.T, g *game.Game, me *game.Player, skip int) {
	t.Helper()
	for i := 0; ; i++ {
		c := PayAmountPrompt(g, me.ID)
		if c == nil {
			return
		}
		n := 0
		if i >= skip {
			n = 1
		}
		if err := g.ResolvePayAmount(c.ID, me.ID, n); err != nil {
			t.Fatalf("attack split answer %d: %v", n, err)
		}
	}
}

func CalamityTokens(g *game.Game, controller uuid.UUID, name string) []*game.Card {
	var out []*game.Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == controller && c.IsToken() && c.Name == name {
			out = append(out, c)
		}
	}
	return out
}
