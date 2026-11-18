package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blood Fountain — Artifact {B} (Innistrad: Crimson Vow):
//
//	"When this artifact enters, create a Blood token.
//	 {3}{B}, {T}, Sacrifice this artifact: Return up to two target creature
//	 cards from your graveyard to your hand."
func init() {
	Register(Spec{
		OracleID:     "abfa5753-8c02-4e2f-8f98-b0636907293a",
		Name:         "Blood Fountain",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Blood Fountain — create a Blood token", Do(s82Vow03CreateBlood(1))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{B}, {T}, Sacrifice this artifact: Return up to two target creature cards from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{3}{B}"), TapCost(), SacrificeThis()),
			Targets: TargetCardInGraveyard("up to two target creature cards from your graveyard", YouOwn(), Creature()).WithCount(0, 2),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return returnLegalGraveyardTargetsToHand(NewContext(g, item))
			},
		}},
	})
}
