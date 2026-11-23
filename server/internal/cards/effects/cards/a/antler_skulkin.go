package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Antler Skulkin — Artifact Creature — Scarecrow {5}, 3/3:
//
//	"{2}: Target white creature gains persist until end of turn."
//
// "White" is the creature's colour as the ability is put on the stack
// and again as it resolves (CR 608.2b). Persist is the engine's
// (#2075).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "30572dd1-c982-4872-9f35-a91ed8b238f7",
		Name:         "Antler Skulkin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Target white creature gains persist until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "persist")),
			Cost:    effects.ManaCost("{2}"),
			Targets: effects.TargetCreature("target white creature", effects.OfColor("W")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantEachLegalTargetUntilEOT(effects.NewContext(g, item), game.KeywordPersist,
					"Antler Skulkin — persist until end of turn")
			},
		}},
	})
}
