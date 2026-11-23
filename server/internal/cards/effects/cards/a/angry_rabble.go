package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angry Rabble — Creature — Human Citizen {1}{R}, 2/2 (Marvel's
// Spider-Man):
//
//	"Trample
//	 Whenever you cast a spell with mana value 4 or greater, this creature
//	 deals 1 damage to each opponent.
//	 {5}{R}: Put two +1/+1 counters on this creature. Activate only as a
//	 sorcery."
//
// The cast trigger reads the spell's mana value off the stack (CR 202.3),
// so X counts at its announced value. The pump is Nakia's sorcery-speed
// two-counter ability pointed at itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cf6dc79e-840c-4d47-8f94-c8623528acd2",
		Name:            "Angry Rabble",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.ManaValueGE(4), "Angry Rabble — 1 damage to each opponent", effects.Do(effects.DamageEachOpponent{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{5}{R}: Put two +1/+1 counters on this creature. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{5}{R}"),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 2}),
		}},
	})
}
