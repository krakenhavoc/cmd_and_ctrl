package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chimil, the Inner Sun — Legendary Artifact {6}:
//
//	"Spells you control can't be countered.
//	 At the beginning of your end step, discover 5."
//
// The discover is ADR 0099's (game/discover.go).
//
// Declared simplification, weaker than printed (owner decision 5 of
// ADR 0099): "spells you control can't be countered" is a continuous
// grant over the STACK, and the engine's can't-be-countered check reads
// only a spell's own printed rider and the mana that paid for it
// (game/cant_be_countered.go). Nothing reaches other spells, so Chimil
// ships without its first line and says so.
func init() {
	Register(Spec{
		OracleID:     "119d9671-61ba-4629-89b1-f94bdec5cb74",
		Name:         "Chimil, the Inner Sun",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Chimil doesn't stop your spells from being countered."},
		Discovers:    true,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Chimil, the Inner Sun — discover 5", DiscoverN(5)),
		},
	})
}
