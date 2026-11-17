package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chartooth Cougar — Creature — Cat Beast {5}{R}, 4/4 (Scourge):
//
//	"{R}: This creature gets +1/+0 until end of turn.
//	 Mountaincycling {2} ({2}, Discard this card: Search your library for
//	 a Mountain card, reveal it, put it into your hand, then shuffle.)"
//
// Avarax's firebreathing for {R}, and the shared typecycling ability (CR
// 702.29e). The search finds any card with the Mountain subtype, a nonbasic
// Mountain included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "717152c3-2082-49aa-9274-a1a6abe54b65",
		Name:         "Chartooth Cougar",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{R}: This creature gets +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{R}"),
				Effect:  thisGetsUntilEndOfTurn(1, 0, "Chartooth Cougar — +1/+0 until end of turn"),
			},
			Typecycling("Mountaincycling", "{2}", "a Mountain card", IsLandWithSubtype("Mountain")),
		},
	})
}
