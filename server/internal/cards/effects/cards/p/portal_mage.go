package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Portal Mage — {2}{U} Creature — Human Wizard 2/2:
//
//	"Flash
//	 When this creature enters during the declare attackers step, you
//	 may reselect which player or permanent target attacking creature
//	 is attacking. (It can't attack its controller or their
//	 permanents.)"
//
// Misleading Signpost's trigger on a body — the same
// reselectTargetAttackerTrigger, word for word (CR 508.7, #1329).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4d19ea6d-cabe-4a25-a911-dd744517dd73",
		Name:            "Portal Mage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.ReselectTargetAttackerTrigger("Portal Mage — reselect what target attacking creature is attacking"),
		},
	})
}
