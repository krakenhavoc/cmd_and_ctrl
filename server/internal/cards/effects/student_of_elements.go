package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Student of Elements // Tobita, Master of Winds — {1}{U} Creature —
// Human Wizard 1/1 // Legendary Creature — Human Wizard 3/3, a flip card
// (#3216, ADR 0151):
//
//	Top:     "When this creature has flying, flip it."
//	Flipped: "Creatures you control have flying."
//
// A CR 603.8 state trigger over the Student's own effective abilities
// (ADR 0107 §1): it has none of its own, so the trigger fires when
// something else gives it flying (another Tobita, a lord, an Aura),
// however briefly. If the flying is gone before the
// trigger resolves it flips anyway: the condition is not an intervening
// "if". Tobita is blue and {1}{U} (CR 710.1c).
//
// No simplification.
const studentOfElementsOracleID = "59eab251-af80-4091-98de-9ed413c14cdc"

func init() {
	Register(Spec{
		OracleID:     studentOfElementsOracleID,
		Name:         "Student of Elements",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenState("Student of Elements — flip it",
				func(_ *game.Game, source *game.Card, _ uuid.UUID) bool {
					return game.HasKeyword(source, "flying")
				},
				Do(FlipThis{})),
		},
	})
	Register(Spec{
		OracleID:     studentOfElementsOracleID + "#1",
		Name:         "Tobita, Master of Winds",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			Labelled(KeywordGrant(b16CreaturesYouControl, "flying"), "Creatures you control have flying."),
		},
	})
}
