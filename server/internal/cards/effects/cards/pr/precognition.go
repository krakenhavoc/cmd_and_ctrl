package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Precognition — Enchantment {4}{U} (Tempest, S82 OLD batch 765, #5628):
//
//	"At the beginning of your upkeep, you may look at the top card of target
//	 opponent's library. If you do, you may put that card on the bottom of that
//	 player's library."
//
// The "you may" is the trigger's own (CR 603.5), asked as the ability resolves;
// the target opponent is chosen as it goes on the stack. Looking is Jace, the
// Mind Sculptor's +2 (LookAtLibraryThenPlace with the top-or-bottom placement),
// so declining the first "may" shows nothing and the second "may" is the
// placement prompt: the card goes back on top or to the bottom of its owner's
// library. A target that has left the game, or an empty library, does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6d1a9d06-028f-487a-b490-27100c128ea3",
		Name:         "Precognition",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:        []game.EventKind{game.EventBeginUpkeep},
			AppliesTo:      effects.ByYou,
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Precognition — look at the top card of target opponent's library?"},
			Targets:        effects.TargetPlayer("target opponent", effects.Opponent()),
			Key:            "Precognition — look at the top card of target opponent's library, may put it on the bottom",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				var victim uuid.UUID
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						victim = t.ID
					}
				}
				if victim == uuid.Nil {
					return nil
				}
				return effects.LookAtLibraryThenPlace{
					Owner:     victim,
					N:         1,
					Placement: game.LibraryPlaceTopOrBottom,
					Label:     "Precognition — you may put that card on the bottom of that player's library",
				}.Apply(ctx)
			},
		}},
	})
}
