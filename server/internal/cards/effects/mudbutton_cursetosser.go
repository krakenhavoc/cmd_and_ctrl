package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mudbutton Cursetosser — Creature — Goblin Warlock {B}, 2/1:
//
//	"As an additional cost to cast this spell, behold a Goblin or pay {2}. (To behold a Goblin, choose a Goblin you control or reveal a Goblin card from your hand.)
//	 This creature can't block.
//	 When this creature dies, destroy target creature an opponent controls with power 2 or less."
//
// The additional cost is the either/or branch cost of ADR 0100 §2
// with the behold branch added by its 2026-10-07 amendment
// (BeholdOrPay): the caster announces the branch, names the
// card on reveal_ids, and either shows it to the table or pays {2}
// more.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "b63e5518-61f3-4cde-a799-5f3ee6aba70b",
		Name:           "Mudbutton Cursetosser",
		Completeness:   CompletenessFull,
		AdditionalCost: BeholdOrPay("a", "Goblin", "{2}"),
		Static: []game.StaticAbility{
			RestrictSelf(game.CantBlock),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Mudbutton Cursetosser — destroy target creature an opponent controls with power 2 or less", destroyFirstLegalTarget), TargetCreature("target creature an opponent controls with power 2 or less", OpponentControls(), PowerLE(2))),
		},
	})
}
