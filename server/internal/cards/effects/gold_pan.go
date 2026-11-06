package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gold Pan — Artifact — Equipment {2} (Outlaws of Thunder Junction,
// #4468):
//
//	"When this Equipment enters, create a Treasure token. (It's an
//	 artifact with '{T}, Sacrifice this token: Add one mana of any
//	 color.')
//	 Equipped creature gets +1/+1.
//	 Equip {1} ({1}: Attach to target creature you control. Equip only
//	 as a sorcery.)"
//
// An enters trigger making the catalog Treasure (CR 111.10a), the
// layer 7c attached pump, and equip (CR 702.6a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "91415d12-b13d-4b5c-bb60-6ec6f4fdde6e",
		Name:         "Gold Pan",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Gold Pan — create a Treasure token", Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{1}"),
		},
	})
}
