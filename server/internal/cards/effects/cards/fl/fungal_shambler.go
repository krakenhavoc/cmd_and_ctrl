package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "9991941e-436f-42d4-8b0c-1ee84234774b",
		Name:            "Fungal Shambler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsDamageToAnOpponent("Fungal Shambler — you draw a card and that opponent discards a card",
				effects.Do(
					effects.DrawCards{N: 1},
					effects.Instruction(func(ctx *effects.Context) error {
						return effects.DiscardOne(ctx.Game, ctx.Item, ctx.Item.Trigger.Event.Target)
					}),
				)),
		},
	})
}
