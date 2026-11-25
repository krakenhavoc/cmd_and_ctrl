package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Syndicate Trafficker — Creature — Aetherborn Rogue {1}{B}, 3/1 (Kaladesh,
// #5539):
//
//	"{1}, Sacrifice an artifact: Put a +1/+1 counter on this creature. It
//	 gains indestructible until end of turn."
//
// Barrage Ogre's artifact sacrifice paid as a cost (CR 602.2b) beside mana,
// then Immersturm Predator's counter on the source (skipped if it has left or
// come back as a new object, CR 400.7) and its indestructible grant. The
// Trafficker may pay with itself only if it is an artifact, which it isn't.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2eb8397a-ba8b-497d-b1f5-a1b09c686597",
		Name:         "Syndicate Trafficker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice an artifact: Put a +1/+1 counter on this creature. It gains indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeN(1, "an artifact", effects.Artifact())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := effects.PlusOneCounterOnThisIfStillHere(ctx); err != nil {
					return err
				}
				return effects.ThisGainsKeywordUntilEndOfTurn("indestructible", "Syndicate Trafficker — indestructible until end of turn")(g, item)
			},
		}},
	})
}
