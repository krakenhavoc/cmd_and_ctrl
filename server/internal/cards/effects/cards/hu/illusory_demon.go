package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Illusory Demon — Creature — Demon Illusion {1}{U}{B}, 4/3 (Alara Reborn,
// #5586):
//
//	"Flying
//	 When you cast a spell, sacrifice this creature."
//
// Fires for every spell its controller casts, as the spell is cast, and
// sacrifices the Demon only while it is still the permanent that triggered
// (CR 400.7). The spell it triggered on is not affected.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "91d0a557-7a80-46d5-b823-392a84a6be44",
		Name:            "Illusory Demon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(nil, "Illusory Demon — sacrifice this creature", func(g *game.Game, item *game.StackItem) error {
				return effects.SacrificeThisIfStillHere(effects.NewContext(g, item))
			}),
		},
	})
}
