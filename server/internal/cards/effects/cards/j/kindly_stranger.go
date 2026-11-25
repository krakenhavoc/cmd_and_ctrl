package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kindly Stranger // Demon-Possessed Witch — Creature — Human {2}{B}, 2/3
// // Creature — Human Shaman 4/3 (Shadows over Innistrad, #5543):
//
//	Front: "Delirium — {2}{B}: Transform this creature. Activate only if
//	        there are four or more card types among cards in your graveyard."
//	Back:  "When this creature transforms into Demon-Possessed Witch, you
//	        may destroy target creature."
//
// Delirium is an activation condition (CR 602.5b), read from the
// activator's own graveyard each time the ability is offered and again as it
// is activated. The transform is the in-place verb (CR 701.27), ignored if
// the creature has turned over since the ability went on the stack
// (CR 701.27f). The Witch's trigger is the one that watches this permanent
// turning to its back face (CR 701.27e, Huntmaster of the Fells' shape): its
// target is chosen as it goes on the stack, the choice to destroy is made as
// it resolves, and a target gone by then is not destroyed (CR 608.2b).
//
// No simplification.
func init() {
	const oracle = "7579e1a2-2747-4e37-9a7a-c13b62190195"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Kindly Stranger",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Delirium — {2}{B}: Transform this creature. Activate only if there are four or more card types among cards in your graveyard.",
			Cost:    effects.ManaCost("{2}{B}"),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return effects.DrcDeliriumMet(g, controller)
			},
			Effect: effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Demon-Possessed Witch",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.On(game.EventTransform, effects.ThisTransformedInto(1),
					"Demon-Possessed Witch — you may destroy target creature", effects.DestroyFirstLegalTarget),
				effects.TargetCreature("target creature")),
				"Demon-Possessed Witch — destroy the creature?"),
		},
	})
}
