package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blood Fountain — Artifact {B} (Innistrad: Crimson Vow):
//
//	"When this artifact enters, create a Blood token.
//	 {3}{B}, {T}, Sacrifice this artifact: Return up to two target creature
//	 cards from your graveyard to your hand."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "abfa5753-8c02-4e2f-8f98-b0636907293a",
		Name:         "Blood Fountain",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Blood Fountain — create a Blood token", effects.Do(effects.S82Vow03CreateBlood(1))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{B}, {T}, Sacrifice this artifact: Return up to two target creature cards from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}{B}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCardInGraveyard("up to two target creature cards from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(0, 2),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ReturnLegalGraveyardTargetsToHand(effects.NewContext(g, item))
			},
		}},
	})
}
