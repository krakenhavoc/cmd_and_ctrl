package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oltec Archaeologists — Creature — Human Artificer Scout {4}{W}, 4/4
// (The Lost Caverns of Ixalan, #4470):
//
//	"When this creature enters, choose one —
//	 • Return target artifact card from your graveyard to your hand.
//	 • Scry 3."
//
// A modal enters trigger (CR 603.3c), Daily Bugle Reporters' shape: the
// mode is chosen as the trigger is put on the stack, and only the first
// bullet has a target, a card in your own graveyard, which is skipped if
// it has left by resolution (CR 608.2b). The second is the ordinary scry
// (CR 701.22), declared as the bullet's rider (Momo, Playful Pet's shape).
//
// No simplification.
func init() {
	etb := effects.WhenThisEnters("Oltec Archaeologists — choose one", func(*game.Game, *game.StackItem) error { return nil })
	etb.Modes = effects.ChooseOne(
		effects.ModeDoing("Return target artifact card from your graveyard to your hand.",
			effects.TargetCardInGraveyard("target artifact card from your graveyard", effects.YouOwn(), effects.Artifact()),
			effects.ReturnTheModesGraveyardTargetToHand),
		effects.ModeWithPurpose(effects.ModeSteps("Scry 3.", nil, effects.Scry{N: 3}), game.Purpose{Scry: 3}),
	)
	effects.Register(effects.Spec{
		OracleID:     "b19c4bbf-258f-4436-8ef9-fc905c283776",
		Name:         "Oltec Archaeologists",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{etb},
	})
}
