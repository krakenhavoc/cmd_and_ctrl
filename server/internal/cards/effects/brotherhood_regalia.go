package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brotherhood Regalia — Artifact — Equipment {2}:
//
//	"Equipped creature has ward {2}, is an Assassin in addition to its
//	 other types, and can't be blocked.
//	 Equip legendary creature {1}
//	 Equip {3}"
//
// Three granted clauses, three different mechanisms, none of them
// new:
//
//   - Ward is a TRIGGERED ability (CR 702.21a), not a keyword, so it
//     is granted the way Lavaspur Boots grants haste-plus-ward:
//     WardAttached puts the trigger on the EQUIPMENT, watching for its
//     current host becoming a target.
//   - "Is an Assassin in addition to its other types" is a Layer 4
//     type ADD (not a replace — SetAttachedTypes is for "is a Forest",
//     this card never says "is"). No existing helper adds exactly one
//     subtype to an attached permanent, so it is one inline
//     StaticAbility literal, the same shape Coercive Recruiter's
//     "becomes a Pirate in addition to its other types" uses.
//   - "Can't be blocked" is the CR 509.1b restriction
//     RestrictAttached(game.CantBeBlocked) already carries for Aether
//     Tunnel.
//
// TWO EQUIP ABILITIES, the Blackblade Reforged shape: the cheap one
// narrowed to Legendary() creatures, the plain one open to any.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d169738e-af0b-4488-a399-3aa1d1933aa1",
		Name:         "Brotherhood Regalia",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Layer:     game.Layer4Type,
				AppliesTo: AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !eotHasType(c.Subtypes, "Assassin") {
						c.Subtypes = append(c.Subtypes, "Assassin")
					}
				},
			},
			RestrictAttached(game.CantBeBlocked),
		},
		Triggered: []game.TriggeredAbility{
			WardAttached(WardMana("{2}"), "Brotherhood Regalia — ward {2}"),
		},
		Activated: []ActivatedAbility{
			EquipOnlyAbility("Equip legendary creature {1}", "{1}",
				TargetCreature("target legendary creature you control", YouControl(), Legendary())),
			EquipAbility("{3}"),
		},
	})
}
