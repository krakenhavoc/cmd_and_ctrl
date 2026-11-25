package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Defiant Salvager — Creature — Aetherborn Artificer {2}{B}, 2/2 (Commander
// Legends):
//
//	"Sacrifice an artifact or creature: Put a +1/+1 counter on this
//	 creature. Activate only as a sorcery."
//
// Ravenous Squirrel's sacrifice cost (the Salvager itself is a legal
// choice), at sorcery speed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a8c9ef7-8171-4f6f-bf39-4a08d127e729",
		Name:         "Defiant Salvager",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "Sacrifice an artifact or creature: Put a +1/+1 counter on this creature. Activate only as a sorcery.",
			Cost: game.AbilityCost{
				SacrificeOther: effects.SacrificeSpec("an artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
			},
			SorcerySpeed: true,
			Effect:       effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}
