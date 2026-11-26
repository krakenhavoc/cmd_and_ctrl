package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Qarsi High Priest — {B} Creature — Human Cleric 0/2 (Fate Reforged):
//
//	"{1}{B}, {T}, Sacrifice another creature: Manifest the top card of your
//	 library."
//
// The sacrifice is a cost (Mold Folk's SacrificeAnotherN), so the creature is
// gone before the ability resolves; the manifest takes the top card as it
// resolves (CR 701.40a).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3bcb5ad4-818e-4f10-aaff-188d713b160b",
		Name:         "Qarsi High Priest",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, {T}, Sacrifice another creature: Manifest the top card of your library.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost(),
				effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect: effects.Do(effects.Manifest{N: 1}),
		}},
	})
}
