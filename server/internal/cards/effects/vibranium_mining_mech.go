package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vibranium Mining Mech — {4} Artifact — Vehicle 6/6 (Marvel Super Heroes
// Commander, #4455):
//
//	"Trample
//	 Whenever this Vehicle enters or attacks, create a tapped Vibranium
//	 token. (It's an artifact with indestructible and "{T}: Add {C}. This
//	 mana can't be spent to cast a nonartifact spell.")
//	 {2}: This Vehicle gets +1/+0 until end of turn.
//	 Crew 2"
//
// One enters-or-attacks trigger (Royal Talon Fighter Jet's
// WhenThisEntersOrAttacks) creating Dora Milaje Elite's printed Vibranium
// token tapped. The pump is Akroma's firebreathing on the Vehicle itself and
// works whether or not it is crewed. Crew is Dependable Quinjet's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1bea6d02-186e-46d8-b788-d4cc37dbb480",
		Name:            "Vibranium Mining Mech",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Vibranium Mining Mech — create a tapped Vibranium token",
				Do(CreateTokenAdvanced{Spec: Token(VibraniumToken()).EntersTapped(), N: 1})),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{2}: This Vehicle gets +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{2}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Vibranium Mining Mech — +1/+0"}.Apply(ctx)
				},
			},
			{
				Label:   "Crew 2",
				Cost:    CrewCost(2),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("Vibranium Mining Mech"),
			},
		},
	})
}
