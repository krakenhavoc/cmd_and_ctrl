package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Giant Inheritance — Enchantment — Aura {4}{G}:
//
//	"Enchant creature
//	 Enchanted creature gets +5/+5 and has "Whenever this creature
//	 attacks, create a Monster Role token attached to up to one target
//	 attacking creature." (Enchanted creature gets +1/+1 and has
//	 trample.)
//	 When this Aura is put into a graveyard from the battlefield,
//	 return it to its owner's hand."
//
// Rancor's skeleton (an Aura with a recursion trigger) plus a granted
// triggered ability. The granted trigger is the CREATURE's, so "this
// creature" is the host and the Role is controlled by the host's
// controller (ADR 0093 Decision 4). Its target clause is "up to one
// target attacking creature" — any attacker, the host included, and
// none is a legal choice. The Monster Role is made while the attack is
// declared, so a creature that stops attacking before the trigger
// resolves is no longer a legal target (CR 608.2b).
//
// No simplifications.
const giantInheritanceGrant = "giant-inheritance/monster-role"

func init() {
	Register(Spec{
		OracleID:     "6be21185-f85c-49e8-988a-a76b6613fdaa",
		Name:         "Giant Inheritance",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(5, 5),
			GrantAbilitiesToAttached(giantInheritanceGrant),
		},
		Grants: []AbilityGrant{{
			Key: giantInheritanceGrant,
			Triggered: []game.TriggeredAbility{
				Targeting(
					On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
						return attackDeclared(ev, source)
					}, "Giant Inheritance — create a Monster Role token attached to up to one target attacking creature",
						createRoleOnFirstTarget(RoleMonster)),
					TargetCreature("up to one target attacking creature", AttackingCreature()).WithCount(0, 1)),
			},
			Text: "Whenever this creature attacks, create a Monster Role token attached to up to one target attacking creature.",
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			Key: "Giant Inheritance — return it to your hand",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return ReturnFromGraveyard{Target: item.SourceCardID, Dest: game.ZoneHand}.Apply(NewContext(g, item))
			},
		}},
	})
}
