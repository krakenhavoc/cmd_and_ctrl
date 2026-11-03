package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blighted Blackthorn — Creature — Treefolk Warlock {4}{B}, 3/7:
//
//	"Whenever this creature enters or attacks, you may blight 2. If you
//	 do, you draw a card and lose 1 life. (To blight 2, put two -1/-1
//	 counters on a creature you control.)"
//
// One ability with two trigger conditions (WhenThisEntersOrAttacks). The
// blight is asked as it resolves (Blight, #4683); a creature chosen draws
// the card and loses the life, in that order.
func init() {
	Register(Spec{
		OracleID:     "a294968e-9a7f-4d66-bde0-e775ae5f8197",
		Name:         "Blighted Blackthorn",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Blighted Blackthorn — you may blight 2; if you do, draw a card and lose 1 life", func(g *game.Game, item *game.StackItem) error {
				return Blight{N: 2, May: true, Then: blightedBlackthornDraw}.Apply(NewContext(g, item))
			}),
		},
	})
}

// blightedBlackthornDraw is "if you do, you draw a card and lose 1 life".
// An on-disk key.
var blightedBlackthornDraw = RegisterBlighted("blight/blighted-blackthorn-draw", func(ctx *Context, d BlightDone) error {
	if !d.Did() {
		return nil
	}
	return b36DrawAndLoseOne(ctx.Game, ctx.Item)
})
