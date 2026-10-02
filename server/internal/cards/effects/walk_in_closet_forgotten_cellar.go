package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Walk-In Closet // Forgotten Cellar — Enchantment — Room (ADR 0103,
// ADR 0108 §4):
//
//	Walk-In Closet {2}{G}: "You may play lands from your graveyard."
//	Forgotten Cellar {3}{G}{G}: "When you unlock this door, you may
//	 cast spells from your graveyard this turn, and if a card would be
//	 put into your graveyard from anywhere this turn, exile it instead."
//
// Walk-In Closet is Glacierwood Siege's standing land permission over
// the graveyard, gated on its own door. Forgotten Cellar is the
// spells-only half of Yawgmoth's Will's clause pair, run when the door
// unlocks. "You may" governs the casting, and the replacement is not
// optional, so there is nothing to ask.
//
// No simplification.
func init() {
	Register(Room(RoomSpec{
		OracleID:     "52e77cc3-f8e9-4a20-811b-fe1e46a96ad7",
		Name:         "Walk-In Closet // Forgotten Cellar",
		Completeness: CompletenessFull,
		Left: Door{GatedCastPermissions: []game.CastPermissionGate{{
			Permission: game.CastPermission{
				Zone:   game.ZoneGraveyard,
				Filter: game.PermissionFilter{LandsOnly: true},
				Label:  "Play a land from your graveyard (Walk-In Closet)",
			},
		}}},
		Right: Door{Triggered: []game.TriggeredAbility{
			WhenYouUnlockThisDoor(game.DoorRight, "Forgotten Cellar — cast spells from your graveyard this turn; exile instead of your graveyard",
				Do(GraveyardPlayThisTurn{SpellsOnly: true, Label: "Cast a spell from your graveyard (Forgotten Cellar)"})),
		}},
	}))
}
