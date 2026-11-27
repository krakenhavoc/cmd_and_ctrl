package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Toymaker — Artifact Creature — Spellshaper {2}, 1/1 (Mercadian Masques, S82
// OLD b682, #5619):
//
//	"{1}, {T}, Discard a card: Target noncreature artifact becomes an artifact
//	 creature with power and toughness each equal to its mana value until end
//	 of turn. (It retains its abilities.)"
//
// Xenic Poltergeist's animation (ArtifactBecomesCreatureOfItsManaValue) with
// an until-end-of-turn duration, behind a Spellshaper's mana, tap and discard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eb80f233-76b5-4c72-8a4a-f22e56314690",
		Name:         "Toymaker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Discard a card: Target noncreature artifact becomes an artifact creature with power and toughness each equal to its mana value until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.DiscardACard()),
			Targets: effects.TargetPermanent("target noncreature artifact", effects.And(effects.Artifact(), effects.Noncreature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.ArtifactBecomesCreatureOfItsManaValue(ctx, effects.DurationUntilEndOfTurn(ctx),
					"Toymaker — an artifact creature until end of turn")
			},
		}},
	})
}
