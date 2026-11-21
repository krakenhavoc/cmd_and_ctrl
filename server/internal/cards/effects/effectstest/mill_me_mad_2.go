package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mm2Library replaces a player's library with the given cards, the
// first one on top.
func Mm2Library(p *game.Player, cards ...game.Card) []uuid.UUID {
	p.Library.Cards = nil
	ids := make([]uuid.UUID, 0, len(cards))
	for _, c := range cards {
		if c.InstanceID == uuid.Nil {
			c.InstanceID = uuid.New()
		}
		c.Owner, c.Controller = p.ID, p.ID
		p.Library.PushBottom(c)
		ids = append(ids, c.InstanceID)
	}
	return ids
}

func Mm2Land(name string) game.Card {
	return game.Card{Name: name, TypeLine: "Basic Land — Forest"}
}

func Mm2Spell(name string) game.Card {
	return game.Card{Name: name, TypeLine: "Sorcery"}
}

func Mm2CreatureCard(name string, power int) game.Card {
	return game.Card{Name: name, TypeLine: TestCreatureTypeLine, Power: power, Toughness: power}
}

func Mm2Zone(g *game.Game, id uuid.UUID) game.ZoneKind {
	var kind game.ZoneKind
	g.ReadSnapshot(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			kind = z.Kind
		}
	})
	return kind
}

// mm2Settle passes priority once, then until the stack is empty or a
// prompt waits.
func Mm2Settle(t *testing.T, g *game.Game) {
	t.Helper()
	PassPriorityAroundTable(t, g)
	for i := 0; i < 16 && g.Stack.Size() > 0 && len(g.PendingChoices) == 0; i++ {
		PassPriorityAroundTable(t, g)
	}
}
