package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spell_control_cards_test.go — the ADR 0104 cards: Invert Polarity,
// Aethersnatch, Commandeer, Sudden Substitution and Perplexing Chimera.

const (
	InvertPolarityOracle     = "69e12b1f-0fd9-43a9-b3db-fe08290442c6"
	AethersnatchOracle       = "45892ec2-8996-444f-a61c-5151926baa0b"
	CommandeerOracle         = "ad7d854b-d303-40eb-acd6-03a8023e05e7"
	SuddenSubstitutionOracle = "e5ddacfb-e5a8-4885-a4c9-fa117f321293"
	PerplexingChimeraOracle  = "7d075b8a-a606-4590-b52b-b4ef3a9e342f"
)

// respondWith hands priority to `p` (passing from the active seat) and
// casts `name` from p's hand in response to what is on the stack.
func RespondWith(t *testing.T, g *game.Game, p *game.Player, card game.Card, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	if card.InstanceID == uuid.Nil {
		card.InstanceID = uuid.New()
	}
	card.Owner, card.Controller = p.ID, p.ID
	g.WithWriteLock(func() { p.Hand.PushTop(card) })
	for i := 0; i < 8 && g.Seats[g.Turn.PriorityHolder].ID != p.ID; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority toward %s: %v", p.Name, err)
		}
	}
	if err := g.CastSpell(p.ID, card.InstanceID, params); err != nil {
		t.Fatalf("CastSpell %s: %v", card.Name, err)
	}
	return card.InstanceID
}

func StackItemControllerOf(g *game.Game, id uuid.UUID) uuid.UUID {
	var out uuid.UUID
	g.ReadSnapshot(func() {
		if it := g.StackMeta[id]; it != nil {
			out = it.Controller
		}
	})
	return out
}
