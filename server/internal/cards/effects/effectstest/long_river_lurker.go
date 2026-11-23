package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const LongRiverLurkerOracle = "c65c4d01-29da-47e6-993c-a265cf1f198c"

// lrlCastLurker casts the Lurker from hand, with `target` as its entry
// trigger's target, and settles the stack.
func LrlCastLurker(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Long River Lurker", TypeLine: "Creature — Frog Scout", OracleID: LongRiverLurkerOracle,
			ManaCost: "{2}{U}", Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID})
	})
	FloatForTest(g, me, "UUU")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	PassPriorityAroundTable(t, g)
	AnswerPickTarget(t, g, target)
	PassPriorityAroundTable(t, g)
	return id
}
