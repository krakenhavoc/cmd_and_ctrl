package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c33e99c6-189e-4dcf-8b6a-64937dbad361",
		Name:         "Rimefire Torque",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Rimefire Torque"),
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, dawnBlessedPennantOfChosenTypeEntered,
				"Rimefire Torque — put a charge counter on this artifact",
				Do(CounterOnThis{Kind: game.CounterCharge, N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label: "{T}, Remove three charge counters from this artifact: When you next cast an instant or sorcery spell this turn, copy it. You may choose new targets for the copy.",
			Cost:  Plus(TapCost(), RemoveCountersFromThis(game.CounterCharge, 3)),
			// The copy comes later, off your own next spell: value, no answer.
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return WhenYouNextCast(
					"Rimefire Torque — copy that spell",
					game.CastFilter{Types: []string{"Instant", "Sorcery"}},
					copyTheSpellBody,
				).Apply(NewContext(g, item))
			},
		}},
	})
}
