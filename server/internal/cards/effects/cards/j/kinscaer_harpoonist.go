package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kinscaer Harpoonist — {3}{U} Creature — Kithkin Soldier 2/2 (Shadowmoor, S82
// older sets):
//
//	"Flying
//	 Whenever this creature attacks, you may have target creature lose flying
//	 until end of turn."
//
// Crash Landing's loss of flying until end of turn, as an optional attack trigger
// with a target chosen as the trigger goes on the stack. A target that is gone
// or illegal as the trigger resolves does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8ee98d6a-6a7e-411d-818c-816d84cfb7cd",
		Name:            "Kinscaer Harpoonist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WheneverThisAttacks("Kinscaer Harpoonist — target creature loses flying until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.UntilEndOfTurn(ctx, id, nil, "Kinscaer Harpoonist — loses flying", game.RemoveKeywordsMod("flying"))
				}), effects.TargetCreature("target creature")),
				"Kinscaer Harpoonist — have target creature lose flying until end of turn?"),
		},
	})
}
