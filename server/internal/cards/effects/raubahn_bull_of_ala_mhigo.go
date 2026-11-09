package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raubahn, Bull of Ala Mhigo — Legendary Creature — Human Warrior {1}{R},
// 2/2:
//
//	"Ward—Pay life equal to Raubahn's power.
//	 Whenever Raubahn attacks, attach up to one target Equipment you
//	 control to target attacking creature."
//
// The ward is WardLifeEqualToPower: the life is read as the ward resolves,
// from Raubahn's power then or its last-known power (the printed ruling).
// The attack trigger is Barret's two-clause attach with the second clause
// narrowed to an attacking creature; each clause is re-checked on its own
// at resolution, and an Equipment that cannot be attached to that creature
// does not move (CR 701.3b).
func init() {
	Register(Spec{
		OracleID:     "d81d5749-ab4d-42c5-8903-0b400f10fe33",
		Name:         "Raubahn, Bull of Ala Mhigo",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardLifeEqualToPower(), "Raubahn, Bull of Ala Mhigo — ward, pay life equal to Raubahn's power"),
			Targeting(
				WheneverThisAttacks("Raubahn, Bull of Ala Mhigo — attach up to one target Equipment you control to target attacking creature",
					AttachClauseTargets),
				TwoSlotAttachTargets(
					"up to one target Equipment you control", []CardPredicate{YouControl()},
					"target attacking creature", []CardPredicate{b41Attacking()},
				).WithCount(0, 1)),
		},
	})
}
