package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Planar Atlas — Artifact {2} (Jumpstart 2022, #5948):
//
//	"This artifact enters tapped.
//	 When this artifact enters, you may look at the top four cards of your
//	 library. If you do, reveal up to one land card from among them, then
//	 put that card on top of your library and the rest on the bottom in a
//	 random order.
//	 {T}: Add {C}."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "4f14bf38-13ec-461d-8a1b-340843b5d95d",
		Name:          "Planar Atlas",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.PainlessColorless()},
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.WhenThisEnters("Planar Atlas — you may look at the top four cards of your library", effects.GapB144AtlasLook),
				"Planar Atlas — look at the top four cards of your library?"),
		},
	})
}
