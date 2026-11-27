package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Corpse Blockade — Creature — Zombie {2}{B}, 1/4 (Gatecrash, #5565):
//
//	"Defender
//	 Sacrifice another creature: This creature gains deathtouch until end of
//	 turn."
//
// The sacrifice is a cost (CR 602.2b) of a creature other than the Blockade
// (Baron Vladimir's "another"), so a second Blockade is a legal payment. The
// grant is Mishra's Foundry's scoped effect on the source: it follows the
// object as the ability resolves and does nothing for a Blockade that left
// in response or came back as a new object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5cb2e859-7737-4b25-a6be-b72e7b5bfada",
		Name:            "Corpse Blockade",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice another creature: This creature gains deathtouch until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost:    effects.SacrificeAnotherN(1, "another creature", effects.Creature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"deathtouch"},
					Label: "Corpse Blockade — gains deathtouch until end of turn"}.Apply(ctx)
			},
		}},
	})
}
