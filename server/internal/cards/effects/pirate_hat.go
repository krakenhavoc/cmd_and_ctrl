package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pirate Hat — Artifact — Equipment {1}{U} (The Lost Caverns of Ixalan,
// #4470):
//
//	"Equipped creature gets +1/+1 and has "Whenever this creature
//	 attacks, draw a card, then discard a card."
//	 Equip Pirate {1}
//	 Equip {2}"
//
// The attack trigger is a granted ability (ADR 0093), White Mage's
// Staff's shape: it belongs to the equipped creature, goes away with the
// Equipment, and the loot is its controller's (b16DrawThenDiscard, the
// discard chosen after the draw). "Equip Pirate {1}" is Gemcutter
// Buccaneer's restricted equip (CR 702.6c), "Equip {2}" the ordinary one.
//
// No simplification.
func init() {
	const grantKey = "pirate-hat/attack-loot"
	Register(Spec{
		OracleID:     "b9080653-87cb-4443-a4a3-36637d5ad165",
		Name:         "Pirate Hat",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: grantKey,
			Triggered: []game.TriggeredAbility{
				WheneverThisAttacks("Pirate Hat — draw a card, then discard a card",
					func(g *game.Game, item *game.StackItem) error { return b16DrawThenDiscard(g, item, 1) }),
			},
			Text: "Whenever this creature attacks, draw a card, then discard a card.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantAbilitiesToAttached(grantKey),
		},
		Activated: []ActivatedAbility{
			EquipOnlyAbility("Equip Pirate {1}", "{1}", TargetCreature("target Pirate you control", YouControl(), Subtype("Pirate"))),
			EquipAbility("{2}"),
		},
	})
}
