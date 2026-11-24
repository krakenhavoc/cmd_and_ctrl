package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bannerhide Krushok — Creature — Beast {3}{G}, 4/4 (Modern Horizons 2,
// #5501):
//
//	"Trample
//	 Reinforce 2—{1}{G} ({1}{G}, Discard this card: Put two +1/+1 counters
//	 on target creature.)
//	 Scavenge {5}{G}{G} ({5}{G}{G}, Exile this card from your graveyard: Put
//	 a number of +1/+1 counters equal to this card's power on target
//	 creature. Scavenge only as a sorcery.)"
//
// Reinforce is Mosquito Guard's hand ability and scavenge is the keyword
// (scavenge.go); both are written once and work only from their zone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e3191093-1aaa-4002-a603-7c8d132b6a31",
		Name:            "Bannerhide Krushok",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{
			effects.MorReinforce(2, "{1}{G}", "Reinforce 2—{1}{G} ({1}{G}, Discard this card: Put two +1/+1 counters on target creature.)"),
			effects.Scavenge("{5}{G}{G}"),
		},
	})
}
