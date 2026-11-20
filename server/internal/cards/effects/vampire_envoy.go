package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vampire Envoy — Creature — Vampire Cleric Ally {2}{B}, 1/4 (Oath of the Gatewatch, #5545):
//
//	"Flying
//	 Whenever this creature becomes tapped, you gain 1 life."
//
// Chrome Companion's trigger: any tap of it, an attack's included, gains
// its controller 1 life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cfc50aef-b17b-4fc1-af07-fe43b89689b7",
		Name:            "Vampire Envoy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Vampire Envoy — you gain 1 life", Do(GainLife{Amount: 1})),
		},
	})
}
