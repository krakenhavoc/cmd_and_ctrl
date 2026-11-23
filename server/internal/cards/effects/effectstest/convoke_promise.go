package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// convoke_promise_test.go — #5647: "the next spell you cast this turn has
// convoke" (Wand of the Worldsoul) and the engine's record that a spell
// had convoke as it became cast (StackItem.HadConvoke).

const WandOfTheWorldsoulOracle = "3d175a59-c1b4-43b8-8a1f-f7cbcefe130e"

// cpHadConvoke reads the cast-time record for a spell on the stack.
func CpHadConvoke(g *game.Game, spell uuid.UUID) (had bool) {
	g.WithWriteLock(func() { had = g.SpellHadConvokeForEffect(spell) })
	return had
}

// cpActivateWand puts a Wand under `me` and resolves its promise.
func CpActivateWand(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	wand := PushCatalogPermanent(g, me.ID, "Wand of the Worldsoul", "Artifact", WandOfTheWorldsoulOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, wand, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activating Wand of the Worldsoul: %v", err)
	}
	PassPriorityAroundTable(t, g)
	if got := NsPromises(me); len(got) != 1 || !got[0].Convoke {
		t.Fatalf("the promise: %+v", got)
	}
}
