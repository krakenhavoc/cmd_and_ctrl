package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dauthi Mindripper — {3}{B} Creature — Dauthi Minion, 2/1 (Tempest, S82 OLD batch
// 770, #5628):
//
//	"Shadow
//	 Whenever this creature attacks and isn't blocked, you may sacrifice it. If
//	 you do, defending player discards three cards."
//
// Shadow is a printed keyword. The trigger fires once blockers are declared and
// nothing blocked this attacker (WhenAttacksAndIsNotBlockedEffect); the defending
// player is the one the trigger's declaration named. "You may" is asked as it
// resolves. The discard is the "if you do" half: it is told whether the sacrifice
// really happened (SacrificePermanent.ThenTail), so a Mindripper that is already
// gone makes no one discard. The defending player chooses the three cards (CR
// 701.9a), or discards what they have.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ecae5eba-8407-4c71-aeae-ab10fb1e5226",
		Name:            "Dauthi Mindripper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Triggered:       []game.TriggeredAbility{effects.S82OldB770MindripperTrigger()},
	})
}
