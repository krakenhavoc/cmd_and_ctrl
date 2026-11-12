package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thrashing Frontliner — Creature — Phyrexian Lizard {1}{R}, 2/2 (March
// of the Machine, #5649):
//
//	"Trample
//	 Whenever this creature attacks a battle, it gets +1/+1 until end of
//	 turn."
//
// Triggers when it is declared as an attacker attacking a battle (CR
// 508.3a), not when it attacks a player or a planeswalker. The pump is
// for this object only: nothing if it has left the battlefield by the
// time the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e75f1834-10b8-4362-adbe-847ad9909dee",
		Name:            "Thrashing Frontliner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacksABattle("Thrashing Frontliner — it gets +1/+1 until end of turn",
				thisGetsUntilEndOfTurn(1, 1, "Thrashing Frontliner — +1/+1")),
		},
	})
}
