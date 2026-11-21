package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Windshaper Planetar — {4}{W} Creature — Angel 4/4:
//
//	"Flash
//	 Flying
//	 When this creature enters during the declare attackers step, for
//	 each attacking creature, you may reselect which player or
//	 permanent that creature is attacking. (It can't attack its
//	 controller or their permanents.)"
//
// The untargeted, every-attacker form of CR 508.7 (#1329):
// reselectEachAttackerTrigger asks one "you may" per attacking
// creature, chained, so each attacker is its own decision. Untargeted,
// so a hexproof attacker is redirected too, and a table with no
// attackers still puts the trigger on the stack (it then asks nothing).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c47050c2-9b46-427e-85d6-058fd6a61e67",
		Name:            "Windshaper Planetar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.ReselectEachAttackerTrigger("Windshaper Planetar — reselect what each attacking creature is attacking"),
		},
	})
}
