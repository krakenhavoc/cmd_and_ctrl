package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ace's Baseball Bat — Legendary Artifact — Equipment {2}:
//
//	"Equipped creature gets +3/+0.
//	 As long as equipped creature is attacking, it has first strike and
//	 must be blocked by a Dalek if able.
//	 Equip legendary creature {1}
//	 Equip {3}"
//
// The Masamune's shape: first strike and the block requirement both read
// live combat state, so they apply only while the host has been declared as
// an attacker (DependsOnAttackingStatus wakes the layer cache as attackers
// are declared and again when combat ends). The requirement is "must be
// blocked if able" narrowed to Dalek blockers (#6087,
// MustBeBlockedBySubtypeWhere): the defending player must block it with a
// Dalek if one can (CR 509.1c), and any other creature may still block it.
//
// Two equip abilities, the Brotherhood Regalia shape: the cheap one narrowed
// to legendary creatures, the plain one open to any.
//
// No simplification.
func init() {
	attacking := func(host *game.Card) bool { return host.AttackingTarget != uuid.Nil }
	firstStrike := effects.GrantToAttachedWhile(func(host *game.Card, _ *game.Game, _ *game.Card) bool {
		return attacking(host)
	}, "first strike")
	firstStrike.Label = "As long as equipped creature is attacking, it has first strike."
	firstStrike.DependsOnAttackingStatus = true
	dalek := effects.MustBeBlockedBySubtypeWhere("Dalek", func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.AttachedToSource(target, g, source) && attacking(target)
	})
	dalek.Label = "As long as equipped creature is attacking, it must be blocked by a Dalek if able."
	dalek.DependsOnAttackingStatus = true

	effects.Register(effects.Spec{
		OracleID:     "3f47ec32-4ee9-4bc0-96f2-310a939cb973",
		Name:         "Ace's Baseball Bat",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 0),
			firstStrike,
			dalek,
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipOnlyAbility("Equip legendary creature {1}", "{1}",
				effects.TargetCreature("target legendary creature you control", effects.YouControl(), effects.Legendary())),
			effects.EquipAbility("{3}"),
		},
	})
}
