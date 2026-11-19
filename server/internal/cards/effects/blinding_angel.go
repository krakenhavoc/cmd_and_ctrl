package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blinding Angel — Creature — Angel {3}{W}{W}, 2/4 (Nemesis):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, that player
//	 skips their next combat phase."
//
// The damaged player is read off the damage event and skips one combat
// phase, the next one of theirs that would begin, usually the combat of
// their next turn (DamagedPlayerSkipsNextCombatPhase). A turn with an
// additional combat phase loses only the first, and two hits skip two.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cb479ce1-59e8-4531-9a21-f7c97d334817",
		Name:            "Blinding Angel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Blinding Angel — that player skips their next combat phase",
				DamagedPlayerSkipsNextCombatPhase),
		},
	})
}
