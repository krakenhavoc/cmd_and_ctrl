package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Foul Orchard — Land (Shadows over Innistrad, #5543):
//
//	"This land enters tapped.
//	 {T}: Add {B} or {G}."
//
// The plain tapped dual, the Guildgates' shape without the Gate type:
// SelfEntersTapped and dualManaAbility (guildgates.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "ad6a2776-801f-4743-8268-6d654122171e",
		Name:          "Foul Orchard",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("B", "G")},
	})
}
