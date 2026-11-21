package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// castPawprintSpell puts the spell in the active player's hand, floats
// exactly its cost and casts it with `modes`.
func CastPawprintSpell(t *testing.T, g *game.Game, name, typeLine, cost, mana, oracle string, modes []int, targets []game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
		Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
	AdvanceToMain(t, g)
	FloatMana(t, g, me, mana)
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{Modes: modes, Targets: targets, Strict: true})
}

func ModeTargetRef(id uuid.UUID, occurrence int) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id, Mode: occurrence}
}
