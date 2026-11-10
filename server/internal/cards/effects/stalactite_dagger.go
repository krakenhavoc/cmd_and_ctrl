package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stalactite Dagger — Artifact — Equipment {2} (Lorwyn Eclipsed, #4458):
//
//	"When this Equipment enters, create a 1/1 colorless Shapeshifter
//	 creature token with changeling. (It's every creature type.)
//	 Equipped creature gets +1/+1 and is all creature types.
//	 Equip {2}"
//
// The token is the printed tecl Shapeshifter row Crib Swap makes. "Is
// all creature types" is Maskwood Nexus' layer 4 fact
// (applyAllCreatureTypes) on the equipped creature, not a changeling
// keyword in layer 6, so a lord's keyword half sees it whatever entered
// first, and a later layer 4 subtype set (an Aura making it an Elk)
// overwrites it by timestamp.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "064c3269-c012-4fc7-84c3-5f147a82a03c",
		Name:         "Stalactite Dagger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Stalactite Dagger — create a 1/1 colorless Shapeshifter creature token with changeling",
				Do(CreateToken{Template: TokenCard("1/1 colorless Shapeshifter with changeling"), N: 1})),
		},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			{
				Label:     "Equipped creature is all creature types.",
				Layer:     game.Layer4Type,
				AppliesTo: AttachedToSource,
				Apply:     applyAllCreatureTypes,
			},
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
