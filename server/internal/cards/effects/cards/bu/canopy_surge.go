package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Canopy Surge — Sorcery {1}{G} (Invasion, S82 OLD #5616):
//
//	"Kicker {2} (You may pay an additional {2} as you cast this spell.)
//	 Canopy Surge deals 1 damage to each creature with flying and each
//	 player. If this spell was kicked, it deals 4 damage to each creature
//	 with flying and each player instead."
//
// Ifh-Bíff Efreet's damage body (the spell is the source, CR 120.3), its size
// read off the cast. The "instead" replaces the 1, so a kicked cast deals 4
// and never 5. The creatures are read as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "a151c45f-0b9e-41af-90b3-fafa12e3bbce",
		Name:          "Canopy Surge",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{2}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := 1
			if ctx.WasKicked() {
				n = 4
			}
			return effects.ThisDealsDamageToEachCreatureMatchingAndEachPlayer(effects.HasKeyword("flying"), n)(ctx.Game, item)
		},
	})
}
