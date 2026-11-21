package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reveal_x_cost_cards_test.go — #2598: "Reveal X <colour> cards from
// your hand" as an activated ability's cost (game.RevealCardsCost, ADR
// 0020's 2026-10-08 amendment) and the five Martyrs it pays for.

const (
	MartyrOfBonesOracle  = "a099d8fa-d51e-4dbc-a03f-6f912097d41e"
	MartyrOfSandsOracle  = "f2048140-b691-4753-8239-1aa75059e389"
	MartyrOfSporesOracle = "46c128d5-92b7-4132-a865-4db299879579"
	MartyrOfFrostOracle  = "196cacb0-a26a-4540-9143-57867a638085"
	MartyrOfAshesOracle  = "7f00bc45-7c65-4455-9db0-58bd79bcdb4b"
)

// martyrTable seats a Martyr on seat 0's battlefield (not summoning
// sick: the ability has no {T}) with `mana` generic mana in the pool,
// and a hand of two black cards, one white card and one red card.
func MartyrTable(t *testing.T, name, oracle string, mana int) (g *game.Game, me *game.Player, martyr uuid.UUID, hand map[string]uuid.UUID) {
	t.Helper()
	g = NewCatalogGame(t)
	me = g.Seats[0]
	martyr = PushCatalogPermanent(g, me.ID, name, "Creature — Human Wizard", oracle, false)
	for i := 0; i < mana; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	hand = map[string]uuid.UUID{
		"black1": CcHandCard(me, "Black One", "Sorcery", "{1}{B}"),
		"black2": CcHandCard(me, "Black Two", "Instant", "{B}"),
		"white":  CcHandCard(me, "White One", "Sorcery", "{W}"),
		"red":    CcHandCard(me, "Red One", "Sorcery", "{R}"),
	}
	return g, me, martyr, hand
}

func Ids(xs ...uuid.UUID) []uuid.UUID { return xs }

func OnBattlefieldByID(g *game.Game, id uuid.UUID) bool { return g.Battlefield.Contains(id) }
