package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hobB10Stack makes cards the whole of p's library, cards[0] on top.
func HobB10Stack(p *game.Player, cards ...game.Card) {
	p.Library.Cards = nil
	for i := len(cards) - 1; i >= 0; i-- {
		p.Library.PushTop(cards[i])
	}
}

// hobB10Filler makes p's library twenty cards that are not lands, enough to
// last through the turns a Saga test walks.
func HobB10Filler(p *game.Player) {
	cards := make([]game.Card, 0, 20)
	for i := 0; i < 20; i++ {
		cards = append(cards, HobB10Card(p, "Filler", "Sorcery", ""))
	}
	HobB10Stack(p, cards...)
}

// hobB10Card is a library or hand card of p's.
func HobB10Card(p *game.Player, name, typeLine, cost string) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID}
}

// hobB10Has reports whether c, as the layers leave it, has the keyword.
func HobB10Has(c game.Card, keyword string) bool { return game.HasKeyword(&c, keyword) }

// hobB10Treasures gives who n Treasure tokens.
func HobB10Treasures(g *game.Game, who uuid.UUID, n int) {
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(who, effects.TreasureToken(), n) })
}

// hobB10NextChapter walks to seat 0's next precombat main, which adds a lore
// counter, and lets the chapter settle.
func HobB10NextChapter(t *testing.T, g *game.Game) {
	t.Helper()
	AdvanceToPrecombatMainOf(t, g, 0)
	PassPriorityAroundTable(t, g)
}
