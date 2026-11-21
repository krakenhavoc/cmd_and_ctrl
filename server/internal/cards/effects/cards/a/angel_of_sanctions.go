package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "3c49e98d-d33b-47fa-8b07-b38b367de2c7",
		Name:            "Angel of Sanctions",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches:        []game.EventKind{game.EventETB},
			AppliesTo:      effects.Self,
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Angel of Sanctions — exile target nonland permanent an opponent controls?"},
			Targets:        effects.TargetPermanent("target nonland permanent an opponent controls", effects.Nonland(), effects.OpponentControls()),
			Key:            label,
			Effect:         effects.ExileChosenTargetUntilThisLeaves("Angel of Sanctions — the exiled permanent returns when Angel of Sanctions leaves the battlefield"),
		}},
		Activated: []effects.ActivatedAbility{effects.Embalm("{5}{W}")},
	})
}
