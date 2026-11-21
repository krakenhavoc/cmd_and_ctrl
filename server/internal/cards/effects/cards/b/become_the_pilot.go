package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Become the Pilot — Enchantment — Aura {3}{U}{U} (Doctor Who):
//
//	"Enchant noncommander creature
//	 You control enchanted creature.
//	 Enchanted creature gets +2/+2 and can't be blocked unless it's
//	 attacking its owner or a permanent its owner controls."
//
// The control change is Control Magic's layer-2 effect, the pump is a layer
// 7c modify, and the evasion is a block rule on the enchanted creature that
// holds unless it attacks its OWNER or a permanent its owner controls (a
// planeswalker or battle, read off the attack target). A creature that is
// not attacking is never refused: the rule is about blocks only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e079dd3-3397-491f-a6c5-7be5ade687d3",
		Name:         "Become the Pilot",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.Not(effects.ACommander())),
		Static: []game.StaticAbility{
			effects.ControlAttachedBySource(),
			effects.PumpAttached(2, 2),
		},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedWhile(effects.OnAttached(), becomeThePilotNotAttackingItsOwner),
		},
	})
}

// becomeThePilotNotAttackingItsOwner holds while the enchanted creature is
// attacking something that is not its owner nor a permanent its owner
// controls.
func becomeThePilotNotAttackingItsOwner(g *game.Game, _ uuid.UUID, attacker game.Card) bool {
	target := attacker.AttackingTarget
	if target == uuid.Nil || target == attacker.Owner {
		return false
	}
	if p, ok := g.LookupCardForEffect(target); ok && p.Controller == attacker.Owner {
		return false
	}
	return true
}
