package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Defiler of Souls — Creature — Demon {3}{B}{B}{R}, 5/5 (Alara Reborn, #5586):
//
//	"Flying
//	 At the beginning of each player's upkeep, that player sacrifices a
//	 monocolored creature of their choice."
//
// Every player's upkeep, the Defiler's controller's included. The player whose
// upkeep it is chooses among their own monocolored creatures (exactly one colour
// as the layers say, colourless excluded, CR 105.2a); nothing happens with none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "34894b92-7ebd-49be-9a67-7dc6428aaab9",
		Name:            "Defiler of Souls",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Defiler of Souls — that player sacrifices a monocolored creature of their choice",
				effects.S82OldB504ActivePlayerSacrificesAMonocoloredCreature),
		},
	})
}
