package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rush of Battle — Khans of Tarkir (S82 OLD b348, #5554):
//
//	Sorcery {3}{W}: Creatures you control get +2/+1 until end of turn. Warrior
//	creatures you control gain lifelink until end of turn.
//
// Both sets are fixed as the spell resolves (CR 611.2c).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e203a201-2567-4684-ae9f-fd4e91c6dc37",
		Name:         "Rush of Battle",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2, Toughness: 1, Label: "Rush of Battle — +2/+1"},
			effects.GrantKeywordUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl(), effects.HasSubtype("Warrior")), Keywords: []string{"lifelink"},
				Label: "Rush of Battle — Warriors gain lifelink"},
		),
	})
}
