package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Battle of Bywater — Sorcery {1}{W}{W} (EDHREC rank 1834):
//
//	"Destroy all creatures with power 3 or greater. Then create a
//	 Food token for each creature you control. (It's an artifact with
//	 "{2}, {T}, Sacrifice this token: You gain 3 life.")"
//
// The go-wide deck's one-sided wrath: the big things die and every
// small creature you kept becomes a Food. Power is the current power
// (PowerGE reads counters and anthems), and the Food count is taken
// AFTER the destruction, so a creature of yours that survived — or
// was too small to be swept — is what pays.
//
// The batched sweep. DestroyAllMatching destroys the big creatures as
// one simultaneous event (CR 700.4), so a "whenever another creature
// dies" watcher caught in the wipe sees every death, and it leaves
// indestructible creatures on the battlefield (#446 / #470), where
// one of yours counts for a Food like any other creature you control.
//
// The Food count cannot simply read the board after the sweep. A
// commander of yours with power 3 or more is destroyed, but CR 903.9
// queues its owner's command-zone prompt and the card stays on the
// battlefield until they answer, which is after the Foods are made.
// So the Then clause leaves out every swept card that is still on
// the battlefield: it was destroyed and is not a creature you
// control any more, whichever zone its owner picks.
func init() {
	Register(Spec{
		OracleID:     "a94c191d-a938-458e-bc1b-2f44fd8873a3",
		Name:         "The Battle of Bywater",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{
				Match: And(Creature(), PowerGE(3)),
				Then: func(ctx *Context, swept []game.Card, _ int) error {
					return b17FoodPerCreatureYouControl(ctx, swept)
				},
			}.Apply(ctx)
		},
	})
}
