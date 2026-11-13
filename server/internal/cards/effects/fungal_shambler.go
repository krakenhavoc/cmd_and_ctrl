package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fungal Shambler — Creature — Fungus Beast {4}{B}{G}{U}, 6/4
// (Apocalypse, #5614):
//
//	"Trample
//	 Whenever this creature deals damage to an opponent, you draw a card
//	 and that opponent discards a card."
//
// Trample rides PrintedKeywords. The trigger is Thieving Otter's
// (WheneverThisDealsDamageToAnOpponent: any damage, combat or not, to an
// opponent of its controller), with Dimir Cutpurse's two halves in the
// printed order: the draw, then the damaged opponent discards a card of
// their choice (CR 701.9b). An opponent with an empty hand discards
// nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9991941e-436f-42d4-8b0c-1ee84234774b",
		Name:            "Fungal Shambler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsDamageToAnOpponent("Fungal Shambler — you draw a card and that opponent discards a card",
				Do(
					DrawCards{N: 1},
					Instruction(func(ctx *Context) error { return discardOne(ctx.Game, ctx.Item, ctx.Item.Trigger.Event.Target) }),
				)),
		},
	})
}
