package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc17Cast puts a card in the player's hand (they must be the active
// seat) and casts it from a main phase, returning its id and the cast
// error. The spell is left on the stack.
func Msc17Cast(t *testing.T, g *game.Game, p *game.Player, c game.Card, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	AdvanceToMain(t, g)
	p.Hand.PushTop(c)
	return c.InstanceID, g.CastSpell(p.ID, c.InstanceID, params)
}

// msc17Play casts a card and resolves everything it puts on the stack.
func Msc17Play(t *testing.T, g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	t.Helper()
	id, err := Msc17Cast(t, g, p, c, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	PassPriorityAroundTable(t, g)
	return id
}

// msc17UntilATriggerIsOnTheStack passes priority until a triggered ability
// sits on the stack, leaving it unresolved.
func Msc17UntilATriggerIsOnTheStack(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 16; i++ {
		for _, it := range g.StackMeta {
			if it != nil && it.Kind == game.StackItemTriggered {
				return
			}
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("no triggered ability reached the stack")
}

func Msc17TargetPlayer(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetPlayer, ID: id, Mode: 0}
}
