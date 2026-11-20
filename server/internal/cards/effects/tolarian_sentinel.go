package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tolarian Sentinel — Creature — Human Spellshaper {3}{U}, 1/3 (Time
// Spiral, #5596):
//
//	"Flying
//	 {U}, {T}, Discard a card: Return target permanent you control to its
//	 owner's hand."
//
// A Spellshaper (nemSpellshaperCost) whose spell is Light the Way's
// second bullet: the target must still be a permanent you control as it
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bef15918-c0e7-4f6d-a589-f91f3fa11ef1",
		Name:            "Tolarian Sentinel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{U}, {T}, Discard a card: Return target permanent you control to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    nemSpellshaperCost("{U}"),
			Targets: TargetPermanent("target permanent you control", YouControl()),
			Effect:  bounceTheTargetToHand,
		}},
	})
}
