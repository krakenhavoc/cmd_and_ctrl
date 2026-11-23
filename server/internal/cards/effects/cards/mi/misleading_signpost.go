package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Misleading Signpost — {2}{U} Artifact:
//
//	"Flash
//	 When this artifact enters during the declare attackers step, you
//	 may reselect which player or permanent target attacking creature
//	 is attacking. (It can't attack its controller or their
//	 permanents.)
//	 {T}: Add {U}."
//
// The deck card (#1306) that #1329 was filed for, and the proof card
// for CR 508.7: the trigger is reselectTargetAttackerTrigger
// (attack_reselect.go), which asks through the engine's
// QueueReselectAttackForEffect. The reminder text is CR 508.7c, and
// the engine enforces it against the ATTACKING creature's controller
// — so a defending player flashing this in can push an attack onto
// any other opponent of the attacker, and never back onto the
// attacker's own side.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "eaffdf95-8e20-408d-99fa-3adc9e19523d",
		Name:            "Misleading Signpost",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.ReselectTargetAttackerTrigger("Misleading Signpost — reselect what target attacking creature is attacking"),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
	})
}
