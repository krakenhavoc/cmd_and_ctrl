package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "2844d274-cb60-46c1-a73b-e763c633e496",
		Name:            "Alchemist's Assistant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated: []ActivatedAbility{
			Renew("{1}{B}", "Put a lifelink counter on target creature.",
				TargetCreature("target creature"), renewCountersOnTarget(game.CounterLifelink, 1)),
		},
	})
}
