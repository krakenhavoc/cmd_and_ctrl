package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Avengers Quinjet — {5} Artifact — Vehicle 4/4 (Marvel Super Heroes
// Commander, #4455):
//
//	"Flying
//	 Whenever this Vehicle enters or attacks, choose one —
//	 • You may put a Hero creature card from your hand onto the
//	   battlefield.
//	 • Return target Hero creature card from your graveyard to your hand.
//	 Crew 3"
//
// Fateshaper Aspirant's modal trigger on WhenThisEntersOrAttacks (one
// ability, two conditions). The first bullet is Quicksilver Amulet's
// PutFromHandOntoBattlefield narrowed to Hero creature cards, the
// second the shared graveyard-to-hand body with its own target. Crew is
// Dependable Quinjet's CrewCost / CrewEffect.
//
// No simplification.
func init() {
	trigger := WhenThisEntersOrAttacks("Avengers Quinjet — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = ChooseOne(
		ModeDoing("You may put a Hero creature card from your hand onto the battlefield.",
			nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return PutFromHandOntoBattlefield{
					Match:    And(Creature(), HasSubtype("Hero")),
					Optional: true,
					Label:    "Avengers Quinjet — you may put a Hero creature card from your hand onto the battlefield",
				}.Apply(ctx)
			}),
		ModeDoing("Return target Hero creature card from your graveyard to your hand.",
			TargetCardInGraveyard("target Hero creature card in your graveyard", YouOwn(), Creature(), HasSubtype("Hero")),
			ReturnTheModesGraveyardTargetToHand),
	)
	Register(Spec{
		OracleID:        "905e4bd1-f5f2-4691-b13f-ce2628d26a38",
		Name:            "Avengers Quinjet",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{trigger},
		Activated: []ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Avengers Quinjet"),
		}},
	})
}
