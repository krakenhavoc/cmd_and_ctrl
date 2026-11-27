package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Light of Sanction — Enchantment {1}{W}{W} (Ravnica: City of Guilds):
//
//	"Prevent all damage that would be dealt to creatures you control by sources
//	 you control."
//
// A prevention static (CR 615) over the Light's controller's creatures, narrowed
// to a source that controller also controls, read off the source as the damage
// would be dealt (its last-known information). A source an opponent controls, or
// damage to a player, is untouched; combat and noncombat damage are both
// prevented, and "you control" is read live, so a stolen source or creature
// changes the answer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c34cf404-729d-46ef-8661-1095e5581766",
		Name:         "Light of Sanction",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.PreventAllDamageDealtTo(effects.PreventionStatic{
				To:    effects.ToCreaturesYouControl,
				From:  effects.FromASourceYouControl(),
				Label: "Light of Sanction — prevent all damage that would be dealt to creatures you control by sources you control",
			}),
		},
	})
}
