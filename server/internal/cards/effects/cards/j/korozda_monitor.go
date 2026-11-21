package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Korozda Monitor — Creature — Lizard {2}{G}{G}, 3/3 (Return to Ravnica, #5566):
//
//	"Trample / Scavenge {5}{G}{G} ({5}{G}{G}, Exile this card from
//	 your graveyard: Put a number of +1/+1 counters equal to this
//	 card's power on target creature. Scavenge only as a sorcery.)"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
// Scavenge is the shared graveyard ability (Deadbridge Goliath,
// scavenge.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "55195474-c7ea-4c0a-b504-af15e64ec7e9",
		Name:            "Korozda Monitor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated:       []effects.ActivatedAbility{effects.Scavenge("{5}{G}{G}")},
	})
}
