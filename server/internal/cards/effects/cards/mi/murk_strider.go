package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Murk Strider — Creature — Eldrazi Processor {3}{U}, 3/2 (Battle for
// Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 When this creature enters, you may put a card an opponent owns from exile
//	 into that player's graveyard. If you do, return target creature to its
//	 owner's hand."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d) and
// re-checked as it is returned (CR 608.2b); it may be Murk Strider itself.
//
// No simplification.
func init() {
	const name = "Murk Strider"
	effects.Register(effects.Spec{
		OracleID:        "386b8c1f-620c-4561-a9d5-d01e2eb5f12a",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{
			effects.OldB305WhenThisEntersMayProcess(name, "murk-strider",
				"return target creature to its owner's hand",
				effects.TargetCreature("target creature"),
				func(ctx *effects.Context) error {
					return effects.BounceTheTarget(ctx.Item, ctx)
				}),
		},
	})
}
