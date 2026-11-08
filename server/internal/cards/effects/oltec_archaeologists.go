package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	etb := WhenThisEnters("Oltec Archaeologists — choose one", func(*game.Game, *game.StackItem) error { return nil })
	etb.Modes = ChooseOne(
		ModeDoing("Return target artifact card from your graveyard to your hand.",
			TargetCardInGraveyard("target artifact card from your graveyard", YouOwn(), Artifact()),
			ReturnTheModesGraveyardTargetToHand),
		ModeWithPurpose(ModeSteps("Scry 3.", nil, Scry{N: 3}), game.Purpose{Scry: 3}),
	)
	Register(Spec{
		OracleID:     "b19c4bbf-258f-4436-8ef9-fc905c283776",
		Name:         "Oltec Archaeologists",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{etb},
	})
}
