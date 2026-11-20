package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Solar Transformer — Artifact {2}:
//
//	"This artifact enters tapped.
//	 When this artifact enters, you get {E}{E}{E} (three energy
//	 counters).
//	 {T}: Add {C}.
//	 {T}, Pay {E}: Add one mana of any color."
//
// ADR 0129 §5: Aether Hub's two mana rows on an artifact that enters
// tapped. The coloured half is planned in the auto-tapper's energy tier.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cb377c27-39a6-42b4-8c89-80a9a2350e81",
		Name:         "Solar Transformer",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Energy: 3},
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Solar Transformer", 3),
		},
		ManaAbilities: effects.AnyColorForEnergyRows(1),
	})
}
