package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	cardsn "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/n"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func NezumiShortfangRow() cards.Card {
	return FlipRow(cardsn.NezumiShortfangOracleID, "{1}{B}", []string{"B"}, nil,
		cards.CardFace{Name: "Nezumi Shortfang", TypeLine: "Creature — Rat Rogue",
			OracleText: "{1}{B}, {T}: Target opponent discards a card. Then if that player has no cards in hand, flip this creature.",
			Power:      "1", Toughness: "1"},
		cards.CardFace{Name: "Stabwhisker the Odious", TypeLine: "Legendary Creature — Rat Shaman",
			OracleText: "At the beginning of each opponent's upkeep, that player loses 1 life for each card fewer than three in their hand.",
			Power:      "3", Toughness: "3"})
}

// flipFillHand replaces a player's hand with n plain cards and returns their
// ids.
func FlipFillHand(g *game.Game, p *game.Player, n int) []uuid.UUID {
	g.WithWriteLock(func() { p.Hand.Cards = nil })
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		ids = append(ids, HandCardForTest(p, "Filler", "Sorcery", ""))
	}
	return ids
}
