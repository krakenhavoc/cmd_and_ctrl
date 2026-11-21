package effectstest

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- shared helpers ----------------------------------------------------------

func Ecl6Seats(g *game.Game) (*game.Player, *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// ecl6Cast puts a catalog card with printed P/T into p's hand and casts it
// where the cursor stands, without the strict mana gate.
func Ecl6Cast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string, power, toughness int, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
			Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
		})
	})
	if err := g.CastSpell(p.ID, id, params); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// ecl6ResolveSpell passes priority until the stack is empty or a prompt
// opens, leaving the prompt for the caller.
func Ecl6ResolveSpell(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 12 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				return
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
}

func Ecl6NextTurn(t *testing.T, g *game.Game) {
	t.Helper()
	AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
}
