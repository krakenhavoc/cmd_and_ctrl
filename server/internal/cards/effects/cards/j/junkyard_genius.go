package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Junkyard Genius — Creature — Human Artificer {1}{B}{R}, 2/2 (The
// Brothers' War, #5482):
//
//	"When this creature enters, create a tapped Powerstone token.
//	 {1}{B}{R}, Sacrifice another creature or artifact: Until end of
//	 turn, other creatures you control get +1/+0 and gain menace and
//	 haste."
//
// The activated ability is Earthshaker Giant's until-end-of-turn bundle:
// the creatures are the ones you control when it resolves (CR 611.2c),
// never the Genius itself, and the sacrificed permanent is already gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68d4d753-5e53-48a6-8ba2-56020893c239",
		Name:         "Junkyard Genius",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Junkyard Genius — create a tapped Powerstone token",
				effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.PowerstoneToken()).EntersTapped(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}{R}, Sacrifice another creature or artifact: Until end of turn, other creatures you control get +1/+0 and gain menace and haste.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant | game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}{R}"), effects.SacrificeAnotherN(1, "another creature or artifact", effects.Or(effects.Creature(), effects.Artifact()))),
			Effect:  junkyardGeniusRally,
		}},
	})
}

func junkyardGeniusRally(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	self := ctx.Source()
	match := effects.And(effects.Creature(), effects.YouControl(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID != self })
	if err := (effects.BoostUntilEOT{Match: match, Power: 1, Label: "Junkyard Genius — +1/+0"}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{Match: match, Keywords: []string{"menace", "haste"}, Label: "Junkyard Genius — menace and haste until end of turn"}.Apply(ctx)
}
