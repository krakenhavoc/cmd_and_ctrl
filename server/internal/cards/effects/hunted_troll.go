package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hunted Troll — Creature — Troll Warrior {2}{G}{G}, 8/4 (Ravnica: City of
// Guilds, #5600):
//
//	"When this creature enters, target opponent creates four 1/1 blue
//	 Faerie creature tokens with flying.
//	 {G}: Regenerate this creature."
//
// Hunted Horror's targeted enters trigger (the opponent creates the
// Faeries, CR 111.2) and River Boa's regeneration row (CR 701.19a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1f789fcf-3df6-45a6-a732-9f43e33718d6",
		Name:         "Hunted Troll",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ravB4TargetOpponentCreates("Hunted Troll — target opponent creates four 1/1 blue Faeries with flying", "1/1 blue Faerie with flying", 4),
		},
		Activated: []ActivatedAbility{regenerateThisRow("{G}")},
	})
}
