package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whitewater Naiads — Enchantment Creature — Nymph {3}{U}{U}, 4/4 (Journey into Nyx, #5558):
//
//	"Constellation — Whenever this creature or another enchantment you
//	 control enters, target creature can't be blocked this turn."
//
// Constellation is Eidolon of Blossoms' condition: its own entry (it is an
// enchantment) or another enchantment entering under your control. Neurok
// Invisimancer's unblockability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6c73e533-9ed1-4df0-af12-c0174c0b52b9",
		Name:         "Whitewater Naiads",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB, constellationTriggers, "Whitewater Naiads — target creature can't be blocked this turn (constellation)",
				targetCantBeBlockedThisTurn("Whitewater Naiads")), TargetCreature("target creature")),
		},
	})
}
