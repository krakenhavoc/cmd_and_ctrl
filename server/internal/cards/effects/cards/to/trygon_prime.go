package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trygon Prime — Creature — Tyranid {2}{G}{U}, 4/4 (Warhammer 40,000
// Commander, S82 GAP batch 193, #5948):
//
//	"Subterranean Assault — Whenever this creature attacks, put a +1/+1 counter
//	 on it and a +1/+1 counter on up to one other target attacking creature.
//	 That creature can't be blocked this turn."
//
// "Up to one other target attacking creature" is Restless Vinestalk's
// optional, other-than-the-source clause narrowed to attackers; the Prime's
// own counter needs it still on the battlefield, and a pick that has left
// combat or the battlefield by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	attack := effects.WheneverThisAttacks("Trygon Prime — a +1/+1 counter on it and on up to one other target attacking creature; that creature can't be blocked this turn",
		effects.GapB193TrygonPrimeAttack)
	attack.Targets = effects.Another(effects.UpToOneTargetCreature("up to one other target attacking creature", effects.AttackingCreature()))
	effects.Register(effects.Spec{
		OracleID:     "6b2d1d7b-a4b4-46c3-a50b-df7e60f39ee2",
		Name:         "Trygon Prime",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{attack},
	})
}
