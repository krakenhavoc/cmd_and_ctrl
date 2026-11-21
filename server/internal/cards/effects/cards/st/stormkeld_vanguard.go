package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stormkeld Vanguard // Bear Down — Creature — Giant Warrior {4}{G}{G},
// 6/7 // Sorcery — Adventure {1}{G} (Wilds of Eldraine, #4471):
//
//	Stormkeld Vanguard — "This creature can't be blocked by creatures
//	 with power 2 or less."
//	Bear Down — "Destroy target artifact or enchantment. (Then exile this
//	 card. You may cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). The block restriction is Bristlebane Outrider's, judged as
// blockers are declared on each would-be blocker's current power (CR
// 509.1b). Bear Down is Abolish's target and destroy.
//
// No simplification.
const StormkeldVanguardOracleID = "8a31c49f-5b3f-4f84-a08e-58b30a6bff7c"

func init() {
	effects.Register(effects.Spec{
		OracleID:     StormkeldVanguardOracleID,
		Name:         "Stormkeld Vanguard",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerLE(2), "creatures with power 2 or less"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     StormkeldVanguardOracleID + "#1",
		Name:         "Bear Down",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
