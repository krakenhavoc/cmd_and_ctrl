package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viashino Skeleton — Creature — Lizard Skeleton {3}{R}, 2/1 (Shards of
// Alara, S82 OLD batch 510, #5588):
//
//	"{1}{B}, Discard a card: Regenerate this creature."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a85e893c-6027-42eb-8d05-14c745cc8c95",
		Name:         "Viashino Skeleton",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.S82OldB510RegenerateRow("{1}{B}, Discard a card: Regenerate this creature.",
				effects.Plus(effects.ManaCost("{1}{B}"), effects.DiscardACard())),
		},
	})
}
