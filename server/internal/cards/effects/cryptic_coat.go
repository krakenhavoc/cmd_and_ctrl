package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cryptic Coat — Artifact — Equipment {2}{U} (Murders at Karlov Manor,
// #4469):
//
//	"When this Equipment enters, cloak the top card of your library, then
//	 attach this Equipment to it.
//	 Equipped creature gets +1/+0 and can't be blocked.
//	 {1}{U}: Return this Equipment to its owner's hand."
//
// The cloak is cloak_action.go (a face-down 2/2 with ward {2}); the
// attach is Armored Skyhunter's, skipped quietly when the Coat has left
// (CR 701.3b) while the cloak still happens (ruling). There is no equip
// ability: the Coat only attaches through its own trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "73b8adea-1786-44fc-b190-bf240b07f5b6",
		Name:         "Cryptic Coat",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 0),
			RestrictAttached(game.CantBeBlocked),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Cryptic Coat — cloak the top card of your library, then attach this Equipment to it",
				func(g *game.Game, item *game.StackItem) error {
					id, err := CloakTopOfLibrary{}.Do(NewContext(g, item))
					if err != nil || id == uuid.Nil {
						return err
					}
					return g.AttachSourceForEffect(item, game.TargetRef{Kind: game.TargetCard, ID: id})
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}: Return this Equipment to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{1}{U}"),
			Effect:  returnThisPermanentToOwnersHand,
		}},
	})
}
