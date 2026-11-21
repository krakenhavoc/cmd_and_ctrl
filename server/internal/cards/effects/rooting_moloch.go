package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rooting Moloch — Creature — Lizard {4}{R}, 4/5 (Jumpstart 2022, #5948):
//
//	"When this creature enters, exile target card with a cycling ability
//	 from your graveyard. Until the end of your next turn, you may play that
//	 card.
//	 Cycling {2}"
//
// The target is a card in your graveyard with a cycling row (typecycling
// counts, CR 702.29e), re-checked as the trigger resolves. The exiled card
// carries a play permission (land or spell) until the end of your next turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4056c1dd-9db4-4ff6-9dc5-a69e8da1d2a6",
		Name:         "Rooting Moloch",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Cycling("{2}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Rooting Moloch — exile target card with a cycling ability from your graveyard; until the end of your next turn you may play it",
					gapB158ExileTargetPlayUntilEndOfYourNextTurn("Rooting Moloch")),
				TargetCardInGraveyard("target card with a cycling ability from your graveyard", YouOwn(), gapB158HasCyclingAbility)),
		},
	})
}
