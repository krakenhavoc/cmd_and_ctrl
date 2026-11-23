package effectstest

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// eoTable is a catalog game in the active player's main phase, with the
// active player's and the next seat's hands emptied so each test says
// exactly what can pay.
func EoTable(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := NewCatalogGame(t)
	AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for _, p := range []*game.Player{me, opp} {
		for p.Hand.Size() > 0 {
			if _, err := p.Hand.PopTop(); err != nil {
				t.Fatal(err)
			}
		}
	}
	return g, me, opp
}

// eoHand puts a card of the type in p's hand.
func EoHand(p *game.Player, name, typeLine string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = typeLine
	p.Hand.PushTop(c)
	return c.InstanceID
}

// eoArtifact puts a plain artifact under owner.
func EoArtifact(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return PushCatalogPermanent(g, owner, name, "Artifact", "", false)
}

// eoPrompt is the open either payment addressed to chooser, or nil.
func EoPrompt(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == chooser {
			if a := c.PayAction(); a != nil && a.Kind == game.PayActionEither {
				return c
			}
		}
	}
	return nil
}

// eoPay answers the open either payment with branch `option`.
func EoPay(t *testing.T, g *game.Game, chooser uuid.UUID, option int, cards ...uuid.UUID) {
	t.Helper()
	c := EoPrompt(g, chooser)
	if c == nil {
		t.Fatalf("no either payment for %s: %+v", chooser, g.PendingChoices)
	}
	if err := g.ResolvePayUnlessWithOption(c.ID, chooser, true, &option, cards); err != nil {
		t.Fatalf("pay option %d: %v", option, err)
	}
}

// eoOfferedOptions is the pay_option of every "pay" move the enumerator
// offers chooser on the open either payment.
func EoOfferedOptions(t *testing.T, g *game.Game, chooser uuid.UUID) map[int]bool {
	t.Helper()
	out := map[int]bool{}
	for _, m := range legal.EnumerateFor(g, chooser) {
		if m.Type != "resolve_choice" {
			continue
		}
		var p struct {
			Apply     *bool `json:"apply"`
			PayOption *int  `json:"pay_option"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		if p.PayOption != nil {
			out[*p.PayOption] = true
		}
	}
	return out
}

// eoActivationBranches is the cost_branch of every activation of
// `source` the enumerator offers seat; -1 marks a move with none.
func EoActivationBranches(t *testing.T, g *game.Game, seat, source uuid.UUID) map[int]bool {
	t.Helper()
	out := map[int]bool{}
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type != "activate_ability" || m.Source != source {
			continue
		}
		var p struct {
			CostBranch *int `json:"cost_branch"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		if p.CostBranch == nil {
			out[-1] = true
			continue
		}
		out[*p.CostBranch] = true
	}
	return out
}

// eoTopAbility is the activated ability on the stack, or nil.
func EoTopAbility(g *game.Game) *game.StackItem {
	for _, it := range g.StackMeta {
		if it != nil && it.Kind == game.StackItemActivated {
			return it
		}
	}
	return nil
}
