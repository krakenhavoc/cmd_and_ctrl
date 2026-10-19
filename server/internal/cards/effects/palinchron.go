package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Palinchron — Creature — Illusion {5}{U}{U}, 4/5:
//
//	"Flying
//	 When this creature enters, untap up to seven lands.
//	 {2}{U}{U}: Return this creature to its owner's hand."
//
// The enters trigger is Peregrine Drake's with a bigger N: UntapUpToLands
// lets the controller choose, as the trigger resolves, up to seven lands
// on the battlefield — any player's, since the clause says neither
// "target" nor "you control". Nothing is targeted, so hexproof and
// shroud are irrelevant. The bounce is the shared self-bounce
// (returnThisPermanentToOwnersHand) behind a {2}{U}{U} cost; paying it in
// response to removal is the point of the card, hence the AnswerSurvive
// purpose, and replaying the creature re-triggers the untap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0023888e-7bec-43e0-8dee-d1a4eb94b372",
		Name:            "Palinchron",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Palinchron — untap up to seven lands", Do(UntapUpToLands{
				N:        7,
				Question: "Palinchron — untap up to seven lands",
			})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}{U}: Return this creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{2}{U}{U}"),
			Effect:  returnThisPermanentToOwnersHand,
		}},
	})
}
