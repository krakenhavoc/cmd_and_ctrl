package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angel of Sanctions — Creature — Angel {3}{W}{W}, 3/4:
//
//	"Flying
//	 When this creature enters, you may exile target nonland permanent an
//	 opponent controls until this creature leaves the battlefield.
//	 Embalm {5}{W}"
//
// Oblivion Ring's one-shot (exileChosenTargetUntilThisLeaves, CR 610.3): the
// permanent returns as a new object under its owner's control when the Angel
// leaves, and if the Angel has left before the trigger resolves nothing is
// exiled (CR 610.3b). The "may" is answered as the trigger resolves (CR
// 603.5). Embalm (CR 702.128a) makes a white Zombie Angel token copy with no
// mana cost whose own enters trigger exiles another permanent.
//
// No simplification.
func init() {
	const label = "Angel of Sanctions — you may exile target nonland permanent an opponent controls until this creature leaves the battlefield"
	Register(Spec{
		OracleID:        "3c49e98d-d33b-47fa-8b07-b38b367de2c7",
		Name:            "Angel of Sanctions",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches:        []game.EventKind{game.EventETB},
			AppliesTo:      Self,
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Angel of Sanctions — exile target nonland permanent an opponent controls?"},
			Targets:        TargetPermanent("target nonland permanent an opponent controls", Nonland(), OpponentControls()),
			Key:            label,
			Effect:         exileChosenTargetUntilThisLeaves("Angel of Sanctions — the exiled permanent returns when Angel of Sanctions leaves the battlefield"),
		}},
		Activated: []ActivatedAbility{Embalm("{5}{W}")},
	})
}
