package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Darksteel Plate — Legendary Artifact — Equipment for {3} (EDHREC
// rank 918):
//
//	"Indestructible
//	 Equipped creature has indestructible.
//	 Equip {2}"
//
// ADR 0036 decision 10 named this card as one the first attachment
// batch had to cut: "indestructible is not in the engine's
// honoured-keyword table, so the grant would be a silent no-op."
// S25 (#77) put indestructible in the table and gave it a consumer in
// the destruction path, so the cut no longer applies and the card is
// two `GrantToAttached`-shaped lines.
//
// The Plate's own indestructible is a PRINTED keyword rather than a
// static, which is the distinction Spec.PrintedKeywords exists for:
// it is true of the Equipment in every zone the keyword can matter
// in, it survives the Plate being unattached, and it means a board
// wipe that kills the creature leaves the Plate to be re-equipped.
//
// What indestructible does NOT stop is on the engine's side and
// documented there (indestructible.go): sacrifice, exile, -X/-X to
// zero toughness, and legend-rule and state-based-action deaths all
// still work. So a Plated commander is not immortal, and the Plate
// itself still dies to Vandalblast.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b5b4cf54-ed5e-42d0-9d98-5fec76b0b0b8",
		Name:            "Darksteel Plate",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static: []game.StaticAbility{
			GrantToAttached("indestructible"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
