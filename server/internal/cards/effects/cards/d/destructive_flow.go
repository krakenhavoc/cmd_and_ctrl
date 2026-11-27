package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Destructive Flow — Enchantment {B}{R}{G} (Planeshift, S82 OLD b644, #5615):
//
//	"At the beginning of each player's upkeep, that player sacrifices a nonbasic
//	 land of their choice."
//
// Ruthless Winnower's each-upkeep edict over nonbasic lands: the player whose upkeep
// it is chooses their own land, and one with none sacrifices nothing. A basic land
// that lost the basic supertype counts as nonbasic.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f7544a96-46b3-4cdc-8b14-c7b6daecdabd",
		Name:         "Destructive Flow",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Destructive Flow — that player sacrifices a nonbasic land of their choice",
				effects.OldB644TriggerPlayerSacrifices("Destructive Flow — sacrifice a nonbasic land",
					func(c game.Card) bool { return c.IsLand() && !c.HasSupertype("basic") })),
		},
	})
}
