package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whirling Catapult — Artifact {4}:
//
//	"{2}, Exile the top two cards of your library: This artifact deals 1
//	 damage to each creature with flying and each player."
//
// ADR 0109 §7 (#1902): "Exile the top N cards of your library" is a
// cost with nothing to choose. A library of fewer than N cards can't pay
// it (CR 118.3), and it is paid after every other cost (CR 601.2h).
// The damage is Squallmonger's: each creature with flying, read as the
// ability resolves, and every player, the activator included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0987b5e9-0012-4189-8745-45f10e9557f3",
		Name:         "Whirling Catapult",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{2}, Exile the top two cards of your library: This artifact deals 1 damage to each creature with flying and each player.",
			// ADR 0126 §6: only creatures with flying.
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.ExileTopOfLibrary(2)),
			Effect:  effects.ThisDealsDamageToEachCreatureMatchingAndEachPlayer(effects.HasKeyword("flying"), 1),
		}},
	})
}
