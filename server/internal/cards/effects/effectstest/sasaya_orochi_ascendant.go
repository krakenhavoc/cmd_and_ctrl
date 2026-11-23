package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	cardss "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/s"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sasaya_orochi_ascendant_test.go — #3250: "Reveal your hand" as an
// activated ability's cost (game.AbilityCost.RevealHand) and the flip
// card it pays for, down the real import road.

func SasayaRow() cards.Card {
	return FlipRow(cardss.SasayaOracleID, "{1}{G}{G}", []string{"G"}, nil,
		cards.CardFace{Name: "Sasaya, Orochi Ascendant", TypeLine: "Legendary Creature — Snake Monk",
			OracleText: "Reveal your hand: If you have seven or more land cards in your hand, flip Sasaya.",
			Power:      "2", Toughness: "3"},
		cards.CardFace{Name: "Sasaya's Essence", TypeLine: "Legendary Enchantment",
			OracleText: "Whenever a land you control is tapped for mana, add an additional one mana of any type that land produced for each other land you control with the same name as it."})
}

// emptyHand clears the seat's hand.
func EmptyHand(g *game.Game, p *game.Player) {
	g.WithWriteLock(func() { p.Hand.Cards = nil })
}

// --- Sasaya's Essence -------------------------------------------------

func FlippedSasaya(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	id := FlipEntered(t, g, SasayaRow(), me)
	g.WithWriteLock(func() {
		if err := g.FlipPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	return id
}

func PushNamedLand(g *game.Game, p *game.Player, name, typeLine string) uuid.UUID {
	return B31Push(g, p.ID, name, typeLine, "", "", 0, 0)
}
