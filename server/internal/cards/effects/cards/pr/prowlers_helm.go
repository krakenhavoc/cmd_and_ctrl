package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prowler's Helm — Artifact — Equipment, {2}:
//
//	"Equipped creature can't be blocked except by Walls.
//	 Equip {2}"
//
// ADR 0045's reference card for a catalog PAIR rule on an attached
// creature (#750, addendum Decision 11): CantBeBlockedExceptBy on
// OnAttached. The rule belongs to the Helm, so it follows the Helm —
// re-equipping moves it, and the equipped creature losing its
// abilities does not remove it (the Helm's ability is the Helm's).
// "Walls" is an effective creature type, so a changeling can block.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "603ef53f-6191-4229-b85f-adad6f4503ce",
		Name:         "Prowler's Helm",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedExceptBy(effects.OnAttached(), effects.OfCreatureType("Wall"), "Walls"),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{2}"),
		},
	})
}
