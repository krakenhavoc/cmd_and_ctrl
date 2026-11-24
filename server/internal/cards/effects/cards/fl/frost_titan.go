package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Frost Titan — Creature — Giant {4}{U}{U}, 6/6:
//
//	"Whenever this creature becomes the target of a spell or ability an
//	 opponent controls, counter that spell or ability unless its controller
//	 pays {2}.
//	 Whenever this creature enters or attacks, tap target permanent. It
//	 doesn't untap during its controller's next untap step."
//
// The first ability is ward {2} in everything but the word: a printed
// triggered ability (the card has no ward keyword to read), built with the
// same trigger body as ward so the tax is paid or the spell or ability is
// countered. The second taps any target permanent and freezes it with
// TapAndFreeze; a target that is gone by resolution is skipped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c5a2443-a038-44d3-83a4-5e9a6db7b1a7",
		Name:         "Frost Titan",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Frost Titan — whenever this creature becomes the target of a spell or ability an opponent controls, counter that spell or ability unless its controller pays {2}"),
			effects.Targeting(
				effects.WhenThisEntersOrAttacks("Frost Titan — tap target permanent; it doesn't untap during its controller's next untap step",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Frost Titan — it doesn't untap during its controller's next untap step"}.Apply(ctx)
					}),
				effects.TargetPermanent("target permanent")),
		},
	})
}
