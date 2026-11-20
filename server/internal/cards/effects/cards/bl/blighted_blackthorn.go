package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "a294968e-9a7f-4d66-bde0-e775ae5f8197",
		Name:         "Blighted Blackthorn",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Blighted Blackthorn — you may blight 2; if you do, draw a card and lose 1 life", func(g *game.Game, item *game.StackItem) error {
				return effects.Blight{N: 2, May: true, Then: blightedBlackthornDraw}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

// blightedBlackthornDraw is "if you do, you draw a card and lose 1 life".
// An on-disk key.
var blightedBlackthornDraw = effects.RegisterBlighted("blight/blighted-blackthorn-draw", func(ctx *effects.Context, d effects.BlightDone) error {
	if !d.Did() {
		return nil
	}
	return effects.B36DrawAndLoseOne(ctx.Game, ctx.Item)
})
