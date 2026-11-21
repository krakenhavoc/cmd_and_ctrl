package testsg

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// graveyard_exile_tokens_test.go — "if a CARD would be put into an
// opponent's graveyard" does not touch a token (CR 108.2): the token dies
// (CR 700.4) and its death triggers fire. Rest in Peace says "card or
// token" and still takes it. Found in an agent-played game, where an
// opponent's Dauthi Voidwalker ate Teysa's sacrificed Soldier token and
// every death trigger with it.

// zulaportBoard seats Zulaport Cutthroat, a Soldier token and a Bear card
// for me, and hate under its controller's name.
func zulaportBoard(t *testing.T, hateOwner int, hateName, hateType, hateOracle string) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue Ally", effectstest.ZulaportCutthroatOracle, false)
	effectstest.PushCatalogPermanent(g, g.Seats[hateOwner].ID, hateName, hateType, hateOracle, false)
	return g, me, opp
}

func TestLeylineOfTheVoidLetsATokenDie(t *testing.T) {
	g, me, opp := zulaportBoard(t, 1, "Leyline of the Void", "Enchantment", effectstest.LeylineOfTheVoidOracle)
	token := effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Human Soldier"))
	life := opp.Life

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(token) })
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life %d → %d: a token is not a card, so it died and Zulaport should drain 1", life, opp.Life)
	}
	for _, c := range g.Exile.Cards {
		if c.InstanceID == token {
			t.Error("Leyline of the Void exiled a token; it replaces only cards")
		}
	}
}

func TestLeylineOfTheVoidStillExilesACard(t *testing.T) {
	g, me, opp := zulaportBoard(t, 1, "Leyline of the Void", "Enchantment", effectstest.LeylineOfTheVoidOracle)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	life := opp.Life

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	exiled := false
	for _, c := range g.Exile.Cards {
		exiled = exiled || c.InstanceID == bear
	}
	if !exiled {
		t.Error("Leyline of the Void did not exile the Bear card")
	}
	if opp.Life != life {
		t.Errorf("opponent life %d → %d: an exiled card did not die, so nothing should drain", life, opp.Life)
	}
}

func TestRestInPeaceStillExilesAToken(t *testing.T) {
	g, me, opp := zulaportBoard(t, 1, "Rest in Peace", "Enchantment", effectstest.RestInPeaceOracle)
	token := effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Human Soldier"))
	life := opp.Life

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(token) })
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("opponent life %d → %d: Rest in Peace says \"card or token\", so the token never died", life, opp.Life)
	}
}
