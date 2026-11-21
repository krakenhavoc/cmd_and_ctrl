package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2B23Cast casts a creature with printed P/T so it survives entry.
func Cmdr2B23Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, p, tough int) uuid.UUID {
	t.Helper()
	return AvB1Cast(t, g, g.Seats[0], game.Card{Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tough}, game.CastSpellParams{})
}
