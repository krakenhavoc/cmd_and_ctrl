package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Town Gossipmonger // Incited Rabble — {W} Creature — Human 1/1 // Creature
// — Human 2/3 (Innistrad Remastered, S82 GAP, #5948):
//
//	Front: "{T}, Tap an untapped creature you control: Transform this
//	        creature."
//	Back:  "This creature attacks each combat if able.
//	        {2}: This creature gets +1/+0 until end of turn."
//
// The second tap is a separate cost that is not the {T} symbol
// (CR 302.6), so a creature that arrived this turn may pay it, and the
// Gossipmonger can't pay for itself (CR 118.3). The back face's forced
// attack is the shared AttacksEachCombat static (CR 508.1d), and its pump is
// a Layer 7c effect on itself.
//
// No simplification.
func init() {
	const oracle = "8a8e225b-14e4-464f-bebe-89b4c9be7fe1"
	const label = "Incited Rabble — +1/+0 until end of turn"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Town Gossipmonger",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Tap an untapped creature you control: Transform this creature.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(TapCost(), TapAnUntapped("an untapped creature you control", Creature())),
			Effect:  Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Incited Rabble",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{AttacksEachCombat()},
		Activated: []ActivatedAbility{{
			Label:   "{2}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{2}"),
			Effect:  thisGetsUntilEndOfTurn(1, 0, label),
		}},
	})
}
