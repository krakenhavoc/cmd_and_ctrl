package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	trigger := effects.WhenThisEntersOrAttacks("Avengers Quinjet — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = effects.ChooseOne(
		effects.ModeDoing("You may put a Hero creature card from your hand onto the battlefield.",
			nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.PutFromHandOntoBattlefield{
					Match:    effects.And(effects.Creature(), effects.HasSubtype("Hero")),
					Optional: true,
					Label:    "Avengers Quinjet — you may put a Hero creature card from your hand onto the battlefield",
				}.Apply(ctx)
			}),
		effects.ModeDoing("Return target Hero creature card from your graveyard to your hand.",
			effects.TargetCardInGraveyard("target Hero creature card in your graveyard", effects.YouOwn(), effects.Creature(), effects.HasSubtype("Hero")),
			effects.ReturnTheModesGraveyardTargetToHand),
	)
	effects.Register(effects.Spec{
		OracleID:        "905e4bd1-f5f2-4691-b13f-ce2628d26a38",
		Name:            "Avengers Quinjet",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{trigger},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    effects.CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Avengers Quinjet"),
		}},
	})
}
