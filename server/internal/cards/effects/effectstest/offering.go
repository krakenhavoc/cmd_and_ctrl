package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// offeringParams is a cast that announces the offering over `sac`.
func OfferingParams(strict bool, sac ...uuid.UUID) game.CastSpellParams {
	return game.CastSpellParams{OptionalCosts: []int{0}, SacrificeIDs: sac, Strict: strict, AutoTap: strict}
}

// offeringPrice is what the engine's one pricer charges for `card` under
// `params`.
func OfferingPrice(t *testing.T, g *game.Game, p *game.Player, card uuid.UUID, params game.CastSpellParams) string {
	t.Helper()
	var out string
	g.WithWriteLock(func() {
		c, ok := g.LookupCardForEffect(card)
		if !ok {
			t.Fatal("the card is gone")
		}
		price, err := g.PriceCastForEffect(p.ID, c, params)
		if err != nil {
			t.Fatalf("PriceCastForEffect: %v", err)
		}
		out = price.Total.String()
	})
	return out
}
