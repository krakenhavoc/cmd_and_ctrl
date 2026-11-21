package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alchemist's Assistant — Creature — Monkey {1}{B}, 2/1 (Tarkir:
// Dragonstorm, #4463):
//
//	"Lifelink
//	 Renew — {1}{B}, Exile this card from your graveyard: Put a lifelink
//	 counter on target creature. Activate only as a sorcery."
//
// Lifelink rides PrintedKeywords; renew is renew.go, and the lifelink
// counter is a keyword counter (CR 122.1b), which grants lifelink for as
// long as it stays on the creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2844d274-cb60-46c1-a73b-e763c633e496",
		Name:            "Alchemist's Assistant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated: []effects.ActivatedAbility{
			effects.Renew("{1}{B}", "Put a lifelink counter on target creature.",
				effects.TargetCreature("target creature"), effects.RenewCountersOnTarget(game.CounterLifelink, 1)),
		},
	})
}
