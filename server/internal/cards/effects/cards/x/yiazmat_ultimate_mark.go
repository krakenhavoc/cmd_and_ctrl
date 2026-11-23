package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yiazmat, Ultimate Mark — the back face of Sidequest: Hunt the Mark,
// Legendary Creature — Dragon, 5/6 (Final Fantasy, #4462):
//
//	"{1}{B}, Sacrifice another creature or artifact: Yiazmat gains
//	 indestructible until end of turn. Tap it."
//
// Immersturm Predator's outlet with a mana payment and a creature-or-artifact
// sacrifice (Acolyte of Aclazotz's cost). The grant and the tap are one
// resolution; the tap happens whether or not Yiazmat has indestructible, and
// a Yiazmat that has left does nothing. Registered under "<oracle_id>#1".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     effects.SidequestHuntTheMarkOracleID + "#1",
		Name:         "Yiazmat, Ultimate Mark",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice another creature or artifact: Yiazmat gains indestructible until end of turn. Tap it.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost: effects.Plus(effects.ManaCost("{1}{B}"), game.AbilityCost{
				SacrificeOther: effects.Another(effects.SacrificeSpec("another creature or artifact", effects.Or(effects.Creature(), effects.Artifact()))),
			}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"indestructible"},
					Label:    "Yiazmat, Ultimate Mark — indestructible until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
				return effects.TapTarget{Target: item.SourceCardID}.Apply(ctx)
			},
		}},
	})
}
