package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- shared helpers ------------------------------------------------------------

// trekB39Hand puts a card in p's hand with a mana cost and P/T.
func TrekB39Hand(g *game.Game, p *game.Player, name, typeLine, cost, oracle string, power, toughness int) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
			Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
			KnownBy: map[uuid.UUID]bool{p.ID: true},
		})
	})
	return id
}

// trekB39Activate activates a catalog ability strictly with `mana` floated.
func TrekB39Activate(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, mana string, targets ...game.TargetRef) error {
	t.Helper()
	if mana != "" {
		FloatMana(t, g, p, mana)
	}
	return g.ActivateCatalogAbility(p.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: targets})
}

func TrekB39Named(g *game.Game, controller uuid.UUID, name string) (cards, tokens int) {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			if effects.IsToken(c) {
				tokens++
			} else {
				cards++
			}
		}
	}
	return cards, tokens
}
