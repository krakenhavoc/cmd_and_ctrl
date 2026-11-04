package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gutsplitter Gang — Creature — Goblin Berserker {3}{B}, 6/6:
//
//	"At the beginning of your first main phase, you may blight 2. If you
//	 don't, you lose 3 life. (To blight 2, put two -1/-1 counters on a
//	 creature you control.)"
//
// "Your first main phase" is the precombat main (CR 505.1a). The blight
// is asked as the trigger resolves (Blight, #4683). "If you don't" is no
// creature chosen: a decline, or nothing that can take the counters
// (CR 701.68b), loses the 3 life.
func init() {
	Register(Spec{
		OracleID:     "322c1c8f-ef8a-4a9f-b9ce-927720bad1d6",
		Name:         "Gutsplitter Gang",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourPrecombatMain("Gutsplitter Gang — you may blight 2; if you don't, lose 3 life", func(g *game.Game, item *game.StackItem) error {
				return Blight{N: 2, May: true, Then: gutsplitterGangLoseLife}.Apply(NewContext(g, item))
			}),
		},
	})
}

// gutsplitterGangLoseLife is "if you don't, you lose 3 life". An on-disk
// key.
var gutsplitterGangLoseLife = RegisterBlighted("blight/gutsplitter-gang-lose-life", func(ctx *Context, d BlightDone) error {
	if d.Did() {
		return nil
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -3)
})
