package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prognostic Sphinx — Creature — Sphinx {3}{U}{U}, 3/5 (Duskmourn
// Commander, #5368):
//
//	"Flying
//	 Discard a card: This creature gains hexproof until end of turn. Tap
//	 it.
//	 Whenever this creature attacks, scry 3."
//
// The discard is the cost (Mondo Gecko's DiscardACard) and both halves
// of the effect are the resolving ability's own, so a Sphinx that was
// already tapped still gains hexproof. The attack trigger is the shared
// attack trigger with a scry for its body (Lothlórien Lookout).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "92c470c9-1373-4f3a-9805-5e9613874793",
		Name:            "Prognostic Sphinx",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Scry: 3},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Prognostic Sphinx — scry 3", Do(Scry{N: 3})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Discard a card: This creature gains hexproof until end of turn. Tap it.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    DiscardACard(),
			Effect:  cmdr2S22HexproofAndTapThis("Prognostic Sphinx — hexproof until end of turn"),
		}},
	})
}
