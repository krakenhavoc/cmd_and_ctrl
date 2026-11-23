package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rimefire Torque — Artifact {1}{U} (Lorwyn Eclipsed, #4458):
//
//	"As this artifact enters, choose a creature type.
//	 Whenever a permanent you control of the chosen type enters, put a
//	 charge counter on this artifact.
//	 {T}, Remove three charge counters from this artifact: When you next
//	 cast an instant or sorcery spell this turn, copy it. You may choose
//	 new targets for the copy."
//
// The choice is the CR 614.12 creature-type prompt Kindred Discovery
// uses (ChooseCreatureTypeAsEnters), landing on Card.NamedTribe. The
// counter trigger is Dawn-Blessed Pennant's "a permanent you control of
// the chosen type enters" (read through the layered view, so a
// changeling or a kindred permanent counts) with the footprinted
// CounterOnThis primitive; before a type is chosen nothing matches.
//
// The activation pays {T} and the three counters at announce, then, as
// it resolves, sets up Doublecast's CR 603.7 delayed trigger: it fires
// on the first instant or sorcery its controller casts this turn and
// copies it (CR 707.10, not cast), with new targets allowed (CR
// 707.10c). If none is cast it ends with the turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c33e99c6-189e-4dcf-8b6a-64937dbad361",
		Name:         "Rimefire Torque",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseCreatureTypeAsEnters("Rimefire Torque"),
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.DawnBlessedPennantOfChosenTypeEntered,
				"Rimefire Torque — put a charge counter on this artifact",
				effects.Do(effects.CounterOnThis{Kind: game.CounterCharge, N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{T}, Remove three charge counters from this artifact: When you next cast an instant or sorcery spell this turn, copy it. You may choose new targets for the copy.",
			Cost:  effects.Plus(effects.TapCost(), effects.RemoveCountersFromThis(game.CounterCharge, 3)),
			// The copy comes later, off your own next spell: value, no answer.
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.WhenYouNextCast(
					"Rimefire Torque — copy that spell",
					game.CastFilter{Types: []string{"Instant", "Sorcery"}},
					effects.CopyTheSpellBody,
				).Apply(effects.NewContext(g, item))
			},
		}},
	})
}
