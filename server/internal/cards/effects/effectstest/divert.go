package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// divert_test.go — #2365: the guarded pay-unless with a custom decline.

const DivertOracle = "561aed01-322b-4699-b7b0-f4416075aa9b"

// boltThenDivert has the active seat (the future payer) Bolt seat 2 and
// seat 1 answer with Divert. Seat 0 is asked once Divert resolves.
func BoltThenDivert(t *testing.T, g *game.Game) (bolt, divert uuid.UUID) {
	t.Helper()
	victim := g.Seats[2].ID
	bolt = CastCatalogSpell(t, g, "Lightning Bolt", "Instant", LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})
	caster := g.Seats[1]
	divert = uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: divert, Name: "Divert", TypeLine: "Instant", OracleID: DivertOracle,
		Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, divert, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}},
	}); err != nil {
		t.Fatalf("CastSpell Divert: %v", err)
	}
	return bolt, divert
}
