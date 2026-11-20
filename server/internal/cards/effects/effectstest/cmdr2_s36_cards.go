package effectstest

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s36CastFaceDown casts a card from the active seat's hand face down by
// its alternative cost key ("morph" or "megamorph") and resolves it.
func S36CastFaceDown(t *testing.T, g *game.Game, name, typeLine, oracle, key string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{AlternativeCost: key}); err != nil {
		t.Fatalf("cast %s face down: %v", name, err)
	}
	PassPriorityAroundTable(t, g)
	if c := FindCardAnywhere(t, g, id); !c.FaceDownIsPermanent() {
		t.Fatalf("%s did not enter face down", name)
	}
	return id
}

func S36TurnUp(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, mana ...string) {
	t.Helper()
	MkmMPTurnFaceUp(t, g, p, id, mana...)
}
